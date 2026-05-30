# Phase 03: Offline Resilience & Polish - Research

**Researched:** 2026-05-30
**Domain:** Offline-first SPAs (IndexedDB, Service Workers), Mobile-responsive CSS, Dark mode theming
**Confidence:** HIGH

## Summary

Phase 3 adds offline resilience and UI polish to the Field Day Logger. The frontend gains an IndexedDB persistence layer via Dexie.js 4.4.3, a Service Worker for app shell caching (SvelteKit built-in), and an offline QSO queue that syncs via a new `POST /api/sync` endpoint. The UI becomes mobile-responsive with CSS media queries at 480px/768px, large touch targets (48×48px minimum), and a dark mode theme toggled via CSS custom properties with localStorage persistence.

The existing codebase uses Svelte 5 runes (`$state`, `$derived`) exclusively — no Svelte stores. Dexie's `liveQuery()` returns an Observable that satisfies the Svelte store contract, but the project convention is to avoid stores. Instead, IndexedDB reads should populate `$state` arrays via `$effect` or manual subscription in `onMount`. The Go backend follows a consistent handler pattern (`func handler(db *sql.DB, hub *ws.Hub, w, r)`) with chi router wiring in `main.go`.

**Primary recommendation:** Use Dexie.js 4.4.3 for IndexedDB, SvelteKit's built-in Service Worker support (no Workbox), CSS custom properties in `app.css` for theming, and media queries (not container queries) for responsive layout. Keep `$state` arrays as the reactive source of truth; populate from IndexedDB on load and via WebSocket events.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Offline QSO buffering | Browser / Client | — | IndexedDB is browser-local; no server involvement |
| Offline dupe checking | Browser / Client | API / Backend | Exact-match check against IndexedDB when offline; server-side for similarity |
| Service Worker app shell | Browser / Client | — | SW runs entirely in browser; caches static build output |
| Batch sync on reconnect | Browser / Client | API / Backend | Client initiates POST; server processes batch insert |
| Connection status display | Browser / Client | — | Reads WebSocket state + queue length locally |
| Dark mode toggle | Browser / Client | — | CSS custom properties + localStorage; no server involvement |
| Mobile-responsive layout | Browser / Client | — | CSS media queries only; no server involvement |
| Debounce / rate limiting | Browser / Client | — | Client-side guard on submit handler |
| `/api/sync` endpoint | API / Backend | Database / Storage | Go handler processes batch, inserts into SQLite |
| `client_id` column | Database / Storage | — | New column in SQLite schema for UUID dedup |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Dexie.js | 4.4.3 | IndexedDB wrapper | 13K+ GitHub stars, 8+ years old, 2M+ weekly downloads, battle-tested offline-first DB. `liveQuery()` for observable queries. |
| SvelteKit Service Worker | Built-in (2.61.1) | SW bundling + registration | SvelteKit auto-bundles `src/service-worker.js`, provides `$service-worker` module with `build`/`files`/`version`. No external dependency. |
| CSS Custom Properties | Native (CSS3) | Theming system | Standard CSS — no library. `[data-theme='dark']` attribute toggle pattern. |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `@sveltejs/adapter-static` | 3.0.10 | SPA build output | Already configured. Produces static `frontend/build/` for `embed.FS`. SW file ends up in build output automatically. |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Dexie.js | Raw IndexedDB API | Raw API is verbose, callback-heavy, requires manual transaction management. Dexie reduces ~10:1 code ratio. |
| Built-in SW | Workbox + vite-pwa-plugin | Workbox adds ~30KB bundle, requires plugin config. SvelteKit built-in is zero-config, smaller, and uses Vite-hashed filenames from `$service-worker`. |
| CSS variables | CSS-in-JS / Tailwind | No build-time dependency. Already using scoped `<style>` blocks — CSS variables compose naturally with this pattern. |

**Installation:**
```bash
npm install dexie@^4.4.3
```

**Version verification:**
```
dexie: 4.4.3 (published 2026-05-27) — confirmed via npm view
@sveltejs/kit: 2.61.1 — confirmed via npm view
svelte: 5.56.0 — confirmed via npm view
```

## Package Legitimacy Audit

> **Required** — slopcheck was not available on this system. All packages below are tagged `[ASSUMED]`. The planner must gate each install behind a `checkpoint:human-verify` task.

| Package | Registry | Age | Downloads | Source Repo | slopcheck | Disposition |
|---------|----------|-----|-----------|-------------|-----------|-------------|
| dexie | npm | 8+ yrs | 2M+/wk | github.com/dexie/Dexie.js | [UNAVAILABLE] | Flagged — planner must add checkpoint |

**Packages removed due to slopcheck [SLOP] verdict:** none

**Packages flagged as suspicious [SUS]:** none

