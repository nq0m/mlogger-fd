# Phase 02: Multi-User & Real-Time — Research

**Researched:** 2026-05-29
**Domain:** Go WebSocket broadcasting + SvelteKit real-time client + shared station configuration
**Confidence:** HIGH

## Summary

Phase 2 adds multi-operator real-time collaboration to the Field Day Logger. The Phase 1 baseline provides a working single-user system: Go backend with chi router, SQLite in WAL mode, SvelteKit SPA with Svelte 5 runes, and a REST API for QSO CRUD, dupe checking, stats, and Cabrillo export. Phase 2 transforms this into a multi-user system by adding WebSocket-based real-time broadcasting, a shared station configuration table, and per-client operator identity.

The core architectural addition is a **WebSocket hub** on the Go backend that broadcasts new QSOs to all connected clients within 1 second. When any operator submits a QSO via the existing `POST /api/qso` endpoint, the handler pushes the new QSO (as JSON) to the hub, which fans it out to every connected WebSocket client. Clients receive these events and merge incoming QSOs into the shared `$state` store, triggering reactive UI updates in the LogTable and StatsBar.

Station configuration (CONF-01 through CONF-03) uses a new `station_config` SQLite table holding a single row: callsign, class, ARRL section, transmitter count, and power level. A `PUT /api/station-config` endpoint allows any operator to update the configuration; a `GET /api/station-config` endpoint returns it. The config persists across server restarts (CONF-03). Operator identity (CONF-02) is a simple text input stored in client-side state only, sent with each `POST /api/qso` as the existing `operator` field.

**Primary recommendation:** Add `github.com/gorilla/websocket` to the Go backend for WebSocket support. Create a Hub that manages connected clients and broadcasts QSO events. Add a `station_config` table and REST endpoints. Wire the SvelteKit frontend to connect a WebSocket and merge incoming QSOs into the existing `$state` store — no new frontend dependencies needed.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| WebSocket upgrade + connection management | API / Backend | — | Go owns the HTTP server; chi middleware and gorilla/websocket handle the upgrade |
| QSO event broadcast (SYNC-02) | API / Backend | — | Server is the single source of truth; broadcasts on successful QSO insert |
| Multi-client QSO persistence (SYNC-01) | API / Backend | Database / Storage | SQLite WAL mode already supports concurrent readers; single writer serialized by Go |
| Station config persistence (CONF-03) | Database / Storage | API / Backend | New `station_config` single-row table; REST endpoints mediate access |
| Station config UI (CONF-01) | Browser / Client | API / Backend | Client renders form; server provides GET/PUT API |
| Operator identity (CONF-02) | Browser / Client | — | Per-session text input; stored in client `$state`, sent with each QSO |
| Live scoreboard sync | Browser / Client | API / Backend | WebSocket broadcasts new QSOs; clients re-fetch stats on receipt |
| WebSocket event handling | Browser / Client | — | Native `WebSocket` API in browser; no library needed |
| Cabrillo export (with real config) | API / Backend | Database / Storage | Now reads real station config instead of hardcoded "N0CALL" |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go | 1.25.0 | Backend runtime | Already in go.mod; single binary, low memory for RPi |
| `modernc.org/sqlite` | v1.51.0 | SQLite driver (CGo-free) | Already in go.mod; Phase 1 baseline; pure Go, ARM cross-compilation |
| `github.com/go-chi/chi/v5` | v5.3.0 | HTTP router | Already in go.mod; WebSocket upgrade integrates as chi handler |
| `github.com/gorilla/websocket` | v1.5.3 | WebSocket protocol | 44,585 importers [VERIFIED: pkg.go.dev], mature, BSD-2-Clause, stable API since v1.0 |
| `net/http` (stdlib) | go1.25.0 | HTTP server | Already used; WebSocket upgrade piggybacks on existing HTTP server |
| `encoding/json` (stdlib) | go1.25.0 | JSON serialization | Already used; WebSocket messages are JSON-encoded QSOs |
| `sync` (stdlib) | go1.25.0 | Concurrent access control | `sync.RWMutex` for Hub client set; `sync.Map` or `channel` for broadcast |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `embed` (stdlib) | go1.25.0 | Embed SPA static files | Already used; no changes needed for Phase 2 |
| `log/slog` (stdlib) | go1.25.0 | Structured logging | Log WebSocket connect/disconnect events |
| `database/sql` (stdlib) | go1.25.0 | Database interface | Station config CRUD; QSO queries unchanged |
| WebSocket API (browser) | — | Client-side WebSocket | No npm package needed; native `new WebSocket()` in all modern browsers |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `gorilla/websocket` v1.5.3 | `github.com/coder/websocket` v1.8.14 | coder/websocket: cleaner context-based API, zero-alloc reads, concurrent writes. BUT: only 790 importers, younger ecosystem. gorilla/websocket: 44K importers, battle-tested, simpler API that maps directly to our broadcast-only use case. For a Field Day tent deployment, gorilla is the safer choice. |
| `gorilla/websocket` | `nhooyr.io/websocket` v1.8.17 | **Do not use.** Deprecated — README says "Use github.com/coder/websocket instead" [VERIFIED: pkg.go.dev/nhooyr.io/websocket] |
| Hub with `sync.Map` | Hub with `chan broadcast` | Channels are more idiomatic Go and avoid map iteration under lock. Use a `chan []byte` for the Hub's broadcast channel; each client goroutine selects on `ctx.Done()` and the broadcast channel. |
| Custom broadcast mechanism | SSE (Server-Sent Events) | SSE is simpler (no library needed) but unidirectional only. WebSocket allows future bidirectional features (e.g., client→server typing indicators). WebSocket is the documented decision from PROJECT.md. |
| Client WebSocket library | `reconnecting-websocket` npm package | Reconnection logic is simple enough to hand-roll (~15 lines); no npm dependency needed. Phase 3 offline resilience already handles reconnect patterns. |

