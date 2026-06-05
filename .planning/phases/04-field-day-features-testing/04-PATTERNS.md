# Phase 04: Field Day Features & Testing - Pattern Map

**Mapped:** 2026-06-04
**Files analyzed:** 14 (new or modified)
**Analogs found:** 14 / 14
**Confidence:** HIGH — every file has a close analog in the existing codebase

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/db/schema.sql` | schema | file-I/O | `internal/db/schema.sql` | exact (modification) |
| `internal/model/bonus.go` | model | config/data | `internal/model/config.go` | exact |
| `internal/handler/bonus.go` | controller | request-response CRUD | `internal/handler/config.go` | exact |
| `internal/handler/backup.go` | controller | file-I/O streaming | `internal/handler/export.go` | exact |
| `internal/handler/stats.go` | controller | request-response read | `internal/handler/stats.go` | exact (modification) |
| `internal/cabrillo/cabrillo.go` | utility | text generation | `internal/cabrillo/cabrillo.go` | exact (modification) |
| `main.go` | wiring | config | `main.go` | exact (modification) |
| `frontend/src/lib/api.js` | utility | request-response | `frontend/src/lib/api.js` | exact (modification) |
| `frontend/src/lib/stores/qso.svelte.js` | store | event-driven | `frontend/src/lib/ws.svelte.js` | role-match |
| `frontend/src/lib/components/BonusTracker.svelte` | component | event-driven | `frontend/src/lib/components/StationConfig.svelte` | exact |
| `frontend/src/lib/audio.svelte.js` | utility | event-driven | `frontend/src/lib/ws.svelte.js` | role-match |
| `frontend/src/routes/+page.svelte` | view | layout | `frontend/src/routes/+page.svelte` | exact (modification) |
| `internal/handler/simtest/simtest_test.go` | test | batch/concurrent | `internal/handler/ws_test.go` | role-match |
| `frontend/src/lib/components/StatsBar.svelte` | component | data display | `frontend/src/lib/components/StatsBar.svelte` | exact (modification) |

## Pattern Assignments

---

### 1. `internal/db/schema.sql` — add `bonus_claims` table

**Role:** schema (SQL DDL)
**Data flow:** declarative definition → DB migration on startup
**Closest analog:** `internal/db/schema.sql` (same file, existing CREATE TABLE blocks)

**Existing pattern** (lines 1-28):
```sql
CREATE TABLE IF NOT EXISTS qsos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TEXT NOT NULL,
    callsign TEXT NOT NULL,
    band TEXT NOT NULL,
    mode TEXT NOT NULL,
    sent_exchange TEXT NOT NULL,
    recv_exchange TEXT NOT NULL,
    client_id TEXT UNIQUE,
    operator TEXT,
    is_dupe INTEGER NOT NULL DEFAULT 0,
    points INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_qsos_callsign ON qsos(callsign);
CREATE INDEX IF NOT EXISTS idx_qsos_timestamp ON qsos(timestamp);
CREATE INDEX IF NOT EXISTS idx_qsos_band_mode ON qsos(band, mode);

