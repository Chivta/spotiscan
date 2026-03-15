# SpotiScan

[![CI](https://github.com/chivta/spotiscan/actions/workflows/ci-cd.yaml/badge.svg)](https://github.com/chivta/spotiscan/actions/workflows/ci-cd.yaml)
[![codecov](https://codecov.io/gh/chivta/spotiscan/branch/main/graph/badge.svg)](https://codecov.io/gh/chivta/spotiscan)

Take control of your Spotify library. SpotiScan scans your playlists and liked songs to identify tracks by Russian artists and remove them with one click.

## Features

- **Scan playlists** - Analyze any playlist for Russian artists
- **Scan liked songs** - Find Russian tracks in your saved music
- **Filter & browse** - View Russian artists and tracks separately
- **Selective deletion** - Choose individual tracks to remove or select all filtered results
- **Spotify integration** - Secure OAuth login, no data stored

## Tech Stack

**Frontend:** React 19 + TypeScript + Vite\
**Backend:** Go + Gin\
**Storage:** PostgreSQL + Redis\
**Deployment:** Docker Compose / Kubernetes (ArgoCD)

## Quick Start

### With Docker
```bash
docker-compose up
```
Frontend: http://localhost:3000\
Backend: http://localhost:8080

### Local Development

**Backend:**
```bash
cd backend
air  # Hot reload dev server (port 8080)
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev  # Dev server (port 5173)
```

## Requirements

- Docker & Docker Compose (for containerized setup)
- Node.js 18+ (for local frontend dev)
- Go 1.25+ (for local backend dev)
- PostgreSQL 14+ (for local database)
- Redis (optional — used for artist name caching and rate limiting; app falls back to DB if unavailable)

## Setup

1. Create `.env` in the root directory:
```
DB_URL=postgres://user:password@localhost:5432/spotiscan
SPOTIFY_CLIENT_ID=your_spotify_client_id
SPOTIFY_CLIENT_SECRET=your_spotify_client_secret
SPOTIFY_REDIRECT_URI=http://localhost:3000/callback
FRONTEND_URL=http://localhost:3000
```

2. Set up Spotify OAuth at https://developer.spotify.com/dashboard

3. Run migrations:
```bash
cd backend
goose -dir migrations postgres "$DB_URL" up
```

## Project Structure

```
spotiscan/
├── frontend/              # React app
│   └── src/
│       ├── pages/         # Dashboard, Landing
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
│   └── migrations/        # Database schema
└── k8s/                   # Kubernetes manifests (kustomize)
```

## API Endpoints

- `GET /api/auth/start` - Initiate Spotify OAuth
- `GET /api/me` - Current user info
- `GET /api/user/playlists` - User's playlists
- `GET/DELETE /api/playlist/:id/rucontent` - Scan/clean playlist
- `GET/DELETE /api/user/liked-songs/rucontent` - Scan/clean liked songs

## License

MIT
