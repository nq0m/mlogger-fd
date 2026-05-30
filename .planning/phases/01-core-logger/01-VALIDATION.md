---
phase: 1
slug: core-logger
status: draft
nyquist_compliant: true
wave_0_complete: false
created: 2026-05-29
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (Go stdlib) + vitest (SvelteKit frontend) |
| **Config file** | vitest.config.ts (frontend), none for Go (stdlib) |
| **Quick run command** | `go test ./...` (backend), `npx vitest run` (frontend) |
| **Full suite command** | `go test -v -race ./... && npx vitest run --reporter=verbose` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./...` (or `npx vitest run` if frontend task)
- **After every plan wave:** Run full suite
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| TBD (populated during planning) | 01 | 1 | REQ-QSO-01 | — | N/A | unit | `go test ./...` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `backend/db/db_test.go` — stubs for SQLite schema and queries
- [ ] `backend/handlers/handlers_test.go` — stubs for REST API handlers
- [ ] `backend/scoring/scoring_test.go` — stubs for points calculation and dupe detection
- [ ] `frontend/src/lib/api.test.ts` — stubs for API client module
- [ ] `frontend/src/lib/validation.test.ts` — stubs for callsign/field validation
- [ ] `vitest.config.ts` — vitest configuration for SvelteKit project
- [ ] Go module init + `go test ./...` baseline passing

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Keyboard shortcut behavior (Ctrl+Enter, Tab order) | QSO-04 | Browser key events require DOM-level testing | Verify Tab moves between fields, Ctrl+Enter submits form |
| Cabrillo file format validity | EXPR-01, EXPR-02 | Requires ARRL format validation | Open downloaded .cbr file, verify header fields and QSO lines match spec |
| Inline edit UI behavior | QSO-03 | Requires DOM interaction for edit/save/cancel | Click row, edit fields, save/cancel in place |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 30s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
