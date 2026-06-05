---
phase: 4
name: "Field Day Features & Testing"
researcher: gsd-researcher
date: "2026-06-04"
status: complete
---

# Phase 4: Field Day Features & Testing — Research

**Researched:** 2026-06-04
**Domain:** ARRL Field Day contest rules, Web Audio API, SQLite WAL backup, Go integration testing
**Confidence:** HIGH

## Summary

Phase 4 delivers Field Day-specific features on top of the existing Phase 1–3 foundation: a bonus points tracker with the official 2026 ARRL Field Day bonus list, Web Audio API audio alerts for QSO confirmation and dupe warnings, one-click SQLite database backup via file streaming, a Go integration test simulating multi-client continuous logging, and a manual field test. Research confirms the 2026 bonus rules are unchanged from 2025 [VERIFIED: arrl.org/field-day-rules], making the list stable. Web Audio API can play WAV files embedded in Go's `embed.FS` via standard URL fetching — no external dependencies needed. SQLite WAL mode already configured in the project ensures backup reads don't block live QSO writes. The httptest pattern used in existing tests (e.g., `ws_test.go`) scales to multi-client simulation. All integration points with existing code (stats.go, cabrillo.go, StationConfig pattern, Chi router, $state store pattern) are fully mapped.

**Primary recommendation:** Build the bonus tracker as a StationConfig-patterned expandable panel, serve audio as static assets under `frontend/static/audio/` for embed.FS access, stream the SQLite file directly for backup (WAL mode protects read consistency), and build the simulation test by extending the existing handler test pattern with goroutine-based simulated clients.

## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01:** Use 2026 ARRL Field Day bonus points list — predefined, fixed, hardcoded.
- **D-02:** Toggle + count per bonus item. Boolean bonuses use simple on/off toggle. Counted bonuses use toggle plus number input.
- **D-03:** Bonus tracker UI lives as an expandable header panel, following StationConfig pattern. "Bonuses" button in header-right between StationConfig and Export.
- **D-04:** Bonus claims persist server-side in new SQLite `bonus_claims` table, fetched/saved via REST API (`GET /api/bonuses`, `PUT /api/bonuses`), with localStorage backup using hybrid pattern from OperatorSelector.
- **D-05:** Bonus points must be reflected in score calculation (stats endpoint) and Cabrillo export (CLAIMED-SCORE line and SOAPBOX/X-BONUS lines).
- **D-06:** Pre-recorded sound files (.wav or .mp3) — user provides the audio files. Confirmation beep and dupe buzz are distinct audio files.
- **D-07:** Mute toggle in header bar (speaker icon near theme toggle), persisted in localStorage, default unmuted.
- **D-08:** Sounds fire only for own QSOs — confirmation on successful create (local submit or sync success), dupe buzz on form dupe detection. No sounds for other operators' QSOs from WebSocket.
- **D-09:** One-click backup button in header-right (next to Cabrillo Export), with brief confirmation toast following StationConfig save-feedback pattern.
- **D-10:** Timestamped filename: `fdlogger_backup_{YYYYMMDD}_{HHMMSS}.db`.
- **D-11:** Stream the live SQLite file as-is. WAL mode ensures readers don't block writers — safe during active logging. No VACUUM INTO needed.
- **D-12:** Scripted Go integration test for 2-hour simulation — lives in internal test package, run as part of `go test ./...`. Uses `httptest` for in-process server testing.
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

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within Phase 4 scope.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Bonus claim state/UI | Browser (Svelte) | API (Go) | Toggle UI in browser; persistence via REST API; localStorage backup for resilience |
| Bonus scoring calculation | API (Go) | — | Score and Cabrillo are computed server-side in stats.go / cabrillo.go |
| Audio playback | Browser (Web Audio API) | — | Client-side only; audio files served as static assets from server |
| Mute preference | Browser (localStorage) | — | Purely client-side UI state; no server involvement |
| Database backup download | API (Go) | — | Server reads SQLite file, streams to client via HTTP download |
| Simulation test | API (Go test) | — | In-process Go test using httptest and real SQLite |
| Field test | Manual / physical | — | Human-operated outdoor test; no automation |

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| UX-03 | Audio alerts via Web Audio API for new QSO confirmation and dupe warning | Web Audio API section below; audio trigger points mapped in QsoEntryForm.svelte |
| BON-01 (promoted) | Bonus points tracker with list of FD bonus opportunities and claim toggles | 2026 ARRL bonus list below; StationConfig panel pattern documented |
| BON-02 (promoted) | Bonus point summary included in score calculation | Scoring integration section; stats.go line 45 and cabrillo.go integration points |
| BKUP-01 (promoted) | One-click database backup (copy SQLite file) | WAL streaming backup pattern; Cabrillo export download pattern reuse |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Web Audio API | Built-in (browser) | Audio playback for QSO alerts | Native browser API, no dependencies, works offline |
| Go `net/http/httptest` | stdlib | Integration test server | Already used in existing handler tests; in-process, no ports |
| SQLite WAL mode | Already configured | Safe concurrent read during backup | Already in db.go (`pragma=journal_mode(WAL)`) |
| Go `io.Copy` | stdlib | Stream SQLite file for backup | Simple, no external deps; WAL guarantees read consistency |
| Chi router (existing) | v5.3.0 | New API routes for bonuses, backup | Already in go.mod; consistent with existing route patterns |
| SvelteKit static adapter (existing) | ^3.0.10 | Serve audio files as static assets | Already in svelte.config.js; builds to `frontend/build/` embedded via `embed.FS` |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `localStorage` | Built-in | Mute preference persistence, bonus claims backup | Client-side only; already used for theme and operator |
| `crypto.randomUUID()` | Built-in | Client ID generation for offline QSOs | Already used in qso.svelte.js; no import needed |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Web Audio API (built-in) | HTML5 `<audio>` element | `<audio>` has limited programmatic control, no fine-grained timing, no polyphonic playback |
| Go `io.Copy` for backup | `VACUUM INTO` | VACUUM INTO requires exclusive lock which blocks writers — D-11 explicitly excludes this |
| Manual simulation test | Gomega/Ginkgo test framework | Adds dependency; stdlib testing + testify assertions sufficient for this scope |

**Installation:** No new packages required. All functionality uses existing dependencies or browser built-in APIs.

## Package Legitimacy Audit

No new external packages are introduced in this phase. All functionality relies on:
- Existing Go stdlib (`net/http`, `io`, `database/sql`, `net/http/httptest`)
- Existing Go dependencies (chi/v5, gorilla/websocket, modernc.org/sqlite)
- Existing frontend dependencies (SvelteKit, Svelte 5, Dexie)
- Browser built-in APIs (Web Audio API, localStorage, crypto)

