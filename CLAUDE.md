# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Spotiscan is a full-stack web application that scans Spotify playlists and liked songs to identify and remove tracks by Russian artists. Uses Spotify OAuth2 for authentication.

**Stack:** Go/Gin (backend), React/TypeScript/Vite (frontend), PostgreSQL, Docker

## Development Commands

### Backend (Go)
```bash
cd backend
air                                          # Hot reload dev server (port 8080)
go build -o tmp/main cmd/server/main.go      # Manual build
```

### Frontend (React)
```bash
cd frontend
npm run dev                                  # Dev server (port 5173, proxies /api to :8080)
npm run build                                # Production build to dist/
```

### Docker
```bash
docker-compose up                            # Frontend :3000, Backend :8080
```

### Database Migrations
```bash
cd backend
goose -dir migrations postgres "$DB_URL" up  # Run migrations
```

## Architecture

### Authentication Flow
1. User clicks login → `GET /api/auth/start` → redirects to Spotify OAuth
2. Spotify callback → `GET /api/auth/callback` → exchanges code for tokens
3. Backend creates session, sets `session_token` cookie (httpOnly, 7 days)
4. Frontend checks `GET /api/me` on load to verify authentication
5. AuthMiddleware auto-refreshes expired Spotify tokens

### API Endpoints
```
GET    /api/auth/start                    # Initiate OAuth
GET    /api/auth/callback                 # OAuth callback
POST   /api/logout                        # Clear session
GET    /api/me                            # Current user info
GET    /api/playlist/:id/rucontent        # Scan playlist for Russian content
DELETE /api/playlist/:id/rucontent        # Remove Russian tracks from playlist
GET    /api/user/liked-songs/rucontent    # Scan liked songs
DELETE /api/user/liked-songs/rucontent    # Remove Russian tracks from liked songs
```

### Backend Structure
- `cmd/server/main.go` - Entry point, router setup
- `internal/handlers/` - HTTP handlers
- `internal/services/` - Business logic
- `internal/middlewares/` - Auth middleware
- `pkg/db/` - Database operations
- `pkg/spotify/` - Spotify API wrapper

### Frontend Structure
- `src/App.jsx` - Router with auth state
- `src/pages/Landing.jsx` - Login page with Aurora background
- `src/pages/Dashboard.tsx` - Main scanner interface

## Environment Variables

Required in `.env` (see `.example.env`):
```
DB_URL                  # PostgreSQL connection string
FRONTEND_URL            # Frontend origin (CORS + OAuth redirect)
SPOTIFY_CLIENT_ID       # Spotify app credentials
SPOTIFY_CLIENT_SECRET   # Spotify app credentials
SPOTIFY_REDIRECT_URI    # Must match Spotify app config
```

## Key Implementation Notes

- **Russian artist detection** in `backend/internal/services/spotify.go` uses placeholder logic - needs real implementation
- **Spotify scopes:** user-read-private, user-library-read, playlist-modify-public, playlist-modify-private
- **CORS:** Only allows FRONTEND_URL origin with credentials
- **Vite proxy:** Dev mode proxies `/api/*` to backend at localhost:8080


## Files Not to Modify Without Explicit Permission
- `backend/` - NEVER edit if wasn't requested explicitly