CREATE TABLE IF NOT EXISTS station_config (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    callsign TEXT NOT NULL DEFAULT 'N0CALL',
    class TEXT NOT NULL DEFAULT '1D',
    arrl_section TEXT NOT NULL DEFAULT 'EMA',
    transmitter_count INTEGER NOT NULL DEFAULT 1,
    power_level TEXT NOT NULL DEFAULT 'LOW',
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
```

**New block to append:**
```sql
CREATE TABLE IF NOT EXISTS bonus_claims (
    bonus_id TEXT PRIMARY KEY,
    claimed INTEGER NOT NULL DEFAULT 0,
    count INTEGER NOT NULL DEFAULT 0,
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
```

**Adaptation notes:**
- Follow `CREATE TABLE IF NOT EXISTS` pattern for idempotent migration
- No index needed (small table, fewer than 20 rows)
- `bonus_id` is a TEXT primary key (e.g., `emergency_power`, `media_publicity`)
- `claimed` is INTEGER 0/1 (SQLite has no native boolean)
- `count` is INTEGER for counted bonuses; 0 for boolean-only
- Use `datetime('now')` default matching existing tables

---

### 2. `internal/model/bonus.go` — BonusClaim struct, default list, validation

**Role:** model (Go struct definitions + business logic)
**Data flow:** hardcoded bonus list → struct definitions → validation logic
**Closest analog:** `internal/model/config.go` (struct definition, default factory, validation function)

**Analog excerpt** — `internal/model/config.go` (lines 1-40):
```go
package model

type StationConfig struct {
    Callsign         string `json:"callsign"`
    Class            string `json:"class"`
    ARRLSection      string `json:"arrl_section"`
    TransmitterCount int    `json:"transmitter_count"`
    PowerLevel       string `json:"power_level"`
    UpdatedAt        string `json:"updated_at,omitempty"`
}

func DefaultStationConfig() StationConfig {
    return StationConfig{
        Callsign:         "N0CALL",
        Class:            "1D",
        ARRLSection:      "EMA",
        TransmitterCount: 1,
        PowerLevel:       "LOW",
    }
}

func ValidateStationConfig(cfg StationConfig) string {
    if cfg.Callsign == "" {
        return "callsign is required"
    }
    // ... field validation ...
    return ""
}
```

**Core pattern to replicate:**

1. **Bonus item definition** — a struct for immutable bonus metadata:
```go
type BonusItem struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    RuleRef     string `json:"rule_ref"`
    Points      int    `json:"points"`
    IsCounted   bool   `json:"is_counted"`
    MaxCount    int    `json:"max_count,omitempty"`
    DefaultCount int   `json:"default_count,omitempty"`
}
```

2. **Bonus claim state** — a struct for per-bonus claim data from DB:
```go
type BonusClaim struct {
    BonusID   string `json:"bonus_id"`
    Claimed   bool   `json:"claimed"`
    Count     int    `json:"count"`
    UpdatedAt string `json:"updated_at,omitempty"`
}
```

3. **Default bonus list** — hardcoded constant slice (D-01 locked):
```go
var DefaultBonuses = []BonusItem{
    {ID: "emergency_power", Name: "100% Emergency Power", RuleRef: "7.3.1", Points: 100, IsCounted: true, MaxCount: 20},
    {ID: "media_publicity", Name: "Media Publicity", RuleRef: "7.3.2", Points: 100, IsCounted: false},
    // ... 18 items per RESEARCH.md § 2026 ARRL Field Day Bonus Points List
}
```

4. **Validation function** patterned after `ValidateStationConfig`:
```go
func ValidateBonusClaims(claims map[string]BonusClaim) string {
    // Validate bonus_id exists in DefaultBonuses
    // Validate count is non-negative and ≤ MaxCount
    return ""
}
```

5. **BonusClaim as map alias** — the API uses `map[string]struct{...}` from JSON, not a flat slice. Match the RESEARCH.md API format: `{"emergency_power": {"claimed": true, "count": 3}, ...}`.

**No analog found nuance:** The `DefaultBonuses` constant slice pattern has no exact analog in `model/config.go` (which has just one struct, not a list). Pattern the slice literal after Go convention with inline struct literals. The map-based API request/response format differs from `StationConfig`'s flat struct — use `map[string]BonusClaim` for the handler layer.

---

### 3. `internal/handler/bonus.go` — GET/PUT /api/bonuses handler

**Role:** controller (HTTP handler for CRUD)
**Data flow:** HTTP request → JSON decode → DB read/write → JSON encode response
**Closest analog:** `internal/handler/config.go` (GET/PUT StationConfig — same pattern)

**Imports pattern** — `internal/handler/config.go` (lines 1-9):
```go
package handler

import (
    "database/sql"
    "encoding/json"
    "net/http"

    "github.com/jeremy/mlogger-fd/internal/model"
)
```

**GET handler pattern** — `internal/handler/config.go` (lines 11-30):
```go
func GetStationConfig(db *sql.DB, w http.ResponseWriter, r *http.Request) {
    var cfg model.StationConfig
    err := db.QueryRow(`SELECT callsign, class, arrl_section, transmitter_count, power_level
        FROM station_config WHERE id = 1`).Scan(
        &cfg.Callsign, &cfg.Class, &cfg.ARRLSection,
        &cfg.TransmitterCount, &cfg.PowerLevel,
    )
    if err == sql.ErrNoRows {
        cfg = model.DefaultStationConfig()
    } else if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(cfg)
}
```

**PUT handler pattern** — `internal/handler/config.go` (lines 32-64):
```go
func PutStationConfig(db *sql.DB, w http.ResponseWriter, r *http.Request) {
    var cfg model.StationConfig
    if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
        return
    }

    if msg := model.ValidateStationConfig(cfg); msg != "" {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": msg})
        return
    }

    _, err := db.Exec(`INSERT OR REPLACE INTO station_config
        (id, callsign, class, arrl_section, transmitter_count, power_level, updated_at)
        VALUES (1, ?, ?, ?, ?, ?, datetime('now'))`,
        cfg.Callsign, cfg.Class, cfg.ARRLSection,
        cfg.TransmitterCount, cfg.PowerLevel,
    )
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(cfg)
}
```

**Adaptation notes for `handler/bonus.go`:**
- **GET:** Query all rows from `bonus_claims`, build a `map[string]BonusClaim` from rows. If table is empty, return an empty map `{}` (or populate from `DefaultBonuses` with all `claimed: false, count: 0`).
- **PUT:** Decode incoming `map[string]map[string]interface{}` body. Iterate over keys, validate each bonus_id against `DefaultBonuses`, then `INSERT OR REPLACE INTO bonus_claims` in a transaction.
- **Error format:** Use the same `map[string]string{"error": "..."}` pattern for 400/500 responses.
- **Response format:** Return the full map after save (mirrors config.go returning the saved struct).
- **DB:** Takes `*sql.DB` — follow the handler function signature pattern.

---

### 4. `internal/handler/backup.go` — GET /api/backup/db handler

**Role:** controller (file streaming download)
**Data flow:** HTTP request → open SQLite file → set Content-Disposition → `io.Copy` stream → response
**Closest analog:** `internal/handler/export.go` (Cabrillo file download)

**Imports pattern** — `internal/handler/export.go` (lines 1-10):
```go
package handler

import (
    "database/sql"
    "fmt"
    "net/http"
    // ... project imports
)
```

**Core download pattern** — `internal/handler/export.go` (lines 12-29):
```go
func ExportCabrillo(db *sql.DB, w http.ResponseWriter, r *http.Request) {
    result, err := cabrillo.Generate(db)
    if err != nil {
        http.Error(w, "Failed to generate Cabrillo file", http.StatusInternalServerError)
        return
    }

    // Read callsign from station_config for filename, fall back to n0call
    callsign := "n0call"
    var c string
    if err := db.QueryRow("SELECT callsign FROM station_config WHERE id = 1").Scan(&c); err == nil && c != "" {
        callsign = strings.ToLower(c)
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s_field_day.cbr\"", callsign))
    w.Write([]byte(result))
}
```

**Adaptation for `handler/backup.go`:**

```go
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
    // Optional: flush WAL to main DB
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
    io.Copy(w, f)
}
```

**Key differences from analog:**
- `Content-Type` is `application/octet-stream` (binary), not `text/plain`
- Uses `io.Copy` for streaming (no `w.Write([]byte(...))`)
- Needs `dbPath` parameter (passed via closure in `main.go`)
- Uses `time.Now().UTC().Format("20060102_150405")` for D-10 timestamped filename (YYYYMMDD_HHMMSS)
- `os.Open` instead of in-memory generation — follows the same `defer f.Close()` pattern

---

### 5. `internal/handler/stats.go` — modify score to include bonus points

**Role:** controller (read-only stats computation)
**Data flow:** SQL queries → compute score with bonus → JSON response
**Closest analog:** `internal/handler/stats.go` (same file — modify existing)

**Current score calculation** — `internal/handler/stats.go` (lines 37-45):
```go
var multiplier int
if err := db.QueryRow("SELECT COUNT(DISTINCT band || '_' || mode) FROM qsos WHERE is_dupe = 0").Scan(&multiplier); err != nil {
    multiplier = 1
}
if multiplier < 1 {
    multiplier = 1
}

