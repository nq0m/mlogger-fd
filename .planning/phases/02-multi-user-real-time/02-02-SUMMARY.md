---
phase: 02-multi-user-real-time
plan: 02
subsystem: api
tags: [gorilla/websocket, WebSocket, hub, broadcast, real-time, Go, TDD]

# Dependency graph
requires:
  - phase: 02-01
    provides: station-config endpoints, model.ValidateStationConfig, handler.GetStationConfig, handler.PutStationConfig
provides:
  - WebSocket Hub with channel-based broadcast fan-out to all connected clients
  - ServeWS HTTP-to-WebSocket upgrade handler at /ws
  - Real-time qso_created JSON broadcast after each QSO insert
  - gorilla/websocket v1.5.3 dependency
affects:
  - 02-03 (frontend WebSocket client + live UI updates)
  - 03-01 (offline sync — depends on WS event stream)

# Tech tracking
tech-stack:
  added: [github.com/gorilla/websocket v1.5.3]
  patterns:
    - "Channel-based Hub pattern: single Run() goroutine processes Register/Unregister/broadcast channels with RWMutex-protected client map"
    - "Per-client buffered send channels (64) with non-blocking select/default for slow-client protection"
    - "Closure-based dependency injection (hub passed to handler via chi route closure)"
    - "WritePump goroutine per connection, ReadPump blocks until disconnect"

key-files:
  created:
    - internal/ws/hub.go - Hub struct, Client struct, NewHub(), Run(), Broadcast(), ReadPump(), WritePump()
    - internal/ws/hub_test.go - 8 unit tests covering register, unregister, broadcast, fan-out, non-blocking send, concurrency, WritePump integration
    - internal/handler/ws.go - ServeWS HTTP handler with gorilla/websocket upgrader
    - internal/handler/ws_test.go - 5 tests covering upgrade, broadcast, nil-hub compat, concurrency, multi-message
  modified:
    - internal/handler/qso.go - CreateQSO signature extended with *ws.Hub, broadcast logic after DB insert
    - internal/handler/qso_test.go - Updated CreateQSO calls to pass hub=nil; added busy_timeout + cache=shared to test DB
    - main.go - Hub init, /ws route, hub passed to CreateQSO closure

key-decisions:
  - "Exported Hub.Register/Unregister channels for cross-package access from handler package"
  - "gorilla/websocket v1.5.3 after human-verified package legitimacy (44K+ importers, BSD-2-Clause)"
  - "CheckOrigin returns true for all origins (trusted LAN per AGENTS.md threat model accept)"
  - "Broadcast errors logged as slog.Warn — do not fail the HTTP request (non-critical path)"
  - "test DB uses file::memory:?cache=shared + SetMaxOpenConns(1) for concurrent test support"
  - "Station-config routes (GET/PUT /api/station-config) wired in main.go completing Plan 02-01 wiring"

patterns-established:
  - "TDD execution for both tasks: RED (failing test commit) → GREEN (implementation commit)"
  - "Hub goroutine is single writer to clients map — Register/Unregister/Broadcast all processed in one select loop"
  - "Client cleanup: ReadPump exits → defer unregisters client and closes conn; WritePump exits when Send channel closed by hub"

requirements-completed:
  - SYNC-01
  - SYNC-02

# Metrics
duration: 12 min
completed: 2026-05-30
---

# Phase 02 Plan 02: Real-Time WebSocket Infrastructure Summary

**WebSocket Hub with channel-based broadcast fan-out, real-time qso_created JSON delivery to all connected clients**

## Performance

- **Duration:** 12 min
- **Started:** 2026-05-30T04:02:47Z
- **Completed:** 2026-05-30T04:15:01Z
- **Tasks:** 3 (1 checkpoint, 2 auto/tdd)
- **Files modified:** 9

## Accomplishments

