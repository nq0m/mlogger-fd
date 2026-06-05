# Phase 4: Field Day Features & Testing - Context

**Gathered:** 2026-06-04
**Status:** Ready for planning

## Phase Boundary

Delivers Field Day-specific features (bonus tracking, audio alerts, one-click database backup) and real-world testing. The bonus tracker uses the 2026 ARRL Field Day bonus list with toggle/count claims and integrates into scoring and Cabrillo export. Audio alerts play pre-recorded sound files for QSO confirmation and dupe warning. Backup downloads the live SQLite file with timestamped filenames. A scripted Go integration test validates core data integrity under simulated load, and a manual field test verifies the system outdoors.

## Implementation Decisions

### Bonus Tracker
- **D-01:** Use the current year's (2026) official ARRL Field Day bonus points list — predefined, fixed list. Researcher should look up the 2026 list and hardcode it.
- **D-02:** Toggle + count per bonus item. Boolean bonuses (e.g., emergency power) use a simple on/off toggle. Counted bonuses (e.g., youth participants, formal traffic messages) use a toggle plus a number input field.
- **D-03:** Bonus tracker UI lives as an expandable header panel, following the StationConfig pattern. A "Bonuses" button in header-right (between StationConfig and Export) toggles an inline panel showing the bonus list. Consistent with existing header controls.
- **D-04:** Bonus claims persist server-side in a new SQLite `bonus_claims` table, fetched/saved via REST API (`GET /api/bonuses`, `PUT /api/bonuses`), with localStorage backup using the hybrid pattern from OperatorSelector. Server is the source of truth, localStorage provides resilience against page reloads.
- **D-05:** Bonus points must be reflected in score calculation (stats endpoint) and Cabrillo export (CLAIMED-SCORE line and SOAPBOX/X-BONUS lines).

### Audio Feedback
- **D-06:** Pre-recorded sound files (.wav or .mp3) — user provides the audio files. Confirmation beep and dupe buzz are distinct audio files.
- **D-07:** Mute toggle in the header bar (speaker icon near theme toggle), persisted in localStorage, default unmuted.
- **D-08:** Sounds fire only for own QSOs — confirmation on successful create (local submit or sync success), dupe buzz on form dupe detection. No sounds for other operators' QSOs from WebSocket.

### Backup
- **D-09:** One-click backup button in header-right (next to Cabrillo Export), with brief confirmation toast ("Backup downloaded") following the StationConfig save-feedback pattern.
- **D-10:** Timestamped filename: `fdlogger_backup_{YYYYMMDD}_{HHMMSS}.db`.
- **D-11:** Stream the live SQLite file as-is. WAL mode ensures readers don't block writers — safe during active logging. No VACUUM INTO needed.

### Testing Strategy
- **D-12:** Scripted Go integration test for the 2-hour simulation — lives in internal test package, run as part of `go test ./...`. Uses `httptest` for in-process server testing.
- **D-13:** Simulation validates core data integrity: no QSOs lost in logging or sync, dupe detection correct, stats remain accurate across the run, sync between simulated clients works without data corruption.
- **D-14:** Field test is a minimal checklist — set up in a park, log QSOs from 2+ devices, verify everything works. No formal test documentation required.

### the agent's Discretion
- Specific 2026 ARRL FD bonus list items and point values (researcher should look up official list)
- Exact SQLite schema for `bonus_claims` table (row per bonus, columns for claimed boolean + count integer)
- API request/response format for bonus claims
- Audio file format, loading approach (Web Audio API decodeAudioData from static assets), and playback logic
- Exact toast implementation for backup confirmation (reuse StationConfig feedback pattern)
- Simulation test: number of simulated clients, QSO rate, test duration, specific assertions on integrity checks
- Bonus tracker component structure (reuses StationConfig expand/collapse pattern)
- Mute toggle icon choice (speaker/speaker-off), exact position in header

## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Foundation
- `.planning/PROJECT.md` — Project context, constraints (RPi 4, SQLite, LAN-only, gloves/wet-fingers UI). Core value: log must never be lost.
- `.planning/REQUIREMENTS.md` — Phase 4 requirement: UX-03. Also: BON-01, BON-02, BKUP-01 are v2 deferred but included in Phase 4 per roadmap.
- `.planning/ROADMAP.md` § Phase 4 — Scope anchor, success criteria, key deliverables
- `.planning/STATE.md` — Phase 1-3 decisions: object-based `$state` pattern, WebSocket reconnect, CSS variable system, 48×48px touch targets, gorilla/websocket

### Prior Phase Context
- `.planning/phases/01-core-logger/01-CONTEXT.md` — D-06 (one-click Cabrillo export, no preview) → backup download follows same pattern. D-07 (single Go binary embed.FS). Three-panel layout.
- `.planning/phases/03-offline-resilience-polish/03-CONTEXT.md` — D-13–D-16 (CSS variables, dark mode, toggle in header). D-09–D-12 (mobile layout, touch targets). Offline sync patterns.