*slopcheck was unavailable at research time — all packages tagged `[ASSUMED]`. Planner must gate each install behind a `checkpoint:human-verify` task.*

**Postinstall script check:** `npm view dexie scripts.postinstall` returned empty — no suspicious postinstall detected.

## Architecture Patterns

### System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        BROWSER (Client)                         │
│                                                                 │
│  ┌──────────┐   ┌──────────────┐   ┌─────────────────────────┐ │
│  │ QSO Form │──▶│ Write Router │──▶│  online?  ──YES──▶ fetch │ │
│  │          │   │ (api.js)     │   │           │      POST    │ │
│  └──────────┘   └──────┬───────┘   │           │              │ │
│                        │           │  offline  │              │ │
│                        │           │   ──NO──▶ │              │ │
│                        │           └─────┬─────┘              │ │
│                        ▼                 ▼                     │ │
│              ┌──────────────────┐  ┌──────────┐               │ │
│              │  $state qsos[]   │  │ IndexedDB│               │ │
│              │  (reactive UI)   │  │ (Dexie)  │               │ │
│              └──────────────────┘  └────┬─────┘               │ │
│                                         │                      │ │
│  ┌──────────────────────────────────────┘                      │ │
│  │                                                              │ │
│  │  ┌──────────────────┐    ┌──────────────────────────────┐  │ │
│  │  │ Sync Trigger     │◀───│ wsState.connected → true     │  │ │
│  │  │ (ws.svelte.js)   │    │ + 30s periodic retry          │  │ │
│  │  └────────┬─────────┘    └──────────────────────────────┘  │ │
│  │           │                                                  │ │
│  │           ▼                                                  │ │
│  │  ┌──────────────────┐                                       │ │
│  │  │ POST /api/sync   │──▶ Server                            │ │
│  │  │ [queued QSOs]    │                                       │ │
│  │  └──────────────────┘                                       │ │
│  │                                                              │ │
│  │  ┌──────────────────────────────────────────────────────┐   │ │
│  │  │ Service Worker (src/service-worker.js)               │   │ │
│  │  │   - Cache-first: /_app/* (build assets)              │   │ │
│  │  │   - Network-first: /api/* (API calls)                │   │ │
│  │  │   - Stale-while-revalidate: / (index.html SPA shell) │   │ │
│  │  └──────────────────────────────────────────────────────┘   │ │
│  └──────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     GO SERVER (RPi / Laptop)                     │
│                                                                 │
│  ┌───────────────┐   ┌──────────────────────┐                  │
│  │ POST /api/sync│──▶│ handler.SyncQSOs()   │                  │
│  │               │   │ - Validate each QSO  │                  │
│  └───────────────┘   │ - Check client_id    │                  │
│                      │   dedup              │                  │
│  ┌───────────────┐   │ - Batch INSERT       │                  │
│  │ GET /ws       │   │ - Broadcast via hub  │                  │
│  └───────────────┘   │ - Return mappings    │                  │
│                      └────────┬─────────────┘                  │
│                               │                                 │
│                               ▼                                 │
│                      ┌──────────────────┐                      │
│                      │  SQLite (WAL)    │                      │
│                      │  qsos table      │                      │
│                      │  + client_id col  │                      │
│                      └──────────────────┘                      │
└─────────────────────────────────────────────────────────────────┘
```

### Recommended Project Structure (new files only)
```
frontend/src/
├── lib/
│   ├── db.js              # Dexie DB setup + schema (NEW)
│   └── sync.svelte.js     # Sync manager: queue, flush, retry (NEW)
├── service-worker.js      # SvelteKit SW: cache strategies (NEW)
├── app.css                # ADD: CSS custom properties + dark mode (MODIFY)
├── routes/+page.svelte    # ADD: queue count + dark mode toggle (MODIFY)

internal/
├── db/
│   └── schema.sql         # ADD: client_id TEXT UNIQUE column (MODIFY)
├── handler/
│   └── sync.go            # POST /api/sync handler (NEW)
└── go.mod                 # No new Go deps needed

main.go                    # ADD: r.Post("/sync", ...) route (MODIFY)
```

### Pattern 1: Dexie.js Database Setup with Svelte 5

**What:** A `.js` module that creates and exports the Dexie database instance with table schemas. In Svelte 5, IndexedDB reads populate `$state` arrays rather than using Svelte stores.

**When to use:** Any component that needs to read/write QSOs to IndexedDB.

**Example (from official Dexie docs):**
```javascript
// frontend/src/lib/db.js
import Dexie from 'dexie';

const db = new Dexie('FDLogger');

db.version(1).stores({
  // queued_qsos: QSOs created offline, waiting for sync
  queued_qsos: 'client_id, created_at',
  // cached_qsos: Full QSO history mirror for offline dupe checking
  cached_qsos: 'client_id, [callsign+band+mode], timestamp'
});

export { db };
```

**Svelte 5 integration pattern:**
```javascript
// In a .svelte.js module or component <script>:
import { db } from '$lib/db.js';

// $state array for UI reactivity, populated from IndexedDB
export const cachedQsos = $state([]);
export const queueLength = $state(0);

export async function loadCachedQsos() {
  cachedQsos.splice(0, cachedQsos.length, ...await db.cached_qsos.orderBy('timestamp').reverse().toArray());
}

export async function enqueueQso(qso) {
  const clientId = crypto.randomUUID();
  await db.queued_qsos.put({ client_id: clientId, qso, created_at: new Date().toISOString() });
  queueLength = await db.queued_qsos.count();
  return clientId;
}
```

**Key design decision:** Keep the existing `$state` `qsos` array as the single source of truth for the UI. On page load, populate it from IndexedDB (cached + queued). WebSocket events continue to push to `qsos` as today. IndexedDB is a mirror, not the reactive source.

### Pattern 2: SvelteKit Service Worker (Built-in)

**What:** SvelteKit auto-bundles `src/service-worker.js` and auto-registers it in the built `index.html`. The `$service-worker` module provides `build` (all Vite-generated files), `files` (static/ directory), and `version` (for cache invalidation).

**Sources:** [CITED: https://kit.svelte.dev/docs/service-workers] [CITED: https://kit.svelte.dev/docs/$service-worker]

**Cache strategy for this project:**
- **App shell (`/`, `/_app/*`):** Cache-first. These are hashed filenames — immutable.
- **API requests (`/api/*`):** Network-first, no cache. These are dynamic data. 
- **Static files:** Cache-first (from `$service-worker` `files` array).
- **Everything else:** Network-first, fallback to `/index.html` for SPA routing.

**File location:** `frontend/src/service-worker.js` (SvelteKit convention)

```javascript
// frontend/src/service-worker.js
/// <reference types="@sveltejs/kit" />
import { build, files, version } from '$service-worker';

const CACHE = `fdlogger-${version}`;
const ASSETS = [...build, ...files];

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE).then((cache) => cache.addAll(ASSETS))
  );
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((keys) => 
      Promise.all(keys.filter(k => k !== CACHE).map(k => caches.delete(k)))
    )
  );
});

self.addEventListener('fetch', (event) => {
  if (event.request.method !== 'GET') return;
  
  const url = new URL(event.request.url);
  
  // API requests: network-first
  if (url.pathname.startsWith('/api/')) {
    return; // Let browser handle — no caching
  }
  
  // App shell / static: cache-first
  event.respondWith(
    caches.match(event.request).then(cached => cached || fetch(event.request))
  );
});
```

**No Workbox needed.** The built-in approach is sufficient for this SPA's needs (cache app shell, let API pass through).

**Svelte 5 auto-registration caveat:** SvelteKit registers the SW via `navigator.serviceWorker.register('./service-worker.js')` in production. For dev, the SW is not bundled (modules-in-SW requirement). This is fine — offline features are tested in production builds.

### Pattern 3: Offline Sync Architecture

**What:** QSOs created while offline go to IndexedDB `queued_qsos` table. On WebSocket reconnect, a sync manager flushes the queue via `POST /api/sync`. A 30-second backup timer retries if the batch POST fails.

**Queue flow:**
```
Create QSO (offline)
  → enqueueQso() in db.js
  → Add to IndexedDB queued_qsos
  → Add to $state qsos[] for optimistic UI
  → Increment queue count

WebSocket reconnects (wsState.connected → true)
  → sync.svelte.js detects transition
  → Calls flushQueue()
  → Reads all queued_qsos from IndexedDB
  → POST /api/sync with array of {client_id, ...qso_fields}
  → On success: clear queued_qsos, update cached_qsos with returned server_ids
  → On failure: leave queue intact, 30s timer will retry
```

**Sync trigger implementation** (in `ws.svelte.js` or new `sync.svelte.js`):
```javascript
import { wsState } from '$lib/ws.svelte.js';
import { db } from '$lib/db.js';

let syncTimer = null;

$effect(() => {
  if (wsState.connected) {
    flushSyncQueue();
    // 30s periodic retry while connected + queue non-empty
    syncTimer = setInterval(async () => {
      const count = await db.queued_qsos.count();
      if (count > 0) flushSyncQueue();
      else if (syncTimer) { clearInterval(syncTimer); syncTimer = null; }
    }, 30000);
  } else {
    if (syncTimer) { clearInterval(syncTimer); syncTimer = null; }
  }
});

async function flushSyncQueue() {
  const queued = await db.queued_qsos.toArray();
  if (queued.length === 0) return;
  
  const response = await fetch('/api/sync', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ qsos: queued.map(q => ({ client_id: q.client_id, ...q.qso })) })
  });
  
  if (response.ok) {
    const { mappings } = await response.json(); // [{client_id, server_id}, ...]
    await db.queued_qsos.clear();
    // Update cached_qsos with server IDs
    // ...
  }
}
```

### Pattern 4: CSS Custom Properties & Dark Mode

**What:** Define all colors as CSS custom properties on `:root` in `app.css`. A `[data-theme='dark']` selector on `<body>` overrides them. Manual toggle persists preference in localStorage. Auto-detect `prefers-color-scheme` on first visit.

**Color palette audit (from current codebase):** 25 distinct colors across 6 files. Key colors to variable-ize:

| Current Value | Variable Name | Used For |
|--------------|---------------|----------|
| #1a3a6b | --color-primary | Header bg, stat value |
| #2266cc | --color-accent | Focus borders, submit btn |
| #1a7a1a | --color-success | Online status, save btn |
| #cc3300 | --color-danger | Offline status, dupe warn, rate value |
| #f5f5f5 | --color-bg | Page background |
| #ffffff | --color-surface | Form bg, config panel, card bg |
| #222222 | --color-text | Primary text |
| #555555 | --color-text-secondary | Labels, secondary text |
| #888888 | --color-text-muted | Empty states, placeholders |
| #e8f0fe | --color-highlight | Stats bar bg, hover bg |
| #c4d7f2 | --color-border-light | Subtle borders |
| #cccccc | --color-border | Input borders |
| #dddddd | --color-border-strong | Section borders |
| #f0f0f0 | --color-bg-alt | Search bar, alternating rows |

**Implementation in `app.css`:**
```css
:root {
  --color-primary: #1a3a6b;
  --color-accent: #2266cc;
  --color-success: #1a7a1a;
  --color-danger: #cc3300;
  --color-bg: #f5f5f5;
  --color-surface: #ffffff;
  --color-text: #222222;
  --color-text-secondary: #555555;
  --color-text-muted: #888888;
  --color-highlight: #e8f0fe;
  --color-border-light: #c4d7f2;
  --color-border: #cccccc;
  --color-border-strong: #dddddd;
  --color-bg-alt: #f0f0f0;
}

[data-theme='dark'] {
  --color-primary: #4a7ab5;
  --color-accent: #5599ee;
  --color-success: #2ecc40;
  --color-danger: #ff4444;
  --color-bg: #1a1a2e;
  --color-surface: #16213e;
  --color-text: #e0e0e0;
  --color-text-secondary: #a0a0a0;
  --color-text-muted: #707070;
  --color-highlight: #1a3a5c;
  --color-border-light: #2a3a5c;
  --color-border: #3a4a6c;
  --color-border-strong: #4a5a7c;
  --color-bg-alt: #1f2d44;
}
```

**Theme toggle logic:**
```javascript
// In +page.svelte or as a $state in a .svelte.js module:
export const theme = $state({ value: 'light' });

function initTheme() {
  const stored = localStorage.getItem('fdlogger_theme');
  if (stored) {
    theme.value = stored;
  } else if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
    theme.value = 'dark';
  }
  applyTheme();
}

function toggleTheme() {
  theme.value = theme.value === 'light' ? 'dark' : 'light';
  localStorage.setItem('fdlogger_theme', theme.value);
  applyTheme();
}

function applyTheme() {
  document.body.setAttribute('data-theme', theme.value);
}
```

### Pattern 5: Mobile Responsive Breakpoints

**What:** CSS media queries at two breakpoints: 480px (phones) and 768px (tablets). Stack & shrink strategy — no JS layout changes.

**Breakpoint rules:**
- **≤480px (phones):** QSO form becomes 2-row wrap. Row 1: callsign (full-width) + exchange (full-width). Row 2: band + mode + submit button. Log table hides Operator and Time columns. Font 13px in table.
- **≤768px (tablets):** Three panels remain but adjust spacing. Form inputs grow to fill width. Stats bar wraps.
- **>768px:** Current desktop layout unchanged.

**Touch target rule:** All interactive elements (buttons, inputs, selects) need `min-height: 48px` and `min-width: 48px` (Material Design guideline). Current inputs at `padding: 10px 12px` plus font-size 16px give ~44px height — need `min-height: 48px`.

**Implementation approach:** Add media queries to each component's `<style>` block where layout changes are needed. No separate CSS files. No CSS-in-JS.

### Anti-Patterns to Avoid

- **Using `localStorage` for QSO data:** IndexedDB is for structured data. localStorage is synchronous, limited to 5-10MB, and string-only. Do not store QSO arrays in localStorage.
- **Workbox dependency:** SvelteKit's built-in SW support is sufficient for app shell caching. Workbox adds unnecessary bundle weight for this use case.
- **`navigator.onLine` for connectivity detection:** This only detects network interface status, not server reachability. Use the existing WebSocket connection state (`wsState.connected`) as the authoritative source.
- **Polling for sync:** The WebSocket reconnect event provides a precise sync trigger. The 30-second timer is a backup for failed batch POSTs, not the primary mechanism.
- **Duplicating QSO data in multiple stores:** Keep `$state qsos[]` as the single UI source. IndexedDB mirrors it; queued QSOs are added to both `qsos[]` and IndexedDB simultaneously.
- **Hand-written IndexedDB transactions:** Dexie.js handles transactions automatically. Use `db.transaction('rw', ...)` for multi-table atomic operations.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| IndexedDB wrapper | Raw IndexedDB API (verbose transactions, manual cursor iteration, error-prone key paths) | Dexie.js 4.4.3 | 10:1 code reduction, automatic transaction scoping, `liveQuery()` for reactive reads, compound indexes |
| Service Worker build integration | Manual webpack/rollup SW bundling, custom `sw-precache` scripts | SvelteKit built-in `src/service-worker.js` | Auto-bundled via Vite, auto-registered in build output, `$service-worker` module provides hashed filenames |
| Dupe checking in IndexedDB | Custom B-tree or Map/Set dedup logic | Dexie compound index `[callsign+band+mode]` + `db.cached_qsos.where(...)` | Dexie indexes handle exact-match lookups efficiently; compound index matches the existing server-side dupe query |
| Batch insert dedup in Go | Manual `SELECT` + `INSERT` loop per item | `INSERT ... ON CONFLICT(client_id) DO NOTHING` | SQLite UPSERT is atomic, one query handles all items. `ON CONFLICT` with UNIQUE index on `client_id` prevents double-insertion |
| CSS theme toggle persistence | Custom cookie/query-string approach | `localStorage` + `[data-theme]` attribute + `prefers-color-scheme` fallback | Standard pattern, zero dependencies, instant on page load (no flash of wrong theme) |
| Queue retry logic | Custom exponential backoff, persistent retry queue | Simple 30s `setInterval` + reconnect-triggered flush | Field Day LAN environment: reconnections are either immediate (WiFi back) or persistent (server reboot). Exponential backoff overcomplicates. |

**Key insight:** The three most deceptively complex problems in this phase are IndexedDB (raw API is famously verbose), Service Worker registration (cache invalidation is tricky), and batch sync deduplication (client_id mapping must survive re-submission of the same queue). Each has a standard, well-tested solution that should not be rebuilt.

## Runtime State Inventory

> This phase is a greenfield feature addition (not a rename/refactor). Existing state is not being renamed.

**Nothing to inventory.** No stored data, live service configs, OS-registered state, secrets, or build artifacts carry names that will change in this phase. New state will be created:
- IndexedDB database `FDLogger` (new)
- localStorage key `fdlogger_theme` (new)
- SQLite column `client_id` on `qsos` table (new addition to existing table)

## Common Pitfalls

### Pitfall 1: Service Worker Caches Stale API Responses
**What goes wrong:** Service Worker's `fetch` handler inadvertently caches `/api/*` responses, causing users to see stale QSO data after reconnect.
**Why it happens:** Default cache-first strategy applied to all GET requests without path filtering.
**How to avoid:** In the SW `fetch` handler, explicitly skip caching for any URL path starting with `/api/`. Let network requests pass through unmodified.
**Warning signs:** QSOs appear after sync but disappear on reload, or dupe warnings show stale data.

### Pitfall 2: Dexie `liveQuery` with Svelte 5 Runes Mode
**What goes wrong:** Dexie's `liveQuery()` returns an Observable with a `.subscribe()` method, which satisfies Svelte's store contract and works with the `$` auto-subscription prefix. However, in Svelte 5 runes mode, mixing store-style auto-subscription (`$myLiveQuery`) with `$state` in the same component can cause confusion about which reactivity model applies where.
**Why it happens:** The project already uses `$state` and `$derived` runes exclusively. Introducing `liveQuery()`'s store-based auto-subscription creates an inconsistent reactivity pattern.
**How to avoid:** Do NOT use `liveQuery()` with the `$` prefix. Instead, manually subscribe in `onMount` and populate `$state` arrays. This keeps all reactive data in the `$state`/`$derived` runes model. Example:
```javascript
import { onMount } from 'svelte';
import { db } from '$lib/db.js';
let cachedQsos = $state([]);

onMount(() => {
  db.cached_qsos.orderBy('timestamp').reverse().toArray().then(data => {
    cachedQsos.splice(0, cachedQsos.length, ...data);
  });
});
```
**Warning signs:** `$` prefix used on Dexie queries, reactivity "works but feels wrong," data appears in UI but doesn't update on mutation.

### Pitfall 3: SQLite `client_id` Column Without UNIQUE Constraint
**What goes wrong:** The sync endpoint receives the same queue twice (e.g., 30s retry fires before first batch POST response returns). Without a UNIQUE constraint, duplicate QSOs are inserted.
**Why it happens:** Network latency or slow batch insert can cause the 30s timer to fire while the previous sync is still in-flight.
**How to avoid:** Add `client_id TEXT UNIQUE` to the `qsos` table. Use `INSERT ... ON CONFLICT(client_id) DO NOTHING` in the batch insert query. The client_id is a UUID from `crypto.randomUUID()` — globally unique.
**Warning signs:** Duplicate QSOs appearing in the log table after a sync, or QSO count jumping by more than the queue size.

### Pitfall 4: Mobile Media Query Cascade Conflicts
**What goes wrong:** Adding media queries to individual component `<style>` blocks creates specificity conflicts when the same element is affected by multiple breakpoint rules.
**Why it happens:** Svelte's scoped CSS adds unique hashes to selectors, but media queries within scoped styles can still conflict when parent and child components both define rules for similar elements at the same breakpoint.
**How to avoid:** Place all responsive layout overrides in the component that owns the elements. Don't split 480px rules for the form across `+page.svelte` and `QsoEntryForm.svelte`. The form layout is owned by `QsoEntryForm.svelte` — all its responsive rules go there.
**Warning signs:** Some mobile styles apply, others don't, and the cascade order is unpredictable between dev builds and production.

### Pitfall 5: Debounce Timer Memory Leak
**What goes wrong:** The existing `debounceTimer` in `LogTable.svelte` is not cleaned up on component destroy. The new sync retry `setInterval` has the same risk.
**Why it happens:** Svelte 5's `onMount` doesn't auto-return cleanup — you must use `$effect` or explicit `return` from `onMount`.
**How to avoid:** Use `$effect` for timers (auto-cleanup on re-run and destroy). For the sync retry:
```javascript
$effect(() => {
  if (!wsState.connected) return;
  const timer = setInterval(trySync, 30000);
  return () => clearInterval(timer);
});
```
**Warning signs:** Console shows errors from timers firing after hot-reload or page navigation.

## Code Examples

Verified patterns from official sources:

### Dexie.js Database Declaration
```javascript
// Source: https://dexie.org/docs/API-Reference (Quick Reference section)
import Dexie from 'dexie';

const db = new Dexie('FDLogger');
db.version(1).stores({
  queued_qsos: 'client_id, created_at',
  cached_qsos: 'client_id, [callsign+band+mode], timestamp'
});

// Compound index [callsign+band+mode] enables efficient dupe queries:
// await db.cached_qsos.where({callsign: 'K1ABC', band: '20M', mode: 'SSB'}).count()
```

### Dexie BulkAdd with Error Handling
```javascript
// Source: https://dexie.org/docs/Table/Table.bulkAdd()
await db.queued_qsos.bulkAdd(queuedItems).catch(error => {
  if (error.name === 'BulkError') {
    console.warn('Some items failed:', error.failures.length);
  }
});
```

### CSS Custom Properties with Dark Mode Toggle
```css
/* Source: https://svelte.dev/docs/svelte/custom-properties (Svelte CSS custom properties) */
/* app.css */
:root {
  --color-bg: #f5f5f5;
  --color-text: #222;
}
[data-theme='dark'] {
  --color-bg: #1a1a2e;
  --color-text: #e0e0e0;
}
/* Component <style> blocks reference variables: */
/* background: var(--color-bg); */
/* color: var(--color-text); */
```

### Go Handler Pattern (from existing codebase)
```go
// Source: internal/handler/qso.go (CreateQSO pattern)
func SyncQSOs(db *sql.DB, hub *ws.Hub, w http.ResponseWriter, r *http.Request) {
    var input struct {
        QSOs []struct {
            ClientID     string `json:"client_id"`
            Callsign     string `json:"callsign"`
            Band         string `json:"band"`
            Mode         string `json:"mode"`
            RecvExchange string `json:"recv_exchange"`
            SentExchange string `json:"sent_exchange"`
            Operator     string `json:"operator"`
        } `json:"qsos"`
    }
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        // ... error handling
    }
    // ... batch insert with ON CONFLICT
}
```

### Chi Router Wiring (from existing main.go pattern)
```go
// Source: main.go (existing route wiring pattern)
r.Route("/api", func(r chi.Router) {
    // ... existing routes ...
    r.Post("/sync", func(w http.ResponseWriter, r *http.Request) {
        handler.SyncQSOs(database, hub, w, r)
    })
})
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Svelte stores (`writable`, `derived`) | Svelte 5 runes (`$state`, `$derived`) | Svelte 5 (2024) | This project already uses runes — maintain consistency |
| Workbox for Service Workers | SvelteKit built-in `$service-worker` | SvelteKit 1.0+ | No external dep, smaller bundle, tighter Vite integration |
| Raw IndexedDB | Dexie.js | 2014+ (stable) | Battle-tested, well-documented, Svelte-compatible via store contract |
| `navigator.onLine` | WebSocket connection state | Always preferred | `onLine` detects interface, not server reachability |