- WebSocket Hub with Register/Unregister/Broadcast channel architecture — single `Run()` goroutine owns client set
- ServeWS HTTP handler upgrades to WebSocket using gorilla/websocket v1.5.3, creates Client, routes to hub
- CreateQSO handler broadcasts full `qso_created` JSON to all connected clients after successful DB insert
- Per-client buffered send channels (64) with non-blocking select/default — slow clients are dropped, not blocking the hub
- SYNC-01 verified: 3 concurrent QSO inserts from separate goroutines succeed without conflicts
- SYNC-02 verified: WebSocket client receives `{"type":"qso_created",...}` message when any operator logs a QSO
- Backward compatibility: CreateQSO with `hub=nil` works for existing tests and standalone usage

## Task Commits

Each task was committed atomically:

1. **Task 1: Install gorilla/websocket v1.5.3** — Package installed and verified (checkpoint:human-verify approved, no separate commit needed)
2. **Task 2: WebSocket Hub (TDD)** — `efead2b` (RED), `82696a4` (GREEN)
3. **Task 3: ServeWS + CreateQSO broadcast + main.go wiring (TDD)** — `c25e193` (RED), `7dbb233` (GREEN)

**Plan metadata:** pending

## Files Created/Modified

- `internal/ws/hub.go` — Hub struct, Client struct, NewHub(), Run(), Broadcast(), ReadPump(), WritePump() with channel-based broadcast pattern
- `internal/ws/hub_test.go` — 8 unit tests: NewHub init, register, unregister, broadcast JSON, fan-out, non-blocking send, concurrent ops, WritePump integration
- `internal/handler/ws.go` — ServeWS handler with gorilla/websocket upgrader (1024 read/write buffers, CheckOrigin: true for trusted LAN)
- `internal/handler/ws_test.go` — 5 tests: upgrade, qso_created broadcast, nil-hub compatibility, concurrent inserts, multiple messages
- `internal/handler/qso.go` — Modified CreateQSO signature to `CreateQSO(db, hub, w, r)`; broadcast logic after DB insert; slog.Warn on broadcast error
- `main.go` — Added `import ws`, `hub := ws.NewHub(); go hub.Run()`, `/ws` route, `CreateQSO(database, hub, w, r)` closure
- `go.mod`, `go.sum` — Added `github.com/gorilla/websocket v1.5.3`

## Decisions Made

- **Exported Hub channels** (Register, Unregister) for cross-package access from handler package
- **gorilla/websocket v1.5.3** chosen over coder/websocket (790 importers) — gorilla has 44K+ importers, BSD-2-Clause, battle-tested
- **CheckOrigin: true** — LAN-only deployment, trusted network per AGENTS.md (threat model T-02-05 accept)
- **Broadcast errors non-fatal** — `slog.Warn` on marshal failure, HTTP response still succeeds (broadcast is best-effort)
- **Station-config routes wired** — completing the Plan 02-01 route registration that was deferred

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed missing busy_timeout pragma in test DB setup**
- **Found during:** Task 3 (CreateQSO broadcast test)
- **Issue:** `setupHandlerTestDB` using bare `:memory:` without `busy_timeout` or `cache=shared` caused concurrent test inserts to fail with "database is locked" errors on in-memory SQLite
- **Fix:** Changed DSN to `file::memory:?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&cache=shared` and added `db.SetMaxOpenConns(1)` to ensure all goroutines share the same in-memory database
- **Files modified:** `internal/handler/qso_test.go`
- **Verification:** TestCreateQSOConcurrent (3 concurrent inserts) passes
- **Committed in:** `c25e193` (Task 3 RED commit)

---

**Total deviations:** 1 auto-fixed (Rule 1 - Bug)
**Impact on plan:** Bug fix essential for SYNC-01 concurrency testing. No scope creep.

## Issues Encountered

None — implementation followed research patterns exactly.

## Next Phase Readiness

- WebSocket infrastructure complete — ready for frontend WebSocket client integration (Plan 02-03)
- Hub accepts connections at `/ws`, broadcasts all QSO events
- Test coverage: 13 tests across ws and handler packages, all passing with `-race`
- `go build ./...` compiles clean

---
*Phase: 02-multi-user-real-time*
*Completed: 2026-05-30*

## Self-Check: PASSED

