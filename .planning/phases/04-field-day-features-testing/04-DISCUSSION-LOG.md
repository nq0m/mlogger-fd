# Phase 4: Field Day Features & Testing - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-06-04
**Phase:** 4-Field Day Features & Testing
**Areas discussed:** Testing Strategy, Bonus Tracker Design, Audio Feedback, Backup Mechanism

---

## Testing Strategy

| Option | Description | Selected |
|--------|-------------|----------|
| Scripted Go integration test | Go program using httptest spawning HTTP clients posting QSOs, verifying sync/dupe/stats. Part of CI. | ✓ |
| Manual test plan with checklist | Written procedure for a human tester. No automation. | |
| Both — automated + manual | Automated script for CI + manual field checklist. | |

**User's choice:** Scripted Go integration test
**Notes:** Simulation test lives in internal test package, run as part of `go test ./...`.

| Option | Description | Selected |
|--------|-------------|----------|
| Core data integrity | No lost QSOs, dupes correctly detected, stats accurate, sync correct. | ✓ |
| Data integrity + performance | Above plus response times, WS latency, memory growth. | |
| Full Field Day scenario | Everything — bonus claims, backup mid-test, offline/reconnect, concurrent operators. | |

**User's choice:** Core data integrity

| Option | Description | Selected |
|--------|-------------|----------|
| Internal test (go test) | Lives in internal package via httptest. In-process. | ✓ |
| Separate cmd tool | Standalone binary hitting running server. More realistic. | |
| Both — unit-level + integration | Quick checks in go test, heavy simulation as cmd tool. | |

**User's choice:** Internal test (go test)

| Option | Description | Selected |
|--------|-------------|----------|
| Minimal checklist — just do it | Set up in park, log, verify. No formal doc. | ✓ |
| Written test plan in repo | Markdown checklist covering all features and scenarios. | |

**User's choice:** Minimal checklist — just do it

---

## Bonus Tracker Design

| Option | Description | Selected |
|--------|-------------|----------|
| Current year's official FD bonuses | 2026 ARRL FD bonus list — emergency power, publicity, public location, etc. Fixed list. | ✓ |
| Fixed historical set | Common FD bonuses that rarely change year to year. | |
| User-configurable list | Start with defaults, let station admin add/edit/remove items. | |

**User's choice:** Current year's (2026) official FD bonuses

| Option | Description | Selected |
|--------|-------------|----------|
| Toggle + count per bonus | Boolean toggle for simple items, number input for counted items (youth, traffic). | ✓ |
| All count-based | Every bonus is a count, 0 = unclaimed. | |
| Checklist + manual notes | Checkboxes for everything, sub-entries for counted items. | |

**User's choice:** Toggle + count per bonus

| Option | Description | Selected |
|--------|-------------|----------|
| Header panel — like StationConfig | Expandable "Bonuses" button in header-right. Same pattern as StationConfig. | ✓ |
| Section between stats and log table | Dedicated panel between StatsBar and LogTable. | |
| Side panel or modal | Opens as modal/panel over log table. | |

**User's choice:** Header panel — like StationConfig

| Option | Description | Selected |
|--------|-------------|----------|
| Server-side — like StationConfig | SQLite bonus_claims table, REST API. One source of truth. | |
| Server-side + localStorage backup | Server is truth, localStorage caches for resilience. | ✓ |

**User's choice:** Server-side + localStorage backup

---

## Audio Feedback

| Option | Description | Selected |
|--------|-------------|----------|
| Simple beeps — different pitch | Web Audio oscillator, ~800Hz confirm / ~400Hz dupe. No files needed. | |
| Pre-recorded sound files | .wav/.mp3 files — ding for confirm, buzz for dupe. | ✓ |

**User's choice:** Pre-recorded sound files

| Option | Description | Selected |
|--------|-------------|----------|
| Generate at build time | Go script generates .wav files during build. | |
| You source the files | User provides .wav/.mp3 files. | ✓ |
| Embed minimal base64 data URIs | Encode tiny sound files as base64 strings in JS. | |

**User's choice:** User sources the files

| Option | Description | Selected |
|--------|-------------|----------|
| Mute toggle in the header bar | Speaker icon next to theme toggle. Persists in localStorage. Default unmuted. | ✓ |
| No controls — always on | Audio always plays. No UI. | |
| Mute toggle + volume slider | Mute plus volume control. | |

**User's choice:** Mute toggle in the header bar

| Option | Description | Selected |
|--------|-------------|----------|
| Own QSOs only | Confirm on create, dupe buzz on form detection. No WS sounds. | ✓ |
| Own QSOs + all QSOs via WS | Confirm on all QSOs including other operators'. | |

**User's choice:** Own QSOs only

---

## Backup Mechanism

| Option | Description | Selected |
|--------|-------------|----------|
| One-click button — like Cabrillo | Immediate download via window.location.href. | |
| One-click + brief confirmation toast | Download + brief "Backup downloaded" feedback. | ✓ |

**User's choice:** One-click + brief confirmation toast

| Option | Description | Selected |
|--------|-------------|----------|
| Station callsign + date | {callsign}_backup_{YYYYMMDD}.db | |
| Timestamped for uniqueness | fdlogger_backup_{YYYYMMDD}_{HHMMSS}.db | ✓ |

**User's choice:** Timestamped for uniqueness

| Option | Description | Selected |
|--------|-------------|----------|
| Just stream the live file | Read and stream current .db file as-is. WAL readers don't block writers. | ✓ |
| SQLite VACUUM INTO for clean copy | Clean, defragmented copy via VACUUM INTO. | |

**User's choice:** Just stream the live file

---

## the agent's Discretion

- Specific 2026 ARRL FD bonus list items and point values
- SQLite schema for `bonus_claims` table
- API request/response format for bonus claims
- Audio file format, loading approach, and playback logic
- Toast implementation for backup confirmation
- Simulation test: number of clients, QSO rate, assertions
- Bonus tracker component structure
- Mute toggle icon and exact position in header

## Deferred Ideas

None — discussion stayed within Phase 4 scope.
