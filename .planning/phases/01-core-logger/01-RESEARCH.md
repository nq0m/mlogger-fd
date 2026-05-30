# Phase 01: Core Logger - Research

**Researched:** 2026-05-29
**Domain:** Go backend + SvelteKit SPA + SQLite — single-user Field Day contest logging
**Confidence:** HIGH

## Summary

Phase 1 delivers a single-user QSO logger as a self-contained Go binary embedding a SvelteKit SPA. The Go backend uses `modernc.org/sqlite` (CGo-free SQLite driver) with WAL mode for the database, `chi` for HTTP routing (REST API at `/api/*`, SPA fallback at `/*`), and `embed.FS` to serve the SvelteKit static build. The SvelteKit frontend is configured as an SPA with `adapter-static` and a `fallback` page (typically `200.html` or `index.html`), implementing three panels (entry form, stats bar, log table) using Svelte 5 runes (`$state`) for reactive state management.

The database schema follows the planning document: `qsos` table with columns for callsign, band, mode, exchange, timestamp, points, and is_dupe flag. Indexes on `callsign`, `timestamp`, and `(band, mode)` support efficient dupe checking and log table queries. Points calculation is server-side on insert: CW/digital modes = 2 pts, phone modes = 1 pt, dupes = 0 pts.

The REST API provides QSO CRUD, dupe checking, stats aggregation, and Cabrillo export. The frontend communicates with the backend via `fetch()` calls to `/api/*` endpoints. Client-side state is managed in Svelte module-level `$state` runes (safe in SPA mode with SSR disabled).

**Primary recommendation:** Scaffold the Go project first with `chi` router, `modernc.org/sqlite` driver, and `embed.FS`; establish the SvelteKit SPA project with `adapter-static` and SPA fallback; implement the database schema and `/api/qsos` POST endpoint as the walking skeleton connecting all three layers.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| QSO persistence (CRUD) | API / Backend | — | SQLite operations are server-side; Go owns the database |
| Dupe detection | API / Backend | Browser / Client | Server-side SQL query is authoritative; client caches for blur-triggered feedback |
| Points calculation | API / Backend | — | Business logic runs server-side on insert to prevent client manipulation |
| Rate/score computation | API / Backend | Browser / Client | Server computes stats from database; client can derive from local QSO list |
| Cabrillo export | API / Backend | — | Generated server-side from database rows, served as downloadable file |
| QSO entry form | Browser / Client | — | Purely client-side UI with keyboard navigation |
| Stats display | Browser / Client | — | Derived from API responses; rendered client-side |
| SPA static serving | API / Backend | — | Go's `embed.FS` serves pre-built static files |
| Log table with pagination | Browser / Client | API / Backend | Client renders; server provides offset-based paginated data |

## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01:** QSO form auto-clears all fields on successful submit and returns focus to the callsign field — optimized for rapid sequential entry during contest conditions.
- **D-02:** Dupe check fires on callsign field blur AND on form submit. Blur-triggered check catches most dupes before the operator finishes filling the form. Submit-triggered check is the final guard.
- **D-03:** Callsign validation is lenient — warn on empty or single-character input, but accept anything that looks like a callsign. DX stations have diverse formats; strict FCC/ITU validation would block valid entries. Submit is always allowed even with a validation warning.
- **D-04:** Single-page three-panel layout: QSO entry form at top (always accessible), stats bar in the middle (rate, score, band/mode counts — always visible), scrollable log table at the bottom. This works on both desktop and mobile without tab switching. No separate tabs or pages for logging vs viewing.
- **D-05:** QSO editing is inline — click a row in the log table to switch it to edit mode, save or cancel in place. No modal or separate edit page.
- **D-06:** Cabrillo export is a one-click button with no preview. Generates and downloads the file immediately. Operators export once before the ARRL submission deadline; verification happens by opening the downloaded file.
- **D-07:** Single Go binary embeds the SvelteKit static build via `embed.FS` and serves both the REST API (`/api/*`) and the SPA (`/*` falling back to `index.html`). No nginx, no separate static file server, no Docker — one binary, one systemd unit.

### the agent's Discretion