score := rawPoints * multiplier
```

**Current response** — `internal/handler/stats.go` (lines 65-74):
```go
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]interface{}{
    "total":      total,
    "raw_points": rawPoints,
    "multiplier": multiplier,
    "score":      score,
    "rate_10min": rate10min,
    "rate_1hr":   rate1hr,
    "breakdown":  breakdown,
})
```

**Modification approach:**

Add bonus points query between multiplier and score calculation:
```go
var bonusPoints int
if err := db.QueryRow(`SELECT COALESCE(SUM(CASE
    WHEN bonus_id = 'emergency_power' THEN claimed * count * 100
    WHEN bonus_id = 'message_handling' THEN claimed * MIN(count, 10) * 10
    WHEN bonus_id = 'youth_participation' THEN claimed * MIN(count, 5) * 20
    WHEN bonus_id = 'gota_bonus' THEN claimed * count * 5 + CASE WHEN claimed AND count >= 10 THEN 100 ELSE 0 END
    WHEN bonus_id = 'web_submission' THEN claimed * 50
    WHEN bonus_id = 'safety_officer' THEN claimed * 100
    WHEN bonus_id = 'site_responsibilities' THEN claimed * 50
    ELSE claimed * ? 
END), 0) FROM bonus_claims`, 100).Scan(&bonusPoints); err != nil {
    bonusPoints = 0
}
```

Then:
```go
// score = (rawPoints * multiplier) + bonusPoints  // ARRL rules: bonus AFTER multiplier
score := (rawPoints * multiplier) + bonusPoints
```

Add `"bonus_points": bonusPoints` to the response map.

**Adaptation notes:**
- **Crucial:** RESEARCH.md flags that ARRL rules (section 7.3) say bonus points are added AFTER the multiplier: `score = (rawPoints * multiplier) + bonusPoints`, NOT `(rawPoints + bonusPoints) * multiplier`. This contradicts CONTEXT.md's phrase. Implement the ARRL-correct formula.
- New `bonus_points` field in the JSON response.
- The bonus query uses `COALESCE(SUM(...), 0)` for idempotent behavior when `bonus_claims` table is empty.
- Use the same error handling pattern: if `QueryRow` fails, `bonusPoints = 0` (safe fallback per RESEARCH.md mid-600s).

---

### 6. `internal/cabrillo/cabrillo.go` — modify for bonus points

**Role:** utility (text generation for export)
**Data flow:** read DB → generate Cabrillo text → return string
**Closest analog:** `internal/cabrillo/cabrillo.go` (same file — modify existing)

**Current score section** — `internal/cabrillo/cabrillo.go` (lines 67-79):
```go
var rawPoints int
var multiplier int
if err := db.QueryRow("SELECT COALESCE(SUM(points), 0) FROM qsos WHERE is_dupe = 0").Scan(&rawPoints); err != nil {
    rawPoints = 0
}
if err := db.QueryRow("SELECT COUNT(DISTINCT band || '_' || mode) FROM qsos WHERE is_dupe = 0").Scan(&multiplier); err != nil {
    multiplier = 1
}
if multiplier < 1 {
    multiplier = 1
}
score := rawPoints * multiplier
buf.WriteString(fmt.Sprintf("CLAIMED-SCORE: %d\n", score))
```

**Header output pattern** — `internal/cabrillo/cabrillo.go` (lines 57-65):
```go
buf.WriteString("START-OF-LOG: 3.0\n")
buf.WriteString("CREATED-BY: FDLogger v1.0\n")
buf.WriteString("CONTEST: ARRL-FIELD-DAY\n")
buf.WriteString(fmt.Sprintf("CALLSIGN: %s\n", callsign))
buf.WriteString(fmt.Sprintf("ARRL-SECTION: %s\n", section))
// ... more headers ...
```

**Modification approach:**
1. After multiplier calculation, add the same bonus points query as in `stats.go`.
2. Change `score := rawPoints * multiplier` to `score := (rawPoints * multiplier) + bonusPoints`.
3. After CLAIMED-SCORE line, iterate over `bonus_claims` rows to add SOAPBOX comment lines:

```go
// After CLAIMED-SCORE, before QSO lines:
bonusRows, err := db.Query("SELECT bonus_id, claimed, count FROM bonus_claims WHERE claimed = 1 ORDER BY bonus_id")
if err == nil {
    defer bonusRows.Close()
    for bonusRows.Next() {
        var bid string
        var claimed, count int
        bonusRows.Scan(&bid, &claimed, &count)
        // Lookup name from default list (or use bid as fallback)
        // buf.WriteString(fmt.Sprintf("SOAPBOX: Bonus: %s = %d pts\n", name, pts))
    }
    buf.WriteString(fmt.Sprintf("SOAPBOX: Total Bonus Points = %d\n", bonusPoints))
}
```

**Adaptation notes:**
- Use the same error handling pattern: `if err != nil { bonusPoints = 0 }` — safe fallback.
- For the bonus name lookup, either import `model.DefaultBonuses` or hardcode a name map in `cabrillo.go`.
- SOAPBOX lines go after CLAIMED-SCORE and before QSO entries (consistent with Cabrillo format).
- Keep using `buf.WriteString(fmt.Sprintf(...))` pattern — don't switch to `fmt.Fprintf`.

---

### 7. `main.go` — add routes for bonuses and backup

**Role:** wiring (router setup)
**Data flow:** HTTP routing table configuration
**Closest analog:** `main.go` (same file — modify existing)

**Existing route pattern** — `main.go` (lines 43-73):
```go
r.Route("/api", func(r chi.Router) {
    r.Get("/health", handler.HealthCheck)
    r.Get("/check-dupe", func(w http.ResponseWriter, r *http.Request) {
        handler.CheckDupeHandler(database, w, r)
    })
    r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
        handler.GetStats(database, w, r)
    })
    r.Get("/export/cabrillo", func(w http.ResponseWriter, r *http.Request) {
        handler.ExportCabrillo(database, w, r)
    })
    r.Get("/station-config", func(w http.ResponseWriter, r *http.Request) {
        handler.GetStationConfig(database, w, r)
    })
    r.Put("/station-config", func(w http.ResponseWriter, r *http.Request) {
        handler.PutStationConfig(database, w, r)
    })
    r.Post("/sync", func(w http.ResponseWriter, r *http.Request) {
        handler.SyncQSOs(database, hub, w, r)
    })
    r.Route("/qso", func(r chi.Router) {
        r.Post("/", func(w http.ResponseWriter, r *http.Request) {
            handler.CreateQSO(database, hub, w, r)
        })
        r.Get("/", func(w http.ResponseWriter, r *http.Request) {
            handler.ListQSOs(database, w, r)
        })
        r.Put("/{id}", func(w http.ResponseWriter, r *http.Request) {
            handler.UpdateQSO(database, w, r)
        })
    })
})
```

**Routes to add** — insert after `/station-config` block and before `/sync`:
```go
r.Get("/bonuses", func(w http.ResponseWriter, r *http.Request) {
    handler.GetBonuses(database, w, r)
})
r.Put("/bonuses", func(w http.ResponseWriter, r *http.Request) {
    handler.PutBonuses(database, w, r)
})
r.Get("/backup/db", func(w http.ResponseWriter, r *http.Request) {
    handler.DownloadBackup(database, dbPath, w, r)
})
```

**Adaptation notes:**
- Closure pattern identical to existing: `func(w http.ResponseWriter, r *http.Request) { handler.Xxx(database, w, r) }`
- Backup handler needs `dbPath` in addition to `database`
- Order: place bonuses routes near `station-config` (both are config-like), backup near `export/cabrillo` (both are downloads)
- `r.Get("/backup/db", ...)` — sub-path under `/api` via the `r.Route("/api", ...)` wrapper

---

### 8. `frontend/src/lib/api.js` — add bonus and backup API functions

**Role:** utility (HTTP API client)
**Data flow:** function call → fetch → JSON parse → return
**Closest analog:** `frontend/src/lib/api.js` (same file — existing getStationConfig/putStationConfig)

**Analog excerpt** — `frontend/src/lib/api.js` (lines 66-85):
```javascript
export async function getStationConfig() {
    const res = await fetch(`${BASE_URL}/api/station-config`);
    if (!res.ok) {
        throw new Error('Failed to fetch station config');
    }
    return res.json();
}