**Deprecated/outdated:**
- **`applicationCache` (AppCache):** Removed from browsers. Service Workers are the replacement.
- **`localStorage` for structured data:** Still functional but IndexedDB is the correct tool for QSO storage. localStorage is synchronous and blocks the main thread.
- **`@sveltejs/adapter-auto`:** Already replaced with `@sveltejs/adapter-static` in this project's `svelte.config.js`. Static adapter is correct for the `embed.FS` deployment model.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Dexie.js 4.4.3 is compatible with Svelte 5 runes mode without conflicts | Standard Stack | Low — Dexie is framework-agnostic; only interaction point is the Observable/store contract, which Svelte 5 still supports |
| A2 | SvelteKit's built-in Service Worker support works correctly with `@sveltejs/adapter-static` (SPA mode) | Service Worker | Low — officially documented combination; static adapter is the standard SPA output method |
| A3 | IndexedDB has sufficient storage for ~3000 QSOs on all target devices (phones, tablets, RPi browser) | Local Cache | Low — stated assumption in CONTEXT.md (1-2MB for 3000 QSOs); IndexedDB limits are typically 50% of disk space, well over 100MB on any modern device |
| A4 | `crypto.randomUUID()` is available in all target browsers (used for client_id generation) | Offline Sync | Low — supported in all modern browsers since Chrome 92, Firefox 95, Safari 15.4 (2021-2022) |
| A5 | The dark mode palette values in this research are reasonable defaults for tent/dim-lighting use | Dark Mode | Medium — palette values are [ASSUMED] based on training knowledge. User may want different dark mode colors. The planner should note that these are starting points. |
| A6 | 480px and 768px breakpoints cover the target device range (phones 320-430px, tablets 600-900px) | Mobile Responsive | Low — stated in CONTEXT.md D-09. These are standard breakpoints for mobile-first design. |