**Installation:**
```bash
# Go backend — add gorilla/websocket (no frontend npm packages needed)
go get github.com/gorilla/websocket@v1.5.3
```

**Version verification:**
```bash
go list -m github.com/gorilla/websocket@latest  # v1.5.3 [VERIFIED: proxy.golang.org]
```

## Package Legitimacy Audit

> slopcheck unavailable — all packages tagged `[ASSUMED]`. Planner must add `checkpoint:human-verify` before each install.

| Package | Registry | Age | Downloads | Source Repo | slopcheck | Disposition |
|---------|----------|-----|-----------|-------------|-----------|-------------|
| `github.com/gorilla/websocket` | Go mod | 10+ yrs | 44,585 importers | github.com/gorilla/websocket | [ASSUMED] | Approved (well-known) |

**Packages removed due to slopcheck [SLOP] verdict:** none (slopcheck unavailable)
**Packages flagged as suspicious [SUS]:** none

*slopcheck was unavailable at research time — all packages above are tagged `[ASSUMED]` and the planner must gate each install behind a `checkpoint:human-verify` task.*

## Architecture Patterns

### System Architecture Diagram (Phase 2 additions in bold)

```
┌──────────────────────────────────────────────────────────────────┐
│                       Browser #1 (Client)                        │
│                                                                  │
│  ┌──────────────┐  ┌────────────┐  ┌─────────────┐              │
│  │ QSO Entry    │  │ Stats Bar  │  │ Log Table   │              │
│  │ Form         │  │            │  │             │              │
│  └──────┬───────┘  └─────┬──────┘  └──────┬──────┘              │
│         │                │                │                      │
│  ┌──────┴────────────────┴────────────────┴───────┐              │
│  │            Svelte $state (shared)              │              │
│  │  qsos[], stats, operatorName, stationConfig    │              │
│  └──────┬──────────────────────────────┬─────────┘              │
│         │ fetch('/api/*')              │ ws://host/ws            │
│  ┌──────┴──────┐              ┌────────┴────────┐              │
│  │ API Client  │              │ **WebSocket     │              │
│  │ (api.js)    │              │  Listener**     │              │
│  └──────┬──────┘              └────────┬────────┘              │
└─────────┼──────────────────────────────┼───────────────────────┘
          │ HTTP (REST)                   │ WebSocket (real-time)
          │                               │
┌─────────┼──────────────────────────────┼───────────────────────┐
│         │       Go Binary (Server)      │                       │
│  ┌──────┴──────────────────────────────┴──────────────────┐   │
│  │                    chi Router                           │   │
│  │  /api/* ──► Handlers      /ws ──► **WebSocket Handler**│   │
│  │  /* ──────► embed.FS SPA                                │   │
│  └──────┬─────────────────────────────────────────────────┘   │
│         │                                                      │
│  ┌──────┴──────────────┐    ┌───────────────────────────┐     │
│  │  Existing Handlers  │    │  **WebSocket Hub**        │     │
│  │  (qso, stats,       │    │  ┌─────────────────────┐  │     │
│  │   dupe, export,     │    │  │ clients map[*Conn]  │  │     │
│  │   health)           │    │  │ broadcast chan []byte│  │     │
│  │                     │    │  │ register/unregister  │  │     │
│  │  **NEW: config.go** │    │  │ channels             │  │     │
│  │  GET/PUT station    │    │  └─────────────────────┘  │     │
│  │  config             │    │                           │     │
│  └──────┬──────────────┘    └───────────┬───────────────┘     │
│         │                               │  ↑ broadcast on     │
│         │                               │  │ QSO create       │
│  ┌──────┴───────────────────────────────┴──┴──────────────┐   │
│  │              modernc.org/sqlite                         │   │
│  │  ┌────────────┐  ┌───────────────────┐                 │   │
│  │  │ qsos table │  │**station_config** │                 │   │
│  │  │ (existing) │  │ (NEW single row)  │                 │   │
│  │  └────────────┘  └───────────────────┘                 │   │
│  └───────────────────────┬────────────────────────────────┘   │
│                          │                                     │
│                   ┌──────┴──────┐                              │
│                   │  SQLite DB  │                              │
│                   │  (WAL mode) │                              │
│                   └─────────────┘                              │
└──────────────────────────────────────────────────────────────┘

  Browser #2, #3, ... — identical client structure connecting to the
  same server via WebSocket. Each receives the same broadcast events.
```

**Data Flow:** Operator on Device A submits QSO via POST /api/qso → handler inserts into SQLite → handler pushes QSO JSON to Hub.broadcast channel → Hub goroutine iterates clients → each client's write goroutine sends JSON over its WebSocket → Device B and Device C WebSocket listeners receive event → merge into client $state → reactive UI updates.

### Recommended Project Structure (Phase 2 additions in **bold**)

