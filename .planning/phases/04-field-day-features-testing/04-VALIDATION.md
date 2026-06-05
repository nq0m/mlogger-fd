---
phase: 04
slug: field-day-features-testing
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-06-04
---

# Phase 04 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` (stdlib) + vitest for frontend |
| **Config file** | `go test ./...` (Go), vitest.config.ts (frontend) |
| **Quick run command** | `go test ./internal/handler/ -run TestSimulation -timeout 120s` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~120 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/handler/ -count=1 -timeout 30s`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd-verify-work`:** Full suite must be green + simulation test passing
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 04-01-01 | 01 | 1 | BON-01 | T-04-01 | Validate bonus_id against known list; reject unknown IDs | integration | `go test ./internal/handler/ -run TestBonuses` | ❌ W0 | ⬜ pending |
| 04-01-02 | 01 | 1 | BON-02 | T-04-02 | Correct ARRL formula for bonus points | unit | `go test ./internal/handler/ -run TestStats -count=1` | ❌ W0 | ⬜ pending |
| 04-02-01 | 02 | 2 | UX-03 | — | N/A | unit | `vitest run src/lib/audio.test.js` | ❌ W0 | ⬜ pending |
| 04-02-02 | 02 | 2 | BKUP-01 | T-04-03 | Generic error messages; no filesystem paths exposed | integration | `go test ./internal/handler/ -run TestBackup` | ❌ W0 | ⬜ pending |
| 04-03-01 | 03 | 3 | D-13 | — | N/A | integration | `go test ./internal/handler/simtest/ -run TestSimulation -timeout 120s` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/handler/bonus_test.go` — covers GetBonuses/PutBonuses handlers
- [ ] `internal/handler/backup_test.go` — covers DownloadBackup handler
- [ ] `internal/handler/simtest/simtest_test.go` — multi-client simulation test
- [ ] `internal/model/bonus_test.go` — bonus claim validation, default list structure
- [ ] `internal/handler/stats_test.go` (modify) — add bonus_points assertion to existing tests
- [ ] `internal/cabrillo/cabrillo_test.go` (modify) — add CLAIMED-SCORE with bonus assertion
- [ ] `internal/db/schema.sql` — add `bonus_claims` table to test DB setup functions
- [ ] `frontend/src/lib/audio.test.js` — mock Web Audio API

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Mute toggle visual feedback | UX-03 | Audio state change is DOM-visual; no programmatic check possible | Click mute icon, verify icon changes and no audio plays on QSO submit |
| Audio playback on QSO confirm | UX-03 | Web Audio API playback requires real browser context; synthetic mocks cannot verify actual sound | Submit QSO, verify confirmation sound plays (subject to autoplay policy) |
| Field test outdoor setup | D-14 | Physical deployment | Set up in park with 2+ devices, log QSOs, verify sync and data integrity |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
