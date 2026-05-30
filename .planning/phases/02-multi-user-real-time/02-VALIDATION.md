---
phase: 02
slug: multi-user-real-time
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-05-29
---

# Phase 02 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go: `testing` (stdlib) + `httptest`; Frontend: `vitest` (via SvelteKit/Vite) |
| **Config file** | Go: none (convention: `_test.go`); Frontend: `vitest.config.ts` (in frontend/) |
| **Quick run command** | `go test ./... -count=1` (Go); `cd frontend && npx vitest run` (Frontend) |
| **Full suite command** | `go test ./... -v -count=1 -race` (Go); `cd frontend && npx vitest run` (Frontend) |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./... -count=1 && cd frontend && npx vitest run`
- **After every plan wave:** Full suite with `-race` flag
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 02-01-01 | 01 | 1 | CONF-01 | T-02-01 | Validate station config fields before DB write | integration | `go test ./internal/handler/ -run TestPutStationConfig -v` | ❌ W0 | ⬜ pending |
| 02-01-02 | 01 | 1 | CONF-03 | — | Parameterized SQL for station config insert | integration | `go test ./internal/handler/ -run TestStationConfigPersistence -v` | ❌ W0 | ⬜ pending |
| 02-01-03 | 01 | 1 | CONF-01 | — | GET returns stored config; defaults if unconfigured | integration | `go test ./internal/handler/ -run TestGetStationConfig -v` | ❌ W0 | ⬜ pending |
| 02-01-04 | 01 | 1 | CONF-02 | — | QSO create sends operator field from client | integration | `go test ./internal/handler/ -run TestCreateQSOWithOperator -v` | ❌ W0 | ⬜ pending |
| 02-01-05 | 01 | 1 | — | T-02-02 | WebSocket origin check on trusted LAN | integration | `go test ./internal/ws/ -run TestWebSocketUpgrade -v` | ❌ W0 | ⬜ pending |
| 02-02-01 | 02 | 2 | SYNC-01 | — | Hub registers multiple clients; concurrent QSO inserts succeed | integration | `go test ./internal/ws/ -run TestHubBroadcast -v` | ❌ W0 | ⬜ pending |
| 02-02-02 | 02 | 2 | SYNC-02 | — | WebSocket broadcasts full QSO JSON on create | unit | `go test ./internal/ws/ -run TestHubRegister -v` | ❌ W0 | ⬜ pending |
| 02-02-03 | 02 | 2 | SYNC-02 | T-02-03 | Write pump uses buffered channels; non-blocking sends prevent head-of-line blocking | unit | `go test ./internal/ws/ -run TestWritePump -v` | ❌ W0 | ⬜ pending |
| 02-02-04 | 02 | 2 | — | — | Handler closure pattern for hub dependency injection | integration | `go test ./internal/handler/ -run TestCreateQSOBroadcast -v` | ❌ W0 | ⬜ pending |
| 02-03-01 | 03 | 2 | SYNC-02 | — | Client WebSocket listener updates qsos $state | unit (frontend) | `cd frontend && npx vitest run src/lib/ws.test.js` | ❌ W0 | ⬜ pending |
| 02-03-02 | 03 | 2 | CONF-02 | — | OperatorSelector persists to localStorage | unit (frontend) | `cd frontend && npx vitest run src/lib/components/OperatorSelector.test.js` | ❌ W0 | ⬜ pending |
| 02-04-01 | 04 | 2 | — | — | Cabrillo export reads real station config from DB | unit | `go test ./internal/cabrillo/ -run TestCabrilloWithConfig -v` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/ws/hub_test.go` — Hub register/unregister/broadcast tests
- [ ] `internal/handler/config_test.go` — Station config GET/PUT tests
- [ ] `internal/handler/qso_test.go` — Update existing to test WebSocket broadcast
- [ ] `internal/cabrillo/cabrillo_test.go` — Update to test real station config
- [ ] `frontend/src/lib/ws.test.js` — WebSocket client unit test (mock WebSocket)
- [ ] `frontend/src/lib/components/OperatorSelector.test.js` — Operator selector + localStorage test
- Framework install: All frameworks already installed from Phase 1

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Two browsers logging simultaneously see each other's QSOs | SYNC-01, SYNC-02 | Multi-client WebSocket interaction requires real browsers | Open two browser tabs, log QSO in tab A, verify it appears in tab B's log table within 1 second |
| Station config visible to all clients | CONF-01 | Multi-client state sharing | Set station config in one browser, open second browser, verify config is visible |
| Operator identity per session | CONF-02 | Client-side state isolation | Set operator to "K1ABC" in tab A, "N2XYZ" in tab B, log QSOs, verify correct operator on each |
| Station config survives server restart | CONF-03 | Persistence across process boundaries | Set config, restart server, verify GET returns same values |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