```
fdlogger/
├── go.mod
├── go.sum
├── main.go                          # Add: /ws route, Hub initialization
├── internal/
│   ├── db/
│   │   ├── db.go                    # Add: station_config table migration
│   │   └── schema.sql               # Add: CREATE TABLE station_config
│   ├── handler/
│   │   ├── qso.go                   # MODIFY: broadcast to Hub after create
│   │   ├── dupe.go
│   │   ├── stats.go
│   │   ├── export.go                # MODIFY: read real station config
│   │   ├── health.go
│   │   ├── **config.go**            # NEW: GET/PUT /api/station-config
│   │   ├── **ws.go**                # NEW: WebSocket upgrade handler
│   │   ├── qso_test.go
│   │   ├── stats_test.go
│   │   ├── **config_test.go**       # NEW: station config handler tests
│   │   └── **ws_test.go**           # NEW: WebSocket hub tests
│   ├── model/
│   │   ├── qso.go                   # MODIFY: add StationConfig struct
│   │   └── **config.go**            # NEW: StationConfig model + validation
│   ├── qso/
│   │   ├── points.go
│   │   ├── dupe.go
│   │   └── dupe_test.go
│   ├── cabrillo/
│   │   ├── cabrillo.go              # MODIFY: use real station config
│   │   └── cabrillo_test.go
│   └── **ws/**                       # NEW: WebSocket hub package
│       ├── **hub.go**               # Hub: client registry, broadcast, Run()
│       └── **hub_test.go**          # Hub unit tests
├── frontend/
│   ├── package.json
│   ├── svelte.config.js
│   ├── vite.config.ts
│   ├── src/
│   │   ├── app.html
│   │   ├── app.css
│   │   ├── routes/
│   │   │   ├── +layout.svelte
│   │   │   ├── +layout.js
│   │   │   └── +page.svelte         # MODIFY: add operator selector, config panel
│   │   ├── lib/
│   │   │   ├── components/
│   │   │   │   ├── QsoEntryForm.svelte    # MODIFY: add operator field
│   │   │   │   ├── StatsBar.svelte
│   │   │   │   ├── LogTable.svelte
│   │   │   │   ├── **StationConfig.svelte** # NEW: config form (admin)
│   │   │   │   └── **OperatorSelector.svelte** # NEW: per-session operator input
│   │   │   ├── stores/
│   │   │   │   └── qso.svelte.js    # MODIFY: add WebSocket listener, operator state
│   │   │   ├── api.js               # MODIFY: add config API functions
│   │   │   └── **ws.js**            # NEW: WebSocket client module
│   └── build/
└── Makefile
```

### Architecture Patterns

### Pattern 1: WebSocket Hub with Channel-Based Broadcast

**What:** A Hub struct manages all connected WebSocket clients. The `CreateQSO` handler pushes new QSO JSON to the Hub's broadcast channel. The Hub's `Run()` goroutine receives from this channel and writes to every connected client. Each client connection runs a write pump goroutine that reads from its own buffered channel and writes to the WebSocket connection.

**When to use:** For broadcasting server-side events (QSO creations) to all connected clients in real-time. The Hub pattern is the canonical gorilla/websocket approach from the official chat example.

**Example:**
```go
// Source: [CITED: pkg.go.dev/github.com/gorilla/websocket + gorilla/websocket chat example]
package ws

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // LAN-only, trusted network per AGENTS.md
	},
}

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	mu       sync.RWMutex
	clients  map[*Client]bool
	broadcast chan []byte
	register   chan *Client
	unregister chan *Client
}

// Client represents a single WebSocket connection.
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),  // buffered for brief bursts
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the Hub's main loop. Must be called in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			slog.Info("websocket client connected", "total", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			slog.Info("websocket client disconnected", "total", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client's send buffer is full — drop message for this client
					// Field Day QSO rate is low enough this shouldn't happen
					close(client.send)
					delete(h.clients, client) // remove under lock? careful...
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients. Safe for concurrent use.
func (h *Hub) Broadcast(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	h.broadcast <- data
	return nil
}
```

### Pattern 2: WebSocket Upgrade Handler (chi Integration)

**What:** A chi handler function that upgrades the HTTP connection to WebSocket, creates a Client, registers it with the Hub, and starts read/write pumps.

**When to use:** Mounted at `GET /ws` in the chi router. The client connects via `new WebSocket('ws://' + location.host + '/ws')`.

**Example:**
```go
// Source: [CITED: pkg.go.dev/github.com/gorilla/websocket]
func ServeWS(hub *ws.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}

	client := &ws.Client{
		Hub:  hub,
		Conn: conn,
		Send: make(chan []byte, 64),
	}
	hub.Register <- client

	// Start write pump in a goroutine
	go client.WritePump()

	// Read pump runs in the current goroutine (blocks until disconnect)
	client.ReadPump()
}
```

### Pattern 3: CreateQSO Handler Broadcasts to Hub

**What:** After successfully inserting a QSO into SQLite, the handler constructs a full QSO response (including the new DB-assigned `id`) and broadcasts it to the Hub before returning the HTTP response.

**When to use:** In the `CreateQSO` handler, after `db.Exec` and `result.LastInsertId()`. The existing response JSON is also sent over WebSocket.

**Example:**
```go
// MODIFICATION to internal/handler/qso.go CreateQSO function
// After the existing insert and LastInsertId:

qsoResponse := map[string]interface{}{
	"type":          "qso_created",
	"id":            id,
	"timestamp":     now,
	"callsign":      input.Callsign,
	"band":          input.Band,
	"mode":          input.Mode,
	"recv_exchange": input.RecvExchange,
	"sent_exchange": input.SentExchange,
	"operator":      input.Operator,
	"is_dupe":       isDupe,
	"points":        points,
}

// Broadcast to all WebSocket clients
if hub != nil {
	hub.Broadcast(qsoResponse)
}
```

### Pattern 4: Station Configuration SQLite Table + REST API

**What:** A single-row table ensures exactly one configuration exists (enforced by the application, not SQLite constraints — a simple upsert). GET returns the current config; PUT stores new config. The Cabrillo export handler now reads real station data instead of hardcoding "N0CALL".

**When to use:** Station admin configuration (CONF-01, CONF-03). The config is global, not per-user.

**Schema:**
```sql
-- Source: [CITED: field-day-logger-plan.md + ARRL Field Day Cabrillo spec]
CREATE TABLE IF NOT EXISTS station_config (
    id INTEGER PRIMARY KEY CHECK (id = 1),  -- enforce single row
    callsign TEXT NOT NULL DEFAULT 'N0CALL',
    class TEXT NOT NULL DEFAULT '1D',
    arrl_section TEXT NOT NULL DEFAULT 'EMA',
    transmitter_count INTEGER NOT NULL DEFAULT 1,
    power_level TEXT NOT NULL DEFAULT 'LOW',
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
```