export async function putStationConfig(data) {
    const res = await fetch(`${BASE_URL}/api/station-config`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
    });
    const json = await res.json();
    if (!res.ok) {
        throw new Error(json.error || 'Failed to save station config');
    }
    return json;
}
```

**Functions to add:**

```javascript
export async function getBonuses() {
    const res = await fetch(`${BASE_URL}/api/bonuses`);
    if (!res.ok) {
        throw new Error('Failed to fetch bonuses');
    }
    return res.json();
}

export async function putBonuses(data) {
    const res = await fetch(`${BASE_URL}/api/bonuses`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
    });
    const json = await res.json();
    if (!res.ok) {
        throw new Error(json.error || 'Failed to save bonuses');
    }
    return json;
}

export function downloadBackup() {
    window.location.href = `${BASE_URL}/api/backup/db`;
}
```

**Adaptation notes:**
- `getBonuses()` and `putBonuses()` follow the EXACT same pattern as `getStationConfig()` / `putStationConfig()` — identical structure, error handling, and `BASE_URL` usage
- `downloadBackup()` follows the Cabrillo export pattern from `+page.svelte` line 40-42: `window.location.href = '/api/export/cabrillo'` — triggers browser download via navigation
- Use `BASE_URL` prefix consistently with all existing functions

---

### 9. `frontend/src/lib/stores/qso.svelte.js` — add bonusClaims state

**Role:** store (reactive Svelte 5 module-level state)
**Data flow:** shared state → components read/write → server sync
**Closest analog:** `frontend/src/lib/ws.svelte.js` (object-based `$state` export pattern)

**Analog excerpt** — `frontend/src/lib/ws.svelte.js` (line 7):
```javascript
// Use object-based $state since Svelte 5 forbids reassigning exported $state variables
export const wsState = $state({ connected: false });
```

**Also check:** `frontend/src/lib/sync.svelte.js` (line 5):
```javascript
export const queueState = $state({ queueLength: 0, syncing: false });
```

**Addition to `qso.svelte.js`:**
```javascript
// Bonus claims: { bonus_id: { claimed: bool, count: int } }
export const bonusClaims = $state({});
```

**Usage pattern (in components that import `bonusClaims`):**
```javascript
import { bonusClaims } from '$lib/stores/qso.svelte.js';

// READ: property access — reactive in Svelte 5
console.log(bonusClaims.emergency_power?.claimed);

