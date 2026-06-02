# ruscan — Project Guide

Project-specific conventions and architecture for AI agents. General coding style and Go conventions are in the global `CLAUDE.md` and `GOLANG.md` (read those first if working on Go code).

---

## What This Project Is

**ruscan** scans Spotify content (playlists, albums, tracks, artists) and identifies tracks featuring Russian artists. It exists to help users audit their music libraries.

The backend is a Go monorepo with four independently deployed services. The frontend is a React SPA. Everything runs on Kubernetes in production; `docker-compose.yml` is dev only.

---

## Services

### `api`
HTTP REST server on port `8080`. Handles auth, scan requests, job polling, and artist suggestions. Accepts a scan request → publishes a `ContentFetchJob` to the `spotify` queue → returns a job ID. The client polls `/api/jobs/{jobId}` until `done` or `failed`.

### `spotify-gateway`
Consumes the `spotify` queue. Calls the Spotify API to fetch playlist/album/track/artist content. Deduplicates tracks and artist refs, then publishes a `ContentScanJob` to the `scanner` queue.

### `scan-worker`
Consumes the `scanner` queue. Looks up each artist in the `ru_artists` table. Writes the filtered result (only Russian-artist tracks) to Redis under `jobs:{jobID}` with a 10-minute TTL.

### `scraper`
Runs as a Kubernetes CronJob (Sunday 03:00 UTC). One-shot: no server, exits after completion. Scrapes artists from LastFM, MusicBrainz, and Phonkersbase and upserts them into `ru_artists`. Pushes a `scraper_ru_artists_total` gauge to Prometheus Pushgateway.

**Flow:** `api` → `spotify` queue → `spotify-gateway` → `scanner` queue → `scan-worker` → Redis job result

---

## Queue Topology (RabbitMQ)

Both queues are **quorum queues** (durable, replicated). Each has a dead-letter exchange and dead-letter queue.

| Queue | DLX | Dead queue | Message type | Producer → Consumer |
|---|---|---|---|---|
| `spotify` | `spotify.dlx` | `spotify.dead` | `ContentFetchJob` | api → spotify-gateway |
| `scanner` | `scanner.dlx` | `scanner.dead` | `ContentScanJob` | spotify-gateway → scan-worker |

Delivery limit: 3 attempts per message. After 3 failures the message routes to the `.dead` queue.

Queue names are constants in `internal/shared/domain/`:
- `domain.SpotifyQueueName = "spotify"`
- `domain.ScannerQueueName = "scanner"`

---

## Protobuf Messages (`backend/proto/scan_job.proto`)

All inter-service queue messages are serialized as proto3.

```
ContentFetchJob  { id, user_id, resource_type (ResourceType enum), resource_id }
ContentScanJob   { scan_job_id, tracks (Track[]), artists (ArtistRef[]) }
Track            { external_id, name, image_url, artists (ArtistRef[]) }
ArtistRef        { external_id, name }

ResourceType: INVALID=0, PLAYLIST_ID=1, TRACK_ID=2, ALBUM_ID=3, ARTIST_ID=4, ARTIST_NAME=5
```

`external_id` is always a Spotify ID (22-character alphanumeric string). `id`/`scan_job_id` are UUID v4.

---

## Domain Models

**`ru_artists`** — the source of truth for Russian artists.
Fields: `id` (serial), `name` (unique), `description_ua`, `description_en`, `source` (lastfm|musicbrainz|phonkersbase), `source_url`, `country` (ISO code), `confirmed` (bool).

**`users`** — registered accounts.
Fields: `id`, `email` (unique), `password_hash`.
Roles: `user`, `admin`, `anon` (anonymous, not stored in DB).

**`admins`** — `user_id` references `users(id)` with cascade delete. A user is an admin if a row exists here.

**`refresh_tokens`** — `id`, `user_id`, `token_hash` (SHA256 of the raw token), `expires_at`. Index on `user_id`.

**`artist_insert_suggestions`** / **`artist_delete_suggestions`** — `id`, `creator_id`, `artist_name`, `description`, `state` (pending|approved|declined), `decline_reason`, `created_at`, `updated_at`.

**Job (Redis-only)** — stored as a hash at `jobs:{jobID}`. Fields: `status` (pending|processing|done|failed), `created_at` (RFC3339), `user_id`, `data` (JSON result). TTL: 10 minutes.

---

## Auth

**Access token**: JWT, 15-minute lifetime, signed with `JWT_SECRET` (min 32 chars). Claims: `UserID` (string), `Role`. Stored in `jwt` cookie (HTTP-only, SameSite=Lax, Secure in prod).

**Refresh token**: 32 random bytes, base64-encoded, SHA256-hashed before DB storage. 30-day lifetime. Stored in `refresh_token` cookie. On an expired JWT, the auth middleware silently rotates both tokens.