- Points calculation table (which modes are 1pt vs 2pt) — hardcoded as described in the planning doc: CW, RTTY, FT8, FT4, PSK31, MFSK, JT65, JT9, OLIVIA, DOMINO = 2pts; SSB, FM, AM = 1pt.
- Keyboard shortcuts for the form (Tab order, Ctrl+Enter to submit) — standard web form behavior, no custom keybinding system.
- Rate meter time windows — last 10 minutes, last 1 hour, and overall session. Show current rate prominently, peak rate as secondary.
- Error handling — toast or inline notification for API errors. Network errors show "Could not reach server" with retry hint. Validation errors appear under the relevant field.
- Pagination for the log table — offset-based with 50 QSOs per page. Load more on scroll or page button.

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within Phase 1 scope. Following ideas are already in the roadmap backlog for later phases:
- Multi-user WebSocket sync (Phase 2)
- Offline IndexedDB buffer (Phase 3)
- Mobile-responsive layout (Phase 3)
- Dark mode (Phase 3)
- Bonus points tracker (Phase 4)
- Audio alerts (Phase 4)

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| QSO-01 | Operator can log a QSO by entering callsign, band, mode, received exchange using a form with Tab/Enter keyboard navigation | Form pattern (Pattern 1), keyboard shortcuts (Pattern 3), POST /api/qso endpoint |
| QSO-02 | QSO form validates callsign format and required fields before submission | Lenient validation per D-03, client-side warning before submit |
| QSO-03 | Operator can search and edit previously logged QSOs | Inline edit pattern (D-05), PUT /api/qso/:id, GET /api/qso with search params |
| QSO-04 | Keyboard shortcuts (Ctrl+Enter to submit, Tab between fields) supported for rapid entry | Standard HTML form Tab order, Ctrl+Enter keydown handler |
| DUPE-01 | Dupe check warns in real-time if callsign already worked on same band AND mode before submission | Dupe detection algorithm (Pattern 4), GET /api/check-dupe endpoint |
| DUPE-02 | Partial call similarity warning when entered call is similar to a previously logged call | LIKE-based similarity query, client-side Levenshtein or prefix matching |
| DUPE-03 | Dupe QSOs are logged but marked as duplicate with zero points | is_dupe=1, points=0 set server-side on POST |
| SCOR-01 | Live rate meter displays QSOs per hour, peak rate, and running total | GET /api/stats with time-windowed COUNT queries |
| SCOR-02 | Live score display shows raw points, multiplier, bonus points, and estimated total | GET /api/stats aggregation; Phase 1 multiplier=1 (no power config yet) |
| SCOR-03 | Band/mode breakdown panel shows QSO count per band+mode combination | GROUP BY band, mode query in /api/stats |
| EXPR-01 | One-click Cabrillo export generates valid ARRL Field Day format with all QSOs and station info | Cabrillo format specification (Appendix A of planning doc), GET /api/export/cabrillo |
| EXPR-02 | Cabrillo export includes bonus points claimed and correct header metadata | Phase 1: bonus=0, station info from hardcoded defaults or minimal config |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go | 1.22+ | Backend runtime | Single binary, low memory (~5 MB), perfect for RPi deployment |
| `modernc.org/sqlite` | v1.51.0 | SQLite driver (CGo-free) | Pure Go, no C compiler needed, cross-compilation for ARM (RPi) works out of the box |
| `github.com/go-chi/chi/v5` | v5.3.0 | HTTP router | Lightweight (~1000 LOC), stdlib-compatible, URL params, middleware support |
| `embed` (stdlib) | go1.26.3 | Embed SPA static files | Standard library, no external dependency, read-only FS interface |
| `net/http` (stdlib) | go1.26.3 | HTTP server + file serving | Standard library; chi wraps it; `http.FileServer` serves embedded content |
| SvelteKit | 2.61.1 | Frontend framework | Small bundles (~8 KB gzip), reactive, excellent for SPAs |
| Svelte | 5.56.0 | UI components | Runes-based reactivity (`$state`, `$derived`, `$effect`) |
| Vite | 8.0.14 | Build tooling | Fast HMR, production bundling, SvelteKit's build system |
| `@sveltejs/adapter-static` | 3.0.10 | Static SPA build | Generates static files for Go's `embed.FS`; fallback page for SPA routing |
| SQLite | 3.53.1 | Database | Single-file, WAL mode, zero config, built into `modernc.org/sqlite` |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `encoding/json` (stdlib) | go1.26.3 | JSON request/response parsing | All API endpoints |
| `database/sql` (stdlib) | go1.26.3 | Database interface | All database operations via `modernc.org/sqlite` driver |
| `log/slog` (stdlib) | go1.26.3 | Structured logging | Server-side request logging, error reporting |
| `time` (stdlib) | go1.26.3 | Timestamp handling | QSO timestamps (ISO 8601 UTC), rate window calculations |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `modernc.org/sqlite` | `github.com/mattn/go-sqlite3` | mattn driver needs CGo (C compiler + cross-compilation pain for ARM); modernc is pure Go |
| `chi` | `net/http` ServeMux (Go 1.22+) | Go 1.22+ ServeMux supports method routing and path params; chi adds middleware composition and sub-routers; either works but chi is more expressive for REST APIs |
| `adapter-static` SPA fallback | `adapter-node` | Node adapter requires Node runtime on server; static adapter produces embeddable files |
| SvelteKit SPA | Preact + Vite | SvelteKit provides routing, layout system, build tooling out of the box; Preact is lighter but requires manual setup |

**Installation:**
```bash
# Go backend (initialize go module in project)
go mod init github.com/example/fdlogger
go get modernc.org/sqlite@v1.51.0
go get github.com/go-chi/chi/v5@v5.3.0

# Frontend (initialize SvelteKit project)
npm create svelte@latest frontend
cd frontend
npm install
npm install -D @sveltejs/adapter-static@3.0.10
```

**Version verification:**
```bash
npm view @sveltejs/kit version          # 2.61.1
npm view @sveltejs/adapter-static version  # 3.0.10
npm view svelte version                 # 5.56.0
npm view vite version                   # 8.0.14
```
Go packages verified via: [VERIFIED: pkg.go.dev]

## Package Legitimacy Audit

> slopcheck unavailable — all packages tagged `[ASSUMED]`. Planner must add `checkpoint:human-verify` before each install.

| Package | Registry | Age | Downloads | Source Repo | slopcheck | Disposition |
|---------|----------|-----|-----------|-------------|-----------|-------------|
| `@sveltejs/kit` | npm | 3+ yrs | 500K+/wk | github.com/sveltejs/kit | [ASSUMED] | Approved (well-known) |
| `@sveltejs/adapter-static` | npm | 3+ yrs | 200K+/wk | github.com/sveltejs/kit | [ASSUMED] | Approved (well-known) |
| `svelte` | npm | 8+ yrs | 1M+/wk | github.com/sveltejs/svelte | [ASSUMED] | Approved (well-known) |
| `vite` | npm | 5+ yrs | 10M+/wk | github.com/vitejs/vite | [ASSUMED] | Approved (well-known) |
| `modernc.org/sqlite` | Go mod | 5+ yrs | — | gitlab.com/cznic/sqlite | [ASSUMED] | Approved (well-known) |
| `github.com/go-chi/chi/v5` | Go mod | 8+ yrs | — | github.com/go-chi/chi | [ASSUMED] | Approved (well-known) |

**Packages removed due to slopcheck [SLOP] verdict:** none (slopcheck unavailable)
**Packages flagged as suspicious [SUS]:** none

*slopcheck was unavailable at research time — all packages above are tagged `[ASSUMED]` and the planner must gate each install behind a `checkpoint:human-verify` task.*

## Architecture Patterns

### System Architecture Diagram

```
┌──────────────────────────────────────────────────────────────┐
│                    Browser (Client)                          │
│                                                              │
│  ┌─────────────────┐  ┌──────────────┐  ┌────────────────┐  │
│  │  QSO Entry Form  │  │  Stats Bar   │  │  Log Table     │  │
│  │  (callsign,      │  │  (rate,      │  │  (paginated,   │  │
│  │   band, mode,    │  │   score,     │  │   inline edit, │  │
│  │   exchange)      │  │   breakdown) │  │   search)      │  │
│  └────────┬─────────┘  └──────┬───────┘  └───────┬────────┘  │
│           │                   │                   │           │
│           └───────────────────┼───────────────────┘           │
│                               │                               │
│                    ┌──────────┴──────────┐                    │
│                    │   Svelte $state      │                    │
│                    │   (qsos[], stats)    │                    │
│                    └──────────┬──────────┘                    │
│                               │                               │
│                    ┌──────────┴──────────┐                    │
│                    │   API Client Module  │                    │
│                    │   (fetch /api/*)     │                    │
│                    └──────────┬──────────┘                    │
└───────────────────────────────┼──────────────────────────────┘
                                │ HTTP (LAN, same origin)
                                │
┌───────────────────────────────┼──────────────────────────────┐
│                    Go Binary (Server)                         │
│                               │                               │
│  ┌────────────────────────────┴───────────────────────────┐  │
│  │                    chi Router                           │  │
│  │                                                        │  │
│  │  /api/* ──────────► Handler functions                  │  │
│  │  /* (fallback) ───► embed.FS static file server        │  │
│  └──┬─────────────────────────────────────────────────────┘  │
│     │                                                         │
│  ┌──┴──────────┐  ┌──────────────┐  ┌────────────────────┐   │
│  │  Handlers   │  │  Points Calc │  │  Cabrillo Gen      │   │
│  │  (CRUD,     │  │  (mode→pts   │  │  (fixed-width       │   │
│  │   dupe,     │  │   mapping)   │  │   text format)     │   │
│  │   stats)    │  │              │  │                    │   │
│  └──┬──────────┘  └──────┬───────┘  └────────┬───────────┘   │
│     │                    │                    │               │
│  ┌──┴────────────────────┴────────────────────┴──────────┐   │
│  │              modernc.org/sqlite                        │   │
│  │              (CGo-free SQLite driver)                   │   │
│  └────────────────────────┬──────────────────────────────┘   │
│                           │                                   │
│                    ┌──────┴──────┐                            │
│                    │  SQLite DB  │                            │
│                    │  (WAL mode) │                            │
│                    │  fdlogger.db│                            │
│                    └─────────────┘                            │
└──────────────────────────────────────────────────────────────┘
```