// WRITE: mutate properties, NOT reassign
bonusClaims.emergency_power = { claimed: true, count: 3 };
// DO NOT: bonusClaims = { emergency_power: {...} };  // BROKEN in Svelte 5
```

**Adaptation notes:**
- MUST use object-based `$state`: `export const bonusClaims = $state({})` — Svelte 5 forbids reassigning exported `$state` primitives
- Property mutation (`bonusClaims.emergency_power = ...`) works correctly
- `bonusClaims` is an object (map), not a Map or array. Keys are bonus_id strings.
- Follow the same module convention: `.svelte.js` extension (already in `qso.svelte.js`)

---

### 10. `frontend/src/lib/components/BonusTracker.svelte` — expandable bonus panel

**Role:** component (UI widget with toggle + form)
**Data flow:** user interaction → toggle expand → display/update bonus list → API save → feedback toast
**Closest analog:** `frontend/src/lib/components/StationConfig.svelte` (identical expandable panel pattern)

**Analog excerpt** — `StationConfig.svelte` (lines 1-54, script section):
```svelte
<script>
    import { onMount } from 'svelte';
    import { getStationConfig, putStationConfig } from '$lib/api.js';

    let expanded = $state(false);
    let callsign = $state('');
    let cls = $state('');
    let section = $state('');
    let txCount = $state(1);
    let power = $state('LOW');
    let saved = $state(false);
    let saveTimer;

    function toggle() {
        expanded = !expanded;
    }

    async function loadConfig() {
        try {
            const cfg = await getStationConfig();
            callsign = cfg.callsign || '';
            cls = cfg.class || '';
            section = cfg.arrl_section || '';
            txCount = cfg.transmitter_count || 1;
            power = cfg.power_level || 'LOW';
        } catch {
            // silently use defaults
        }
    }

    async function handleSubmit(e) {
        e.preventDefault();
        try {
            await putStationConfig({
                callsign,
                class: cls,
                arrl_section: section,
                transmitter_count: txCount,
                power_level: power,
            });
            saved = true;
            if (saveTimer) clearTimeout(saveTimer);
            saveTimer = setTimeout(() => { saved = false; }, 2000);
        } catch {
            // silently handle error
        }
    }

    onMount(() => {
        loadConfig();
    });
