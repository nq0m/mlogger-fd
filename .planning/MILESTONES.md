# Milestones

## v1.0.0 v1.0 MVP (Shipped: 2026-06-05)

**Phases completed:** 4 phases, 19 plans, 43 tasks

**Key accomplishments:**

- Walking Skeleton — Go chi router + SvelteKit SPA + SQLite with end-to-end QSO CRUD
- Server-side dupe checking with exact band+mode match + partial call similarity, client-side blur/submit warnings
- Real-time rate meter, score display, and band/mode breakdown — updating after every QSO
- Searchable log with prefix matching, click-to-edit inline rows, and pagination
- Valid ARRL Field Day Cabrillo v3.0 generation with fixed-width QSO lines and one-click download
- Vitest test infrastructure for StationConfig, WebSocket client, and OperatorSelector with Svelte 5 + jsdom — 8 passing tests across 3 test files, clearing the path for TDD-driven Plans 02-01 through 02-03.
- Complete station config vertical slice: SQLite table → REST API → Svelte UI form with validation and persistence.
- WebSocket Hub with channel-based broadcast fan-out, real-time qso_created JSON delivery to all connected clients
- Real-time QSO sync via WebSocket with deduplication, localStorage-backed operator identity, and live connection status indicator in the SPA header
- Cabrillo export reads station callsign, class, ARRL section, and power from station_config with N0CALL/NH/LOW/1D defaults; export filename uses real lowercased callsign
- Offline QSO buffering via Dexie.js IndexedDB, batch POST /api/sync with client_id dedup, and auto-sync triggered by WebSocket reconnect
- IndexedDB cache with compound [callsign+band+mode] index, offline dupe check against cached + queued QSOs, and real-time cache update via WebSocket
- 14 CSS custom properties on :root, [data-theme='dark'] override block, sun/moon toggle in header, zero hardcoded colors in components
- CSS media queries at 500px/768px breakpoints, 48px touch targets, SvelteKit Service Worker with versioned app shell caching, debounce guard confirmed
- Three vitest test files with placeholder source files for BonusTracker, audio, and QsoEntryForm audio trigger contracts — resolves Nyquist violations 8a, 8c, and 8d with 7 passing and 6 skipped tests
- SQLite bonus_claims table, 18-item ARRL 2026 bonus list, GET/PUT /api/bonuses handlers with validation, and Chi route wiring — all TDD with 32 tests
- Expandable ★ Bonuses panel with 18 bonus items, localStorage persistence, ARRL-correct scoring, and Cabrillo SOAPBOX lines
- Web Audio API sound feedback for QSO confirmation and dupe warnings, with persistent mute toggle in header bar
- One-click SQLite backup via io.Copy streaming with timestamped filename, Go integration test simulating 3 clients × 210 QSOs with data integrity assertions, and field test checklist for outdoor deployment verification.

---