### Recommended Project Structure
```
fdlogger/
├── go.mod                    # Go module definition
├── go.sum
├── main.go                   # Entry point: starts HTTP server
├── internal/
│   ├── db/
│   │   ├── db.go             # Database connection, WAL setup, migrations
│   │   └── schema.sql        # CREATE TABLE statements
│   ├── handler/
│   │   ├── qso.go            # QSO CRUD handlers (POST/GET/PUT/DELETE /api/qso)
│   │   ├── dupe.go           # Dupe check handler (GET /api/check-dupe)
│   │   ├── stats.go          # Stats handler (GET /api/stats)
│   │   ├── export.go         # Cabrillo export handler (GET /api/export/cabrillo)
│   │   └── health.go         # Health check (GET /api/health)
│   ├── model/
│   │   └── qso.go            # QSO struct, JSON tags, validation
│   ├── qso/
│   │   ├── points.go         # Points calculation logic
│   │   └── dupe.go           # Dupe detection queries
│   └── cabrillo/
│       └── cabrillo.go       # Cabrillo format generation
├── frontend/                 # SvelteKit project (separate npm project)
│   ├── package.json
│   ├── svelte.config.js      # adapter-static with fallback
│   ├── vite.config.ts
│   ├── src/
│   │   ├── app.html          # SPA shell template
│   │   ├── app.css           # Global styles
│   │   ├── routes/
│   │   │   ├── +layout.svelte  # Root layout (ssr=false, three-panel)
│   │   │   └── +page.svelte    # Main page (composes panels)
│   │   ├── lib/
│   │   │   ├── components/
│   │   │   │   ├── QsoEntryForm.svelte   # Entry form
│   │   │   │   ├── StatsBar.svelte       # Rate + score + breakdown
│   │   │   │   └── LogTable.svelte       # Paginated log with inline edit
│   │   │   ├── stores/
│   │   │   │   └── qso.svelte.js         # Shared QSO state ($state)
│   │   │   └── api.js                    # API client module (fetch wrappers)
│   │   └── static/                        # Static assets (favicon, etc.)
│   └── build/                             # Output directory (embedded by Go)
└── Makefile                  # Build: frontend → Go binary
```

### Pattern 1: Go REST API with chi Router

**What:** chi provides idiomatic REST routing with method-based handlers, URL parameters, middleware chains, and sub-router mounting.

**When to use:** All API endpoint definitions. Use `r.Route("/api", ...)` to group all API routes under the `/api` prefix, and `r.Get("/*", ...)` for the SPA fallback.

**Example:**
```go
// Source: [CITED: pkg.go.dev/github.com/go-chi/chi/v5]
package main

import (
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", healthHandler)
		r.Route("/qso", func(r chi.Router) {
			r.Post("/", createQSO)
			r.Get("/", listQSOs)
			r.Put("/{id}", updateQSO)
			r.Delete("/{id}", deleteQSO)
		})
		r.Get("/check-dupe", checkDupe)
		r.Get("/stats", getStats)
		r.Get("/export/cabrillo", exportCabrillo)
	})

	// SPA fallback — serve embedded static files
	// See Pattern 2 for the SPA serving approach

	http.ListenAndServe(":8080", r)
}
```

### Pattern 2: Embed SvelteKit SPA in Go Binary

**What:** Go's `embed.FS` embeds the SvelteKit build output at compile time. The Go HTTP server serves static files from the embedded FS and falls back to `index.html` for client-side routing.

**When to use:** Single-binary deployment (D-07). The SvelteKit SPA build goes to `frontend/build/`. Go embeds that directory and serves it for all non-API routes.

**Example:**
```go
// Source: [CITED: pkg.go.dev/embed]
package main

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed frontend/build/*
var staticFiles embed.FS

func spaHandler() http.Handler {
	// Strip the "frontend/build" prefix from embedded paths
	content, err := fs.Sub(staticFiles, "frontend/build")
	if err != nil {
		panic(err)
	}
	fs := http.FileServer(http.FS(content))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try serving the exact file
		path := r.URL.Path

		// If the file doesn't exist in the embedded FS, serve index.html
		// (SPA client-side routing fallback)
		f, err := content.Open(path)
		if err != nil {
			// Rewrite to index.html for SPA routing
			r.URL.Path = "/"
		} else {
			f.Close()
		}
		fs.ServeHTTP(w, r)
	})
}
```

**SvelteKit SPA configuration:**
```js
// Source: [CITED: kit.svelte.dev/docs/adapter-static + single-page-apps]
// frontend/svelte.config.js
import adapter from '@sveltejs/adapter-static';

export default {
	kit: {
		adapter: adapter({
			fallback: 'index.html',  // SPA fallback page
			pages: 'build',
			assets: 'build'
		})
	}
};
```

```svelte
<!-- frontend/src/routes/+layout.svelte -->
<script>
  export const ssr = false;  // Disable SSR for SPA mode
</script>

<slot />
```

### Pattern 3: SQLite WAL Mode Setup

**What:** Enable WAL mode on database open for concurrent read performance. Configure busy timeout and foreign keys.

**When to use:** Database initialization in `internal/db/db.go`. Call once on startup.

