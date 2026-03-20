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
GET    /api/user/playlists                # Fetch user's playlists
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
- `src/App.tsx` - Router with auth state, `RequireAuth` wrapper for protected routes
- `src/pages/Landing.tsx` - Landing page with Aurora background and login CTA
- `src/pages/Dashboard.tsx` - Main scanner interface with playlist selection and track filtering
- `src/components/Aurora.tsx` - WebGL-based animated gradient background
- `src/components/GradientText.tsx` - Animated gradient text effect
- `src/components/Header.tsx` - Simple navigation header with sign-out button
- `src/components/AnimatedList.tsx` - Reusable scrollable list with keyboard navigation
- `src/types/models.ts` - TypeScript interfaces (User, Artist, Track, Playlist, RuContent)

## Frontend Implementation Details

### Auth Pattern
Frontend uses a custom fetch wrapper (`useAuthFetch` in `src/App.tsx`) that intercepts 401 responses globally and redirects to landing page. Session validation happens on app load via `GET /api/me`.

### Dashboard State Management
The Dashboard component (645 lines) uses React hooks for state. Key features:
- **Playlist selection:** Users can paste Spotify URLs/IDs or pick from `GET /api/user/playlists`
- **Track filtering:** `RuContent` response contains tracks and artist list; client-side rendering with highlight for Russian artists
- **Selective deletion:** Users can select individual tracks before calling `DELETE` endpoint
- **Search filtering:** Built-in search for user's playlists with `AnimatedList` rendering

### Component Architecture
- Use inline styles with TypeScript `CSSProperties` (avoid CSS files in most components)
- `Aurora` component renders WebGL canvas for animated background via `ogl` library
- `AnimatedList` provides virtualized scrolling with arrow key navigation
- Components pass styling objects (`cardStyle`, `buttonStyle`, etc.) for consistent theming

## Environment Variables

Required in `.env` (see `.example.env`):
```
DB_URL                  # PostgreSQL connection string
FRONTEND_URL            # Frontend origin (CORS + OAuth redirect)
SPOTIFY_CLIENT_ID       # Spotify app credentials
SPOTIFY_CLIENT_SECRET   # Spotify app credentials
```

## Key Implementation Notes

- **Russian artist detection** in `backend/internal/services/spotify.go` uses placeholder logic - needs real implementation
- **Spotify scopes:** user-read-private, user-library-read, playlist-modify-public, playlist-modify-private
- **CORS:** Only allows FRONTEND_URL origin with credentials
- **Vite proxy:** Dev mode proxies `/api/*` to backend at localhost:8080
- **Sign-out endpoint:** Uses `/api/logout` (not `/api/signout` as currently called in Dashboard.tsx line 211 - this is a bug)
- **Playlist ownership check:** `Playlist.Owned` field controls if tracks can be deleted; `RuContent.AbleToDelete` also gates delete UI

## Frontend Testing Notes

When testing the Dashboard:
- Invalid playlist IDs (not alphanumeric) trigger validation error
- Already-scanned playlists show "already been scanned" error on re-scan
- Playlists user doesn't own show results but delete buttons are disabled
- Track artist highlighting matches on `Artist.ID` in Russian artist list