</script>
```

**Analog excerpt** — `StationConfig.svelte` (lines 54-117, template + style):
```svelte
<div class="station-config">
    <button class="config-toggle" onclick={toggle} aria-label="Config">
        <span class="toggle-icon">⚙</span> Config
    </button>

    {#if expanded}
        <div class="config-panel">
            <form onsubmit={handleSubmit}>
                <div class="config-form">
                    <div class="field">
                        <label for="cfg-callsign">Callsign</label>
                        <input id="cfg-callsign" type="text" bind:value={callsign} placeholder="N0CALL" />
                    </div>
                    <!-- ... more fields ... -->
                </div>
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

**Panel positioning from CSS** — `StationConfig.svelte` (lines 147-158):
```css
.config-panel {
    position: absolute;
    top: 44px;
    right: 8px;
    background: var(--color-surface);
    border: 1px solid var(--color-border-light);
    border-radius: 8px;
    padding: 16px;
    box-shadow: 0 4px 16px rgba(0,0,0,0.12);
    z-index: 100;
    min-width: 260px;
}
```

**Toggle button CSS** — `StationConfig.svelte` (lines 127-137):
```css
.config-toggle {
    padding: 4px 12px;
    font-size: 14px;
    font-weight: 600;
    border: 2px solid var(--color-surface);
    border-radius: 6px;
    background: transparent;
    color: var(--color-surface);
    cursor: pointer;
    white-space: nowrap;
}
```

**Adaptation notes for `BonusTracker.svelte`:**

1. **State:** Replace individual form fields with `bonusList` (array from imported `DefaultBonuses` data — or fetched from server with metadata) and `claims` (reactive map of bonus_id → `{claimed, count}`). Use `bonusClaims` from store.
2. **Load on mount:** Call `getBonuses()` from api.js, populate `bonusClaims`.
3. **Iterate bonus list:** Use `{#each bonusList as item}` — each row renders a toggle checkbox + optional number input for counted bonuses.
4. **Toggle button:** Same style class (`config-toggle` → rename to `bonus-toggle` or reuse), icon: `★` (star) per RESEARCH.md header diagram, label: "Bonuses".
5. **Panel styling:** Copy exact `position: absolute; top: 44px; right: 8px; z-index: 100` positioning. Adjust `right` offset if the Bonuses button is to the left of Export button.
6. **Save feedback:** Copy the exact `saved` + `setTimeout` pattern from lines 41-43.
7. **localStorage backup:** On save, also write to `localStorage.setItem('fdlogger_bonus_claims', JSON.stringify(bonusClaims))`. On mount, check localStorage first for instant display, then fetch from server.

**Key differences:**
- Multiple items (list iteration) vs single config form
- Mix of toggle checkbox + optional number input per item
- JSON shape is a map `{bonus_id: {claimed, count}}`, not a flat struct
- Added localStorage persistence (from OperatorSelector pattern) for resilience

---

### 11. `frontend/src/lib/audio.svelte.js` — audio utility with mute state

**Role:** utility (Web Audio API wrapper)
**Data flow:** function call → create/lazy-init AudioContext → fetch + decode audio file → play buffer → mute state check
**Closest analog:** `frontend/src/lib/ws.svelte.js` (object-based `$state` export, module-level state management)

**Analog excerpt** — `frontend/src/lib/ws.svelte.js` (lines 1-15):
```javascript
// WebSocket client module for real-time multi-user QSO sync
import { qsos, fetchStats } from '$lib/stores/qso.svelte.js';
import { addToCache } from '$lib/db.js';

// Use object-based $state since Svelte 5 forbids reassigning exported $state variables
export const wsState = $state({ connected: false });

let ws = null;
let reconnectTimer = null;
let shouldReconnect = true;
```

**Core module pattern:**
```javascript
// audio.svelte.js

// Object-based $state for mute (Svelte 5 pattern)
export const audioState = $state({ muted: false });

// Lazy-initialized AudioContext (browsers require user gesture)
let audioCtx = null;
const buffers = {};  // Cache decoded AudioBuffers

function ensureContext() {
    if (!audioCtx) {
        audioCtx = new AudioContext();
    }
    if (audioCtx.state === 'suspended') {
        audioCtx.resume();
    }
}

async function loadSound(name) {
    if (buffers[name]) return buffers[name];
    const response = await fetch(`/audio/${name}.wav`);
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

**Adaptation notes:**
- Object-based `$state` export EXACTLY matches `ws.svelte.js` pattern: `export const audioState = $state({ muted: false })`.
- Module-level private variables (`let audioCtx`, `const buffers`) follow the `let ws = null` pattern.
- localStorage pattern matches `OperatorSelector.svelte` lines 4-8 and the theme toggle in `+page.svelte` lines 15-17: read on init, save on change.
- Key prefix convention: `fdlogger_muted` matches `fdlogger_operator`, `fdlogger_theme`.
- `.svelte.js` extension required for `$state` usage.
- Audio trigger points: `playSound('confirm')` in `QsoEntryForm.svelte` after successful create; `playSound('dupe')` in `handleCheckDupe()` when dupe detected.

---

### 12. `frontend/src/routes/+page.svelte` — add Bonus, Backup, Mute to header

**Role:** view (page layout)
**Data flow:** component composition → header bar layout
**Closest analog:** `frontend/src/routes/+page.svelte` (same file — modify existing)

**Current header-right** — `+page.svelte` (lines 61-64):
```svelte
<div class="header-right">
    <StationConfig />
    <button class="export-btn" onclick={exportCabrillo}>Export Cabrillo</button>
</div>
```

**Current header-left** — `+page.svelte` (lines 46-57):
```svelte
<div class="header-left">
    <button class="theme-toggle" onclick={toggleTheme} aria-label="Toggle dark mode">
        {theme === 'light' ? '☀' : '☾'}
    </button>
    <h1 class="title">FD Logger</h1>
    <span class="ws-status" class:online={wsState.connected} class:offline={!wsState.connected}>
        {wsState.connected ? '● Live' : '● Disconnected'}
    </span>
    {#if queueState.queueLength > 0}
        <span class="queue-count">{queueState.syncing ? 'Syncing...' : `${queueState.queueLength} queued`}</span>
    {/if}
</div>
```

**Script section additions:**
```svelte
<script>
    // ... existing imports ...
    import BonusTracker from '$lib/components/BonusTracker.svelte';
    import { audioState, toggleMute } from '$lib/audio.svelte.js';
    import { downloadBackup } from '$lib/api.js';

    // ... existing code ...
</script>
```

**Template modifications:**

1. In `header-left`, add mute toggle between theme toggle and title:
```svelte
<button class="theme-toggle" onclick={toggleMute} aria-label="Toggle audio">
    {audioState.muted ? '🔇' : '🔊'}
</button>
```

2. In `header-right`, add Bonuses and Backup buttons:
```svelte
<div class="header-right">
    <StationConfig />
    <BonusTracker />
    <button class="export-btn" onclick={exportCabrillo}>Export Cabrillo</button>
    <button class="export-btn" onclick={downloadBackup}>↓ Backup</button>
</div>
```
NOTE: `exportCabrillo` is the existing function; `downloadBackup` is new.

3. Add backup toast near the backup button or in header-right. Use StationConfig's 2-second `$state` boolean pattern:
```svelte
let backupToast = $state(false);
let backupTimer;

function handleBackup() {
    downloadBackup();
    backupToast = true;
    if (backupTimer) clearTimeout(backupTimer);
    backupTimer = setTimeout(() => { backupToast = false; }, 2000);
}
```

**Adaptation notes:**
- Mute button reuses `.theme-toggle` CSS class (same styling, transparent with border)
- Backup button reuses `.export-btn` CSS class (same styling as Export Cabrillo)
- Backup toast uses the same `$state` boolean + `setTimeout` pattern from `StationConfig.svelte` lines 41-43
- BonusTracker component inserted between StationConfig and Export buttons — its toggle button uses same `.config-toggle` styling
- No new CSS classes needed — reuse existing `.theme-toggle`, `.export-btn`, `.config-toggle`

---

### 13. `internal/handler/simtest/simtest_test.go` — multi-client simulation test

**Role:** test (Go integration test with httptest)
**Data flow:** goroutine clients → submit QSOs via HTTP/WS → verify server data integrity
**Closest analog:** `internal/handler/ws_test.go` (httptest.Server + gorilla/websocket + concurrent QSO pattern)

**Test DB setup pattern** — `internal/handler/qso_test.go` (lines 14-39):
```go
func setupHandlerTestDB(t *testing.T) *sql.DB {
    t.Helper()
    db, err := sql.Open("sqlite", "file::memory:?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&cache=shared")
    if err != nil {
        t.Fatalf("failed to open test DB: %v", err)
    }
    db.SetMaxOpenConns(1)
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS qsos (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        timestamp TEXT NOT NULL,
        callsign TEXT NOT NULL,
        band TEXT NOT NULL,
        mode TEXT NOT NULL,
        sent_exchange TEXT NOT NULL,
        recv_exchange TEXT NOT NULL,
        operator TEXT,
        is_dupe INTEGER NOT NULL DEFAULT 0,
        points INTEGER NOT NULL DEFAULT 0,
        created_at TEXT NOT NULL DEFAULT (datetime('now'))
    )`)
    if err != nil {
        t.Fatalf("failed to create table: %v", err)
    }
    t.Cleanup(func() { db.Close() })
    return db
}
```

**WebSocket + httptest.Server pattern** — `internal/handler/ws_test.go` (lines 44-60):
```go
func TestCreateQSOBroadcast(t *testing.T) {
    db := setupHandlerTestDB(t)
    hub := ws.NewHub()
    go hub.Run()
    time.Sleep(10 * time.Millisecond)

    // Create test server that handles /ws and /api/qso
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if strings.HasPrefix(r.URL.Path, "/ws") {
            ServeWS(hub, w, r)
            return
        }
        CreateQSO(db, hub, w, r)
    }))
    defer srv.Close()
    // ...
}
```

**Concurrent goroutine pattern** — `internal/handler/ws_test.go` (lines 144-173):
```go
func TestCreateQSOConcurrent(t *testing.T) {
    db := setupHandlerTestDB(t)
    hub := ws.NewHub()
    go hub.Run()
    time.Sleep(10 * time.Millisecond)

    errChan := make(chan error, 3)

    for i := 0; i < 3; i++ {
        go func(callsign string) {
            body := `{"callsign":"` + callsign + `","band":"20M","mode":"SSB","recv_exchange":"2A NH"}`
            req := httptest.NewRequest("POST", "/api/qso", strings.NewReader(body))
            req.Header.Set("Content-Type", "application/json")
            rec := httptest.NewRecorder()

            CreateQSO(db, hub, rec, req)
            if rec.Code != http.StatusCreated {
                errChan <- fmt.Errorf("expected 201, got %d", rec.Code)
                return
            }
            errChan <- nil
        }(string(rune('A' + i)) + "1ZZZ")
    }

    for i := 0; i < 3; i++ {
        if err := <-errChan; err != nil {
            t.Errorf("concurrent QSO insert failed: %v", err)
        }
    }
}
```

**Simulation test structure** (package `internal/handler/simtest`):

```go
package simtest

import (
    // stdlib testing, httptest, database/sql, etc.
    "github.com/go-chi/chi/v5"
    "github.com/gorilla/websocket"
    "github.com/jeremy/mlogger-fd/internal/handler"
    "github.com/jeremy/mlogger-fd/internal/ws"
    _ "modernc.org/sqlite"
)

// setupSimRouter builds a chi router with all production routes wired
func setupSimRouter(db *sql.DB, hub *ws.Hub) chi.Router {
    r := chi.NewRouter()
    r.Route("/api", func(r chi.Router) {
        r.Get("/check-dupe", func(w http.ResponseWriter, req *http.Request) {
            handler.CheckDupeHandler(db, w, req)
        })
        r.Get("/stats", func(w http.ResponseWriter, req *http.Request) {
            handler.GetStats(db, w, req)
        })
        r.Post("/qso", func(w http.ResponseWriter, req *http.Request) {
            handler.CreateQSO(db, hub, w, req)
        })
        // ... all routes ...
    })
    r.Get("/ws", func(w http.ResponseWriter, req *http.Request) {
        handler.ServeWS(hub, w, req)
    })
    return r
}

func TestSimulation(t *testing.T) {
    // 1. Setup in-memory SQLite with full schema (qsos + station_config + bonus_claims)
    // 2. Start httptest.Server with chi router
    // 3. Start ws.Hub
    // 4. Spawn 3 goroutine clients, each:
    //    - Connects WebSocket
    //    - Submits ~70 mixed QSOs (unique + deliberate duplicates)
    //    - Tracks received WS broadcasts
    // 5. Run until all clients finish (use sync.WaitGroup)
    // 6. Assert integrity:
    //    - GET /api/stats: total matches non-dupe count
    //    - GET /api/qso?limit=9999: total rows match submitted
    //    - Every submitted non-dupe QSO present in list
    //    - Dupe QSOs have is_dupe=true, points=0
    //    - All clients received = total broadcast count
    //    - Score correct: (rawPoints * multiplier) + bonusPoints
}
```

**Adaptation notes:**
- Package: `internal/handler/simtest` (separate package from `handler` to avoid circular imports; use `handler.` prefix for handler functions)
- Go file naming: `simtest_test.go` per RESEARCH.md convention (test-only package)
- DB setup: use `_1pragma=journal_mode(WAL)&cache=shared` pattern from `qso_test.go` line 16; also create `station_config` and `bonus_claims` tables
- `httptest.NewServer` pattern from `ws_test.go` — NOT `httptest.NewRequest` (need real HTTP for WebSocket upgrade)
- Use `sync.WaitGroup` for goroutine coordination (same as `errChan` pattern but for completion tracking)
- WebSocket client: use `gorilla/websocket.DefaultDialer.Dial()` as in `ws_test.go` line 27
- Assertions: use `testing` stdlib (table-driven or inline) — no testify dependency per RESEARCH.md
- **Do NOT re-implement handler logic** — the simulation routes to real handlers via chi router wired with `handler.CreateQSO(db, hub, w, r)` etc.

---

### 14. `frontend/src/lib/components/StatsBar.svelte` — add bonus points display

**Role:** component (data display)
**Data flow:** reactive `stats` object → rendered DOM
**Closest analog:** `frontend/src/lib/components/StatsBar.svelte` (same file — modify existing)

**Current stat blocks** — `StatsBar.svelte` (lines 20-44):
```svelte
<div class="stats-bar">
    <div class="stat rate">
        <span class="stat-label">Rate</span>
        <span class="stat-value">{stats.rate_10min}</span>
        <span class="stat-unit">/hr</span>
    </div>
    <div class="stat">
        <span class="stat-label">QSOs</span>
        <span class="stat-value">{stats.total}</span>
    </div>
    <div class="stat">
        <span class="stat-label">Pts</span>
        <span class="stat-value">{stats.raw_points}</span>
    </div>
    <div class="stat">
        <span class="stat-label">Mult</span>
        <span class="stat-value">{stats.multiplier}</span>
    </div>
    <div class="stat score">
        <span class="stat-label">Score</span>
        <span class="stat-value">{stats.score}</span>
    </div>
    <!-- ... -->
</div>
```

**Stat block CSS** — `StatsBar.svelte` (lines 86-117):
```css
.stat {
    display: flex;
    gap: 4px;
    align-items: baseline;
}

.stat-label {
    color: var(--color-text-secondary);
    font-size: 11px;
    text-transform: uppercase;
    font-weight: 600;
}

.stat-value {
    font-size: 20px;
    font-weight: 700;
    color: var(--color-primary);
}

.stat-unit {
    font-size: 11px;
    color: var(--color-text-secondary);
}
```

**Addition:** Insert a new bonus stat block between "Pts" and "Mult" (logical flow: raw QSO points → bonus points → multiplier → total score):
```svelte
<div class="stat bonus">
    <span class="stat-label">Bonus</span>
    <span class="stat-value">{stats.bonus_points || 0}</span>
</div>
```

**Add CSS for bonus color:**
```css
.bonus .stat-value {
    color: #cc8800; /* gold/amber to distinguish from raw points and score */
}
/* In dark mode, use a brighter gold */
:global([data-theme='dark']) .bonus .stat-value {
    color: #ffaa00;
}
```

**Adaptation notes:**
- `stats.bonus_points` — new field added to the `$state` object in `qso.svelte.js` (line 3-11, add `bonus_points: 0` default) and returned from server `stats.go`
- The `Object.assign(stats, data)` in `fetchStats()` (line 69 of `qso.svelte.js`) will automatically populate `bonus_points` when the server returns it — no store change needed beyond the default initialization
- Reuse identical stat block structure: `.stat > .stat-label + .stat-value`
- Score stat already exists and will naturally show the updated total

---

## Shared Patterns

### Authentication
**Source:** N/A — no auth in this project (trusted LAN per PROJECT.md constraints)
**Apply to:** No files

### Error Handling (Go Handlers)
**Source:** `internal/handler/config.go` (lines 20-23, 54-58), `internal/handler/qso.go` (lines 40-45)
**Apply to:** `internal/handler/bonus.go`, `internal/handler/backup.go`, `internal/handler/stats.go`, `internal/cabrillo/cabrillo.go`

**Pattern:**
```go
// JSON error response for 500
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusInternalServerError)
json.NewEncoder(w).Encode(map[string]string{"error": "database error"})

// JSON error response for 400
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusBadRequest)
json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})

// Safe fallback for bonus query failures
if err != nil {
    bonusPoints = 0  // don't block scoring if bonus table missing
}
```

### Svelte 5 $state Object Export Pattern
**Source:** `frontend/src/lib/ws.svelte.js` (line 7), `frontend/src/lib/sync.svelte.js` (line 5)
**Apply to:** `audio.svelte.js` (`audioState`), `qso.svelte.js` (`bonusClaims`)

**Pattern:**
```javascript
export const audioState = $state({ muted: false });
// DO mutate: audioState.muted = true
// DO NOT reassign: audioState = { muted: true }
```

### localStorage Persistence
**Source:** `OperatorSelector.svelte` (lines 4-8), `+page.svelte` (lines 15-17)
**Apply to:** `BonusTracker.svelte`, `audio.svelte.js`, mute toggle in `+page.svelte`

**Pattern:**
```javascript
// Read on init
let value = $state(localStorage.getItem('fdlogger_key') || defaultValue);

// Save on change
localStorage.setItem('fdlogger_key', value);
```

### Chi Router Closure Wiring
**Source:** `main.go` (lines 43-73)
**Apply to:** `main.go` (modification — add new routes)

**Pattern:**
```go
r.Route("/api", func(r chi.Router) {
    r.Get("/bonuses", func(w http.ResponseWriter, r *http.Request) {
        handler.GetBonuses(database, w, r)
    })
    r.Put("/bonuses", func(w http.ResponseWriter, r *http.Request) {
        handler.PutBonuses(database, w, r)
    })
})
```

### Expandable Header Panel (Svelte)
**Source:** `StationConfig.svelte` (entire file)
**Apply to:** `BonusTracker.svelte`

**Template structure:**
```svelte
<button class="{prefix}-toggle" onclick={toggle}>★ Label</button>
{#if expanded}
    <div class="{prefix}-panel">  <!-- position: absolute; top: 44px; right: 8px; z-index: 100 -->
        <form onsubmit={handleSubmit}>
            <!-- form fields -->
            <div class="{prefix}-actions">
                <button type="submit" class="save-btn">Save</button>
                {#if saved}<span class="saved-msg">Saved!</span>{/if}
            </div>
        </form>
    </div>
{/if}
```

### Test DB Setup (Go)
**Source:** `internal/handler/qso_test.go` (lines 14-39)
**Apply to:** `internal/handler/simtest/simtest_test.go`

**Pattern:**
```go
func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()
    db, err := sql.Open("sqlite", "file::memory:?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&cache=shared")
    if err != nil {
        t.Fatalf("failed to open test DB: %v", err)
    }
    db.SetMaxOpenConns(1)
    // Run CREATE TABLE IF NOT EXISTS statements for all needed tables
    t.Cleanup(func() { db.Close() })
    return db
}
```

### File Download (Go)
**Source:** `internal/handler/export.go` (lines 12-29)
**Apply to:** `internal/handler/backup.go`

**Pattern:**
```go
w.Header().Set("Content-Type", "application/octet-stream")
w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
// Stream content to w
```

---

## No Analog Found

All 14 files have close analogs in the existing codebase. No file lacks a pattern reference.

The only pattern adaptation needed is:
- **`internal/model/bonus.go`** uses a constant slice of structs (`DefaultBonuses`) which has no exact analog in `model/config.go` (which has a single struct). The pattern follows standard Go conventions.
- **`internal/handler/simtest/simtest_test.go`** is larger in scope than `ws_test.go` but follows the same fundamental patterns (httptest.Server, gorilla/websocket, goroutine-based concurrency).
- **`internal/handler/backup.go`** uses `io.Copy` instead of `w.Write` — the same file-download response header pattern from `export.go` applies.

## Metadata

**Analog search scope:**
- `internal/handler/` — 13 files read (config.go, export.go, stats.go, qso.go, ws_test.go, stats_test.go, qso_test.go)
- `internal/cabrillo/` — 2 files read (cabrillo.go, cabrillo_test.go)
- `internal/model/` — 3 files read (config.go, qso.go, config_test.go)
- `internal/db/` — 2 files read (db.go, schema.sql)
- `frontend/src/routes/` — 1 file read (+page.svelte)
- `frontend/src/lib/components/` — 4 files read (StationConfig.svelte, QsoEntryForm.svelte, OperatorSelector.svelte, StatsBar.svelte)
- `frontend/src/lib/` — 3 files read (api.js, ws.svelte.js, sync.svelte.js)
- `frontend/src/lib/stores/` — 1 file read (qso.svelte.js)
- Root: `main.go`, `go.mod`, `frontend/src/app.css`

**Files scanned:** 33 (all key files + associated tests/config)
**Pattern extraction date:** 2026-06-04
**Module path:** `github.com/jeremy/mlogger-fd`