**Example:**
```go
// Source: [CITED: sqlite.org/wal.html + modernc.org/sqlite docs]
package db

import (
	"database/sql"
	_ "modernc.org/sqlite"
)

func Open(dsn string) (*sql.DB, error) {
	// Use URI format to set pragmas at connection time
	// _pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)
	dsnWithPragmas := dsn + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)"

	db, err := sql.Open("sqlite", dsnWithPragmas)
	if err != nil {
		return nil, err
	}

	// Verify WAL mode is active
	var mode string
	db.QueryRow("PRAGMA journal_mode").Scan(&mode)
	// mode should be "wal"

	db.SetMaxOpenConns(1) // SQLite serializes writes; single writer is safe
	db.SetMaxIdleConns(1)

	return db, nil
}
```

> **Critical:** SQLite in WAL mode with `modernc.org/sqlite` has a known WAL-reset bug fixed in SQLite 3.51.3. `modernc.org/sqlite` v1.51.0 ships SQLite 3.53.1 which includes the fix. [VERIFIED: pkg.go.dev/modernc.org/sqlite + sqlite.org/wal.html §11]

### Pattern 4: Dupe Detection Algorithm

**What:** Check if a callsign exists on the same band AND mode. Query the database before insert. Return dupe status with optional partial match warning.

**When to use:** On callsign blur (D-02) and on form submit (D-02). Two queries: exact dupe check and partial call similarity check.

**Example:**
```go
// Source: [CITED: field-day-logger-plan.md §2.7 + §4.4]
func CheckDupe(db *sql.DB, callsign, band, mode string) (isDupe bool, similarCalls []string, err error) {
	// Exact dupe: same callsign + band + mode
	err = db.QueryRow(
		"SELECT COUNT(*) FROM qsos WHERE callsign = ? AND band = ? AND mode = ? AND is_dupe = 0",
		callsign, band, mode,
	).Scan(&count)
	if count > 0 {
		isDupe = true
	}

	// Partial call similarity: callsigns that share a prefix
	// (e.g., "K1X" matches "K1XX")
	rows, err := db.Query(
		"SELECT DISTINCT callsign FROM qsos WHERE callsign != ? AND (callsign LIKE ? OR ? LIKE callsign || '%') LIMIT 5",
		callsign, callsign+"%", callsign,
	)
	// ... collect similar calls
	return
}
```

### Pattern 5: Svelte 5 Client-Side State (SPA Mode)

**What:** Use Svelte 5 `$state` runes in a shared `.svelte.js` module for client-side state. Since SSR is disabled (SPA mode), module-level state is safe — no shared-server-state risks.

**When to use:** Managing the QSO list, stats, form state, and pagination in the client.

**Example:**
```svelte
<!-- Source: [CITED: svelte.dev/docs/svelte/$state + kit.svelte.dev/docs/state-management] -->
<!-- frontend/src/lib/stores/qso.svelte.js -->
<script>
  // Module-level $state — safe in SPA mode (ssr=false)
  export const qsos = $state([]);
  export const stats = $state({
    total: 0,
    rate: 0,
    peakRate: 0,
    score: 0,
    rawPoints: 0,
    multiplier: 1,
    breakdown: {}  // { "20M_CW": 5, "40M_SSB": 3, ... }
  });
  export const currentPage = $state(0);
  export const hasMore = $state(true);
</script>
```

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| HTTP routing | Custom ServeMux with manual path parsing | `chi` router | URL params, middleware composition, route groups — 1000 LOC of well-tested code |
| SQLite driver | Custom CGo bindings or raw C library calls | `modernc.org/sqlite` | Pure Go, no C compiler, cross-compiles to ARM (RPi), 3500+ importers |
| Embedded static files | Custom file-loading or external asset server | `embed.FS` (stdlib) | Compile-time embedding, read-only, stdlib-supported, zero dependencies |
| JSON serialization | Manual string building | `encoding/json` (stdlib) | Type-safe, battle-tested, handles edge cases |
| SPA routing fallback | Custom URL rewriting logic | `http.FileServer` + fallback to `index.html` | Standard pattern for SPAs; well-understood |
| Points calculation | Hardcoded in SQL or client-side | Go function with mode→points map | Single source of truth, testable, prevents client manipulation |
| Rate window calculations | Client-side Date math | Server-side SQL with time-range COUNT | Accurate regardless of client clock skew; server is source of truth |
| Cabrillo formatting | Template engine | `text/template` (stdlib) or `fmt.Sprintf` with fixed-width fields | Cabrillo is fixed-width text; templates overcomplicate. Use direct string building. |
| Client-side API calls | Raw `fetch` with error handling scattered everywhere | Single `api.js` module wrapping `fetch` with error handling, JSON parsing, and base URL | DRY, consistent error handling, easy to add auth/retry later |

**Key insight:** The Go standard library already provides `embed`, `net/http`, `encoding/json`, and `database/sql`. Only three external dependencies are needed: `chi` (routing ergonomics), `modernc.org/sqlite` (CGo-free SQLite), and the SvelteKit frontend stack. Resist the urge to add more — each dependency is a future maintenance burden on a Raspberry Pi in a tent.

## Runtime State Inventory

> Phase 1 is greenfield — no existing runtime state exists. This section is included for completeness.

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | None — greenfield project, no existing SQLite database | Create new database with schema migration |
| Live service config | None — no services configured | N/A |
| OS-registered state | None — no systemd units exist yet | Create systemd unit during deployment |
| Secrets/env vars | None — no secrets configured | May add `FDLOGGER_DB_PATH` or `FDLOGGER_PORT` env vars for configuration |
| Build artifacts | None — greenfield project | Build process creates: Go binary, `frontend/build/` directory |

**Nothing found in any category.** All state will be created fresh by this phase.

## Common Pitfalls

### Pitfall 1: CGo Cross-Compilation with mattn/go-sqlite3
**What goes wrong:** Using `mattn/go-sqlite3` requires CGo and a C cross-compiler for ARM (Raspberry Pi). Builds fail with cryptic errors when cross-compiling from x86_64 to ARM.
**Why it happens:** mattn's driver wraps the C SQLite library. Cross-compilation needs `CC=arm-linux-gnueabihf-gcc` and the correct C toolchain.
**How to avoid:** Use `modernc.org/sqlite` which is a pure Go SQLite implementation. Cross-compilation works with `GOOS=linux GOARCH=arm64 go build` — no C toolchain needed.
**Warning signs:** Build error mentioning "cgo", "exec: gcc: executable file not found", or "SQLITE_ENABLE_STAT4".

