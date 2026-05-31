---
phase: 03-offline-resilience-polish
plan: 02
subsystem: ui
tags: [dexie, indexeddb, offline, dupe-check, svelte5]

requires:
  - phase: 03-01
    provides: Dexie.js IndexedDB layer, POST /api/sync, wsState
provides:
  - IndexedDB cached_qsos table with compound [callsign+band+mode] index
  - Offline dupe checking against cached + queued QSOs
  - Cache population on page load from GET /api/qso?limit=9999
  - Real-time cache update via WebSocket qso_created events
affects: [offline-resilience-polish]

tech-stack:
  added: []
  patterns:
    - Compound Dexie index [callsign+band+mode] for efficient dupe queries
    - Offline/online branching in handleCheckDupe based on wsState.connected

key-files:
  created: []
  modified:
    - frontend/src/lib/db.js
    - frontend/src/lib/stores/qso.svelte.js
    - frontend/src/lib/ws.svelte.js
    - frontend/src/lib/components/QsoEntryForm.svelte
    - frontend/src/routes/+page.svelte

key-decisions:
  - "Bumped Dexie schema to version(2) to add cached_qsos table alongside existing queued_qsos"
  - "Callsigns normalized to uppercase in cache for case-insensitive dupe matching"
  - "Queued (unsynced) QSOs included in offline dupe check via JS iteration (queue is small, <50 items)"
  - "Offline dupe does NOT include similar_calls — partial similarity remains server-only per D-06"

patterns-established:
  - "Compound Dexie index [callsign+band+mode] for keyed lookups"
  - "Offline/online dupe check branching pattern in Svelte component"

requirements-completed: [SYNC-06]

duration: 5min
completed: 2026-05-30
---

# Phase 3 Plan 02: Offline Dupe Checking Summary

**IndexedDB cache with compound [callsign+band+mode] index, offline dupe check against cached + queued QSOs, and real-time cache update via WebSocket**

## Performance

- **Duration:** ~5 min
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- Dexie schema bumped to version 2 with new `cached_qsos` table and `[callsign+band+mode]` compound index
- `offlineDupeCheck()` queries IndexedDB cache and queued QSOs for exact callsign+band+mode matches
- Cache populated on page load via `GET /api/qso?limit=9999` with `bulkPut` for idempotent upsert
- WebSocket `qso_created` events update the cache in real-time so dupe data stays current
- QsoEntryForm `handleCheckDupe` branches: offline uses IndexedDB, online uses server API unchanged

## Task Commits

1. **Task 1: IndexedDB Cache Table + Population** - `65bac83` (feat)
2. **Task 2: Offline Dupe Check Integration in QsoEntryForm** - `20f10f7` (feat)

## Files Created/Modified
- `frontend/src/lib/db.js` — Added version(2) with cached_qsos table, loadCachedQsos, populateCache, addToCache, offlineDupeCheck
- `frontend/src/lib/stores/qso.svelte.js` — Added loadCache() with top-level auto-call for page load cache population
- `frontend/src/lib/ws.svelte.js` — Added addToCache() call in qso_created handler for real-time cache updates
- `frontend/src/lib/components/QsoEntryForm.svelte` — handleCheckDupe branches on wsState.connected for offline vs online dupe
- `frontend/src/routes/+page.svelte` — Added loadCache() call in onMount

## Decisions Made
- Bumped Dexie version from 1 to 2 to add cached_qsos — Dexie auto-upgrades on schema change
- Callsigns normalized to uppercase in cache for case-insensitive matching
- Queued QSO iteration uses JS filtering (not Dexie compound query) since queued_qsos stores nested objects
- Offline branch returns `{ is_dupe: true/false }` only — no `similar_calls` field, consistent with D-06 decision that partial similarity is server-only

## Deviations from Plan
None — plan executed as written.

## Issues Encountered
None

## User Setup Required
None

## Next Phase Readiness
- Offline dupe checking complete — operators get DUPE warnings in both online and offline modes
- Plan 03-03 (dark mode) has no dependency on this plan beyond the existing components
- Manual verification recommended: offline mode + same callsign/band/mode should show DUPE warning

---
*Phase: 03-offline-resilience-polish*
*Completed: 2026-05-30*
