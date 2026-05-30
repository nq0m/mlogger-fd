# Phase 3: Offline Resilience & Polish - Context

**Gathered:** 2026-05-30
**Status:** Ready for planning

## Phase Boundary

Delivers offline resilience (IndexedDB buffering with auto-sync, local dupe checking, Service Worker app shell) and UI polish (mobile-responsive layout with touch-friendly controls, dark mode theme). The system survives network drops — operators continue logging uninterrupted and buffered QSOs sync automatically within 5 seconds of reconnection. Mobile devices get usable controls with 48×48px minimum touch targets.

## Implementation Decisions

### Offline Sync & Queue

- **D-01:** Queued QSOs sync via batch POST to a new `/api/sync` endpoint. Server receives an array, inserts all, returns assigned server IDs mapped to client UUIDs. One round-trip, no ordering issues.
- **D-02:** Each QSO gets a client-generated UUID (`crypto.randomUUID()`) at creation time. UUID is stored in IndexedDB as primary key. The server stores it as a `client_id` column. No fragile local-to-server ID mapping.
- **D-03:** Pending queue count displayed next to the connection indicator in the header bar (e.g., "3 queued"). Changes to "Syncing..." during batch POST, clears on success. Follows existing `● Live / ● Disconnected` pattern.
- **D-04:** Sync fires immediately when `wsState.connected` flips to true (WebSocket reconnected = server reachable). Additionally retries every 30 seconds if queue is non-empty and connected. Covers failed batch POSTs without requiring another disconnect cycle.

### Local Cache & Offline Dupe

- **D-05:** IndexedDB caches the full QSO history (all QSOs from all operators). ~3000 QSOs max for a Field Day event fits comfortably in IndexedDB (~1-2MB). Dexie.js manages the store.
- **D-06:** Offline dupe checking uses exact match only: same callsign + band + mode against IndexedDB. Partial call similarity (Levenshtein/prefix matching) remains server-side and online-only — it's a soft warning, not the hard dupe rule. Can add parity later.
- **D-07:** Locally queued (not-yet-synced) QSOs are included in offline dupe checking. Prevents double-logging the same station during extended offline periods — consistent with Field Day dupe rules.
- **D-08:** IndexedDB cache is populated on page load via `GET /api/qso?limit=9999` (full fetch), then kept current via existing WebSocket `qso_created` events. No polling needed — WebSocket already broadcasts every new QSO.

### Mobile Layout

- **D-09:** Stack & shrink strategy with CSS media queries at 480px and 768px breakpoints. All three panels stack vertically as now but with responsive sizing — inputs grow to full width, stats wrap, table shrinks. Single scrollable page, no JS layout changes.
- **D-10:** All interactive elements must have minimum 48×48px touch targets (Material Design). Current inputs at 44px need bumped padding to 12px. Buttons and selects need explicit min-height.
- **D-11:** QSO entry form becomes full-width 2-row wrap on screens <500px: row 1 = callsign + exchange (both full-width), row 2 = band + mode + submit button. Removes the fixed 140px input widths via media query.
- **D-12:** Log table hides Operator and Time columns on narrow mobile screens (CSS only, no DOM changes). Keeps Callsign, Band, Mode, Exchange, Pts — the 5 essential contest columns. Reduces font-size to 13px.

### Dark Mode

- **D-13:** CSS custom properties defined on `:root` for the full color palette. A `[data-theme='dark']` override on `<body>` provides the inverted values. Manual toggle persists preference in localStorage. No `prefers-color-scheme` media queries in components — the toggle sets the attribute.
- **D-14:** Toggle button placed in the header bar, far left — between the "FD Logger" title and the WS connection status indicator. Small icon (sun/moon), always accessible on every screen size.
- **D-15:** On first visit (no localStorage preference), check `prefers-color-scheme`. If OS is in dark mode, start dark. Otherwise start light. The manual toggle overrides and persists — subsequent visits use stored preference.
- **D-16:** All hardcoded colors across every component migrated to CSS variables in a single pass. The current palette uses ~10-12 distinct colors shared across components — define once in `app.css`, reference in component `<style>` blocks.

### the agent's Discretion

