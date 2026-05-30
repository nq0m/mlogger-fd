---
phase: 03
slug: offline-resilience-polish
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-05-30
---

# Phase 03 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | vitest 4.1.7 + jsdom 29.1.1 + @testing-library/svelte 5.3.1 (frontend), go test (backend) |
| **Config file** | `frontend/vitest.config.ts` |
| **Quick run command** | `npx vitest run --reporter=verbose` |
| **Full suite command** | `npx vitest run` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `npx vitest run --reporter=verbose`
- **After every plan wave:** Run `npx vitest run`
- **Before `/gsd-verify-work`:** Full suite must be green + manual browser testing for UX
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 03-{P}-{T} | {P} | {W} | SYNC-03 | T-03-01 | IndexedDB buffers QSOs when server unreachable | unit | `npx vitest run src/lib/db.test.js -t "buffers qso when offline"` | ❌ W0 | ⬜ pending |
| 03-{P}-{T} | {P} | {W} | SYNC-04 | T-03-02 | Batch sync flushes queue on reconnect | integration | `npx vitest run src/lib/sync.test.js -t "flushes queue on connect"` | ❌ W0 | ⬜ pending |
| 03-{P}-{T} | {P} | {W} | SYNC-05 | N/A | Connection indicator shows online/offline state | unit | `npx vitest run src/routes/page.test.js -t "shows connection status"` | ❌ W0 | ⬜ pending |
| 03-{P}-{T} | {P} | {W} | SYNC-06 | T-03-03 | Dupe check uses IndexedDB when offline | unit | `npx vitest run src/lib/db.test.js -t "dupe check uses IndexedDB"` | ❌ W0 | ⬜ pending |
| 03-{P}-{T} | {P} | {W} | UX-01 | N/A | Mobile-responsive layout with 48px touch targets | manual | Manual: browser resize + touch simulator | N/A | ⬜ pending |
| 03-{P}-{T} | {P} | {W} | UX-02 | N/A | Dark mode renders correctly across all components | manual | Manual: toggle + visual inspection | N/A | ⬜ pending |
| 03-{P}-{T} | {P} | {W} | UX-04 | N/A | SW serves cached SPA when server unreachable | manual | Manual: stop server, reload page | N/A | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*
*Note: Task IDs filled in after plans are created.*

---

## Wave 0 Requirements

- [ ] `frontend/src/lib/db.test.js` — stubs for SYNC-03, SYNC-06 (IndexedDB buffering + dupe check)
- [ ] `frontend/src/lib/sync.test.js` — stubs for SYNC-04 (batch sync on reconnect)
- [ ] `frontend/src/routes/page.test.js` — stubs for SYNC-05 (connection indicator component)
- [ ] `internal/handler/sync_test.go` — Go handler test for POST /api/sync
- [ ] `npm install dexie@^4.4.3` — Dexie.js not yet installed
- [ ] Vitest config update for IndexedDB mocking if dexie doesn't work with jsdom shim

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Mobile-responsive layout | UX-01 | Requires real viewport resize across devices | Resize browser to 480px/768px widths; verify form wraps, touch targets ≥48px |
| Dark mode rendering | UX-02 | Requires visual inspection of all components | Toggle dark mode; verify all components render correctly across all pages |
| Service Worker offline load | UX-04 | Requires server stop + browser refresh | Stop fdlogger server; reload page in browser; verify SPA loads from cache |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