**Packages removed:** none
**Packages flagged:** none

## Architecture Patterns

### System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                        FIELD DAY LOGGER                             │
│                                                                     │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────────────────┐│
│  │ Operator 1   │   │ Operator 2   │   │ Operator N (n clients)   ││
│  │ (Phone/Tablet)│   │ (Phone/Tablet)│   │ (Phone/Tablet)           ││
│  └──────┬───────┘   └──────┬───────┘   └───────────┬──────────────┘│
│         │                  │                       │                │
│  ┌──────┴──────────────────┴───────────────────────┴──────────────┐│
│  │                    SPA (Browser Tier)                           ││
│  │                                                                 ││
│  │  ┌──────────────────────────────────────────────────────────┐  ││
│  │  │  Header Bar                                              │  ││
│  │  │  [☀] [Live] [Operator] [⚙Config] [★Bonuses] [Export] [↓Backup]││
│  │  │  [🔊/🔇 mute toggle]                                     │  ││
│  │  └──────────────────────────────────────────────────────────┘  ││
│  │  ┌──────────────┐  ┌──────────────────┐  ┌──────────────────┐  ││
│  │  │ QsoEntryForm  │  │  StatsBar        │  │  LogTable         │  ││
│  │  │ (audio: dupe) │  │ (bonus in score) │  │                   │  ││
│  │  └──────┬───────┘  └──────────────────┘  └──────────────────┘  ││
│  │         │                                                       ││
│  │  ┌──────┴──────────────────────────────────────────────────┐   ││
│  │  │  Stores + Audio                                         │   ││
│  │  │  qso.svelte.js ($state: qsos, stats, bonusClaims)       │   ││
│  │  │  audio.svelte.js ($state: muted → audioContext mgmt)    │   ││
│  │  │  ws.svelte.js ($state: connected)                       │   ││
│  │  │  sync.svelte.js ($state: queue)                         │   ││
│  │  │  db.js (Dexie IndexedDB cache)                          │   ││
│  │  │  api.js (fetch wrappers: bonuses, backup)               │   ││
│  │  └──────────────────────────────────────────────────────────┘   ││
│  └─────────────────────────────────────────────────────────────────┘│
│         │ REST + WebSocket                                          │
│  ┌──────┴──────────────────────────────────────────────────────────┐│
│  │                    Go Server (API Tier)                          ││
│  │                     chi/v5 Router                                ││
│  │                                                                 ││
│  │  /api/bonuses    GET  → handler.GetBonuses(db)                  ││
│  │                   PUT  → handler.PutBonuses(db)                  ││
│  │  /api/backup/db  GET  → handler.DownloadBackup(dbPath)          ││
│  │  /api/stats       GET  → handler.GetStats(db)  [bonus added]    ││
│  │  /api/export/cabrillo GET → handler.ExportCabrillo(db) [bonus+  ││
│  │  /api/qso         POST → handler.CreateQSO(db,hub)              ││
│  │  /ws              GET  → handler.ServeWS(hub)                   ││
│  │                                                                 ││
│  │  ┌─────────────────┐  ┌──────────────────┐  ┌──────────┐      ││
│  │  │ handler/bonus.go│  │ handler/backup.go │  │ stats.go │      ││
│  │  │ (new)           │  │ (new)             │  │ (modified)│      ││
│  │  └────────┬────────┘  └────────┬─────────┘  └──────────┘      ││
│  │           │                    │                                ││
│  └───────────┼────────────────────┼────────────────────────────────┘│
│              │                    │                                 │
│  ┌───────────┴────────────────────┴────────────────────────────────┐│
│  │                    Storage Tier                                  ││
│  │                                                                 ││
│  │  ┌─────────────────────────────────────────────────────────┐   ││
│  │  │  SQLite (WAL mode)                                       │   ││
│  │  │  fdlogger.db                                             │   ││
│  │  │  ├── qsos                     (existing)                 │   ││
│  │  │  ├── station_config           (existing)                 │   ││
│  │  │  └── bonus_claims             (NEW)                     │   ││
│  │  └─────────────────────────────────────────────────────────┘   ││
│  └─────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────┘
```

### Recommended Project Structure (new files only)

```
frontend/
├── static/
│   └── audio/                     # NEW: user-provided audio files
│       ├── confirm.wav
│       └── dupe.wav
└── src/
    └── lib/
        ├── audio.svelte.js        # NEW: Web Audio API utility
        ├── components/
        │   └── BonusTracker.svelte # NEW: expandable bonus panel
        ├── api.js                 # MODIFIED: add get/putBonuses, downloadBackup
        └── stores/
            └── qso.svelte.js      # MODIFIED: add bonusClaims state, bonus_points to stats

internal/
├── handler/
│   ├── bonus.go                   # NEW: GET/PUT /api/bonuses handler
│   ├── backup.go                  # NEW: GET /api/backup/db handler
│   ├── stats.go                   # MODIFIED: score includes bonus
│   └── simtest/                   # NEW: simulation test
│       └── simtest_test.go        # Multi-client integration test
├── cabrillo/
│   └── cabrillo.go                # MODIFIED: CLAIMED-SCORE + bonus lines
├── db/
│   └── schema.sql                 # MODIFIED: add bonus_claims table
└── model/
    └── bonus.go                   # NEW: BonusClaim struct, default list

main.go                             # MODIFIED: add /api/bonuses, /api/backup/db routes
```

### Pattern 1: Expandable Header Panel (for BonusTracker)

**What:** A toggle button in the header bar that shows/hides an absolutely-positioned panel with form controls. The StationConfig component is the canonical example.

**When to use:** Any feature that lives as an inline dropdown panel from the header bar. BonusTracker follows this exactly.

**Example (from StationConfig.svelte, lines 54–117):**
```svelte
<script>
	let expanded = $state(false);
	let saved = $state(false);
	let saveTimer;

	function toggle() { expanded = !expanded; }

	async function handleSubmit(e) {
		e.preventDefault();
		// ... API call ...
		saved = true;
		if (saveTimer) clearTimeout(saveTimer);
		saveTimer = setTimeout(() => { saved = false; }, 2000);
	}
</script>

<div class="station-config">
	<button class="config-toggle" onclick={toggle} aria-label="Config">
		<span class="toggle-icon">⚙</span> Config
	</button>
	{#if expanded}
		<div class="config-panel">
			<form onsubmit={handleSubmit}>
				<!-- form fields -->
				<div class="config-actions">
					<button type="submit" class="save-btn">Save</button>
					{#if saved}
						<span class="saved-msg">Saved!</span>
					{/if}
				</div>
			</form>
		</div>
	{/if}
</div>
```

**Key patterns to replicate:**
- Panel uses `position: absolute; top: 44px; right: 8px; z-index: 100`
- Save feedback toast: 2-second `$state` boolean + `setTimeout` clear
- Toggle button uses transparent background with 2px solid var(--color-surface) border
- `min-height: 48px` inherited from app.css global rule

### Pattern 2: Object-Based $state for Exported Reactivity

**What:** Svelte 5 forbids reassigning exported `$state` primitives. Workaround: wrap in an object and mutate properties.

**When to use:** Any module-level reactive state that is imported by components (audio mute, bonus claims, ws state).

**Examples from codebase:**
```javascript
// ws.svelte.js (line 7)
export const wsState = $state({ connected: false });

// sync.svelte.js (line 5)
export const queueState = $state({ queueLength: 0, syncing: false });

// Usage in components: wsState.connected = true  (property mutation, OK)
// NOT: wsState = { connected: true }             (reassignment, FORBIDDEN)
```

**New states to create following this pattern:**
```javascript
// audio.svelte.js
export const audioState = $state({ muted: false });

// qso.svelte.js (addition)
export const bonusClaims = $state({});  // { bonus_id: { claimed: bool, count: int } }
```

### Pattern 3: Chi Router Handler Registration

**What:** Closures in `main.go` wrapping handler functions that take `*sql.DB` (and optionally `*ws.Hub`).

**When to use:** Every new API endpoint.

**Example (from main.go lines 43–74):**
```go
r.Route("/api", func(r chi.Router) {
    // GET with query params
    r.Get("/check-dupe", func(w http.ResponseWriter, r *http.Request) {
        handler.CheckDupeHandler(database, w, r)
    })
    // GET simple
    r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
        handler.GetStats(database, w, r)
    })
    // PUT with body
    r.Put("/station-config", func(w http.ResponseWriter, r *http.Request) {
        handler.PutStationConfig(database, w, r)
    })
})
```

**New routes to add:**
```go
r.Get("/bonuses", func(w http.ResponseWriter, r *http.Request) {
    handler.GetBonuses(database, w, r)
})
r.Put("/bonuses", func(w http.ResponseWriter, r *http.Request) {
    handler.PutBonuses(database, w, r)
})
r.Get("/backup/db", func(w http.ResponseWriter, r *http.Request) {
    handler.DownloadBackup(database, dbPath, w, r)  // needs dbPath for file access
})
```

### Pattern 4: File Download Response (for Backup)

**What:** Set `Content-Disposition: attachment` header and write content to response writer — identical to Cabrillo export pattern.

**Example (from handler/export.go lines 12–28):**
```go
func ExportCabrillo(db *sql.DB, w http.ResponseWriter, r *http.Request) {
    result, err := cabrillo.Generate(db)
    // ... error handling ...
    w.Header().Set("Content-Type", "text/plain")
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s_field_day.cbr\"", callsign))
    w.Write([]byte(result))
}
```

**Backup download adaptation:**
```go
func DownloadBackup(db *sql.DB, dbPath string, w http.ResponseWriter, r *http.Request) {
    now := time.Now().UTC().Format("20060102_150405")
    filename := fmt.Sprintf("fdlogger_backup_%s.db", now)
    w.Header().Set("Content-Type", "application/octet-stream")
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
    f, _ := os.Open(dbPath)
    defer f.Close()
    io.Copy(w, f)
}
```

### Pattern 5: localStorage Persistence (OperatorSelector Pattern)

**What:** Read from localStorage on component init, save on every keystroke/change.

**Example (from OperatorSelector.svelte lines 4–8):**
```svelte
let operator = $state(localStorage.getItem('fdlogger_operator') || '');

