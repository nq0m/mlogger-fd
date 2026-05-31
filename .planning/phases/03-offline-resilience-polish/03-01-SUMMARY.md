---
phase: 03-offline-resilience-polish
plan: 01
subsystem: api
tags: [dexie, indexeddb, offline, sync, websocket, svelte5]

requires:
  - phase: 02-multi-user-real-time
    provides: WebSocket hub, real-time QSO broadcast, operator identity
provides:
  - POST /api/sync batch endpoint with client_id dedup
  - Dexie.js IndexedDB layer for offline QSO queue
  - Offline write path with optimistic UI
  - Auto-sync trigger on WebSocket reconnect with 30s retry
  - Connection indicator with queue count in header
affects: [offline-resilience-polish]

tech-stack:
  added: [dexie@^4.4.3]
  patterns:
    - Object-based $state for exported reactive values (syncState)
    - $effect for WebSocket-triggered sync auto-start
    - ON CONFLICT(client_id) DO NOTHING for idempotent batch sync

key-files:
  created:
    - internal/handler/sync.go
    - internal/handler/sync_test.go
    - frontend/src/lib/db.js
    - frontend/src/lib/sync.svelte.js
  modified:
    - internal/db/schema.sql
    - internal/model/qso.go
    - main.go
    - frontend/package.json
    - frontend/src/lib/api.js
    - frontend/src/lib/ws.svelte.js
    - frontend/src/lib/stores/qso.svelte.js
    - frontend/src/lib/components/QsoEntryForm.svelte
    - frontend/src/routes/+page.svelte

key-decisions:
  - "Used object-based $state (syncState) for exported queue/syncing values following wsState pattern"
  - "Sync trigger via $effect in sync.svelte.js rather than direct import in ws.svelte.js to avoid circular dependency"
  - "Dynamic import for db.js in qso.svelte.js to break potential circular module chain"
  - "1-second debounce on QSO submit to prevent accidental double-taps"

patterns-established:
  - "Object-based $state pattern for cross-module reactive primitives"
  - "Dexie.js schema: version(1).stores with simple key columns"
  - "Go batch handler pattern: iterate, validate all, insert with ON CONFLICT, broadcast each"

requirements-completed: [SYNC-03, SYNC-04, SYNC-05]

duration: 15min
completed: 2026-05-30
---

# Phase 3 Plan 01: Core Offline QSO Loop Summary

**Offline QSO buffering via Dexie.js IndexedDB, batch POST /api/sync with client_id dedup, and auto-sync triggered by WebSocket reconnect**

## Performance

- **Duration:** ~15 min
- **Tasks:** 3
- **Files modified:** 12 (5 created, 7 modified)

## Accomplishments
- Server `/api/sync` endpoint accepts batched QSOs, validates each, inserts with `ON CONFLICT(client_id) DO NOTHING`, broadcasts via WebSocket
- Dexie.js IndexedDB layer stores offline QSOs with client-generated UUIDs
- Svelte 5 `$effect` watches `wsState.connected` — fires immediate sync on reconnect + 30s periodic retry
- QsoEntryForm catches fetch failures and routes to IndexedDB offline path with optimistic UI
- Connection indicator in header shows "N queued" / "Syncing..." alongside Live/Disconnected status

## Task Commits

1. **Task 1: Dexie.js Package Verification + Install** - `d1b6753` (build)
2. **Task 2: Server Sync Endpoint + Frontend Foundation** - `5cfbacf` (feat)
3. **Task 3: Offline Write Path + Sync Trigger + Connection Indicator** - `b7dc3a1` (feat)

## Files Created/Modified
- `internal/handler/sync.go` — POST /api/sync handler with batch validation, insert, broadcast
- `internal/handler/sync_test.go` — 4 tests: empty body, valid batch, validation failure, client_id dedup
- `internal/db/schema.sql` — Added `client_id TEXT UNIQUE` column to qsos table
- `internal/model/qso.go` — Added `ClientID` field to QSO and CreateQSOInput structs
- `main.go` — Added `r.Post("/sync", ...)` inside `/api` Route block
- `frontend/src/lib/db.js` — Dexie.js FDLogger database with enqueueQso/getQueuedQsos/getQueueCount/clearQueued
- `frontend/src/lib/sync.svelte.js` — Sync manager with flushSyncQueue, refreshQueueCount, $effect watcher
- `frontend/src/lib/api.js` — Added `syncBatch(qsos)` function
- `frontend/src/lib/stores/qso.svelte.js` — Added `addQsoOffline()` for IndexedDB write + optimistic UI
- `frontend/src/lib/components/QsoEntryForm.svelte` — Network error catch routes to offline path, 1s debounce
- `frontend/src/routes/+page.svelte` — Queue count badge ("N queued"/"Syncing...") next to ws-status

## Decisions Made
- Used object-based `$state` for syncState (queueLength, syncing) following existing wsState pattern — Svelte 5 forbids reassigning primitive $state exports
- Sync trigger via `$effect` in sync.svelte.js rather than importing flushSyncQueue into ws.svelte.js — avoids circular dependency between ws ↔ sync modules
- Dynamic import for db.js in qso.svelte.js to break potential circular dependency chain
- 1-second debounce on submit handler prevents accidental double-taps in tent conditions

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Svelte 5 $state export pattern**
- **Found during:** Task 3 (sync.svelte.js creation)
- **Issue:** `export let queueLength = $state(0)` / `export let syncing = $state(false)` rejected by Svelte 5 — cannot export $state primitives that are reassigned
- **Fix:** Changed to object-based pattern: `export const queueState = $state({ queueLength: 0, syncing: false })`, matching existing wsState convention
- **Files modified:** frontend/src/lib/sync.svelte.js, frontend/src/routes/+page.svelte
- **Verification:** `npm run build` succeeds
- **Committed in:** b7dc3a1

---

**Total deviations:** 1 auto-fixed (missing critical — Svelte 5 constraint)
**Impact on plan:** Minor refactor to match existing codebase pattern. No functional change.

## Issues Encountered
None

## User Setup Required
None — no external service configuration required.

## Next Phase Readiness
- Core offline loop is complete: submit → IndexedDB → batch sync → server insert
- Plan 03-02 (offline dupe checking) can build on the IndexedDB cache from this plan
- Manual browser smoke test recommended before proceeding: start server, simulate offline in DevTools, submit QSO, go back online, verify sync

---
*Phase: 03-offline-resilience-polish*
*Completed: 2026-05-30*
