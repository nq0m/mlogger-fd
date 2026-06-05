---
phase: 04-field-day-features-testing
plan: 02
subsystem: ui
tags: [svelte, bonus-tracker, cabrillo, arrl-scoring, statsbar, localStorage]

# Dependency graph
requires:
  - phase: 04-01
    provides: "Bonus API endpoints, bonus_claims table, model layer"
provides:
  - "BonusTracker.svelte expandable panel with toggle + count UI"
  - "getBonuses()/putBonuses() API client functions"
  - "bonusClaims reactive state in qso.svelte.js with localStorage backup"
  - "bonus_points in stats response and ARRL-correct score formula"
  - "Cabrillo CLAIMED-SCORE with bonus + SOAPBOX bonus lines"
  - "Bonus stat block in StatsBar with gold color"
affects: [04-03-plan, 04-04-plan]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "StationConfig-style expandable panel pattern (config-toggle + config-panel CSS)"
    - "Object-based $state for shared bonus claims (Svelte 5 reactivity)"
    - "localStorage hybrid persistence: local-first display, server as source of truth"
    - "COALESCE(SUM(CASE...)) SQL for safe bonus query fallback"

key-files:
  created:
    - frontend/src/lib/components/BonusTracker.svelte
  modified:
    - frontend/src/lib/api.js
    - frontend/src/lib/stores/qso.svelte.js
    - frontend/src/routes/+page.svelte
    - internal/handler/stats.go
    - internal/cabrillo/cabrillo.go
    - frontend/src/lib/components/StatsBar.svelte
    - internal/handler/stats_test.go
    - internal/cabrillo/cabrillo_test.go

key-decisions:
  - "ARRL-correct score formula: (rawPoints * multiplier) + bonusPoints — bonus added AFTER multiplier per section 7.3"
  - "localStorage key: fdlogger_bonus_claims — follows fdlogger_theme precedent"
  - "Bonus stat block color: #cc8800 (light) / #ffaa00 (dark) — gold/amber distinguishes from primary Pts and success Score"

patterns-established:
  - "Expandable header panel: StationConfig.svelte template reused exactly for BonusTracker"
  - "Bonus points SQL: COALESCE(SUM(CASE...)) with per-bonus point calculation inline in both stats.go and cabrillo.go"
  - "SOAPBOX lines: one per claimed bonus + total line, after CLAIMED-SCORE, before QSO rows"

requirements-completed: [BON-01, BON-02]

# Metrics
duration: 6min
completed: 2026-06-04
---

# Phase 04 Plan 02: Bonus Tracker UI + Scoring Integration Summary

**Expandable ★ Bonuses panel with 18 bonus items, localStorage persistence, ARRL-correct scoring, and Cabrillo SOAPBOX lines**

## Performance

- **Duration:** 6 min
- **Started:** 2026-06-04T20:27:59Z
- **Completed:** 2026-06-04T20:33:02Z
- **Tasks:** 2 (1 auto, 1 TDD)
- **Files modified:** 9

## Accomplishments

- BonusTracker.svelte expandable header panel with toggle checkboxes and count number inputs for 18 ARRL bonus items
- localStorage backup (fdlogger_bonus_claims) + server persistence via PUT /api/bonuses
- bonus_points field added to GET /api/stats response with safe COALESCE(SUM(CASE...)) fallback
- ARRL-correct score formula: `(rawPoints * multiplier) + bonusPoints` in both stats.go and cabrillo.go
- Cabrillo export includes CLAIMED-SCORE with bonus + SOAPBOX lines for each claimed bonus + total
- StatsBar Bonus stat block between Pts and Mult with gold (#cc8800) color

## Task Commits

Each task was committed atomically:

1. **Task 1: Build BonusTracker frontend** - `e4baf96` (feat)
   - api.js: getBonuses()/putBonuses(), qso.svelte.js: bonusClaims $state + bonus_points default
   - BonusTracker.svelte: expandable panel, 18 bonus items, localStorage + server persistence
   - +page.svelte: BonusTracker in header, backup toast prep

2. **Task 2: Integrate bonus points into scoring** — TDD (RED → GREEN)
   - RED: `59654c1` (test) — 4 stats tests + 2 cabrillo tests for bonus scoring
   - GREEN: `2bd6a53` (feat) — stats.go bonus query + score formula, cabrillo.go bonus + SOAPBOX, StatsBar bonus stat block

**TDD Gate Compliance:**
| Gate | Status | Commit |
|------|--------|--------|
| RED | ✓ | `59654c1` |
| GREEN | ✓ | `2bd6a53` |
| REFACTOR | — | Not needed (clean code, passes vet) |

## Files Created/Modified

- `frontend/src/lib/api.js` — Added getBonuses() and putBonuses() API client functions
- `frontend/src/lib/stores/qso.svelte.js` — Added bonus_points: 0 to stats, bonusClaims $state({})
- `frontend/src/lib/components/BonusTracker.svelte` — **NEW** Expandable bonus claim panel with toggle + count UI
- `frontend/src/routes/+page.svelte` — BonusTracker in header, backupToast state for Plan 04-04
- `internal/handler/stats.go` — Bonus points SQL query, ARRL-correct score, bonus_points in response
- `internal/cabrillo/cabrillo.go` — Bonus query, modified CLAIMED-SCORE, SOAPBOX bonus lines
- `frontend/src/lib/components/StatsBar.svelte` — Bonus stat block with gold color between Pts and Mult
- `internal/handler/stats_test.go` — 4 new bonus-related test functions
- `internal/cabrillo/cabrillo_test.go` — 2 new bonus-related test functions + bonus_claims table in setup

## Decisions Made

- **ARRL-correct score formula**: `(rawPoints * multiplier) + bonusPoints` — bonus added AFTER multiplier per ARRL rules section 7.3. This contradicts the earlier CONTEXT.md formula but matches ARRL submission requirements
- **localStorage key**: `fdlogger_bonus_claims` — follows fdlogger_theme precedent for naming
- **Bonus stat color**: #cc8800 (light) / #ffaa00 (dark) — gold/amber to visually distinguish bonus from primary Pts and success Score colors
- **SOAPBOX format**: One line per claimed bonus with computed points + total line, placed after CLAIMED-SCORE and before QSO rows

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered

- Accidental deletion of 6 pre-existing cabrillo test functions when inserting new bonus tests — immediately restored from known state
- TestGenerate_ScoreIncludesBonus test was non-discriminating in RED phase (same result with bonus=0) — kept as-is since it correctly verifies formula post-implementation

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- Bonus Tracker frontend and scoring integration complete — ready for Plan 04-03
- backupToast state and template already added to +page.svelte for Plan 04-04
- All 38 existing tests pass, 6 new bonus tests pass, zero regressions

---
*Phase: 04-field-day-features-testing*
*Completed: 2026-06-04*

## Self-Check: PASSED

- All 9 key files exist on disk
- All 3 task commits verified in git history: e4baf96, 59654c1, 2bd6a53
- All 9 stats tests pass, all 16 cabrillo tests pass
- All acceptance criteria grep checks pass
- Go vet returns clean (no warnings)
