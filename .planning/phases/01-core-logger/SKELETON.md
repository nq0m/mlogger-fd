# Walking Skeleton — Field Day Logger

**Phase:** 1
**Generated:** 2026-05-29

## Capability Proven End-to-End

A Field Day operator can enter a QSO (callsign, band, mode, exchange) via a keyboard-navigable form, the QSO persists to SQLite via the Go API server, and the operator sees it in the scrollable log table — all served from a single Go binary embedding the SvelteKit SPA.

## Architectural Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Framework | Go 1.22+ (chi v5 router) + SvelteKit 2 (SPA mode) | Single binary deployment on RPi; SvelteKit provides ~8KB gzip bundles and reactive UI with Svelte 5 runes |
| Data layer | SQLite via modernc.org/sqlite (pure Go, CGo-free) | CGo-free enables trivial cross-compilation to ARM for RPi; WAL mode for concurrent reads; single-file DB easy to back up |
| Auth | None (trusted LAN) | Everyone in the tent is trusted; simple shared password deferred to Phase 2 if open WiFi is used |
| Deployment target | Single Go binary + systemd service on RPi 4 / Linux laptop | One binary embeds SPA via `embed.FS`; systemd manages startup, restart, and logging |
| Directory layout | Go: `internal/{db,handler,model,qso,cabrillo}/`; Frontend: `frontend/src/{routes,lib/{components,stores}}/` | Standard Go internal package structure prevents external imports; SvelteKit convention for routes and shared lib |
| Build system | Makefile with `make dev` (frontend dev + Go build) | `make dev` builds frontend static files, compiles Go binary with embedded FS, and starts server on :8080 |
| State management | Svelte 5 `$state` runes in shared `.svelte.js` modules | SPA mode (`ssr=false`) makes module-level state safe; no store subscriptions needed |

## Stack Touched in Phase 1

- [x] Project scaffold — Go module (`go.mod`, `main.go`), SvelteKit SPA (`package.json`, `svelte.config.js`, `vite.config.ts`), Makefile
- [x] Routing — chi router with `/api/*` prefix group + SPA fallback via `embed.FS` + `http.FileServer`
- [x] Database — SQLite WAL mode via DSN pragmas; `qsos` table with indexes; `POST /api/qso` (write) and `GET /api/qso` (read)
- [x] UI — QSO entry form with Tab navigation + Ctrl+Enter submit (D-01, QSO-04); three-panel layout (D-04); scrollable log table
- [x] Deployment — `make dev` command exercises full stack locally on :8080

## Out of Scope (Deferred to Later Slices)

- Multi-user WebSocket sync (Phase 2)
- Station configuration UI — callsign, class, section, power (Phase 2)
- Operator identity selector (Phase 2)
- Offline IndexedDB buffer + Dexie.js (Phase 3)
- Service Worker offline fallback (Phase 3)
- Mobile-responsive layout + touch targets (Phase 3)
- Dark mode theme (Phase 3)
- Debounce/rate-limiting rapid entry (Phase 3)
- Bonus points tracker (Phase 4)
- Audio alerts via Web Audio API (Phase 4)
- Database backup endpoint (Phase 4)
- Docker deployment (explicitly out of scope per PROJECT.md — single binary preferred)

## Subsequent Slice Plan

Each later phase adds one vertical slice on top of this skeleton without altering its architectural decisions:

- **Phase 2 (Multi-User):** WebSocket-based real-time QSO broadcasting, station config UI, operator identity, shared scoreboard
- **Phase 3 (Offline Resilience):** IndexedDB buffer, Service Worker, offline dupe checking, mobile responsive design, dark mode
- **Phase 4 (Field Day Features):** Bonus tracker, audio alerts, database backup, simulation testing
