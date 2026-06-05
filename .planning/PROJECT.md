# Field Day Logger

## What This Is

A lightweight, mobile-friendly, multi-user web-based logging application purpose-built for ARRL Field Day. Operators on separate devices log QSOs to a shared local server (Raspberry Pi or laptop) over a LAN. Designed for tent-based deployment with offline-first resilience — logging continues uninterrupted when WiFi or power drops.

## Core Value

Operators can log QSOs even when the network goes down, with all data syncing automatically when reconnected. The log must never be lost.

## Current State

**Shipped:** v1.0.0 on 2026-06-05 — 4 phases, 19 plans, ~20,500 LOC (Go + SvelteKit + JS)

**Tech stack:** Go backend (chi router, gorilla/websocket) + SvelteKit SPA frontend + SQLite (WAL mode). Single-binary deploy on Raspberry Pi 4 or Linux laptop. LAN-only, no internet dependency.

**What's built:**
- Full QSO logging with keyboard shortcuts, dupe detection, and inline editing
- Real-time multi-user sync via WebSocket hub (channel-based broadcast)
- Station configuration (callsign, class, section, power) with UI form
- Operator identity per client session (localStorage)
- Offline resilience: Dexie.js IndexedDB buffer, batch POST /api/sync, auto-reconnect
- ARRL Field Day Cabrillo v3.0 export with config-driven headers
- Live scoreboard with rate meter, band/mode breakdown, and bonus points
- Bonus tracker: 18-item ARRL 2026 bonus list with localStorage persistence
- Web Audio API beeps for QSO confirmation and dupe warnings
- Dark mode CSS theme with 14 custom properties, no hardcoded colors
- Mobile-responsive layout (48px touch targets, 500px/768px breakpoints)
- Service Worker with versioned app shell caching
- One-click SQLite database backup
- Full test suites: Go (race-clean), vitest + jsdom, multi-client integration simulation

## Requirements

### Validated

- ✓ Quick-entry QSO log form with keyboard shortcuts — v1.0.0
- ✓ Real-time dupe checking (band+mode) with inline warning — v1.0.0
- ✓ Live rate meter (QSOs/hour, peak rate, running total) — v1.0.0
- ✓ Multi-user LAN support (2-6 operators, shared database) — v1.0.0
- ✓ Offline resilience (local IndexedDB buffer, sync on reconnect) — v1.0.0
- ✓ One-click Cabrillo export in valid ARRL Field Day format — v1.0.0
- ✓ Live score display (raw points, multiplier, bonus, estimated score) — v1.0.0
- ✓ Band/mode breakdown panel — v1.0.0
- ✓ Station configuration (class, section, power, transmitter count) — v1.0.0
- ✓ Dupe warning against locally cached QSOs when offline — v1.0.0
- ✓ Mobile-responsive UI with large touch targets — v1.0.0
- ✓ Dark mode theme — v1.0.0
- ✓ Audio alerts for QSO confirmation and dupe warning — v1.0.0
- ✓ Bonus points tracker (claim/unclaim FD bonuses) — v1.0.0
- ✓ One-click database backup (SQLite file export) — v1.0.0

### Active

_(Ready for next milestone planning)_

### Out of Scope

- CW keyer integration (Winkeyer) — hardware-dependent, deferred
- Voice keyer — deferred
- Bandmap / cluster integration — deferred
- ADIF import/export — deferred
- Multi-contest support (non-FD contests) — future
- Live dashboard / big screen projector view — deferred
- Propagation map — future
- OAuth or complex authentication — LAN-only, shared secret if needed
- Docker deployment — single binary + static files preferred

## Context

**Domain:** Amateur radio contest logging for ARRL Field Day, an annual 27-hour event on the fourth full weekend of June. Stations operate portable (tents, parks) on generator/battery/solar power with minimal infrastructure.

**Technical environment:** A Raspberry Pi 4 or Linux laptop acting as a LAN server. 2-6 client devices (laptops, tablets, phones) connect via WiFi router. Internet is unreliable or absent during the event.

**Key Field Day rules:**
- Exchange format: station class + ARRL section (e.g., "2A NH")
- Scoring: CW/Digital = 2 pts, Phone = 1 pt. Total = (Raw + Bonus) × Power Multiplier
- Dupe rule: same station on same band AND mode = zero points
- Submission format: Cabrillo via ARRL web portal within 30 days

**Prior art:** N1MM (Windows-only, heavy), Cloudlog (general-purpose, not FD-specific). No existing web logger is purpose-built for Field Day's tent-based, offline-first reality.

## Constraints

- **Offline resilience**: Must log without server connectivity, sync when reconnected
- **Hardware**: Runs on Raspberry Pi 4 (4GB) or old Linux laptop
- **Storage**: SQLite single-file database, WAL mode
- **Network**: LAN-only, no internet dependency, no CORS needed
- **UI**: Must work on phones/tablets with gloves or wet/dirty fingers
- **Deployment**: Single Go binary + static SPA files, systemd service
- **Auth**: None for trusted LAN; simple shared password if open WiFi
- **Browser**: Modern browsers with IndexedDB and Service Worker support

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go backend (chi + gorilla/websocket) | Single binary, low memory, easy deploy on RPi | ✓ Good — compiles to ~15MB binary, runs on RPi 4 |
| SvelteKit frontend | Small bundles (~8KB), reactive, excellent offline support | ✓ Good — Svelte 5 runes work well, Vitest + jsdom viable |
| SQLite with WAL mode | No server process, perfect for single-server LAN deployment | ✓ Good — zero-config, concurrent reads work with WAL |
| Offline-first architecture (IndexedDB + Dexie.js) | QSOs written locally first, synced to server when connected | ✓ Good — Dexie.js v4 solid, sync dedup via client_id |
| SPA + WebSockets for real-time | Instant dupe check requires local state; WebSockets for live multi-user updates | ✓ Good — gorilla/websocket hub handles broadcast cleanly |
| Svelte stores for state management | Lightweight reactive state sufficient for ~3000 QSOs max | ✓ Good — $state runes + exported objects for reactivity |
| No auth for LAN use | Everyone in the tent is trusted | ✓ Good — zero friction for operators |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---

_Last updated: 2026-06-05 after v1.0.0 milestone_