### External Specs
- ARRL Field Day 2026 official bonus points list (researcher should find and reference the official ARRL source)

### Existing Code
- `frontend/src/routes/+page.svelte` — Header bar structure (`header-left`, `header-right`). Export button pattern. Theme toggle pattern.
- `frontend/src/lib/components/StationConfig.svelte` — Expandable panel pattern for BonusTracker. Save feedback toast pattern for backup confirmation.
- `frontend/src/lib/components/StatsBar.svelte` — Score display where bonus points integrate. `stats` object shape in `qso.svelte.js`.
- `frontend/src/lib/api.js` — API client pattern. Add `getBonuses()` / `putBonuses()` / `downloadBackup()` functions.
- `frontend/src/lib/stores/qso.svelte.js` — `$state` store pattern. Add `bonusClaims` state.
- `main.go` — Chi router. Add `GET/PUT /api/bonuses`, `GET /api/backup/db` routes.
- `internal/handler/stats.go` — Score calculation. Add bonus points to `score = (raw_points + bonus_points) * multiplier`.
- `internal/cabrillo/cabrillo.go` — Cabrillo export. Add bonus points to CLAIMED-SCORE and bonus claims to output.
- `internal/db/schema.sql` — DB schema. Add `bonus_claims` table.
- `internal/handler/config.go` — Handler pattern reference for bonus API endpoints (GET/PUT with JSON).

## Existing Code Insights

### Reusable Assets
- **StationConfig.svelte** — Expandable header panel with toggle button, inline form, save feedback toast. Directly reusable as BonusTracker component template.
- **+page.svelte header-right** — Button cluster (StationConfig, Export Cabrillo). Add Bonuses button and Backup button here.
- **Cabrillo export flow** (`+page.svelte` → `window.location.href` → `handler/export.go`) — Backup download follows the identical pattern.
- **api.js** — fetch wrapper pattern. New API functions follow `getStationConfig()` / `putStationConfig()` conventions.
- **StatsBar.svelte** — Already displays `stats.score`. Bonus points integrate into this display.
- **scoring logic** (`stats.go`, `cabrillo.go`) — Both calculate `score = raw_points * multiplier`. Both need the same bonus addition.

### Established Patterns
- Svelte 5 runes: `$state`, `$derived`, `$effect` — new state uses these
- Object-based `$state` for exported reactive variables (`wsState.connected`) — use for `audioState` (muted boolean)
- `.svelte.js` extension for non-component files using `$state` — any new audio utility using reactive state must follow this
- CSS variables defined in `app.css` — all new UI inherits dark mode automatically
- Go handler pattern: take `*sql.DB` and optional `*ws.Hub`, return JSON, wrapped in closures in `main.go`
- Chi router: `r.Route("/api", ...)` with method-specific sub-routes
- `Content-Disposition: attachment` for file downloads (Cabrillo export pattern)
- localStorage persistence pattern from OperatorSelector — applies to bonus claims backup and mute preference
- `min-height: 48px; min-width: 48px` global rule in `app.css` — all new interactive elements get touch targets automatically

### Integration Points
- **Bonus in score:** `stats.go` line 45 (`score = rawPoints * multiplier`) becomes `score = (rawPoints + bonusPoints) * multiplier`
- **Bonus in Cabrillo:** `cabrillo.go` lines 67-79 (score calculation) and output formatting (add bonus claim lines)
- **StatsBar display:** Add bonus points field to `stats` `$state` in `qso.svelte.js`, display in StatsBar component
- **New routes in main.go:** `r.Get("/bonuses", ...)`, `r.Put("/bonuses", ...)`, `r.Get("/backup/db", ...)`
- **New DB table:** `bonus_claims` in `schema.sql` — columns for bonus ID, claimed boolean, count integer
- **Audio utility:** New `frontend/src/lib/audio.svelte.js` — loads and plays sound files, respects mute state
- **Audio triggers:** In `QsoEntryForm.svelte` `handleCheckDupe()` (dupe buzz), in `qso.svelte.js` `addQso()` success path or `ws.svelte.js` QSO created handler (confirmation beep, for own QSOs only)
- **Simulation test:** New `internal/simtest/` or integrated into existing handler test package — spawns simulated clients, creates QSOs, verifies integrity after run

## Specific Ideas

- Bonus tracker should feel like the expandable StationConfig panel — same open/close animation, same styling, same save feedback. Operators already know this UI pattern.
- Audio files go in `frontend/static/audio/` directory (e.g., `confirm.wav`, `dupe.wav`). Served as static assets by the Go embed.FS.
- Mute icon uses the same styling as the theme toggle icon — small, in the header, persistent.

## Deferred Ideas

None — discussion stayed within Phase 4 scope.

---

*Phase: 4-Field Day Features & Testing*
*Context gathered: 2026-06-04*