### Pattern 5: Client-Side WebSocket Integration with Svelte 5 $state

**What:** A `ws.js` module manages the WebSocket connection lifecycle. When a `qso_created` message arrives, it's merged into the shared `qsos` array and stats are refreshed. Connection status is tracked in a `$state` variable.

**When to use:** Replaces or augments the current `fetchStats()` polling pattern. On each WebSocket `qso_created` event, call `fetchStats()` to refresh the scoreboard.

**Example:**
```javascript
// Source: [CITED: developer.mozilla.org/en-US/docs/Web/API/WebSocket + svelte.dev/docs/svelte/$state]
// frontend/src/lib/ws.js
import { qsos, stats, fetchStats } from '$lib/stores/qso.svelte.js';

export const wsConnected = $state(false);

let ws = null;
let reconnectTimer = null;

export function connectWebSocket() {
	const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
	const url = `${protocol}//${location.host}/ws`;

	function connect() {
		ws = new WebSocket(url);

		ws.onopen = () => {
			wsConnected = true;
			if (reconnectTimer) {
				clearTimeout(reconnectTimer);
				reconnectTimer = null;
			}
		};

		ws.onmessage = (event) => {
			const data = JSON.parse(event.data);
			if (data.type === 'qso_created') {
				// Add to front of qsos array (newest first)
				qsos.unshift({
					id: data.id,
					timestamp: data.timestamp,
					callsign: data.callsign,
					band: data.band,
					mode: data.mode,
					recv_exchange: data.recv_exchange,
					sent_exchange: data.sent_exchange,
					operator: data.operator,
					is_dupe: data.is_dupe,
					points: data.points,
				});
				// Refresh stats to update scoreboard
				fetchStats();
			}
		};

		ws.onclose = () => {
			wsConnected = false;
			// Reconnect after 2 seconds
			reconnectTimer = setTimeout(connect, 2000);
		};

		ws.onerror = () => {
			// onclose will fire after onerror, reconnect handled there
		};
	}

	connect();
}
```

### Pattern 6: Operator Identity (Client-Side Only)

**What:** A text input (`<input bind:value={operator} />`) in the UI that sets a per-session operator identifier. The value is stored in `$state` and sent with every `POST /api/qso` as the existing `operator` field. No server-side storage needed.

**When to use:** CONF-02. The operator field already exists in the QSO schema from Phase 1. This just adds client-side UI to set it.

### Anti-Patterns to Avoid

- **[Anti-pattern] Broadcasting from multiple goroutines without synchronization:** The `CreateQSO` handler should not iterate clients directly. Always send to `Hub.broadcast` channel and let the Hub's single `Run()` goroutine handle fan-out.
- **[Anti-pattern] Holding the Hub mutex during WebSocket writes:** Writing to a WebSocket connection can block. Use per-client buffered `send` channels so the Hub can quickly iterate clients and move on. The write pump goroutine handles the actual I/O.
- **[Anti-pattern] Using `r.Context()` after WebSocket upgrade:** gorilla/websocket docs warn that `r.Context()` may behave unexpectedly after `Upgrade()`. Use `context.Background()` or a derived context for WebSocket read/write operations. This is a LAN-only server — request-scoped context isn't meaningful post-upgrade.
- **[Anti-pattern] Storing operator identity on the server:** Operator identity should stay client-side. Storing it server-side creates complexity (session management, cleanup on disconnect) with no benefit for Field Day. The operator name is metadata on QSO records, not an authentication mechanism.
- **[Anti-pattern] Using `nhooyr.io/websocket`:** This package is deprecated. Its README explicitly says "Use github.com/coder/websocket instead" [VERIFIED: pkg.go.dev/nhooyr.io/websocket].

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| WebSocket protocol (handshake, framing, ping/pong) | Custom TCP socket with hand-crafted HTTP upgrade | `gorilla/websocket` v1.5.3 | RFC 6455 is 71 pages. gorilla passes the Autobahn test suite. Hand-rolling WebSocket framing is error-prone and unnecessary. |
| Client fan-out / broadcast | Iterating clients from handler goroutines | Hub pattern with `chan []byte` | The Hub ensures single-writer access to the client set. Channels provide natural backpressure. Tested in thousands of Go projects. |
| WebSocket client reconnection | Custom exponential backoff | Simple `setTimeout(connect, 2000)` in onclose | Field Day is LAN — reconnection is about momentary WiFi drops. 2-second fixed delay is sufficient. Phase 3 adds robust offline queue. |
| Multi-writer fan-out with mutex | `sync.Mutex` on client map + direct writes | Per-client buffered `send chan` | Direct writes under mutex can block the Hub on slow clients. Channels decouple broadcast from I/O. |
| Station config validation | Inline validation in handlers | `model.ValidateStationConfig()` function | Consistent with Phase 1's `model.ValidateRequired()` pattern. Testable separately. |

**Key insight:** The broadcast hub pattern is well-established in Go. The gorilla/websocket chat example provides a battle-tested reference. Resist the urge to add complexity (Redis pub/sub, message queues, multiple hubs) — a single in-process Hub with buffered channels is sufficient for 2–6 Field Day operators.

## Common Pitfalls

### Pitfall 1: WebSocket Write Blocking the Hub
**What goes wrong:** A slow WebSocket client (switched WiFi, mobile throttling) blocks the Hub's broadcast loop because the Hub writes directly to connections under a mutex.
**Why it happens:** WebSocket writes can block if the TCP send buffer is full. Holding the Hub mutex during a blocked write freezes all other clients.
**How to avoid:** Use per-client buffered `send` channels (64-buffer is sufficient for Field Day QSO rate). The Hub does a non-blocking send: `select { case client.send <- msg: default: /* drop */ }`. The client's write pump goroutine handles the actual blocking I/O.
**Warning signs:** One operator disconnects and all other clients stop receiving updates for 5+ seconds.

### Pitfall 2: Goroutine Leak from Unclosed WebSocket Read Pump
**What goes wrong:** The read pump goroutine never exits, accumulating goroutines and memory over the 27-hour contest.
**Why it happens:** If the read pump's `for { conn.ReadMessage() }` loop doesn't detect disconnection properly, or if the context isn't cancelled on close.
**How to avoid:** The read pump exits when `ReadMessage()` returns an error (connection closed). The `defer` in the upgrade handler unregisters the client. gorilla/websocket properly returns errors on all disconnect scenarios. Use `conn.SetReadDeadline()` with a reasonable timeout (60s ping/pong) to detect zombie connections.
**Warning signs:** `curl localhost:8080/debug/pprof/goroutine?debug=1` shows growing goroutine count.

### Pitfall 3: SQLite WAL Blocking with Multiple Writers
**What goes wrong:** Two operators submit QSOs simultaneously and one gets a "database is locked" error.
**Why it happens:** SQLite in WAL mode supports concurrent readers but serializes writers. With `SetMaxOpenConns(1)`, multiple concurrent writes queue up. The default busy timeout (5000ms) should handle this for Field Day volumes, but very rapid simultaneous submissions could hit it.
**How to avoid:** `busy_timeout(5000)` is already set in Phase 1's `db.Open()`. For Field Day (max ~6 operators, QSO every few seconds), this is sufficient. If errors appear, increase to `busy_timeout(10000)`. The Go handler already returns 500 on DB errors — the client should retry.
**Warning signs:** "database is locked" errors in server logs during contests with 4+ operators logging simultaneously on the same second.

### Pitfall 4: WebSocket Origin Check Blocking LAN Clients
**What goes wrong:** WebSocket connections fail because the browser sends an Origin header that doesn't match the server hostname (e.g., accessing via IP `192.168.1.5:8080`).
**Why it happens:** gorilla/websocket's default `CheckOrigin` function rejects cross-origin requests where Origin ≠ Host. On a LAN, clients connect via different IP addresses or hostnames.
**How to avoid:** Set `CheckOrigin: func(r *http.Request) bool { return true }` in the Upgrader. This is safe per AGENTS.md ("LAN-only, no CORS needed, trusted network"). If using the simple shared password from AGENTS.md, add that check separately in the WebSocket upgrade handler.
**Warning signs:** Browser console shows "WebSocket connection failed" with no server-side log of the upgrade attempt.

### Pitfall 5: Client Receives Duplicate QSOs from WebSocket + Polling
**What goes wrong:** An operator's own QSO appears twice in their log table — once from the `POST /api/qso` response handler, and again from the WebSocket broadcast.
**Why it happens:** The client adds the QSO from the HTTP response AND from the WebSocket event. Without deduplication, both entries appear.
**How to avoid:** Two strategies: (1) The submitting client skips WebSocket events for QSOs it created (compare `data.operator` with local operator name — imperfect). (2) Use QSO `id` deduplication: the client tracks recently added IDs in a `Set` and skips duplicates. **Recommend option 2:** simpler, no false negatives. Maintain a `const recentIds = new Set()` that's trimmed to last 50 IDs.
**Warning signs:** QSO count in log table doesn't match stats total.

### Pitfall 6: Station Config Race Condition
**What goes wrong:** Two operators open the station config form simultaneously. Operator A saves, then Operator B saves 30 seconds later, overwriting A's changes.
**Why it happens:** No optimistic concurrency control. Last write wins.
**How to avoid:** This is an acceptable risk for Field Day. Station config is set once at the start of the contest and rarely changed. If changes ARE needed mid-contest, it's typically one person doing it. Document this limitation — no need for ETags or version fields for a single-row config table.
**Warning signs:** Configuration reverts unexpectedly. Mitigation: show a toast "Config updated" with the new values so operators notice conflicts.

## Code Examples

### Hub Initialization and Route Wiring in main.go
```go
// Source: [CITED: gorilla/websocket chat example + Phase 1 main.go]
package main