## Open Questions

1. **Single source of truth: IndexedDB vs $state array?**
   - What we know: CONTEXT.md gives the agent discretion. Current code uses `$state qsos[]` + `seenIds` Set. CONTEXT.md D-08 says cache is populated on load via `GET /api/qso?limit=9999` then kept current via WebSocket events.
   - What's unclear: Whether to keep `$state qsos[]` as primary and mirror to IndexedDB, or make IndexedDB the primary with reactive bindings.
   - Recommendation: Keep `$state qsos[]` as primary UI source, mirror to IndexedDB. This minimizes changes to existing component logic and maintains the current reactivity model. IndexedDB is a write-through cache, not the reactive source.

2. **Service Worker update flow for field deployment?**
   - What we know: Service Workers require a page refresh to activate new versions (the "waiting" state). On a LAN with no internet, this is less of a concern since there are no CDN deployments.
   - What's unclear: Whether the deployment workflow needs a "skip waiting" pattern or a manual "new version available" prompt.
   - Recommendation: Use `self.skipWaiting()` in the `install` event and `clients.claim()` in `activate`. In a LAN field deployment, there's exactly one server version at a time — old SW versions with stale caches are not a concern. Keep it simple.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Node.js | Frontend build (Dexie, SvelteKit, Vite) | ✓ | v22.22.3 | — |