**Anonymous users**: JWT with `Role=anon` and a synthetic UUID as UserID. Not stored in DB. 24-hour lifetime. Subject to per-path Redis quota.

**Cookie flag**: `SECURE_COOKIES=true` in production (HTTPS only), `false` in dev.

---

## Rate Limiting

**Global** (all IPs, all scan endpoints): 50 requests per 60 seconds per client IP. Enforced via `ratelimit.lua` — an atomic Redis Lua script that increments a key and returns 0 if the limit is exceeded. Exempt: `/health`, `/metrics`, `/api/auth/*`, `/api/me`.

**Anonymous quota**: 5 requests per 24 hours per path per session. Tracked via `anon:{sessionID}:{path}` keys in Redis. Applies to all `/api/scan/*` paths.

**Spotify API**: 175 requests per 30 seconds, enforced client-side with a sliding window. Respects `Retry-After` on 429; blocks all outgoing requests for the specified duration.

---

## Redis Keys

| Key pattern | Type | TTL | Purpose |
|---|---|---|---|
| `jobs:{jobID}` | Hash | 10 min | Job result storage |
| `ratelimit:{clientIP}` | Integer | 60 s | Global rate limit counter |
| `anon:{sessionID}:{path}` | Integer | 24 h | Anonymous quota per path |

Lua script `backend/scripts/ratelimit.lua` handles the global rate limit atomically (GET → check → INCR → SET EX).

---

## Sentinel Errors

All defined in `internal/shared/domain/errors.go` as `*AppError{HTTPCode, Code}`. The `Code` field is the string sent to clients.

| Sentinel | HTTP | Code |
|---|---|---|
| `ErrInternal` | 500 | `INTERNAL_ERROR` |
| `ErrDatabaseFailure` | 500 | `DATABASE_ERROR` |
| `ErrSpotifyAPIError` | 500 | `SPOTIFY_API_ERROR` |
| `ErrNotFound` | 404 | `NOT_FOUND` |
| `ErrSpotifyNotFound` | 404 | `SPOTIFY_NOT_FOUND` |
| `ErrBadRequest` | 400 | `BAD_REQUEST` |
| `ErrUnauthorized` | 401 | `UNAUTHORIZED` |
| `ErrForbidden` | 403 | `FORBIDDEN` |
| `ErrInvalidCredentials` | 401 | `INVALID_CREDENTIALS` |
| `ErrEmailExists` | 409 | `EMAIL_EXISTS` |
| `ErrArtistExists` | 409 | `ARTIST_EXISTS` |
| `ErrQuotaExceeded` | 429 | `ANON_QUOTA_EXCEEDED` |
| `ErrTooManyRequests` | 429 | `TOO_MANY_REQUESTS` |
| `ErrSuggestionNotPending` | 400 | `SUGGESTION_NOT_PENDING` |

---

## API Endpoints

**No auth:**
- `GET /health` → 204
- `GET /metrics` → Prometheus text

**Auth:**
- `POST /api/auth/signup` — `{email, password}` → sets `jwt` + `refresh_token` cookies
- `POST /api/auth/login` — same
- `GET /api/me` — returns current user or anon marker
- `POST /api/auth/logout` — requires `RoleUser`+; deletes refresh token

**Scan** (rate-limited; anon limited to 5/path/24h):
- `GET /api/scan/{provider}/playlist/{id}`
- `GET /api/scan/{provider}/track/{id}`
- `GET /api/scan/{provider}/album/{id}`
- `GET /api/scan/{provider}/artist/{id}`
- `GET /api/scan/{provider}/artist/name?Name={name}`
- All return `202 { "jobId": "uuid" }`
- `GET /api/jobs/{jobId}` → `{ jobId, status, result, createdAt }`

**Suggestions** (requires `RoleUser`+):
- `GET/POST/PUT/DELETE /api/suggestions/artist-insert`
- `GET/POST/PUT/DELETE /api/suggestions/artist-delete`

**Admin** (requires `RoleAdmin`):
- `GET /api/admin/suggestions/artist-insert`
- `POST /api/admin/suggestions/artist-insert/{id}/approve`
- `POST /api/admin/suggestions/artist-insert/{id}/decline`
- Same pattern for `artist-delete`

Spotify ID params are validated as exactly 22 alphanumeric characters. Artist names 1–255 characters.

---

## Third-Party Integrations

**Spotify** — client credentials OAuth (no user auth). Fetches playlists, albums, tracks, artists by ID, and artists by name. Token cached in memory and refreshed on expiry. Concurrency cap: 5 simultaneous requests (semaphore). 429 → read `Retry-After`, block all requests for that duration. Errors: 404→`ErrSpotifyNotFound`, 400→`ErrBadRequest`, 429→`ErrTooManyRequests`, else→`ErrSpotifyAPIError`.

**LastFM** — `tag.getTopArtists` endpoint, paginated. Requires `LASTFM_API_KEY`. Controlled by `SCRAPE_LASTFM_TOP_ARTISTS_FOR_ALL_TAGS` flag.