import (
	// ... existing imports
	"github.com/jeremy/mlogger-fd/internal/handler"
	"github.com/jeremy/mlogger-fd/internal/ws"
)

func main() {
	// ... existing setup (logger, db, chi router)

	// Create WebSocket hub and start it
	hub := ws.NewHub()
	go hub.Run()

	// Pass hub to handlers that need it
	// (via closure or dependency injection — consistent with Phase 1 closure pattern)

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
		r.Route("/qso", func(r chi.Router) {
			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				handler.CreateQSO(database, hub, w, r)  // MODIFIED: passes hub
			})
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				handler.ListQSOs(database, w, r)
			})
			r.Put("/{id}", func(w http.ResponseWriter, r *http.Request) {
				handler.UpdateQSO(database, w, r)
			})
		})
		// NEW: Station configuration endpoints
		r.Get("/station-config", func(w http.ResponseWriter, r *http.Request) {
			handler.GetStationConfig(database, w, r)
		})
		r.Put("/station-config", func(w http.ResponseWriter, r *http.Request) {
			handler.PutStationConfig(database, w, r)
		})
	})

	// WebSocket endpoint (outside /api — direct path)
	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		handler.ServeWS(hub, w, r)
	})

	// ... rest of main (SPA handler, ListenAndServe)
}
```

### Station Config Handler
```go
// Source: [CITED: Phase 1 handler pattern + ARRL Field Day config spec]
package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/jeremy/mlogger-fd/internal/model"
)