| npm | Package installation | ✓ | 10.9.8 | — |
| Go | Backend server build | ✓ | 1.26.3 | — |
| SQLite | Database (via modernc.org/sqlite, pure Go) | ✓ | Embedded in Go binary | — |
| Chromium/Chrome | Service Worker + IndexedDB testing | Not checked | — | Flag for human — browser needed for manual testing |

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | vitest 4.1.7 + jsdom 29.1.1 + @testing-library/svelte 5.3.1 |
| Config file | `frontend/vitest.config.ts` |
| Quick run command | `npx vitest run --reporter=verbose` |
| Full suite command | `npx vitest run` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| SYNC-03 | Client buffers QSOs to IndexedDB when server unreachable | unit | `npx vitest run src/lib/db.test.js -t "buffers qso when offline"` | ❌ Wave 0 |
| SYNC-04 | Buffered QSOs sync on reconnect | integration | `npx vitest run src/lib/sync.test.js -t "flushes queue on connect"` | ❌ Wave 0 |
| SYNC-05 | Connection status indicator shows online/offline | unit | `npx vitest run src/routes/page.test.js -t "shows connection status"` | ❌ Wave 0 |
| SYNC-06 | Dupe checking works against local cache when offline | unit | `npx vitest run src/lib/db.test.js -t "dupe check uses IndexedDB"` | ❌ Wave 0 |
| UX-01 | Mobile-responsive layout with touch targets | manual-only | Manual browser resize + touch simulator | N/A (manual) |
| UX-02 | Dark mode renders across all components | manual-only | Manual toggle + visual inspection | N/A (manual) |
| UX-04 | App loads from cache when server unavailable | manual-only | Manual: stop server, reload page, verify SW serves cached SPA | N/A (manual) |

