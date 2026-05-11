# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
npm run dev        # Vite dev server on :5173, proxies /api → http://localhost:8080
npm run build      # production build to dist/
```

For the full stack (backend + frontend with hot reload):
```bash
docker-compose up  # from repo root
```

The Vite proxy target can be overridden via `VITE_API_URL` env var.

---

## Architecture

React 18 + TypeScript + Vite SPA. No CSS framework — all styles are inline `React.CSSProperties` objects.

### Routing & auth (`App.tsx`)
`RootRoute` checks `/api/me` on mount and redirects authenticated non-anon users to `/dashboard`. Three user roles: `anon`, `user`, `admin`. Header and page components are rendered as siblings inside `Routes`, not as nested route components. Auth state lives in `AppRoutes` as local `useState`.

### i18n (`i18n.ts`, `context/LanguageContext.tsx`)
All UI strings are in a single `translations` object keyed by `Lang` (`"en" | "uk"`). The English object defines the `T` type; `uk` must satisfy it (`satisfies Record<Lang, T>`). Some values are functions: e.g. `tracksTab: (n: number) => string`. When adding new strings, always add to both locales. Language is persisted in `localStorage` under key `ruscan_lang` and auto-detected from `navigator.language` on first visit. Access via `useLanguage()` hook.

### State management
Local `useState` only — no global store. `LanguageContext` is the only React context.

### Key components
- **`PlaylistScanner`** — accepts Spotify URLs, URIs, bare 22-char playlist IDs, or free-text artist names. Bare Base62 IDs are always treated as playlist IDs (other resource types are ambiguous without a URL/URI). Has an in-memory scan cache keyed by `type:id` or `artist-name:lowercased`. The scanner tab stays mounted and is only hidden with `display: none` to preserve results across tab switches.
- **`ArtistSuggestions`** — tabbed form for insert/delete artist suggestions. `deletePrefillName` prop pre-fills the delete tab when the user clicks "Not Russian?" in the scanner.
- **`AdminPage`** — suggestion review queue, only accessible to `admin` role.
- **`react-bits/`** — third-party animated components (Aurora background, AnimatedList, GradientText) from reactbits.dev; treat as vendored.

### API
All `fetch` calls use `credentials: "include"` for cookie-based auth. Error handling maps backend `code` strings (e.g. `ANON_QUOTA_EXCEEDED`, `SPOTIFY_NOT_FOUND`) to i18n keys defined in `PlaylistScanner.tsx`.

### Scan API (backend contract)
Scan endpoints are async: `GET /api/scan/:provider/:type/:id` returns `202 { jobId }`. Client polls `GET /api/jobs/:jobId` until `status` is `done` or `failed`. See `scan-api.md` for the full contract and polling strategy. Jobs expire after 10 minutes.