func GetStationConfig(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var cfg model.StationConfig
	err := db.QueryRow(`SELECT callsign, class, arrl_section, transmitter_count, power_level 
		FROM station_config WHERE id = 1`).Scan(
		&cfg.Callsign, &cfg.Class, &cfg.ARRLSection,
		&cfg.TransmitterCount, &cfg.PowerLevel,
	)
	if err == sql.ErrNoRows {
		// Return defaults if no config exists yet
		cfg = model.DefaultStationConfig()
	} else if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}

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

	// Upsert: INSERT OR REPLACE ensures single-row semantics
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

### Station Config Model
```go
// Source: [CITED: ARRL Field Day rules + Phase 1 model pattern]
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
	if cfg.Class == "" {
		return "class is required"
	}
	if cfg.ARRLSection == "" {
		return "arrl_section is required"
	}
	if cfg.TransmitterCount < 1 || cfg.TransmitterCount > 20 {
		return "transmitter_count must be between 1 and 20"
	}
	validPower := map[string]bool{"LOW": true, "HIGH": true, "QRP": true}
	if !validPower[cfg.PowerLevel] {
		return "power_level must be LOW, HIGH, or QRP"
	}
	return ""
}
```

### Svelte 5: WebSocket Client Module
```javascript
// Source: [CITED: MDN WebSocket API + Svelte 5 $state docs]
// frontend/src/lib/ws.js
import { qsos, stats, fetchStats } from '$lib/stores/qso.svelte.js';

export const wsConnected = $state(false);

let ws = null;
let reconnectTimer = null;
let shouldReconnect = true;

// Deduplication: track recent QSO IDs to avoid double-display
const recentIds = new Set();
const MAX_RECENT_IDS = 100;

function pruneRecentIds() {
	if (recentIds.size > MAX_RECENT_IDS) {
		const entries = [...recentIds];
		entries.slice(0, entries.length - MAX_RECENT_IDS).forEach(id => recentIds.delete(id));
	}
}

export function connectWebSocket() {
	const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
	const url = `${protocol}//${location.host}/ws`;
	shouldReconnect = true;

	function connect() {
		if (!shouldReconnect) return;

		ws = new WebSocket(url);

		ws.onopen = () => {
			wsConnected = true;
			if (reconnectTimer) {
				clearTimeout(reconnectTimer);
				reconnectTimer = null;
			}
			console.log('WebSocket connected');
		};

		ws.onmessage = (event) => {
			try {
				const data = JSON.parse(event.data);
				if (data.type === 'qso_created') {
					// Deduplicate by QSO ID
					if (recentIds.has(data.id)) return;
					recentIds.add(data.id);
					pruneRecentIds();

					qsos.unshift({
						id: data.id,
						timestamp: data.timestamp,
						callsign: data.callsign,
						band: data.band,
						mode: data.mode,
						recv_exchange: data.recv_exchange,
						sent_exchange: data.sent_exchange,
						operator: data.operator,
						is_dupe: data.is_dupe,
						points: data.points,
					});
					// Refresh stats to keep scoreboard current
					fetchStats();
				}
			} catch (e) {
				console.error('WebSocket message parse error:', e);
			}
		};

		ws.onclose = () => {
			wsConnected = false;
			// Reconnect after 2 seconds (LAN-appropriate)
			reconnectTimer = setTimeout(connect, 2000);
		};

		ws.onerror = (err) => {
			console.error('WebSocket error:', err);
			// onclose will fire after onerror
		};
	}

	connect();
}

export function disconnectWebSocket() {
	shouldReconnect = false;
	if (reconnectTimer) {
		clearTimeout(reconnectTimer);
		reconnectTimer = null;
	}
	if (ws) {
		ws.close();
		ws = null;
	}
	wsConnected = false;
}
```

### Svelte 5: Operator Selector Component
```svelte
<!-- Source: [CITED: svelte.dev/tutorial + CONF-02 requirement] -->
<!-- frontend/src/lib/components/OperatorSelector.svelte -->
<script>
	let operator = $state(localStorage.getItem('fdlogger_operator') || '');

	function saveOperator() {
		localStorage.setItem('fdlogger_operator', operator);
	}

	// Expose operator for parent components
	export function getOperator() {
		return operator || 'OP';
	}
</script>

<div class="operator-selector">
	<label for="operator-input">Operator:</label>
	<input
		id="operator-input"
		type="text"
		bind:value={operator}
		onchange={saveOperator}
		placeholder="Your callsign or name"
		maxlength="20"
	/>
</div>

<style>
	.operator-selector {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 14px;
	}
	.operator-selector input {
		width: 140px;
		padding: 4px 8px;
		border: 1px solid #ccc;
		border-radius: 4px;
		font-size: 14px;
	}
</style>
```

### Svelte 5: WebSocket Connection Status Indicator
```svelte
<!-- Source: [CITED: SYNC-05 placeholder for Phase 3 — Phase 2 prep] -->
<script>
	import { wsConnected } from '$lib/ws.js';
</script>

