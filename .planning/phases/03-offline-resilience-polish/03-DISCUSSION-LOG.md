# Phase 3: Offline Resilience & Polish - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-30
**Phase:** 3-Offline Resilience & Polish
**Areas discussed:** Offline sync & queue, Local cache & offline dupe, Mobile layout, Dark mode

---

## Offline Sync & Queue

| Option | Description | Selected |
|--------|-------------|----------|
| Batch POST /api/sync | Collect all queued QSOs, send in one batch POST to new /api/sync endpoint | ✓ |
| Individual POST with retry | Fire each queued QSO as individual POST /api/qso with retry on failure | |

| Option | Description | Selected |
|--------|-------------|----------|
| Client-generated UUID | crypto.randomUUID() at creation time, stored as client_id on server | ✓ |
| Temporary local index only | IndexedDB auto-increment key, mapped to server ID after sync | |

| Option | Description | Selected |
|--------|-------------|----------|
| Show pending count in status bar | "X queued" next to connection indicator, "Syncing..." during POST | ✓ |
| Dismissible banner | Banner at top showing queue count | |
| No queue indicator | QSOs just appear in log table | |

| Option | Description | Selected |
|--------|-------------|----------|
| On reconnect + periodic (30s) | Fire on wsState.connected flip + every 30s while queue non-empty | ✓ |
| On reconnect only | Fire once when wsState.connected becomes true | |
| Reconnect + manual button | Auto-sync plus "Sync now" button | |

---

## Local Cache & Offline Dupe

| Option | Description | Selected |
|--------|-------------|----------|
| Full QSO history | IndexedDB mirrors all QSOs from server (~3000 max, ~1-2MB) | ✓ |
| Recent window (last 500) | Cache only most recent 500 QSOs | |
| Own operator's QSOs only | Only cache this operator's QSOs | |

| Option | Description | Selected |
|--------|-------------|----------|
| Exact match only | Check callsign + band + mode against IndexedDB | ✓ |
| Full parity with server | Both exact match and partial call similarity (Levenshtein) | |

| Option | Description | Selected |
|--------|-------------|----------|
| Include queued QSOs | Dupe check scans both synced QSOs and local queue | ✓ |
| Only synced QSOs | Only check server-synced QSOs | |

| Option | Description | Selected |
|--------|-------------|----------|
| Page load fetch + WebSocket | GET /api/qso?limit=9999 on load, WebSocket events keep current | ✓ |
| Polling every N seconds | Periodically re-fetch all QSOs | |

---

## Mobile Layout

| Option | Description | Selected |
|--------|-------------|----------|
| Stack & shrink | All panels stack vertically, responsive sizing, 480px/768px breakpoints | ✓ |
| Tabbed view | Three tabs: Log, Stats, Table — tap to switch | |
| Simplified mobile view | Show only QSO form, stats/table behind drawer | |

| Option | Description | Selected |
|--------|-------------|----------|
| 48×48px (Material) | Google Material Design minimum touch target | ✓ |
| 44×44px (Apple) | Apple HIG minimum | |
| No strict minimum | Use responsive sizing, no explicit target | |

| Option | Description | Selected |
|--------|-------------|----------|
| Full-width 2-row wrap | Row 1: callsign + exchange, Row 2: band + mode + submit | ✓ |
| Single row, scroll | Keep desktop layout, horizontal scroll | |
| Simplified form | Drop exchange field on mobile | |

| Option | Description | Selected |
|--------|-------------|----------|
| Hide low-priority columns | Hide Operator and Time, keep 5 essential columns at 13px font | ✓ |
| Card layout | Switch each QSO to a card | |
| Horizontal scroll | Keep all 7 columns, scroll horizontally | |

---

## Dark Mode

| Option | Description | Selected |
|--------|-------------|----------|
| CSS custom properties + toggle | :root vars, [data-theme='dark'] override, localStorage persistence | ✓ |
| prefers-color-scheme only | @media query in every component, no manual toggle | |
| CSS class-based | .dark class on body with nesting overrides | |

| Option | Description | Selected |
|--------|-------------|----------|
| Header bar, far left | Between "FD Logger" title and WS status indicator | ✓ |
| Header bar, far right | Next to Export Cabrillo button | |
| Floating action button | Bottom-right corner icon | |

| Option | Description | Selected |
|--------|-------------|----------|
| Match OS, fallback light | Check prefers-color-scheme on first load, manual toggle persists | ✓ |
| Always start light | Ignore OS preference | |
| Always start dark | Start dark for night vision | |

| Option | Description | Selected |
|--------|-------------|----------|
| All at once, single pass | Define all variables in app.css, replace every component at once | ✓ |
| Incremental, component by component | One component per plan, mixed state during transition | |

---

## the agent's Discretion

- Dexie.js version and exact Dexie table schema
- Specific CSS variable names and dark mode color values
- Media query breakpoint pixel values
- Service Worker implementation (Workbox vs hand-written, cache strategy)
- `/api/sync` endpoint request/response format
- Periodic 30s retry implementation
- Connection indicator UI details
- Server-side `client_id` column addition to QSO table
- Whether reactive `$state` array coexists with IndexedDB or IndexedDB is single source of truth

## Deferred Ideas

None — discussion stayed within Phase 3 scope.