### Sampling Rate
- **Per task commit:** `npx vitest run --reporter=verbose`
- **Per wave merge:** `npx vitest run`
- **Phase gate:** Full test suite green + manual browser testing for UX requirements

### Wave 0 Gaps
- [ ] `frontend/src/lib/db.test.js` — covers SYNC-03, SYNC-06 (IndexedDB buffering + dupe check)
- [ ] `frontend/src/lib/sync.test.js` — covers SYNC-04 (batch sync on reconnect)
- [ ] `frontend/src/routes/page.test.js` — covers SYNC-05 (connection indicator component)
- [ ] `frontend/src/service-worker.js` — Service Worker file (not testable in vitest; manual verification)
- [ ] `internal/handler/sync_test.go` — Go handler test for POST /api/sync
- [ ] Dexie.js install: `npm install dexie@^4.4.3` — package not yet installed
- [ ] Vitest config may need update for Dexie (IndexedDB mocking — dexie exports work with jsdom's indexedDB shim, or fake-indexeddb may be needed)

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | No auth for trusted LAN per AGENTS.md |
| V3 Session Management | no | No sessions — stateless SPA |
| V4 Access Control | no | Trusted LAN — no access control needed |
| V5 Input Validation | yes | Existing validation: `model.ValidateRequired()` in Go, `validateCallsign()` in Svelte. Sync endpoint reuses same validation. |
| V6 Cryptography | no | No cryptographic operations in this phase |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Duplicate QSO insertion via sync replay | Tampering | `client_id TEXT UNIQUE` with `ON CONFLICT DO NOTHING` in SQLite — idempotent batch insert |
| IndexedDB data tampering via browser console | Tampering | Client-side data is non-authoritative; server validates on sync. QSOs with invalid callsign/band/mode are rejected by server validation. |
| Service Worker serving stale/malicious cache | Spoofing | Hashed filenames (`$service-worker` version) ensure cache matches deployed build. No runtime cache of `/api/*` responses. |
| Queue poisoning (injecting fake QSOs into IndexedDB) | Tampering | Server-side validation on `POST /api/sync` (required fields, callsign format) rejects invalid entries. IndexedDB is client-local — attacker would need physical device access. |

## Sources

### Primary (HIGH confidence)
- [Dexie.js Official Docs](https://dexie.org/docs/API-Reference) — Schema syntax, CRUD operations, bulkAdd, Table API
- [Dexie.js liveQuery() Docs](https://dexie.org/docs/liveQuery()) — Observable queries, Svelte integration example
- [SvelteKit Service Workers Docs](https://kit.svelte.dev/docs/service-workers) — Built-in SW support, `$service-worker` module, cache strategies
- [SvelteKit $service-worker Reference](https://kit.svelte.dev/docs/$service-worker) — `build`, `files`, `version`, `prerendered` exports
- [Svelte 5 $state Docs](https://svelte.dev/docs/svelte/$state) — Runes reactivity model, deep Reactivity, passing state across modules
- [Svelte 5 Stores Docs](https://svelte.dev/docs/svelte/stores) — Store contract, `$` prefix auto-subscription, runes vs stores guidance
- [npm registry (dexie)](https://www.npmjs.com/package/dexie) — Version 4.4.3 confirmed, published 2026-05-27
- Existing codebase: `frontend/src/lib/api.js`, `ws.svelte.js`, `stores/qso.svelte.js`, `QsoEntryForm.svelte`, `LogTable.svelte`, `main.go`, `internal/handler/qso.go`, `internal/db/schema.sql`

### Secondary (MEDIUM confidence)
- [MDN: Using Service Workers](https://developer.mozilla.org/en-US/docs/Web/API/Service_Worker_API/Using_Service_Workers) — General SW patterns referenced but not directly cited
- [Material Design Touch Targets](https://m3.material.io/foundations/accessible-design/accessibility-basics) — 48×48px minimum touch target standard

### Tertiary (LOW confidence)
- Dark mode color palette values — [ASSUMED] based on training knowledge for dim-lighting field conditions; not verified against a design system

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — Dexie.js verified via npm registry + official docs; SvelteKit SW verified via official docs; CSS variables are standard CSS
- Architecture: HIGH — existing codebase patterns thoroughly surveyed; Dexie schema pattern verified from official docs; Go handler pattern confirmed from existing code
- Pitfalls: HIGH — based on documented Svelte 5 reactivity model, SvelteKit SW caveats, and SQLite ON CONFLICT semantics

**Research date:** 2026-05-30
**Valid until:** 2026-07-30 (30 days — stable libraries, no fast-moving APIs in this stack)