### Pitfall 2: SvelteKit SSR Leaking State Between Requests
**What goes wrong:** If SSR is accidentally left enabled, module-level `$state` in `.svelte.js` files shares state across all server-side renders, causing data leaks between users.
**Why it happens:** SvelteKit's server-side rendering shares a single Node.js process. Module-level variables are shared across all requests.
**How to avoid:** Set `export const ssr = false` in the root `+layout.svelte` for SPA mode. This disables SSR entirely — the SPA runs only in the browser. Also set `fallback` in `adapter-static` config.
**Warning signs:** Seeing other operators' QSOs appear in your browser when only one person is logging. Check SSR setting first.

### Pitfall 3: SQLite WAL File Growth Without Checkpointing
**What goes wrong:** The WAL file grows unboundedly, consuming disk space and degrading read performance.
**Why it happens:** If a database connection holds a read transaction open too long, the WAL cannot be checkpointed (reset). New writes keep appending to the WAL file.
**How to avoid:** SQLite's automatic checkpointing (at 1000 pages ≈ 4MB) handles normal usage. For Field Day (single writer, occasional readers), this is sufficient. Explicitly close result rows (`rows.Close()`) and use short-lived transactions. Do not hold read transactions across HTTP requests.
**Warning signs:** `*-wal` file growing beyond 10MB. Check with `ls -lh fdlogger.db-wal`.

### Pitfall 4: Client-Side Points Calculation Drift
**What goes wrong:** The frontend computes points client-side but the server computes them differently, causing the displayed score to differ from the exported Cabrillo total.
**Why it happens:** Two implementations of the points logic that diverge.
**How to avoid:** Points calculation is server-side only. The `POST /api/qso` response includes the computed points. The frontend displays what the server returns. The stats endpoint (`GET /api/stats`) also computes from the database. Single source of truth.
**Warning signs:** Score in UI doesn't match QSO point values in the log table.

### Pitfall 5: SPA Fallback Not Serving Assets Correctly
**What goes wrong:** After building with `adapter-static`, the Go server returns 404 for JS/CSS assets, or returns `index.html` for asset requests (MIME type mismatch).
**Why it happens:** The SPA fallback logic is too aggressive — it rewrites every 404 to `index.html`, including missing asset requests.
**How to avoid:** In the Go SPA handler, check if the requested file exists in the embedded FS before falling back to `index.html`. Only fall back for paths that don't match any file. See the `fs.Sub` + `content.Open` pattern in Pattern 2.
**Warning signs:** Browser console shows "Expected JavaScript module but got HTML" errors. MIME type errors for `.js` or `.css` files.

### Pitfall 6: Timezone Confusion in Timestamps
**What goes wrong:** QSO timestamps stored in local time instead of UTC, causing Cabrillo export to have incorrect times, or rate windows to compute incorrectly across timezone boundaries.
**Why it happens:** Using `datetime('now')` in SQLite returns local time by default. The browser's `Date` objects are in the user's local timezone.
**How to avoid:** Store all timestamps as ISO 8601 UTC strings. In Go, use `time.Now().UTC().Format(time.RFC3339)`. In SQLite, use `datetime('now')` which returns UTC when the connection is configured correctly, or compute timestamps in Go and pass them to SQL. The `_time_format=sqlite` DSN parameter configures time handling.
**Warning signs:** Cabrillo times don't match the operator's recollection. Rate windows show spikes at UTC day boundaries.

## Code Examples

