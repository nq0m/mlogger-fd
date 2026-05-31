# Requirements: Field Day Logger

**Defined:** 2026-05-29
**Core Value:** Operators can log QSOs even when the network goes down, with all data syncing automatically when reconnected.

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### QSO Entry

- [ ] **QSO-01**: Operator can log a QSO by entering callsign, band, mode, received exchange using a form with Tab/Enter keyboard navigation
- [ ] **QSO-02**: QSO form validates callsign format and required fields before submission
- [ ] **QSO-03**: Operator can search and edit previously logged QSOs
- [ ] **QSO-04**: Keyboard shortcuts (Ctrl+Enter to submit, Tab between fields) supported for rapid entry

### Dupe Detection

- [ ] **DUPE-01**: Dupe check warns in real-time if callsign already worked on same band AND mode before submission
- [ ] **DUPE-02**: Partial call similarity warning when entered call is similar to a previously logged call (e.g., K1XX vs K1X)
- [ ] **DUPE-03**: Dupe QSOs are logged but marked as duplicate with zero points

### Scoring & Stats

- [ ] **SCOR-01**: Live rate meter displays QSOs per hour, peak rate, and running total
- [ ] **SCOR-02**: Live score display shows raw points, multiplier, bonus points, and estimated total score
- [ ] **SCOR-03**: Band/mode breakdown panel shows QSO count per band+mode combination

### Multi-User & Sync

- [x] **SYNC-01**: Multiple operators on LAN can log to the same server database simultaneously
- [x] **SYNC-02**: New QSOs broadcast via WebSocket to all connected clients in real-time
- [x] **SYNC-03**: Each client buffers QSOs locally via IndexedDB when server is unreachable
- [x] **SYNC-04**: Buffered QSOs auto-sync to server when connection is restored
- [x] **SYNC-05**: Connection status indicator shows online/offline state in the UI
- [x] **SYNC-06**: Dupe checking works against locally cached QSOs when offline

### Station Configuration

- [x] **CONF-01**: Station admin configures station callsign, class, ARRL section, transmitter count, and power level
- [x] **CONF-02**: Operator identity can be set per logging session (callsign or name)
- [x] **CONF-03**: Station configuration persists across server restarts

### Export

- [ ] **EXPR-01**: One-click Cabrillo export generates valid ARRL Field Day format with all QSOs and station info
- [ ] **EXPR-02**: Cabrillo export includes bonus points claimed and correct header metadata

### User Experience

- [x] **UX-01**: Mobile-responsive layout with large touch targets (works on phones and tablets)
- [x] **UX-02**: Dark mode theme for night operation in dimly lit tents
- [ ] **UX-03**: Audio alerts via Web Audio API for new QSO confirmation and dupe warning
- [x] **UX-04**: Service Worker provides offline fallback (app shell cached, works without server)

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Bonus Tracking

- **BON-01**: Bonus points tracker with list of FD bonus opportunities and claim toggles
- **BON-02**: Bonus point summary included in score calculation

### Bandmap

- **BMAP-01**: Scrollable frequency display showing worked and spotted stations by band
- **BMAP-02**: Manual spot entry for stations heard

### Cluster Integration

- **CLUS-01**: Pull DX cluster spots via Telnet connection
- **CLUS-02**: Display spots on bandmap

### CW & Voice Keyer

- **CWV-01**: Programmatic keying via Winkeyer protocol over TCP/serial
- **CWV-02**: Pre-recorded CQ message playback for SSB

### Backup & Data

- **BKUP-01**: One-click database backup (copy SQLite file)
- **BKUP-02**: Print-friendly log summary page
- **BKUP-03**: ADIF export for interoperability with QRZ, LoTW, Club Log

## Out of Scope

| Feature | Reason |
|---------|--------|
| Multi-contest support (Sprint, Sweepstakes, WPX) | Focus on Field Day; template system for v3+ |
| Live dashboard / big screen projector view | Not essential for operators logging; nice-to-have for visitors |
| Propagation map (grey-line, solar data) | External data dependency; complex implementation |
| Satellite QSO mode | Requires grid squares, different exchange format |
| Operator profiles with per-op stats | Adds complexity for marginal benefit during FD |
| Remote operation over internet | Against ARRL rules for portable classes |
| Docker deployment | Overkill; single Go binary + static files sufficient |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| QSO-01 | Phase 1 | Pending |
| QSO-02 | Phase 1 | Pending |
| QSO-03 | Phase 1 | Pending |
| QSO-04 | Phase 1 | Pending |
| DUPE-01 | Phase 1 | Pending |
| DUPE-02 | Phase 1 | Pending |
| DUPE-03 | Phase 1 | Pending |
| SCOR-01 | Phase 1 | Pending |
| SCOR-02 | Phase 1 | Pending |
| SCOR-03 | Phase 1 | Pending |
| SYNC-01 | Phase 2 | Complete |
| SYNC-02 | Phase 2 | Complete |
| SYNC-03 | Phase 3 | Complete |
| SYNC-04 | Phase 3 | Complete |
| SYNC-05 | Phase 3 | Complete |
| SYNC-06 | Phase 3 | Complete |
| CONF-01 | Phase 2 | Complete |
| CONF-02 | Phase 2 | Complete |
| CONF-03 | Phase 2 | Complete |
| EXPR-01 | Phase 1 | Pending |
| EXPR-02 | Phase 1 | Pending |
| UX-01 | Phase 3 | Complete |
| UX-02 | Phase 3 | Complete |
| UX-03 | Phase 4 | Pending |
| UX-04 | Phase 3 | Complete |

**Coverage:**
- v1 requirements: 24 total
- Mapped to phases: 24
- Unmapped: 0

---
*Requirements defined: 2026-05-29*
*Last updated: 2026-05-29 after initial definition*