{#if !wsConnected}
	<div class="ws-indicator offline">● Disconnected</div>
{:else}
	<div class="ws-indicator online">● Live</div>
{/if}

<style>
	.ws-indicator {
		font-size: 11px;
		padding: 2px 8px;
		border-radius: 4px;
		font-weight: 600;
	}
	.online { color: #1a7a1a; }
	.offline { color: #cc3300; }
</style>
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `gorilla/websocket` v1.4.x (archived project) | `gorilla/websocket` v1.5.3 (new maintainer, active) | 2024 | Project was archived Nov 2022, unarchived and maintained under github.com/gorilla/websocket. Still the standard. |
| `nhooyr.io/websocket` | `github.com/coder/websocket` | 2024-09 | nhooyr handed maintenance to Coder. nhooyr.io/websocket is now deprecated with redirect notice [VERIFIED: pkg.go.dev]. |
| Hardcoded "N0CALL" in Cabrillo export | Read from `station_config` table | Phase 2 | Cabrillo export now uses real station callsign/class/section. |
| Single-user `$state` store | Multi-user `$state` store with WebSocket merge | Phase 2 | qsos array grows from both local POSTs AND remote WebSocket events. Deduplication via ID Set. |
| `Object.assign(stats, data)` on timer/poll | WebSocket-triggered `fetchStats()` | Phase 2 | Stats refresh becomes event-driven instead of poll-based. |

**Deprecated/outdated:**
- `nhooyr.io/websocket`: Explicitly deprecated. README says "Use https://github.com/coder/websocket instead". Do not use in any new code.
- Polling `/api/stats` on a timer: Phase 1 used manual refresh after QSO submit. Phase 2 makes it event-driven via WebSocket trigger. Polling remains as a fallback for clients that can't connect via WebSocket.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `gorilla/websocket` v1.5.3 works with Go 1.25.0 on ARM (Raspberry Pi) | Standard Stack | LOW — pure Go, no CGo, used in thousands of ARM deployments |
| A2 | `CheckOrigin: func(r *http.Request) bool { return true }` is safe for LAN-only deployment | Common Pitfalls | LOW — AGENTS.md explicitly says "No auth for trusted LAN; simple shared password if open WiFi". For open WiFi, add a shared-secret check before upgrade. |
| A3 | SQLite `busy_timeout(5000)` is sufficient for 6 operators logging simultaneously | Common Pitfalls | LOW — Field Day QSO rate is 1-4 QSOs/minute per operator; 5-second timeout handles rare write collisions. If issues arise, increase to 10000ms in `db.Open()`. |
| A4 | `INSERT OR REPLACE` on station_config table with `CHECK (id = 1)` correctly enforces single-row semantics | Architecture Patterns | LOW — Standard SQLite pattern. Phase 1 already uses parameterized queries correctly. |
| A5 | Client-side deduplication via QSO `id` Set is sufficient (no server-side "who already received this" tracking) | Common Pitfalls | LOW — Set lookup is O(1), memory is bounded (100 entries max). Server-side tracking would require per-client state on the Hub, adding complexity for no benefit. |
| A6 | `serveWS` handler doesn't need authentication for trusted LAN per AGENTS.md | Standard Stack | LOW — Explicitly stated in AGENTS.md and PROJECT.md constraints. |
| A7 | The existing `operator` field in the QSO schema (from Phase 1) is the correct column to send client-side operator identity | Architecture Patterns | LOW — Field already exists in `qsos` table and `CreateQSOInput` struct. Phase 1 sent it optionally; Phase 2 makes it visible in the UI. |
| A8 | Cabrillo export handler can be modified to read `station_config` table with no breaking changes | Architecture Patterns | LOW — `ExportCabrillo` currently hardcodes "N0CALL". Adding a config lookup is additive. |

## Open Questions

1. **Should we require a shared password for WebSocket connections on open WiFi?**
   - What we know: AGENTS.md says "None for trusted LAN; simple shared password if open WiFi." But Phase 2 doesn't differentiate — all clients connect to WebSocket without auth.
   - What's unclear: Whether to implement a password check in the WebSocket upgrade handler now, or defer to later. Since Phase 2 is LAN-only and AGENTS.md treats LAN as "trusted", the current no-auth approach is correct.
   - Recommendation: No password for Phase 2. If operators use an open WiFi network, the Phase 3/4 shared password mechanism can gate the WebSocket endpoint. Document as known limitation.

2. **Should the WebSocket broadcast include the full QSO object or just a notification to re-fetch?**
   - What we know: Broadcasting the full QSO JSON allows instant UI update without an extra HTTP round-trip. A "notification" approach (just `{type: "qso_created", id: 5}`) would require the client to `GET /api/qso` to get the data.
   - What's unclear: Which is more reliable? Full objects risk stale data if the QSO is edited before the client receives it. Notification approach adds latency.
   - Recommendation: Broadcast full QSO JSON. This matches Field Day's "< 1 second" real-time requirement. QSO edits during the broadcast window are extremely unlikely. The client can always re-fetch if needed.

3. **Should the Cabrillo export use the stored station config or allow overrides?**
   - What we know: CONF-01 requires that station config can be set. The Cabrillo export's header metadata (callsign, class, section) should use the stored values.
   - What's unclear: Whether to add query parameters to override station info per-export (e.g., for testing), or just use the database values always.
   - Recommendation: Use database values only. No overrides. Keep the one-click export simple per D-06. If operators need different headers, they update the config first.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | Backend compiler | ✓ | 1.25.0 | — |
| Node.js | Frontend dev server | ✓ | 22.22.3 | — |
| npm | Frontend package manager | ✓ | 10.9.8 | — |
| SQLite3 | Database | ✓ | 3.53.1 | — |
| `gorilla/websocket` | WebSocket protocol | ✓ (via go get) | v1.5.3 | — |

**Missing dependencies with no fallback:** None — gorilla/websocket will be fetched via `go get`.
**Missing dependencies with fallback:** None.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go: `testing` (stdlib) + `httptest` (stdlib) + `net/http/httptest`; Frontend: `vitest` (via SvelteKit/Vite) |
| Config file | Go: none (convention: `_test.go` files); Frontend: `vitest.config.ts` (in frontend/) |
| Quick run command | `go test ./... -count=1` (Go); `cd frontend && npx vitest run` (Frontend) |
| Full suite command | `go test ./... -v -count=1 -race` (Go); `cd frontend && npx vitest run` (Frontend) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| SYNC-01 | Multiple concurrent QSO inserts succeed without conflicts | unit (integration) | `go test ./internal/handler/ -run TestConcurrentQSOs -v` | ❌ Wave 0 |
| SYNC-02 | WebSocket broadcasts QSO to connected clients on create | unit | `go test ./internal/ws/ -run TestHubBroadcast -v` | ❌ Wave 0 |
| SYNC-02 | Client receives WebSocket message and updates QSO list | unit (frontend) | `cd frontend && npx vitest run src/lib/ws.test.js` | ❌ Wave 0 |
| CONF-01 | PUT /api/station-config stores configuration | integration | `go test ./internal/handler/ -run TestPutStationConfig -v` | ❌ Wave 0 |
| CONF-01 | GET /api/station-config returns stored configuration | integration | `go test ./internal/handler/ -run TestGetStationConfig -v` | ❌ Wave 0 |
| CONF-02 | QSO create sends operator field from client state | integration | `go test ./internal/handler/ -run TestCreateQSOWithOperator -v` | ❌ Wave 0 |
| CONF-03 | Station config survives server restart | integration | `go test ./internal/handler/ -run TestStationConfigPersistence -v` | ❌ Wave 0 |
| — | Hub registers and unregisters clients correctly | unit | `go test ./internal/ws/ -run TestHubRegister -v` | ❌ Wave 0 |
| — | Hub write pump sends messages without blocking | unit | `go test ./internal/ws/ -run TestWritePump -v` | ❌ Wave 0 |
| — | Cabrillo export uses real station config from DB | unit | `go test ./internal/cabrillo/ -run TestCabrilloWithConfig -v` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./... -count=1` (Go) + `cd frontend && npx vitest run` (Frontend)
- **Per wave merge:** Full suite with `-race` flag
- **Phase gate:** All tests green before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] `internal/ws/hub_test.go` — Hub register/unregister/broadcast tests
- [ ] `internal/handler/config_test.go` — Station config GET/PUT tests
- [ ] `internal/handler/qso_test.go` — Update existing to test WebSocket broadcast (or add new test)
- [ ] `internal/cabrillo/cabrillo_test.go` — Update to test real station config
- [ ] `frontend/src/lib/ws.test.js` — WebSocket client unit test (mock WebSocket)
- Framework install: All frameworks already installed from Phase 1

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | No auth for trusted LAN per AGENTS.md |
| V3 Session Management | No | No server-side sessions; operator identity is client-only |
| V4 Access Control | No | All operators have equal access on trusted LAN |
| V5 Input Validation | Yes | Station config validation (callsign, class, section, power level); QSO fields validated per Phase 1 → existing model.ValidateRequired() |
| V6 Cryptography | No | No encryption needed for LAN-only use |

### Known Threat Patterns for Go WebSocket + SvelteKit

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| WebSocket origin spoofing (CSRF via WebSocket) | Spoofing | `CheckOrigin: func(r *http.Request) bool { return true }` for trusted LAN. For open WiFi: validate Origin against known LAN IP range or add shared-secret token. |
| Malformed WebSocket frames crashing the server | Denial of Service | gorilla/websocket handles RFC 6455 frame parsing; `Recoverer` middleware catches panics. Set `ReadLimit` to prevent memory exhaustion from oversized messages. |
| JSON injection in WebSocket messages | Tampering | All WebSocket messages are server-generated (broadcast only). Client only reads, never parses into eval(). `JSON.parse()` is safe in browsers. |
| Station config overwrite by unauthorized operator | Tampering | No auth — acceptable on trusted LAN per AGENTS.md. All operators are trusted. |
| SQL injection in station config queries | Tampering | Parameterized queries (`?` placeholders) used consistently — same pattern as Phase 1 QSO handlers [VERIFIED: existing codebase]. |

## Sources

### Primary (HIGH confidence)
- [VERIFIED: pkg.go.dev/github.com/gorilla/websocket] — gorilla/websocket v1.5.3 API docs: Upgrader, Conn, message types, chat example pattern; verified 44,585 importers, BSD-2-Clause license
- [VERIFIED: proxy.golang.org] — gorilla/websocket@v1.5.3 confirmed as latest tagged version
- [VERIFIED: pkg.go.dev/nhooyr.io/websocket] — Confirmed deprecated: README says "Use github.com/coder/websocket instead"
- [CITED: developer.mozilla.org/en-US/docs/Web/API/WebSocket] — WebSocket browser API for client-side integration
- [CITED: svelte.dev/docs/svelte/$state] — Svelte 5 $state runes for shared client state
- [VERIFIED: existing codebase] — Phase 1 handler patterns: closure-based handler wiring, httptest patterns, chi route mounting

### Secondary (MEDIUM confidence)
- [CITED: gorilla/websocket chat example] — Hub pattern with register/unregister/broadcast channels; canonical reference for Go WebSocket servers
- [CITED: ARRL Field Day Cabrillo spec] — Station config fields (class, section, power level) used in station_config table design

### Tertiary (LOW confidence)
- [WebSearch] — coder/websocket vs gorilla/websocket comparison — verified via pkg.go.dev import counts
- [ASSUMED] — ARM/RPi compatibility of gorilla/websocket (no CGo, should work but untested on target hardware)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — gorilla/websocket is the definitive Go WebSocket library; verified via pkg.go.dev (44K importers, v1.5.3 is latest)
- Architecture: HIGH — Hub pattern is well-established; matches gorilla/websocket chat example; SQLite WAL concurrent readers already configured in Phase 1
- Pitfalls: MEDIUM — WebSocket write blocking and goroutine leaks are well-known; LAN-specific concerns (origin check, concurrent writes) verified against gorilla docs

**Research date:** 2026-05-29
**Valid until:** 2026-06-29 (30 days — gorilla/websocket is stable; station config fields are per ARRL spec which doesn't change mid-year)
