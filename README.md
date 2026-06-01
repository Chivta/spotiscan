# RuScan

[![CI](https://github.com/chivta/ruscan/actions/workflows/ci-cd.yaml/badge.svg)](https://github.com/chivta/ruscan/actions/workflows/ci-cd.yaml)
[![codecov](https://codecov.io/gh/chivta/ruscan/branch/main/graph/badge.svg)](https://codecov.io/gh/chivta/ruscan)

Scan your Spotify playlists to identify tracks by Russian artists.

## Features

- **Scan playlists** - Paste any Spotify playlist URL or ID to find tracks by Russian artists
- **Tracks & Artists tabs** - Browse flagged tracks with album art, or view Russian artists sorted by track count; filter and navigate between them
- **Result caching** - Re-opening a scanned playlist is instant; rescan on demand to refresh

## Tech Stack

**Backend:** Go + Gin\
**Auth:** Email/password, short-living JWT + refresh token\
**Frontend:** React 19 + TypeScript + Vite (vibecoded; some components from [React Bits](https://reactbits.dev))\
**Storage:** PostgreSQL + Redis\
**Deployment:** Kubernetes (FluxCD)

## Quick Start

```bash
docker-compose up
```

Runs both services in dev stage with hot reload inside containers.\
Frontend: http://localhost:5173 \
Backend: http://localhost:8080

## Requirements

- Docker & Docker Compose (for containerized setup)
- Node.js 18+ (for local frontend dev)
- Go 1.26+ (for local backend dev)
- PostgreSQL 14+ (for local database)
- Redis (required env var)

## Monitoring

Create exporter user in postgres DB manualy

```sql
CREATE USER exporter WITH PASSWORD '...';
GRANT pg_monitor TO exporter;
```

## Setup

1. Create `.env` in `backend/`:
```
DB_URL=postgres://user:password@localhost:5432/ruscan
REDIS_URL=redis://localhost:6379
SPOTIFY_CLIENT_ID=your_spotify_client_id
SPOTIFY_CLIENT_SECRET=your_spotify_client_secret
JWT_SECRET=your_jwt_secret_at_least_32_chars
```

2. Get Spotify API credentials at https://developer.spotify.com/dashboard (client credentials only — no OAuth redirect needed).

Database migrations run automatically on startup.

## Project Structure

```
ruscan/
├── frontend/              # React app
│   └── src/
│       ├── pages/         # Landing, AuthPage, Dashboard, AdminPage
│       ├── components/    # Header, Footer, PlaylistScanner, ArtistSuggestions, LanguageSwitcher
│       ├── context/       # LanguageContext
│       ├── types/         # TypeScript interfaces
│       └── i18n.ts        # Translations
├── backend/               # Go services (monorepo)
│   ├── cmd/
│   │   ├── api/           # HTTP API server entry point
│   │   ├── scan-worker/   # RabbitMQ consumer that runs playlist scans
│   │   ├── scraper/       # Periodic artist data scraper (cron job)
│   │   └── spotify-gateway/ # Spotify API proxy with token management & rate limiting
│   ├── internal/
│   │   ├── api/           # HTTP handlers, middlewares, services, config
│   │   ├── scanner/       # Scan job processing logic
│   │   ├── scraper/       # LastFM / MusicBrainz / Phonkersbase scrapers
│   │   ├── spotify/       # Spotify client, rate limiter, token worker
│   │   └── shared/        # Domain models, repository, queue, metrics
│   ├── proto/             # Protobuf definitions (scan job messages)
│   ├── migrations/        # SQL migration files (embedded into binary)
│   └── scripts/           # Lua scripts for Redis rate limiting (embedded into binary)
└── k8s/                   # Kubernetes manifests (Kustomize + FluxCD)
    ├── api/
    ├── frontend/
    ├── scan-worker/
    ├── scraper/
    ├── spotify-gateway/
    ├── postgres/
    ├── rabbitmq/
    └── redis/
```

## API Endpoints

- `POST /api/auth/signup` - Create account
- `POST /api/auth/login` - Log in
- `POST /api/auth/logout` - Log out
- `GET /api/me` - Current user info
- `GET /api/playlist/:id/rucontent` - Scan playlist for Russian content

## License

MIT
