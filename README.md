# SpotiScan

[![CI](https://github.com/chivta/spotiscan/actions/workflows/ci-cd.yaml/badge.svg)](https://github.com/chivta/spotiscan/actions/workflows/ci-cd.yaml)
[![codecov](https://codecov.io/gh/chivta/spotiscan/branch/main/graph/badge.svg)](https://codecov.io/gh/chivta/spotiscan)

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
- Go 1.25+ (for local backend dev)
- PostgreSQL 14+ (for local database)
- Redis (required env var; app degrades gracefully if unavailable — caching and rate limiting disabled)

## Setup

1. Create `.env` in `backend/`:
```
DB_URL=postgres://user:password@localhost:5432/spotiscan
REDIS_URL=redis://localhost:6379
SPOTIFY_CLIENT_ID=your_spotify_client_id
SPOTIFY_CLIENT_SECRET=your_spotify_client_secret
JWT_SECRET=your_jwt_secret_at_least_32_chars
```

2. Get Spotify API credentials at https://developer.spotify.com/dashboard (client credentials only — no OAuth redirect needed).

Database migrations run automatically on startup.

## Project Structure

```
spotiscan/
├── frontend/              # React app
│   └── src/
│       ├── pages/         # Landing, AuthPage, Dashboard
│       ├── components/    # Header, AnimatedList, Aurora, GradientText
│       └── types/         # TypeScript interfaces
├── backend/               # Go API server
│   ├── cmd/               # Entry point
│   ├── internal/
│   │   ├── handlers/      # HTTP handlers
│   │   ├── services/      # Business logic
│   │   ├── middlewares/   # Auth, rate limiting
│   │   ├── repository/    # DB, Redis, Spotify API clients
│   │   ├── models/        # Domain types
│   │   └── config/        # Config loading & validation
│   ├── migrations/        # Database migration files (embeded into binary)
│   └── scripts/           # Lua script for ratelimiting (embeded into binary)
└── k8s/                   # Kubernetes manifests (kustomize)
```

## API Endpoints

- `POST /api/auth/signup` - Create account
- `POST /api/auth/login` - Log in
- `POST /api/auth/logout` - Log out
- `GET /api/me` - Current user info
- `GET /api/playlist/:id/rucontent` - Scan playlist for Russian content

## License

MIT
