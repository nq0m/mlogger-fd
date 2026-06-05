# Project Retrospective

*A living document updated after each milestone. Lessons feed forward into future planning.*

## Milestone: v1.0.0 — MVP

**Shipped:** 2026-06-05
**Phases:** 4 | **Plans:** 19 | **Sessions:** multiple

### What Was Built
- Go backend with SQLite WAL, chi router, gorilla/websocket hub — single binary ~15MB
- SvelteKit SPA with QSO entry, log table, stats dashboard, dark mode, mobile layout
- Real-time multi-user WebSocket sync with channel-based broadcast fan-out
- Offline-first IndexedDB buffer via Dexie.js with dedup, batch sync, and auto-reconnect
- ARRL Field Day Cabrillo v3.0 export with config-driven headers
- Bonus points tracker with 18 predefined ARRL 2026 bonuses
- Web Audio API sound feedback, Service Worker app shell caching, one-click DB backup

### What Worked
- Offline-first architecture: IndexedDB→queue→sync pattern was correct from the start
- SQLite WAL mode on single-file DB: zero config, concurrent reads work perfectly
- Channel-based WebSocket hub with non-blocking per-client sends: clean, race-free
- Svelte 5 $state runes with exported object wrappers: solved reactivity across modules
- TDD: test-first in later phases caught regressions from parallel plan merges
- Each phase built strictly on the previous — no rework or backtracking

### What Was Inefficient
- Svelte 5 compatibility quirks ($state export rules, .svelte.js extension requirement)
- gorilla/websocket vs coder/websocket decision required research (settled on gorilla)
- Vitest + jsdom DOM isolation required explicit afterEach(cleanup) in each file
- Concurrent SQLite in-memory test DBs needed `cache=shared` + `SetMaxOpenConns(1)`
- No automated integration/E2E test harness — multi-client testing was manual

### Patterns Established
- Object-based $state wrappers for exported reactive state (`wsState.connected`, `syncState`)
- ON CONFLICT(client_id) DO NOTHING for idempotent offline sync batches
- Silent fallback on config read errors for non-critical paths (Cabrillo export)
- CSS custom properties on `:root` with `[data-theme='dark']` override block
- localStorage for operator identity, bonus claims — no server session state

### Key Lessons
1. Offline-first is architecture, not a feature — design the entire write path around it from day one
2. Svelte 5's reactivity model works well but requires care with export boundaries
3. SQLite is sufficient for Field Day scale; don't over-engineer with Postgres
4. WebSocket hubs need non-blocking sends or a slow client blocks everyone
5. Service Worker caching strategy needs careful versioning to avoid stale app shells

### Cost Observations
- Model mix: TBD
- Sessions: multiple across 6 days
- Notable: Wave-based parallel execution kept each phase to 1-2 sessions

---

## Cross-Milestone Trends

### Process Evolution

| Milestone | Sessions | Phases | Key Change |
|-----------|----------|--------|------------|
| v1.0.0 | multiple | 4 | Initial build — established conventions and architecture |

### Cumulative Quality

| Milestone | Tests | Coverage | Zero-Dep Additions |
|-----------|-------|----------|-------------------|
| v1.0.0 | 100+ | Go packages + vitest 22+ | gorilla/websocket, Dexie.js |

### Top Lessons (Verified Across Milestones)

1. TDD catches integration conflicts that individual worktree self-checks miss
2. Parallel execution requires careful files_modified overlap checking per wave