function saveOperator() {
    localStorage.setItem('fdlogger_operator', operator);
}
```

**Apply to:**
- Mute preference: `localStorage.getItem('fdlogger_muted')` → default `false`
- Bonus claims backup: `localStorage.getItem('fdlogger_bonus_claims')` → parse JSON or `{}`

### Anti-Patterns to Avoid

- **Direct $state export without object wrapper:** Svelte 5 will break on reassignment. Always wrap in object: `export const fooState = $state({ value: x })`. See STATE.md records for ws.connected and syncState decisions.
- **AudioContext creation in component onMount without user gesture:** Browsers block AudioContext creation until user interaction. Create context lazily on first user gesture (form submit, button click) or resume suspended context.
- **Long-running audio file fetches blocking UI:** Use `fetch()` + `decodeAudioData()` asynchronously; cache decoded AudioBuffers for instant replay.
- **Blocking the WAL checkpoint for backup:** Do NOT use `VACUUM INTO` or exclusive lock. Simple `io.Copy` from the -wal and -shm companion files alongside the main .db file is sufficient, but `cp` of just the .db file is safe because WAL mode guarantees the .db file is always consistent (uncommitted data is in -wal file).
- **Hardcoding station config in scoring:** Bonus scoring must read from `bonus_claims` table, not assume defaults. Missing config → 0 bonus points (safe fallback).
- **Over-engineering simulation test:** Don't need real HTTP ports or network latency simulation — httptest in-process with goroutine-based simulated clients is sufficient for data integrity validation.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Audio playback in browser | Custom audio library | Web Audio API (built-in) | Native, zero deps, works offline, supports scheduling, polyphony, volume control |
| SQLite file download | Custom file streaming protocol | Go `io.Copy` + `Content-Disposition: attachment` | Standard, battle-tested, handles large files efficiently via OS-level sendfile |
| Multi-client test simulation | Real network-based test harness | Go `httptest` + goroutines | In-process, deterministic, fast; existing ws_test.go pattern already validates WebSocket broadcast |
| Bonus list management | Dynamic bonus admin UI | Hardcoded Go constant array | D-01 locked decision; FD rules change annually but list is fixed per year |
| Mute state management | Custom event system | Svelte 5 `$state` object + localStorage | Consistent with existing wsState/queueState patterns |
| Backup confirmation toast | Custom notification widget | 2-second `$state` boolean with `setTimeout` | Already pattern in StationConfig `saved` feedback |

**Key insight:** Everything needed for this phase already exists in the browser, Go stdlib, or established project patterns. No new npm packages, no new Go dependencies.

## Runtime State Inventory

**Not applicable — this is a greenfield feature phase, not a rename/refactor/migration.** No existing runtime state needs migration.

However, note that the `fdlogger.db` SQLite file schema will be extended (new `bonus_claims` table). The existing `db.Open()` function runs `schema.sql` on every startup, and the new `CREATE TABLE IF NOT EXISTS` statement ensures backward compatibility — existing databases won't break.

## 2026 ARRL Field Day Bonus Points List

[VERIFIED: arrl.org/field-day-rules, section 7.3, revised March 1, 2026]

> Note: ARRL states "No changes from 2025 rules" for 2026. The bonus list below is from section 7.3 of the official 2026 Field Day Rules.

| # | Bonus ID | Name | Rule | Points | Type | Available To | Description |
|---|----------|------|------|--------|------|-------------|-------------|
| 1 | `emergency_power` | 100% Emergency Power | 7.3.1 | 100/transmitter | Counted (max 2000) | A, B, C, E, F | 100 pts per transmitter classification on emergency power |
| 2 | `media_publicity` | Media Publicity | 7.3.2 | 100 | Boolean | All classes | Obtaining publicity from local media |
| 3 | `public_location` | Public Location | 7.3.3 | 100 | Boolean | A, B, F | Physically locating operation in a public place |
| 4 | `public_info_table` | Public Information Table | 7.3.4 | 100 | Boolean | A, B, F | Info table with handouts at Field Day site |
| 5 | `message_to_sm` | Message to Section Manager | 7.3.5 | 100 | Boolean | All classes | Formal NTS/ICS-213 message to SM or SEC via RF |
| 6 | `message_handling` | Message Handling | 7.3.6 | 10/msg | Counted (max 100) | All classes | Formal messages originated/relayed/received via RF |
| 7 | `satellite_qso` | Satellite QSO | 7.3.7 | 100 | Boolean | A, B, F | At least one QSO via amateur satellite |
| 8 | `alternate_power` | Alternate Power | 7.3.8 | 100 | Boolean | A, B, E, F | 5+ QSOs using solar/wind/water/battery (no mains/gen) |
| 9 | `w1aw_bulletin` | W1AW Bulletin | 7.3.9 | 100 | Boolean | All classes | Copy special W1AW (or K6KPH) Field Day bulletin via RF |
| 10 | `educational_activity` | Educational Activity | 7.3.10 | 100 | Boolean | A, F (+ D/E with 3+) | Formal educational activity related to amateur radio |
| 11 | `official_visit` | Elected Official Visit | 7.3.11 | 100 | Boolean | All classes | Site visit by elected government official |
| 12 | `agency_visit` | Agency Representative Visit | 7.3.12 | 100 | Boolean | All classes | Site visit by served agency rep (Red Cross, EMA, etc) |
| 13 | `gota_bonus` | GOTA Station Bonus | 7.3.13 | 5/QSO + 100 coach | Counted + Boolean | A, F | 5 pts per GOTA QSO; 100 pts if GOTA Coach present (10+ contacts) |
| 14 | `web_submission` | Web Submission | 7.3.14 | 50 | Boolean | All classes | Submit entry via web app |
| 15 | `youth_participation` | Youth Participation | 7.3.15 | 20/participant | Counted (max 100) | A, C, D, E, F (+ limited B) | 20 pts per participant ≤18 who makes at least 1 QSO |
| 16 | `social_media` | Social Media Promotion | 7.3.16 | 100 | Boolean | All classes | Promote FD on recognized social media platform |
| 17 | `safety_officer` | Safety Officer | 7.3.17 | 100 | Boolean | A only | Designated safety officer verifies checklist |
| 18 | `site_responsibilities` | Site Responsibilities | 7.3.18 | 50 | Boolean | B, C, D, E, F | Person ensures hazard-free site throughout event |

**Total possible bonus points:** 1,650 points (base) + up to 2,000 (emergency power, 20 transmitters) + up to 100 (message handling) + up to 100 (youth) + unlimited GOTA QSO bonuses. Typical maximum for a well-equipped club: ~3,000–4,000 bonus points.

**Hardcoding strategy (D-01):** Define as a Go `const` slice in `internal/model/bonus.go`. Each item has: `ID`, `Name`, `RuleRef`, `Points`, `IsCounted` (boolean toggle vs counted), and `DefaultCount` (0 for booleans, 0 for counted items).

### Bonus Claims SQLite Schema

```sql
CREATE TABLE IF NOT EXISTS bonus_claims (
    bonus_id TEXT PRIMARY KEY,
    claimed INTEGER NOT NULL DEFAULT 0,   -- Boolean: 0/1
    count INTEGER NOT NULL DEFAULT 0,     -- Number input for counted bonuses
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
```

**Rationale:** Row-per-bonus design allows adding/removing bonus types in future years without schema migration. The `count` field is used for counted bonuses (emergency_power transmitter count, message_handling count, youth_participation count, GOTA QSO count) and is 0 for boolean-only bonuses.

### API Request/Response Format

**GET /api/bonuses → 200:**
```json
{
  "emergency_power": { "claimed": true, "count": 3 },
  "media_publicity": { "claimed": true, "count": 0 },
  "social_media": { "claimed": false, "count": 0 },
  "...": { "...": "..." }
}
```

**PUT /api/bonuses (request body → response):** Same shape. Server merges incoming claims and returns the full state.

### Score Calculation Integration (D-05)

**Current:** `stats.go` line 45: `score := rawPoints * multiplier`

**Modified:**
```go
// Read total bonus points from bonus_claims
var bonusPoints int
db.QueryRow(`SELECT 
    (SELECT COALESCE(SUM(CASE 
        WHEN bonus_id = 'emergency_power' THEN claimed * count * 100
        WHEN bonus_id = 'message_handling' THEN claimed * MIN(count, 10) * 10
        WHEN bonus_id = 'youth_participation' THEN claimed * MIN(count, 5) * 20
        WHEN bonus_id = 'gota_bonus' THEN claimed * count * 5
        ELSE claimed * 100  -- all other fixed-100 bonuses
    END), 0) FROM bonus_claims)`).Scan(&bonusPoints)
// Also add GOTA coach bonus (separate boolean in gota row)
// Add 50 for web_submission, safety_officer, site_responsibilities as fixed amounts

score := (rawPoints + bonusPoints) * multiplier
```

**Note:** Per ARRL rules, bonus points are added AFTER the multiplier is applied. The current codebase already multiplies raw points. The correct formula is: `score = (rawPoints * multiplier) + bonusPoints` — bonus is NOT multiplied. The CONTEXT.md says `score = (raw_points + bonus_points) * multiplier` (line 97), but **this contradicts ARRL rules section 7.3** which states "bonus points will be added to the score, after the multiplier is applied." Flagging for user confirmation during discuss-phase or planning.

**Correct formula per ARRL rules:** `final_score = (raw_points * multiplier) + bonus_points`

### Cabrillo Export Integration (D-05)

**CLAIMED-SCORE line:** Replace current `score := rawPoints * multiplier` with `score := (rawPoints * multiplier) + bonusPoints` in `cabrillo.go`.

**SOAPBOX lines:** Add bonus claims as SOAPBOX lines:
```
SOAPBOX: Bonus: 100% Emergency Power (3 transmitters) = 300 pts
SOAPBOX: Bonus: Media Publicity = 100 pts
SOAPBOX: Total Bonus Points = 500
```

**X-BONUS lines (optional, per N3FJP convention):**
```
X-BONUS: emergency_power=300
X-BONUS: media_publicity=100
```

## Web Audio API Pattern

### Audio File Loading from embed.FS

Audio files placed in `frontend/static/audio/` are built to `frontend/build/audio/` by SvelteKit's static adapter and served by Go's embed.FS via the SPA handler (`main.go` line 81: `r.Get("/*", spaHandler())`). Files are accessible at relative paths like `/audio/confirm.wav`.

### Audio Utility Module (`frontend/src/lib/audio.svelte.js`)

```javascript
// Object-based $state for mute (Svelte 5 pattern)
export const audioState = $state({ muted: false });

// Lazy-initialized AudioContext (browsers require user gesture)
let audioCtx = null;
const buffers = {};  // Cache decoded AudioBuffers

function ensureContext() {
    if (!audioCtx) {
        audioCtx = new AudioContext();
        // Restore mute from localStorage
        const stored = localStorage.getItem('fdlogger_muted');
        if (stored !== null) audioState.muted = stored === 'true';
    }
    // Resume if suspended (autoplay policy)
    if (audioCtx.state === 'suspended') {
        audioCtx.resume();
    }
}

async function loadSound(name) {
    if (buffers[name]) return buffers[name];
    const url = `/audio/${name}.wav`;
    const response = await fetch(url);
    const arrayBuffer = await response.arrayBuffer();
    buffers[name] = await audioCtx.decodeAudioData(arrayBuffer);
    return buffers[name];
}

export async function playSound(name) {
    if (audioState.muted) return;
    ensureContext();
    try {
        const buffer = await loadSound(name);
        const source = audioCtx.createBufferSource();
        source.buffer = buffer;
        source.connect(audioCtx.destination);
        source.start(0);
    } catch (e) {
        console.warn('Audio playback failed:', e);
    }
}

export function toggleMute() {
    audioState.muted = !audioState.muted;
    localStorage.setItem('fdlogger_muted', audioState.muted.toString());
}

// Initialize mute from localStorage on module load
if (typeof localStorage !== 'undefined') {
    const stored = localStorage.getItem('fdlogger_muted');
    if (stored !== null) audioState.muted = stored === 'true';
}
```

### Audio Trigger Points (D-08)

1. **Confirmation beep:** In `QsoEntryForm.svelte` `handleSubmit()`, after successful `createQSO()` and before clearing form (line 86): `playSound('confirm')`. Also in WebSocket message handler in `ws.svelte.js` when the message is from the local operator's own client (requires client_id tracking).

2. **Dupe buzz:** In `QsoEntryForm.svelte` `handleCheckDupe()`, when dupe is detected (lines 36–37, 45–46): `playSound('dupe')`.

**Important (D-08):** The WebSocket handler in `ws.svelte.js` receives all QSOs (including other operators'). Do NOT play sounds for remote QSOs. Only play confirmation when the QSO was submitted locally (track with `client_id` or local Submit response).

### Mute Toggle

- **Position:** In `+page.svelte` header-left, between theme toggle and "FD Logger" title
- **Icons:** 🔊 (unmuted) / 🔇 (muted) — Unicode speaker characters
- **Styling:** Same CSS class as `.theme-toggle` (small, transparent, same border)
- **Persistence:** localStorage key `fdlogger_muted`
- **Default:** Unmuted (D-07)

```svelte
<!-- In +page.svelte, header-left -->
<button class="theme-toggle mute-toggle" onclick={toggleMute} aria-label="Toggle audio">
    {audioState.muted ? '🔇' : '🔊'}
</button>
```

### Browser Compatibility Notes

- Web Audio API: supported in all modern browsers (Chrome 35+, Firefox 25+, Safari 6+, Edge 79+) [ASSUMED: based on MDN compatibility tables]
- `AudioContext.resume()`: required for browsers with autoplay policy (Chrome 66+)
- `.wav` format: universally supported (PCM encoded), no codec issues
- Audio files should be short (< 0.5 sec) to avoid latency perception
- First sound playback after page load may have slight delay (~50ms) while AudioContext initializes — subsequent plays are instant

## SQLite WAL Backup Pattern (D-11)

### Why WAL Mode Enables Safe Concurrent Reads

The project already uses WAL mode (`db.go` line 12: `_pragma=journal_mode(WAL)`). In WAL mode:
- Writers write to the `-wal` file (Write-Ahead Log)
- Readers read from the main `.db` file plus any committed `-wal` pages
- The main `.db` file is always in a consistent state (all committed transactions)
- A simple `io.Copy` of the `.db` file produces a consistent snapshot

**Source:** [CITED: sqlite.org/wal.html] — "WAL provides more concurrency as readers do not block writers and a writer does not block readers."

### Backup Implementation

```go
// handler/backup.go (NEW)
package handler

import (
    "database/sql"
    "fmt"
    "io"
    "net/http"
    "os"
    "time"
)

func DownloadBackup(db *sql.DB, dbPath string, w http.ResponseWriter, r *http.Request) {
    // Force WAL checkpoint to flush committed transactions to main DB
    // This is optional but ensures the backup includes all committed data
    db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")

    now := time.Now().UTC().Format("20060102_150405")
    filename := fmt.Sprintf("fdlogger_backup_%s.db", now)

    f, err := os.Open(dbPath)
    if err != nil {
        http.Error(w, "Failed to open database", http.StatusInternalServerError)
        return
    }
    defer f.Close()

    w.Header().Set("Content-Type", "application/octet-stream")
    w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

    if _, err := io.Copy(w, f); err != nil {
        // Client may have disconnected — not a server error
        return
    }
}
```

**Key points:**
- `PRAGMA wal_checkpoint(TRUNCATE)` flushes WAL to main DB before copying — ensures backup has all committed data (optional; safe to skip)
- `io.Copy` streams efficiently (OS-level sendfile on Linux)
- No exclusive locks needed — readers don't block writers in WAL mode
- Timestamped filename per D-10: `fdlogger_backup_20260627_143052.db`

### Backup Button Frontend

```javascript
// api.js (addition)
export function downloadBackup() {
    window.location.href = '/api/backup/db';
}
```

```svelte
<!-- +page.svelte header-right (addition, between Export and right edge) -->
<button class="export-btn" onclick={downloadBackup}>↓ Backup</button>
```

**Toast pattern (StationConfig save-feedback):** Use a 2-second `$state` boolean with `setTimeout`. Show "Backup downloaded" text near the backup button on `downloadBackup()` trigger (detect via `window.location.href` assignment — but since the page doesn't reload for downloads, use a brief toast before the download starts, or just show it permanently for 2 seconds).

## Go Integration Test Pattern (for Simulation)

### Existing Test Infrastructure

All existing Go tests use the same pattern (e.g., `qso_test.go`, `ws_test.go`, `stats_test.go`):
- In-memory SQLite: `sql.Open("sqlite", ":memory:?_pragma=journal_mode(WAL)&cache=shared")`
- `db.SetMaxOpenConns(1)` for concurrent test safety
- `httptest.NewRequest` / `httptest.NewRecorder` for unit tests
- `httptest.NewServer` for WebSocket integration tests
- `ws.NewHub()` + `go hub.Run()` for WebSocket testing
- Real `gorilla/websocket` client dials to `httptest.Server`

### Simulation Test Design (D-12, D-13)

**Package:** `internal/handler/simtest/simtest_test.go` (following Go convention of `_test` suffix in test-only packages)

**Approach:** Scale up the existing `ws_test.go` pattern with goroutine-based simulated clients.

```go
// Structure:
// 1. Start httptest.Server with chi router (all endpoints)
// 2. Start ws.Hub
// 3. Spawn N simulated clients (goroutines)
//    - Each client connects WebSocket, logs QSOs, verifies broadcasts
// 4. Run for target duration or QSO count
// 5. Verify integrity:
//    - Total QSO count matches on server and all clients
//    - No QSOs lost (every submitted QSO appears in GET /api/qso)
//    - Dupe detection correct (known dupes marked correctly)
//    - Stats endpoint returns accurate counts
//    - WebSocket broadcasts received for all QSOs
//    - Offline queue → sync → server consistency
```

**Parameters (agent's discretion):**
- **Simulated clients:** 3 (matching success criteria "3+ clients")
- **QSOs per client:** ~70 (totaling ~210, meeting "200+ QSOs")
- **Test duration:** Target < 60 seconds wall-clock (not real 2 hours — test speed is CPU-bound, not time-bound)
- **Integrity assertions:**
  1. `len(serverQsos) == totalSubmitted`
  2. Every submitted QSO appears in server list
  3. Known duplicate QSOs have `is_dupe=true` and `points=0`
  4. `stats.total == nonDupeCount`
  5. `stats.raw_points` matches expected point calculation
  6. All WebSocket clients received QSO count matches total
- **Dupe simulation:** Mix of unique QSOs and deliberate duplicates
- **Edge cases:** Rapid-fire submissions (tight loops), concurrent submissions from all clients simultaneously

### Test Fixture Pattern

The simulation test should create its own chi router instance with all routes wired, replicating the `main.go` route setup pattern in test code. This ensures the test exercises actual handler code path.

```go
func setupSimRouter(db *sql.DB, hub *ws.Hub) chi.Router {
    r := chi.NewRouter()
    r.Route("/api", func(r chi.Router) {
        r.Get("/check-dupe", func(w http.ResponseWriter, req *http.Request) {
            handler.CheckDupeHandler(db, w, req)
        })
        r.Get("/stats", func(w http.ResponseWriter, req *http.Request) {
            handler.GetStats(db, w, req)
        })
        // ... all routes wired ...
    })
    r.Get("/ws", func(w http.ResponseWriter, req *http.Request) {
        handler.ServeWS(hub, w, req)
    })
    return r
}
```

## Common Pitfalls

### Pitfall 1: Web Audio API Autoplay Policy Blocking
**What goes wrong:** `AudioContext` creation or playback fails silently because the browser's autoplay policy blocks audio before user interaction.
**Why it happens:** Chrome (since v66) and other browsers require a user gesture before creating an `AudioContext` or playing audio.
**How to avoid:** Create `AudioContext` lazily on first user action (form submit or button click). Call `audioCtx.resume()` if state is 'suspended'. The first QSO submission naturally constitutes a user gesture.
**Warning signs:** Console shows "The AudioContext was not allowed to start" or sounds don't play on first QSO.

### Pitfall 2: Bonus Points Multiplied in Score
**What goes wrong:** Score calculation multiplies bonus points by the power multiplier, inflating the claimed score and making it inconsistent with ARRL rules.
**Why it happens:** The CONTEXT.md phrase "score = (raw_points + bonus_points) * multiplier" is ambiguous. ARRL rules section 7.3 explicitly states bonus points are added AFTER the multiplier.
**How to avoid:** Use `score = (rawPoints * multiplier) + bonusPoints`. Verify against known-good calculations.
**Warning signs:** CLAIMED-SCORE in Cabrillo output is significantly higher than expected.

### Pitfall 3: SQLite File Lock During Backup
**What goes wrong:** Backup download fails or blocks QSO logging if the SQLite file is opened with an exclusive lock.
**Why it happens:** Some backup approaches (VACUUM INTO, EXCLUSIVE lock) block concurrent writes. Even simple `os.Open` can fail if SQLite has a reserved lock.
**How to avoid:** WAL mode already configured. Use `os.Open` for the `.db` file only (not the `-wal` or `-shm` companion files) — WAL mode guarantees the `.db` file is always consistent. Optionally run `PRAGMA wal_checkpoint(TRUNCATE)` before copying to flush committed data.
**Warning signs:** Backup download hangs during active logging, or QSO logging fails during backup download.

### Pitfall 4: Svelte 5 $state Export Reassignment
**What goes wrong:** Importing a `$state` variable from a `.svelte.js` module and reassigning it breaks reactivity silently.
**Why it happens:** Svelte 5 forbids reassigning exported `$state` primitives. The compiler may not warn in all cases.
**How to avoid:** Always use object-based `$state` for module-level exports: `export const audioState = $state({ muted: false })`. Mutate properties only: `audioState.muted = true`. This is the established pattern (see `wsState.connected`, `queueState`).
**Warning signs:** Mute toggle doesn't update UI, audio plays when it shouldn't.

### Pitfall 5: Own-QSO Detection Failure
**What goes wrong:** Audio plays for other operators' QSOs received via WebSocket, violating D-08.
**Why it happens:** The WebSocket message handler in `ws.svelte.js` processes all `qso_created` messages. Without `client_id` tracking, it can't distinguish "my QSO" from "someone else's QSO".
**How to avoid:** Mark locally submitted QSOs with a flag before server response arrives. In `QsoEntryForm.svelte` `handleSubmit()`, play confirmation beep immediately on successful `createQSO()` response (where we know it's ours). For the WebSocket path, track recently submitted QSOs via `client_id` Set and skip audio if the QSO wasn't locally submitted.
**Warning signs:** Audio beeps play every time any operator anywhere logs a QSO.

### Pitfall 6: Hardcoded Bonus List Goes Stale
**What goes wrong:** Next year's Field Day uses different bonus rules, making the hardcoded list inaccurate.
**Why it happens:** D-01 locks the list as a hardcoded constant. No update mechanism exists.
**How to avoid:** Accept this as by-design (D-01). Document the file location (`internal/model/bonus.go`) in PROJECT.md for easy annual update. The list uses `bonus_id` strings that won't change year-over-year (ARRL bonus categories are stable).
**Warning signs:** Out-of-sync bonus point values after rule changes (but this is expected — user must update manually pre-FD).

## Scoring Integration Details

### Current Stats Endpoint (stats.go Response Shape)

```json
{
    "total": 42,
    "raw_points": 84,
    "multiplier": 7,
    "score": 588,
    "rate_10min": 120,
    "rate_1hr": 42,
    "breakdown": { "20M_CW": 20, "40M_SSB": 15, "40M_CW": 7 }
}
```

### Modified Response (add bonus_points field)

```json
{
    "total": 42,
    "raw_points": 84,
    "bonus_points": 750,    // NEW
    "multiplier": 7,
    "score": 1338,          // (84 * 7) + 750 = 1338
    "rate_10min": 120,
    "rate_1hr": 42,
    "breakdown": { "20M_CW": 20, "40M_SSB": 15, "40M_CW": 7 }
}
```

### Current Cabrillo Score (cabrillo.go line 78)

```go
score := rawPoints * multiplier
buf.WriteString(fmt.Sprintf("CLAIMED-SCORE: %d\n", score))
```

### Modified Cabrillo Score

```go
// Read bonus points
var bonusPoints int
if err := db.QueryRow(`SELECT COALESCE(SUM(...), 0) FROM bonus_claims`).Scan(&bonusPoints); err != nil {
    bonusPoints = 0
}
score := (rawPoints * multiplier) + bonusPoints
buf.WriteString(fmt.Sprintf("CLAIMED-SCORE: %d\n", score))

// Add bonus claims as SOAPBOX lines
// SOAPBOX: Bonus: 100% Emergency Power (3 transmitters) = 300 pts
// SOAPBOX: Total Bonus Points = 750
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| HTML5 `<audio>` element | Web Audio API | Always — Web Audio API (2011) | Programmatic control, polyphony, fine-grained scheduling |
| WAV files via `<audio>` tag | `fetch()` + `decodeAudioData()` | Always — modern pattern | Non-blocking, cacheable, precise timing |
| `VACUUM INTO` for backup | `io.Copy` + WAL mode | D-11 decision | Non-blocking concurrent reads |
| Scripted `$state` primitive export | Object-wrapped `$state` | Phase 2 (wsState) → all downstream | Svelte 5 compatibility |

**Deprecated/outdated:**
- **`VACUUM INTO` for SQLite backup:** Replaced by `io.Copy` under WAL mode. VACUUM requires exclusive lock.
- **Premultiplied bonus in score:** ARRL rules state bonus is post-multiplier. The CONTEXT.md formula needs correction.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `score = (rawPoints * multiplier) + bonusPoints` (ARRL 7.3) — contradicts CONTEXT.md which says `(rawPoints + bonusPoints) * multiplier` | Scoring Integration | Cabrillo CLAIMED-SCORE wrong per ARRL rules; submission may be rejected or scored incorrectly |
| A2 | Audio file format is `.wav` (PCM) — users can provide any WAV | Audio Feedback | Non-PCM WAV may fail `decodeAudioData()` |
| A3 | Go `embed.FS` serves files at project-root-relative paths under `/frontend/build/` | Audio Loading | If SvelteKit build output path changes, audio 404s |
| A4 | 2026 bonus rules unchanged from 2025 (confirmed by ARRL rules page) | Bonus List | If ARRL publishes a revision after research date, list may be incomplete |
| A5 | `PRAGMA wal_checkpoint(TRUNCATE)` is safe during concurrent writes — it waits for readers to complete | SQLite Backup | If checkpoint blocks for >5s (busy_timeout), backup download times out |
| A6 | Web Audio API `decodeAudioData()` works for WAV files served from same origin (no CORS) | Audio Loading | Cross-origin audio would fail; but this is a LAN-only app with same-origin serving |
| A7 | Browser supports Web Audio API (all modern browsers per ASSUMED compatibility) | Audio Feedback | Very old Android browsers or Opera Mini would not play sounds |
| A8 | User provides audio files — application does not generate or bundle them | Audio Feedback | If user forgets to provide audio files, `fetch()` returns 404 and playback silently fails (caught by try/catch) |

## Open Questions

1. **Score formula: pre-multiplier or post-multiplier bonus addition?**
   - What we know: CONTEXT.md says `(raw_points + bonus_points) * multiplier`. ARRL rules section 7.3 says "bonus points will be added to the score, after the multiplier is applied."
   - What's unclear: User intent — did they mean post-multiplier or was the formula a simplification?
   - Recommendation: Flag in discuss phase. Implement post-multiplier (ARRL-correct) and note the discrepancy. If user wants pre-multiplier, it's a one-line change.

2. **Audio file format preference?**
   - What we know: D-06 says `.wav` or `.mp3`. WAV is simpler (no codec support issues), universally supported.
   - What's unclear: Does the user have audio files already? Do they care about file size (WAV is larger)?
   - Recommendation: Default to `.wav` (no codec decoding issues). Mention that `.mp3` works but requires browser codec support.

3. **Simulation test duration: wall-clock or QSO-count-based?**
   - What we know: D-13 says "2-hour simulation." Actual 2 hours in CI is impractical.
   - What's unclear: Should the test actually run for 2 hours of wall-clock time, or should it run quickly and simulate 2 hours worth of QSO activity (200+ QSOs)?
   - Recommendation: Use QSO-count-based (200+ QSOs, 3 clients) running in ~60 seconds. This validates integrity without CI timeout. Note that a true 2-hour soak test would be done during the manual field test (D-14).

4. **Bonus points tracking: should the bonus list auto-filter by station class?**
   - What we know: Some bonuses only apply to certain classes (e.g., safety_officer only for Class A).
   - What's unclear: Should the UI hide/show bonuses based on configured station class? Or show all and let the operator figure it out?
   - Recommendation: Show all bonuses, but dim/grey-out those not applicable to the configured class. This avoids data loss if class changes and is more transparent.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | All backend code, simulation test | ✓ | 1.25.0 | — |
| Node.js | Frontend build | ✓ | (inferred from package.json) | — |
| SQLite (via modernc.org/sqlite) | Database | ✓ | 1.51.0 (Go lib) | — |
| Browser with Web Audio API | Audio alerts | ✓ | N/A | Silent fallback (no sound) |
| Browser with localStorage | Mute/bonus persistence | ✓ | N/A | State lost on reload |
| `cp` / file I/O | Backup download | ✓ | stdlib | — |

**Missing dependencies with no fallback:** None — all required capabilities exist in the current environment.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing` (stdlib) + vitest for frontend |
| Config file | `go test ./...` (Go), vitest.config.ts (frontend) |
| Quick run command | `go test ./internal/handler/ -run TestSimulation -timeout 120s` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| UX-03 | Audio alert plays on QSO confirm and dupe | Manual / unit | `vitest run src/lib/audio.test.js` | ❌ Wave 0 |
| BON-01 | Bonus tracker with FD list and toggles | Integration | `go test ./internal/handler/ -run TestBonuses` | ❌ Wave 0 |
| BON-02 | Bonus points in score calculation | Unit | `go test ./internal/handler/ -run TestStats -count=1` | ❌ Wave 0 (modify stats_test.go) |
| BKUP-01 | One-click database backup | Integration | `go test ./internal/handler/ -run TestBackup` | ❌ Wave 0 |
| D-13 | Simulation test data integrity | Integration | `go test ./internal/handler/simtest/ -run TestSimulation -timeout 120s` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/handler/ -count=1 -timeout 30s` (bonus handler tests)
- **Per wave merge:** `go test ./... -count=1` (full suite)
- **Phase gate:** Full suite green + simulation test passing before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] `internal/handler/bonus_test.go` — covers GetBonuses/PutBonuses handlers
- [ ] `internal/handler/backup_test.go` — covers DownloadBackup handler
- [ ] `internal/handler/simtest/simtest_test.go` — multi-client simulation test
- [ ] `internal/model/bonus_test.go` — bonus claim validation, default list structure
- [ ] `internal/handler/stats_test.go` (modify) — add bonus_points assertion to existing tests
- [ ] `internal/cabrillo/cabrillo_test.go` (modify) — add CLAIMED-SCORE with bonus assertion
- [ ] `internal/db/schema.sql` — add `bonus_claims` table to test DB setup functions
- [ ] Frontend audio unit test: `frontend/src/lib/audio.svelte.js` — mock Web Audio API

