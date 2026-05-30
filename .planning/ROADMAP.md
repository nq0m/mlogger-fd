# Roadmap: Field Day Logger

**Created:** 2026-05-29
**Phases:** 4
**Requirements:** 24 v1 requirements mapped

---

### Phase 1: Core Logger [COMPLETE]

**Goal:** Single operator can log QSOs, see rate/scoring, and export a valid Cabrillo file.
**Mode:** mvp
**Plans:** 5/5 plans complete

**Success Criteria:**

1. Operator can submit a QSO via form (callsign, band, mode, exchange) using keyboard shortcuts and see it persisted
2. Real-time dupe warning appears before submission when same callsign exists on same band+mode
3. Live rate meter shows QSOs/hour, peak, and running total updating on each QSO
4. Live score display shows raw points, multiplier, and estimated total
5. Band/mode breakdown panel shows QSO counts per band+mode
6. One-click Cabrillo export generates a file acceptable by the ARRL submission portal
7. Logged QSOs can be searched, viewed, and edited

**Requirements:** QSO-01, QSO-02, QSO-03, QSO-04, DUPE-01, DUPE-02, DUPE-03, SCOR-01, SCOR-02, SCOR-03, EXPR-01, EXPR-02
**Plans:**
**Wave 1**

- [x] 01-01-PLAN.md — Walking Skeleton: scaffold Go + SvelteKit, SQLite schema, QSO CRUD API, QSO entry form, log table, dev build

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 01-02-PLAN.md — Dupe Detection + Validation: exact dupe check, partial call similarity, client-side dupe warnings, dupe marking
- [x] 01-03-PLAN.md — Live Stats Dashboard: rate meter, score display, band/mode breakdown, StatsBar component

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 01-04-PLAN.md — QSO Search & Inline Edit: callsign search, inline row editing, pagination (50/page)
- [x] 01-05-PLAN.md — Cabrillo Export: ARRL format generation, one-click download button

**Key Deliverables:**

- Go backend with SQLite schema, REST API for QSO CRUD, dupe checking, points calculation
- SvelteKit SPA shell with QSO entry form, keyboard shortcuts
- Rate meter and score display components
- Band/mode breakdown panel
- Cabrillo export endpoint
- QSO search and edit UI

---

### Phase 2: Multi-User & Real-Time

**Goal:** Multiple operators on the LAN see each other's QSOs in real-time with shared station configuration.
**Mode:** mvp

**Success Criteria:**

1. Two operators on separate devices can log QSOs to the same server and see each other's entries appear in real-time
2. WebSocket broadcasts new QSOs to all connected clients within 1 second
3. Station configuration (callsign, class, section, power, transmitter count) is set once and visible to all clients
4. Operator identity can be selected per client session
5. Live scoreboard updates for all clients when any operator logs a QSO

**Requirements:** SYNC-01, SYNC-02, CONF-01, CONF-02, CONF-03
**Plans:** 1/5 plans executed

**Plans:**
- [x] 02-00-PLAN.md — Wave 0 Test Scaffolding: StationConfig.test.js, ws.test.js, OperatorSelector.test.js stubs
- [ ] 02-01-PLAN.md — Station Configuration: SQLite table → REST API → Svelte UI (CONF-01, CONF-03)
- [ ] 02-02-PLAN.md — WebSocket Hub: gorilla/websocket hub + broadcast in CreateQSO + station config route wiring (SYNC-01, SYNC-02 server)
- [ ] 02-03-PLAN.md — Frontend WebSocket: client WS listener + Operator identity selector (SYNC-02 client, CONF-02)
- [ ] 02-04-PLAN.md — Cabrillo Export: reads real station_config instead of hardcoded N0CALL

**Key Deliverables:**

- WebSocket endpoint for real-time QSO broadcasting
- Client WebSocket listener with UI update on remote QSOs
- Station configuration UI (class, section, power, transmitter count)
- Operator identity selector
- Unified live scoreboard across clients
- Real-time log table updates from WebSocket events

---

### Phase 3: Offline Resilience & Polish

**Goal:** The system survives network drops, works on mobile devices, and is pleasant to use in field conditions.
**Mode:** mvp

**Success Criteria:**

1. Client continues to log QSOs locally when server is unreachable, with a visible offline indicator
2. Buffered QSOs auto-sync to the server within 5 seconds of reconnection
3. Dupe checking works against locally cached QSOs when offline
4. UI is usable on phones and tablets with touch-friendly controls
5. Dark mode renders correctly across all components
6. App loads from cache when server is unavailable (Service Worker)

**Requirements:** SYNC-03, SYNC-04, SYNC-05, SYNC-06, UX-01, UX-02, UX-04

**Key Deliverables:**

- IndexedDB persistence layer via Dexie.js for QSO caching
- Service Worker for offline app shell caching
- Offline QSO queue with batch sync via POST /api/sync
- Local dupe checking against IndexedDB cache
- Connection status indicator component
- Mobile-responsive CSS (large touch targets, responsive layout)
- Dark mode CSS theme
- Debounce/rate-limiting rapid entry to prevent double QSOs

---

### Phase 4: Field Day Features & Testing

**Goal:** Event-ready system with bonus tracking, audio feedback, backup, and real-world testing.
**Mode:** mvp

**Success Criteria:**

1. Bonus points tracker allows claiming/unclaiming FD bonuses and reflects in score
2. Audio alert plays on new QSO confirmation and dupe warning
3. One-click database backup exports the SQLite file
4. System survives a 2-hour continuous logging simulation with 3+ clients
5. Tested in a real outdoor setup (park, tent) with at least 2 operators

**Requirements:** UX-03

**Key Deliverables:**

- Bonus points tracker UI with predefined FD bonus list
- Bonus points reflected in score calculation and Cabrillo export
- Web Audio API beep for QSO confirmation / dupe warning
- Backup database endpoint (one-click download)
- Full Field Day simulation test (200+ QSOs, multiple clients)
- Real-world field test with at least 2 operators
- Bug fixes and UX refinements from testing

---

## Phase Dependencies

```
Phase 1 (Core Logger)
  └─► Phase 2 (Multi-User & Real-Time)
        └─► Phase 3 (Offline Resilience & Polish)
              └─► Phase 4 (Field Day Features & Testing)
```

Phases are sequential. Each builds on the previous.

---

## Coverage Summary

| Phase | Requirements | Status |
|-------|-------------|--------|
| 1 | QSO-01–QSO-04, DUPE-01–DUPE-03, SCOR-01–SCOR-03, EXPR-01–EXPR-02 (12) | Pending |
| 2 | SYNC-01–SYNC-02, CONF-01–CONF-03 (5) | Pending |
| 3 | SYNC-03–SYNC-06, UX-01–UX-02, UX-04 (7) | Pending |
| 4 | UX-03 (1) | Pending |

24/24 v1 requirements mapped. 0 unmapped.

---

*Last updated: 2026-05-29 after initialization*