### Go: Create QSO Handler (POST /api/qso)
```go
// Source: [CITED: field-day-logger-plan.md §4.4-4.5 + chi docs]
func createQSO(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Callsign     string `json:"callsign"`
		Band         string `json:"band"`
		Mode         string `json:"mode"`
		RecvExchange string `json:"recv_exchange"`
		SentExchange string `json:"sent_exchange"`
		Operator     string `json:"operator,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if input.Callsign == "" || input.Band == "" || input.Mode == "" {
		http.Error(w, `{"error":"callsign, band, and mode are required"}`, http.StatusBadRequest)
		return
	}

	db := getDB(r) // from context

	// Dupe check
	isDupe, similarCalls, _ := qso.CheckDupe(db, input.Callsign, input.Band, input.Mode)

	// Calculate points
	points := qso.CalculatePoints(input.Mode, isDupe)

	// Insert
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := db.Exec(
		`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, operator, is_dupe, points, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		now, input.Callsign, input.Band, input.Mode,
		input.SentExchange, input.RecvExchange, input.Operator,
		boolToInt(isDupe), points, now,
	)
	if err != nil {
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":            id,
		"is_dupe":       isDupe,
		"similar_calls": similarCalls,
		"points":        points,
		"timestamp":     now,
	})
}
```

### Go: Stats Handler (GET /api/stats)
```go
// Source: [CITED: field-day-logger-plan.md §4.5 + requirements SCOR-01–SCOR-03]
func getStats(w http.ResponseWriter, r *http.Request) {
	db := getDB(r)
	now := time.Now().UTC()

	// Total QSOs and raw points (excluding dupes)
	var total int
	var rawPoints int
	db.QueryRow("SELECT COUNT(*), COALESCE(SUM(points), 0) FROM qsos WHERE is_dupe = 0").Scan(&total, &rawPoints)

	// Rate: last 10 minutes
	tenMinAgo := now.Add(-10 * time.Minute).Format(time.RFC3339)
	var q10min int
	db.QueryRow("SELECT COUNT(*) FROM qsos WHERE timestamp >= ?", tenMinAgo).Scan(&q10min)
	rate10min := float64(q10min) / 10.0 * 60.0 // QSOs/hour

	// Rate: last 1 hour
	oneHourAgo := now.Add(-1 * time.Hour).Format(time.RFC3339)
	var q1hr int
	db.QueryRow("SELECT COUNT(*) FROM qsos WHERE timestamp >= ?", oneHourAgo).Scan(&q1hr)
	rate1hr := float64(q1hr) // QSOs/hour

	// Band/mode breakdown
	rows, _ := db.Query("SELECT band, mode, COUNT(*) FROM qsos WHERE is_dupe = 0 GROUP BY band, mode ORDER BY band, mode")
	breakdown := make(map[string]int)
	for rows.Next() {
		var band, mode string
		var count int
		rows.Scan(&band, &mode, &count)
		breakdown[band+"_"+mode] = count
	}
	rows.Close()

	// Multiplier: number of unique band+mode combinations worked
	var multiplier int
	db.QueryRow("SELECT COUNT(DISTINCT band || '_' || mode) FROM qsos WHERE is_dupe = 0").Scan(&multiplier)
	if multiplier < 1 {
		multiplier = 1
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"total":      total,
		"raw_points": rawPoints,
		"multiplier": multiplier,
		"score":      rawPoints * multiplier, // Phase 1: no power multiplier, no bonuses
		"rate_10min": math.Round(rate10min),
		"rate_1hr":   math.Round(rate1hr),
		"breakdown":  breakdown,
	})
}
```

### Go: Points Calculation
```go
// Source: [CITED: field-day-logger-plan.md §4.4 + §2.4]
package qso

// Two-point modes (CW and digital)
var twoPointModes = map[string]bool{
	"CW": true, "RTTY": true, "FT8": true, "FT4": true,
	"PSK31": true, "MFSK": true, "JT65": true, "JT9": true,
	"OLIVIA": true, "DOMINO": true,
}

// One-point modes (phone)
var onePointModes = map[string]bool{
	"SSB": true, "FM": true, "AM": true,
}

func CalculatePoints(mode string, isDupe bool) int {
	if isDupe {
		return 0
	}
	mode = strings.ToUpper(mode)
	if twoPointModes[mode] {
		return 2
	}
	return 1 // Default to 1 point for unknown modes (phone default)
}
```

### Svelte 5: QSO Entry Form with Keyboard Shortcuts
```svelte
<!-- Source: [CITED: svelte.dev/docs/svelte/$state + CONTEXT.md D-01–D-03] -->
<script>
  let callsign = $state('');
  let band = $state('20M');
  let mode = $state('SSB');
  let recvExchange = $state('');
  let dupeWarning = $state('');
  let submitting = $state(false);

  const bands = ['160M','80M','40M','20M','15M','10M','6M','2M','70CM'];
  const modes = ['CW','SSB','FM','RTTY','FT8','FT4','PSK31'];

  let callsignInput;

  async function checkDupe() {
    if (callsign.length < 2) return;
    const res = await fetch(`/api/check-dupe?callsign=${callsign}&band=${band}&mode=${mode}`);
    const data = await res.json();
    dupeWarning = data.is_dupe ? 'DUPE: Already worked on this band/mode' : '';
  }

  async function submit() {
    if (submitting) return;
    submitting = true;
    const res = await fetch('/api/qso', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ callsign, band, mode, recv_exchange: recvExchange })
    });
    const data = await res.json();
    if (data.is_dupe) {
      dupeWarning = 'Logged as duplicate (0 points)';
    }
    // D-01: auto-clear and refocus
    callsign = '';
    recvExchange = '';
    dupeWarning = '';
    callsignInput?.focus();
    submitting = false;
  }

  function handleKeydown(e) {
    // D-02: Ctrl+Enter to submit
    if (e.ctrlKey && e.key === 'Enter') {
      e.preventDefault();
      submit();
    }
  }
</script>

<form onsubmit={(e) => { e.preventDefault(); submit(); }} onkeydown={handleKeydown}>
  <input
    bind:this={callsignInput}
    bind:value={callsign}
    onblur={checkDupe}
    placeholder="Callsign"
    tabindex="1"
    autofocus
  />
  {#if dupeWarning}
    <span class="dupe-warning">{dupeWarning}</span>
  {/if}

  <select bind:value={band} tabindex="2">
    {#each bands as b}
      <option value={b}>{b}</option>
    {/each}
  </select>

  <select bind:value={mode} tabindex="3">
    {#each modes as m}
      <option value={m}>{m}</option>
    {/each}
  </select>

  <input bind:value={recvExchange} placeholder="Exchange (e.g., 2A NH)" tabindex="4" />

  <button type="submit" disabled={submitting}>
    Log QSO (Ctrl+Enter)
  </button>
</form>
```

### Go: Cabrillo Export
```go
// Source: [CITED: field-day-logger-plan.md Appendix A]
func exportCabrillo(w http.ResponseWriter, r *http.Request) {
	db := getDB(r)

	var buf bytes.Buffer

	// Header
	buf.WriteString("START-OF-LOG: 3.0\n")
	buf.WriteString("CREATED-BY: FDLogger v1.0\n")
	buf.WriteString("CONTEST: ARRL-FIELD-DAY\n")
	buf.WriteString("CALLSIGN: N0CALL\n") // TODO: from config
	buf.WriteString("CATEGORY-OPERATOR: SINGLE-OP\n")
	buf.WriteString("CATEGORY-POWER: LOW\n")
	buf.WriteString("CATEGORY-STATION: PORTABLE\n")
	buf.WriteString("CLAIMED-SCORE: 0\n") // TODO: compute

	// QSOs
	rows, _ := db.Query("SELECT timestamp, callsign, band, mode, sent_exchange, recv_exchange, is_dupe FROM qsos ORDER BY timestamp")
	for rows.Next() {
		var ts, call, band, mode, sentEx, recvEx string
		var isDupe bool
		rows.Scan(&ts, &call, &band, &mode, &sentEx, &recvEx, &isDupe)

		t, _ := time.Parse(time.RFC3339, ts)
		date := t.Format("2006-01-02")
		timeStr := t.Format("1504")

		freq := bandToFreq(band)            // "20M" → "14000"
		modeCode := modeToCabrillo(mode)    // "SSB" → "PH"
		exchSent := padRight(sentEx, 10)
		exchRecv := padRight(recvEx, 10)

		line := fmt.Sprintf("QSO: %5s %-4s %s %s %-10s %-10s %-10s %-10s\n",
			freq, modeCode, date, timeStr,
			"N0CALL", exchSent, call, exchRecv)
		if isDupe {
			line = strings.Replace(line, "QSO:", "QSO: ---dupe---", 1)
		}
		buf.WriteString(line)
	}
	rows.Close()

	buf.WriteString("END-OF-LOG:\n")

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment; filename=n0call_field_day.cbr")
	w.Write(buf.Bytes())
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `mattn/go-sqlite3` (CGo) | `modernc.org/sqlite` (pure Go) | 2020+ | No C toolchain needed; cross-compilation for ARM works trivially |
| Svelte 4 stores (`writable`, `derived`) | Svelte 5 runes (`$state`, `$derived`) | 2024-10 (Svelte 5) | Simpler API, universal reactivity, no `.subscribe()` boilerplate |
| `gorilla/mux` | `chi` | ~2020 | gorilla/mux is archived (2024); chi is actively maintained and has middleware composition |
| Go 1.21- ServeMux | Go 1.22+ enhanced ServeMux | 2024-02 | Go 1.22+ ServeMux supports `GET /path/{id}` patterns and method routing; could replace chi for simple APIs |
| Rollback journal (DELETE mode) | WAL mode | SQLite 3.7.0 (2010) | Concurrent reads without blocking writes; essential for any multi-reader scenario |

**Deprecated/outdated:**
- `gorilla/mux`: Archived December 2024. Do not use. Use `chi` or Go 1.22+ ServeMux.
- Svelte 4 `writable`/`derivable` stores: Still work in Svelte 5 via compatibility, but new code should use `$state` runes.
- `github.com/mattn/go-sqlite3`: Still maintained but CGo-dependent. Prefer `modernc.org/sqlite` for greenfield projects targeting ARM.

## Assumptions Log

> List all claims tagged `[ASSUMED]` in this research. The planner and discuss-phase use this
> section to identify decisions that need user confirmation before execution.

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | SvelteKit 2.61.1 + Svelte 5.56.0 are compatible and stable together | Standard Stack | MEDIUM — if incompatible, downgrade to SvelteKit 2.x with Svelte 4 stores |
| A2 | `modernc.org/sqlite` v1.51.0 supports all SQLite features needed (WAL, indexes, LIKE for partial call matching) | Standard Stack | LOW — this is a mature library (v1.51) with 3500+ importers |
| A3 | `chi` v5.3.0 is the correct major version for Go 1.22+ modules | Standard Stack | LOW — v5 is the current major version on pkg.go.dev |
| A4 | `adapter-static` fallback to `index.html` works when served by Go's `embed.FS` with `http.FileServer` | Architecture Patterns | MEDIUM — if Go's file server doesn't handle SPA fallback correctly, may need a custom handler |
| A5 | Go 1.22+ ServeMux enhanced routing could replace `chi` for this simple API surface | Alternatives Considered | LOW — either works; chi adds middleware that may be useful |
| A6 | The WAL-reset bug (SQLite 3.51.2 fixed in 3.51.3) is fully addressed by `modernc.org/sqlite` v1.51.0 (ships SQLite 3.53.1) | Common Pitfalls | LOW — confirmed by pkg.go.dev listing SQLite version 3.53.1 |
| A7 | Points calculation table (2-pt vs 1-pt modes) from planning doc is complete for Field Day Phase 1 | Code Examples | MEDIUM — if ARRL changes mode classifications, update the map |
| A8 | Cabrillo frequency mapping (band → kHz) follows standard Field Day conventions | Code Examples | LOW — well-established ARRL format; documented in Appendix A |

## Open Questions (RESOLVED)

1. **Go version on development machine** (RESOLVED)
   - Resolution: Go 1.22+ to be installed via `snap install go --classic` before Phase 1 implementation. Documented as prerequisite in PLAN.md.
   - What we know: Go is NOT installed on the current dev machine (`command -v go` returns empty)
   - What's unclear: Whether Go will be installed before implementation, or if an alternate build environment will be used
   - Recommendation: Install Go 1.22+ via `snap install go --classic` or `apt install golang-go` before Phase 1 implementation begins. Document as a prerequisite in PLAN.md.

2. **SvelteKit project init vs manual scaffold** (RESOLVED)
   - Resolution: Use `npm create svelte@latest` with "Skeleton" template as specified in Plan 01-01 Task 2.
   - What we know: `npm create svelte@latest` is the standard way to start a SvelteKit project. The CONTEXT.md specifies SvelteKit but not whether `npm create` should be used.
   - What's unclear: Whether the project should be scaffolded with `npm create svelte@latest` (which adds demo content and config) and then stripped down, or manually scaffolded
   - Recommendation: Use `npm create svelte@latest` with the "Skeleton" template (no demo app) to get proper config files, then configure for SPA mode with adapter-static.

3. **Go module path for the project** (RESOLVED)
   - Resolution: `github.com/jeremy/mlogger-fd` — specified in Plan 01-01 Task 2.
   - What we know: The project is named "Field Day Logger" but no Go module path is specified
   - What's unclear: What module path to use in `go.mod`
   - Recommendation: Use `github.com/jeremy/mlogger-fd` (matching the repo directory) or a simpler path. Document in PLAN.md.

4. **Station configuration for Phase 1** (RESOLVED)
   - Resolution: Hardcode "N0CALL" for Cabrillo export headers. Phase 2 adds station config UI. Documented in Plan 01-05.
   - What we know: Station config (callsign, class, section) is deferred to Phase 2 per the roadmap. Phase 1 needs a callsign for Cabrillo export headers.
   - What's unclear: Whether to hardcode "N0CALL" (placeholder), use a config file, or add a minimal station config endpoint
   - Recommendation: Hardcode "N0CALL" for Phase 1 Cabrillo export. Phase 2 will add the station configuration UI. Document as known limitation.

5. **Frontend routing needs** (RESOLVED)
   - Resolution: Single `+page.svelte` composing three component panels — as implemented in Plan 01-01 Task 4.
   - What we know: D-04 specifies a single-page three-panel layout — no tab switching, no separate pages
   - What's unclear: Whether SvelteKit's file-based routing is overkill for a single-page app. A single `+page.svelte` with three component imports may be sufficient.
   - Recommendation: Start with a single `+page.svelte` route composing three components. Add SvelteKit routes only if navigation complexity grows (unlikely in Phase 1).

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | Backend compiler/runtime | ✗ | — | **BLOCKING** — must install Go 1.22+ before implementation |
| Node.js | Frontend dev server/build | ✓ | v22.22.3 | — |
| npm | Frontend package management | ✓ | 10.9.8 | — |
| SQLite3 CLI | Database inspection/debugging | ✓ | 3.53.1 | — |
| systemd | Deployment (service unit) | ✗ (dev machine) | — | N/A for dev; required on RPi for production deployment |
| `gsdp` | GSD SDK tooling | ✗ | — | Not needed for research/planning phase |

**Missing dependencies with no fallback:**
- **Go 1.22+** — BLOCKING. Backend cannot be compiled without Go. Must be installed before implementation tasks begin. Install via: `sudo snap install go --classic` or `sudo apt install golang-go` (check version ≥ 1.22).

**Missing dependencies with fallback:**
- None — all other dependencies are available or not needed at dev time

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go: `testing` (stdlib) + `httptest` (stdlib); Frontend: `vitest` (via SvelteKit/Vite) |
| Config file | Go: none (convention: `_test.go` files); Frontend: `vitest.config.ts` (in frontend/) |
| Quick run command | `go test ./... -count=1` (Go); `cd frontend && npx vitest run` (Frontend) |
| Full suite command | `go test ./... -v -count=1 -race` (Go); `cd frontend && npx vitest run` (Frontend) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| QSO-01 | Form submission creates QSO via API | integration | `go test ./internal/handler/ -run TestCreateQSO -v` | ❌ Wave 0 |
| QSO-02 | Callsign validation warns on empty/single-char | unit | `go test ./internal/handler/ -run TestQSOValidation -v` | ❌ Wave 0 |
| QSO-03 | QSO search and edit via GET/PUT endpoints | integration | `go test ./internal/handler/ -run TestQSOUpdate -v` | ❌ Wave 0 |
| QSO-04 | Ctrl+Enter triggers form submit | unit (frontend) | `cd frontend && npx vitest run src/lib/components/QsoEntryForm.test.js` | ❌ Wave 0 |
| DUPE-01 | Dupe check returns true for same callsign+band+mode | unit | `go test ./internal/qso/ -run TestCheckDupe -v` | ❌ Wave 0 |
| DUPE-02 | Partial call similarity returns matching calls | unit | `go test ./internal/qso/ -run TestSimilarCalls -v` | ❌ Wave 0 |
| DUPE-03 | Dupe QSO has is_dupe=1 and points=0 | integration | `go test ./internal/handler/ -run TestDupeQSOZeroPoints -v` | ❌ Wave 0 |
| SCOR-01 | Stats endpoint returns rate_10min, rate_1hr, total | integration | `go test ./internal/handler/ -run TestStatsRate -v` | ❌ Wave 0 |
| SCOR-02 | Stats endpoint returns raw_points, multiplier, score | integration | `go test ./internal/handler/ -run TestStatsScore -v` | ❌ Wave 0 |
| SCOR-03 | Stats endpoint returns band/mode breakdown counts | integration | `go test ./internal/handler/ -run TestStatsBreakdown -v` | ❌ Wave 0 |
| EXPR-01 | Cabrillo export returns valid fixed-width QSO lines | unit | `go test ./internal/cabrillo/ -run TestCabrilloExport -v` | ❌ Wave 0 |
| EXPR-02 | Cabrillo includes correct header metadata | unit | `go test ./internal/cabrillo/ -run TestCabrilloHeader -v` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./... -count=1` (Go) + `cd frontend && npx vitest run` (Frontend)
- **Per wave merge:** Full suite with `-race` flag
- **Phase gate:** All tests green before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] `internal/handler/qso_test.go` — covers QSO-01, QSO-02, QSO-03, DUPE-03
- [ ] `internal/qso/dupe_test.go` — covers DUPE-01, DUPE-02
- [ ] `internal/qso/points_test.go` — covers points calculation
- [ ] `internal/handler/stats_test.go` — covers SCOR-01, SCOR-02, SCOR-03
- [ ] `internal/cabrillo/cabrillo_test.go` — covers EXPR-01, EXPR-02
- [ ] `internal/db/db_test.go` — covers schema creation, WAL mode verification
- [ ] `frontend/src/lib/components/QsoEntryForm.test.js` — covers QSO-04 (keyboard shortcuts)
- [ ] Go test infrastructure: none detected (greenfield) — create `go.mod` and test files
- [ ] Frontend test infrastructure: none detected (greenfield) — install vitest + jsdom