## Security Domain

### Applicable ASVS Categories (Level 1)

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | LAN-only, no auth (per PROJECT.md constraints) |
| V3 Session Management | No | No sessions; stateless REST |
| V4 Access Control | No | Trusted LAN; no access control |
| V5 Input Validation | Yes | Bonus claim counts must be validated (non-negative integers, max 100 for counted bonuses) |
| V6 Cryptography | No | No sensitive data stored |
| V7 Error Handling | Yes | Backup download errors must not expose filesystem paths; use generic error messages |
| V11 Business Logic | Yes | Bonus points calculation must be correct per ARRL rules; intentional inflation detection unnecessary (honor system for Field Day) |

### Known Threat Patterns

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Path traversal in backup download (if dbPath is user-controlled) | Tampering | `dbPath` is a server-side env var (`FDLOGGER_DB_PATH`), not user input — no traversal risk |
| Integer overflow in bonus count inputs | Tampering | Validate count range per bonus type (0–20 for transmitters, 0–10 for messages, 0–5 for youth) |
| Unvalidated bonus_id injection | Tampering | Server should validate bonus IDs against known list; ignore unknown IDs in PUT handler |
| Large audio file DoS | Denial of Service | User-provided files; no validation needed at app level (trusted LAN environment) |
| SQLite file exposure via backup path | Information Disclosure | Backup endpoint streams the file — acceptable on trusted LAN; all data is non-sensitive (callsigns, exchanges) |