**MusicBrainz** — `/ws/2/artist?area={id}` endpoint, paginated (limit 100). Public API; must send `User-Agent: ruscan/1.0 (https://ruscan.chivtar.dev)`. Respects `Retry-After`. Controlled by `SCRAPE_MUSICBRAINZ_ARTISTS_FOR_ALL_REGIONS` flag.

**Phonkersbase** — `https://www.phonkersbase.com/api/artists`, paginated (limit 50). Returns `{data: {items, info}}`. Country mapping: `"ruzzia"`/`"russia"` → `"RU"`, `"ukraine"` → `"UA"`. Controlled by `SCRAPE_PHONKERS_DB_ARTISTS` flag. Same User-Agent header.

---

## Metrics

**Global (all services):**
- `errors_total{component}` (Counter) — incremented automatically by `MetricsHook` on every `log.Error()` call. Component labels: `api`, `scanner`, `scraper`, `spotify-gateway`.

**Per-service instrumentation:**
- `http_requests_total{method,path,status}` — API
- `http_request_duration_seconds{method,path}` — API
- `http_requests_in_flight` — API
- `spotify_api_duration_seconds{method}` — spotify-gateway
- `postgres_latency_seconds{operation}` — shared repo layer (via pgx QueryTracer)
- `redis_latency_seconds{operation}` — shared repo layer (via go-redis Hook)
- `scraper_ru_artists_total` (Gauge) — scraper only; pushed to Pushgateway

The scraper cannot expose a `/metrics` endpoint (it's a one-shot CronJob), so it pushes to `http://prometheus-pushgateway.monitoring.svc.cluster.local:9091` using job name `"scraper"`.

---

## Database Migrations

7 migrations in `backend/migrations/`, Goose format (`+goose Up` / `+goose Down`). Embedded in the binary via `//go:embed migrations/*.sql` and run automatically at startup. Never edit or renumber applied migrations — add a new one.

Sequence: create artists + tokens → auth tables → token index → expand artists table → drop spotify tokens → suggestion tables → admin table.

---

## Config Variables

**api**: `DATABASE_URL`, `REDIS_URL`, `RABBITMQ_URL`, `JWT_SECRET` (min 32 chars), `SECURE_COOKIES`

**spotify-gateway**: `REDIS_URL`, `RABBITMQ_URL`, `SPOTIFY_CLIENT_ID`, `SPOTIFY_CLIENT_SECRET`

**scan-worker**: `DATABASE_URL`, `REDIS_URL`, `RABBITMQ_URL`

**scraper**: `DATABASE_URL`, `LASTFM_API_KEY`, `SCRAPE_LASTFM_TOP_ARTISTS_FOR_ALL_TAGS` (bool), `SCRAPE_PHONKERS_DB_ARTISTS` (bool), `SCRAPE_MUSICBRAINZ_ARTISTS_FOR_ALL_REGIONS` (bool), `PUSHGATEWAY_URL`

All loaded from `.env` (dev) or Kubernetes Secret `app-secrets` (prod) via `secretRef`.

---

## Kubernetes

Namespace: `ruscan`. All resources carry `app: <service-name>` labels.

- `api`, `scan-worker`, `spotify-gateway` → Deployments, Service, ServiceMonitor (scrapes `:8080/metrics` every 15 s)
- `scraper` → CronJob (`0 3 * * 0`, `Forbid` concurrency, 1h deadline, 3 retries)
- `postgres`, `redis`, `rabbitmq` → StatefulSets with VolumeClaimTemplates
- `app-secrets` encrypted with SOPS/age (`k8s/secrets.enc.yaml`)
- Image tags managed by CD via `kustomize edit set image`; placeholder tag in repo

All containers: `allowPrivilegeEscalation: false`, all capabilities dropped, `runAsNonRoot: true`, `readOnlyRootFilesystem: true`, resource requests + limits set.

---

## Local Development

`docker-compose.yml` runs: `api`, `spotify-gateway`, `scan-worker`, `frontend`, `postgres`, `redis`, `rabbitmq`.

`scraper` is an opt-in profile: `docker-compose --profile scraper up scraper`.

All backend services use hot-reload (`air`) with source bind-mounted. Frontend uses Vite dev server with HMR. Backend dev images are Alpine-based; prod images are distroless.

---

## CI/CD

`.github/workflows/ci.yaml` — triggers on push to `main` (excluding `k8s/kustomization.yaml`) and on PRs. Runs `go build`, `go test`, `go vet`, and `npm run build`.

`.github/workflows/cd.yaml` — triggers on successful CI. Diffs changed paths against the current image SHA in `kustomization.yaml`. Builds and pushes only changed components to `ghcr.io/chivta/ruscan/{service}:{sha}`. Commits updated image tags back to `main` as `deploy: update images to {sha}` authored by `github-actions[bot]`.