*(All gaps — greenfield project, no existing test infrastructure)*

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | No auth for trusted LAN (per PROJECT.md) |
| V3 Session Management | No | No sessions — stateless REST API |
| V4 Access Control | No | Single-user Phase 1; LAN-only deployment |
| V5 Input Validation | Yes | Lenient callsign validation (D-03); required field checks on POST /api/qso; band/mode enum validation |
| V6 Cryptography | No | No sensitive data; LAN-only; no encryption needed for Phase 1 |

### Known Threat Patterns for Go + SQLite + SPA

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| SQL injection via callsign/band/mode/exchange fields | Tampering | Use parameterized queries (`?` placeholders in `database/sql`); never string-concatenate user input into SQL |
| XSS via QSO data rendered in log table | Information Disclosure | Svelte auto-escapes `{expression}` output. No `{@html}` for user-provided data. |
| Path traversal via embedded file serving | Information Disclosure | `embed.FS` is read-only and only contains files known at compile time. `http.FileServer` sanitizes paths. |
| JSON injection in API responses | Tampering | `encoding/json` properly escapes all strings. No raw string concatenation for JSON. |
| Mass assignment via JSON body | Tampering | Explicit struct mapping in Go handlers — only decode expected fields. Never decode into a map and pass to SQL. |

## Sources

