---
phase: 04-field-day-features-testing
plan: 01
subsystem: api
tags: [go, sqlite, chi, bonus, arrl, field-day, tdd]

# Dependency graph
requires: []
provides:
  - "bonus_claims SQLite table with idempotent schema migration"
  - "BonusItem and BonusClaim Go structs with JSON serialization"
  - "DefaultBonuses: hardcoded 2026 ARRL Field Day 18-item bonus list (rules §7.3)"
  - "ValidateBonusClaim: count validation per bonus type limits"
  - "CalculateBonusPoints: ARRL-correct post-multiplier scoring formula"
  - "GET /api/bonuses: returns all 18 bonus items with claim state"
  - "PUT /api/bonuses: validates, persists, returns full state; unknown IDs silently skipped"
  - "Chi router wiring in main.go for /api/bonuses (GET and PUT)"
affects: [bonus-tracker-ui, stats-integration, cabrillo-export]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "TDD RED-GREEN cycle: 2 commits per task (test → feat)"
    - "Model validation pattern: error-string returns following ValidateStationConfig convention"
    - "Handler pattern: httptest.NewRecorder direct handler calls following config_test.go pattern"
    - "Chi closure wiring: database capture pattern from existing routes"

key-files:
  created:
    - internal/model/bonus.go - BonusItem/BonusClaim structs, DefaultBonuses, validation, scoring
    - internal/model/bonus_test.go - 24 tests covering model layer
    - internal/handler/bonus.go - GetBonuses/PutBonuses HTTP handlers
    - internal/handler/bonus_test.go - 8 tests covering handler layer
  modified:
    - internal/db/schema.sql - Added bonus_claims table DDL
    - main.go - Added GET/PUT /api/bonuses route wiring

key-decisions:
  - "Bonus points computed post-multiplier per ARRL rules §7.3: score = (rawPoints * multiplier) + bonusPoints"
  - "Unknown bonus IDs silently skipped at handler level (not model) — defense-in-depth: model validates, handler filters"
  - "GOTA coach bonus triggered at count≥10 (not separate boolean) per plan specification"
  - "MaxCount=0 means unlimited (GOTA QSOs) — only enforce non-zero MaxCount limits"

patterns-established:
  - "TDD for model layer: 24 tests covering structs, JSON tags, validation, scoring edge cases"
  - "TDD for handler layer: 8 tests covering GET, PUT, persistence, validation, error handling"
  - "Handler test DB setup: in-memory SQLite with cache=shared, SetMaxOpenConns(1)"

requirements-completed: [BON-01]

# Metrics
duration: 6min
completed: 2026-06-05
---

# Phase 04 Plan 01: Bonus Tracker Backend Summary

**SQLite bonus_claims table, 18-item ARRL 2026 bonus list, GET/PUT /api/bonuses handlers with validation, and Chi route wiring — all TDD with 32 tests**

## Performance

- **Duration:** 6 min
- **Started:** 2026-06-05T01:07:34Z
- **Completed:** 2026-06-05T01:14:04Z
- **Tasks:** 2 (both TDD: RED + GREEN = 4 commits)
- **Files modified:** 6

## Accomplishments
- bonus_claims table with idempotent CREATE TABLE IF NOT EXISTS in schema.sql
- BonusItem and BonusClaim Go structs with validated JSON tags
- Hardcoded 18-item 2026 ARRL Field Day bonus list (DefaultBonuses constant) per rules §7.3
- ValidateBonusClaim: checks bonus_id existence, non-negative count, and per-bonus MaxCount limits
- CalculateBonusPoints: correct ARRL formula with special cases for emergency_power, message_handling, youth_participation, GOTA (with coach bonus), web_submission, safety_officer, site_responsibilities
- GET /api/bonuses: returns full 18-item map with defaults when DB empty
- PUT /api/bonuses: validates, persists in transaction, unknown IDs silently skipped per ASVS V5
- Chi router wiring in main.go following existing closure pattern

## Task Commits

Each TDD task produced RED + GREEN commits:

1. **Task 1: bonus_claims schema + model layer** — `d0b59df` (test), `93fcca1` (feat)
2. **Task 2: bonus API handler + route wiring** — `7f55c3f` (test), `0ef74df` (feat)

## Files Created/Modified
- `internal/db/schema.sql` — Added bonus_claims table (idempotent migration)
- `internal/model/bonus.go` — BonusItem, BonusClaim, DefaultBonuses (18 items), ValidateBonusClaim, CalculateBonusPoints
- `internal/model/bonus_test.go` — 24 tests: JSON tags, DefaultBonuses completeness, validation edge cases, scoring formula
- `internal/handler/bonus.go` — GetBonuses (defaults overlay + DB read), PutBonuses (validate + transaction + full response)
- `internal/handler/bonus_test.go` — 8 tests: empty table GET, data GET, valid PUT, persistence round-trip, unknown ID skip, negative count rejection, MaxCount rejection, malformed JSON handling
- `main.go` — Added GET/PUT /api/bonuses route closures between station-config and sync

## Decisions Made
- Bonus points computed post-multiplier per ARRL rules §7.3: `score = (rawPoints * multiplier) + bonusPoints` — contradicts CONTEXT.md pre-multiplier formula but follows official rules
- Unknown bonus IDs silently skipped at handler level (not model) — defense-in-depth: model validates, handler filters
- GOTA coach bonus triggered at count≥10 (not separate boolean) per plan specification
- MaxCount=0 means unlimited (GOTA QSOs) — only enforce non-zero MaxCount limits

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Unknown bonus IDs were not skipped during handler validation**
- **Found during:** Task 2 (GREEN phase — TestPutBonuses_UnknownBonusID)
- **Issue:** ValidateBonusClaim was called before the unknown-ID filter, causing a 400 error instead of silently skipping
- **Fix:** Reordered validation: check DefaultBonuses membership first, skip unknown IDs, then call ValidateBonusClaim only for known IDs; also added skip check to insertion loop
- **Files modified:** internal/handler/bonus.go
- **Verification:** TestPutBonuses_UnknownBonusID passes (200 OK, invalid_bonus not in response, valid claim persisted)
- **Committed in:** 0ef74df (Task 2 GREEN commit)

---

**Total deviations:** 1 auto-fixed (Rule 1 - Bug)
**Impact on plan:** Validation order fix was necessary for correct behavior per security guidance. No scope creep.

## Issues Encountered
None

## User Setup Required
None — no external service configuration required.

## Next Phase Readiness
- Bonus tracker backend complete — ready for frontend BonusTracker component (Plan 04-02 or 04-03)
- Score calculation `CalculateBonusPoints` ready for integration into stats.go and cabrillo.go in downstream plans
- All 32 tests pass (24 model + 8 handler), no regressions in existing tests

---
*Phase: 04-field-day-features-testing*
*Completed: 2026-06-05*