## Sources

### Primary (HIGH confidence)
- [arrl.org/field-day-rules](https://www.arrl.org/field-day-rules) — 2026 ARRL Field Day rules, section 7.3 bonus points list [VERIFIED: fetched 2026-06-04, states "no changes from 2025 rules"]
- Codebase files read directly from repo:
  - `main.go` — Chi router pattern, embed.FS, handler wiring
  - `internal/handler/qso.go` — CreateQSO handler, dupe detection, broadcast
  - `internal/handler/stats.go` — Score calculation (line 45: `score := rawPoints * multiplier`)
  - `internal/cabrillo/cabrillo.go` — Cabrillo generation, CLAIMED-SCORE
  - `internal/handler/config.go` — GET/PUT handler pattern for bonus handler
  - `internal/handler/export.go` — File download pattern for backup
  - `internal/db/db.go` — WAL mode configuration (line 12)
  - `internal/db/schema.sql` — Existing schema
  - `frontend/src/routes/+page.svelte` — Header bar layout, export button
  - `frontend/src/lib/components/StationConfig.svelte` — Expandable panel pattern
  - `frontend/src/lib/components/QsoEntryForm.svelte` — Audio trigger points
  - `frontend/src/lib/stores/qso.svelte.js` — $state store pattern, stats object
  - `frontend/src/lib/ws.svelte.js` — WebSocket, own-QSO detection needs
  - `frontend/src/lib/api.js` — API client pattern
  - `frontend/src/lib/components/OperatorSelector.svelte` — localStorage pattern
  - `frontend/svelte.config.js` — Static adapter config
  - `internal/handler/ws_test.go` — WebSocket test + concurrent QSO test patterns
  - `internal/handler/stats_test.go` — Stats test with in-memory SQLite
  - `internal/handler/qso_test.go` — Handler test pattern
  - `internal/qso/points.go` — Point calculation per mode

### Secondary (MEDIUM confidence)
- [sqlite.org/wal.html](https://sqlite.org/wal.html) — WAL mode concurrency guarantees [CITED]
- Web Audio API MDN documentation — AudioContext, decodeAudioData, createBufferSource [ASSUMED: based on training knowledge of stable Web API]

### Tertiary (LOW confidence)
- Browser compatibility for Web Audio API — assumed all modern browsers (Chrome 35+, FF 25+, Safari 6+) [ASSUMED]
- Audio file encoding requirements for `decodeAudioData()` — assumed WAV/PCM works universally [ASSUMED]

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all tools already in project or built-in browser APIs
- Architecture: HIGH — existing patterns (StationConfig, Chi router, $state, localStorage) well-documented in codebase
- ARRL bonus list: HIGH — verified from official 2026 rules page
- Pitfalls: MEDIUM — most from established patterns, but audio autoplay policy varies by browser version
- Scoring formula: MEDIUM — discrepancy between CONTEXT.md and ARRL rules needs user confirmation

**Research date:** 2026-06-04
**Valid until:** 2026-07-04 (stable tech; bonus rules valid for 2026 Field Day, June 27–28)