### Primary (HIGH confidence)
- [CITED: pkg.go.dev/modernc.org/sqlite] - SQLite driver docs, v1.51.0, shipped SQLite 3.53.1, DSN pragma parameters, connection hooks
- [CITED: pkg.go.dev/github.com/go-chi/chi/v5] - Router docs, v5.3.0, middleware list, URL params, Route/Mount/Group patterns
- [CITED: pkg.go.dev/embed] - Go embed.FS docs, //go:embed directive, FS.Open/ReadDir/ReadFile, http.FS integration
- [CITED: kit.svelte.dev/docs/adapter-static] - Static adapter configuration, fallback option, prerender settings, build output
- [CITED: kit.svelte.dev/docs/single-page-apps] - SPA mode, ssr=false, fallback page, Apache config reference
- [CITED: kit.svelte.dev/docs/state-management] - Client vs server state, SPA mode state safety, context API
- [CITED: svelte.dev/docs/svelte/$state] - Svelte 5 runes, deep reactivity, $state.raw, passing state across modules
- [CITED: sqlite.org/wal.html] - WAL mode mechanics, checkpointing, concurrency, activation, persistence, WAL-reset bug (§11)
- [CITED: sqlite.org/pragma.html#pragma_journal_mode] - PRAGMA journal_mode syntax, WAL persistence
- [CITED: field-day-logger-plan.md] - Complete system specification: schema (§4.4), API endpoints (§4.5), points calculation (§2.4, §4.4), Cabrillo format (Appendix A), component list (§4.2), scoring formula (§2.5)

### Secondary (MEDIUM confidence)
- npm registry verification for Svelte ecosystem packages (versions confirmed: kit 2.61.1, adapter-static 3.0.10, svelte 5.56.0, vite 8.0.14)
- SQLite version 3.53.1 confirmed installed on dev machine

### Tertiary (LOW confidence)
- Training data knowledge about SvelteKit + Go embedding patterns (not verified against a specific project)
- Go 1.22+ enhanced ServeMux capabilities (not tested; chi is recommended as primary)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all packages verified via official pkg.go.dev and npm registry
- Architecture: HIGH — patterns derived from official Go docs (embed, chi) and SvelteKit docs (adapter-static, SPA mode)
- Pitfalls: MEDIUM — based on official SQLite WAL docs and common Go/SvelteKit ecosystem knowledge; the WAL-reset bug is specifically verified
- Environment: HIGH — audited via `command -v` and `--version` probes on the actual dev machine

**Research date:** 2026-05-29
**Valid until:** 2026-07-01 (ARRL Field Day event; stack is stable)
