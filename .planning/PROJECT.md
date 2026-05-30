# Field Day Logger

## What This Is

A lightweight, mobile-friendly, multi-user web-based logging application purpose-built for ARRL Field Day. Operators on separate devices log QSOs to a shared local server (Raspberry Pi or laptop) over a LAN. Designed for tent-based deployment with offline-first resilience — logging continues uninterrupted when WiFi or power drops.

## Core Value

Operators can log QSOs even when the network goes down, with all data syncing automatically when reconnected. The log must never be lost.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] Quick-entry QSO log form (callsign, band, mode, exchange) with keyboard shortcuts
- [ ] Real-time dupe checking (band+mode) with inline warning
- [ ] Live rate meter (QSOs/hour, peak rate, running total)
- [ ] Multi-user LAN support (2–6 operators, shared database)
- [ ] Offline resilience (local IndexedDB buffer, sync on reconnect)
- [ ] One-click Cabrillo export in valid ARRL Field Day format
- [ ] Live score display (raw points, multiplier, bonus, estimated score)
- [ ] Band/mode breakdown panel
- [ ] Station configuration (class, section, power, transmitter count)
- [ ] Dupe warning against locally cached QSOs when offline
- [ ] Mobile-responsive UI with large touch targets
- [ ] Dark mode theme

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

**Technical environment:** A Raspberry Pi 4 or Linux laptop acting as a LAN server. 2–6 client devices (laptops, tablets, phones) connect via WiFi router. Internet is unreliable or absent during the event.

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
| Go backend (chi + gorilla/websocket) | Single binary, low memory, easy deploy on RPi | — Pending |
| SvelteKit frontend | Small bundles (~8KB), reactive, excellent offline support | — Pending |
| SQLite with WAL mode | No server process, perfect for single-server LAN deployment | — Pending |
| Offline-first architecture (IndexedDB + Dexie.js) | QSOs written locally first, synced to server when connected | — Pending |
| SPA + WebSockets for real-time | Instant dupe check requires local state; WebSockets for live multi-user updates | — Pending |
| Svelte stores for state management | Lightweight reactive state sufficient for ~3000 QSOs max | — Pending |
| No auth for LAN use | Everyone in the tent is trusted | — Pending |

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
*Last updated: 2026-05-29 after initialization*