- Dexie.js version and exact Dexie table schema (table names, indexes)
- Specific CSS variable names and dark mode color palette values
- Exact media query breakpoint pixel values (480px, 768px confirmed as targets)
- Service Worker implementation approach (Workbox vs hand-written, cache strategy — stale-while-revalidate vs cache-first)
- `/api/sync` endpoint request/response JSON format
- Periodic 30s retry implementation (setInterval vs recursive setTimeout)
- Connection indicator specific UI (icon choice, animation, transition)
- The `client_id` column addition to the server-side QSO table schema
- Whether to keep the `$state` qsos array in addition to IndexedDB, or make IndexedDB the single source of truth

## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Foundation
- `.planning/PROJECT.md` — Project context, constraints (RPi 4, SQLite, LAN-only, gloves/wet-fingers UI). Core value: log must never be lost.
- `.planning/REQUIREMENTS.md` — Phase 3 requirements: SYNC-03, SYNC-04, SYNC-05, SYNC-06, UX-01, UX-02, UX-04
- `.planning/ROADMAP.md` § Phase 3 — Scope anchor, success criteria, key deliverables

### Prior Phase Context
- `.planning/phases/01-core-logger/01-CONTEXT.md` — D-01 (auto-clear form on submit), D-04 (three-panel layout), D-07 (single Go binary embed.FS). These constrain mobile adaptation and SW deployment.
- `.planning/STATE.md` — Phase 2 decisions: object-based `$state` pattern (wsState), WebSocket 2-sec reconnect, gorilla/websocket, route wiring patterns

### Existing Code
- `frontend/src/lib/api.js` — Current fetch wrappers. Extend with an `api.syncBatch(qsos)` function.
- `frontend/src/lib/ws.svelte.js` — WebSocket client with reconnect. Hook sync trigger into the `wsState.connected` transition.
- `frontend/src/lib/stores/qso.svelte.js` — Current `$state` stores. Need offline-capable write path that routes to IndexedDB when server unreachable.
- `main.go` — Chi router with `/api/*` namespace. Add `POST /api/sync` route.

## Existing Code Insights

### Reusable Assets
- **`api.js`** — Existing fetch wrapper pattern. Add `syncBatch()` function for the `/api/sync` endpoint.
- **`stores/qso.svelte.js`** — `addQso()`, `fetchStats()` functions. Need offline-aware variants that detect connectivity and route to IndexedDB vs server.
- **`ws.svelte.js`** — `connectWebSocket()` already has reconnect logic; `wsState.connected` provides the signal SyncTrigger uses. Need to add a sync callback on reconnect.
- **`QsoEntryForm.svelte`** — Dupe check on blur + submit via `handleCheckDupe()`. Needs an offline-aware dupe function that queries IndexedDB when server is unreachable.
- **`LogTable.svelte`** — `loadQsos()` fetches from server. Needs ability to read from IndexedDB when offline.
- **Header bar in `+page.svelte`** — `wsStatus` span already shows connection state. Add queue count and dark mode toggle here.

### Established Patterns
- Svelte 5 runes: `$state`, `$derived` (no stores, no `writable`)
- Object-based `$state` for exported reactive variables (`wsState.connected`) — applies to queue state and theme state
- Component-scoped CSS in `<style>` blocks — dark mode variable migration touches only these
- Go handler pattern: take `*db.Database` and `*ws.Hub` as parameters, write JSON responses
- Chi router: `r.Route("/api", ...)` with `r.Post("/", ...)`, `r.Get("/", ...)` sub-routes
- `embed.FS` for static files — Service Worker `.js` file must be in the static build output

### Integration Points
- New `POST /api/sync` endpoint in `internal/handler/` — receives array of queued QSOs, inserts via `db.InsertQSO`, returns `[{client_id, server_id}, ...]`
- New Dexie.js store layer (`frontend/src/lib/db.js` or similar) — manages IndexedDB tables alongside existing reactive `$state` stores
- Service Worker registration in `app.html` or `+layout.svelte` — must be part of the SvelteKit static build
- Connection indicator in header — reads both `wsState.connected` and queue length, displays combined status
- Dupe check path in `QsoEntryForm` — branches: online → `api.checkDupe()`, offline → IndexedDB query
- QSO creation path — branches: online → `api.createQSO()`, offline → IndexedDB queue + optimistic UI update

## Specific Ideas

No specific requirements — open to standard approaches for Dexie.js schema, Service Worker implementation, and CSS variable naming.

## Deferred Ideas

None — discussion stayed within Phase 3 scope.

---

*Phase: 3-Offline Resilience & Polish*
*Context gathered: 2026-05-30*
