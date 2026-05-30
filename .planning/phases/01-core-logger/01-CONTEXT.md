# Phase 1: Core Logger - Context

**Gathered:** 2026-05-29
**Status:** Ready for planning

## Phase Boundary

Delivers a single-user QSO logger: an operator can submit QSOs via a keyboard-navigable form, see live rate/scoring with band/mode breakdown, detect duplicates in real-time, search and edit past entries, and export a valid ARRL Field Day Cabrillo file. Single Go binary serves both the API and the SvelteKit SPA. Multi-user sync, offline resilience, and mobile polish are deferred to later phases.

## Implementation Decisions

### Entry Form Behavior

- **D-01:** QSO form auto-clears all fields on successful submit and returns focus to the callsign field — optimized for rapid sequential entry during contest conditions.
- **D-02:** Dupe check fires on callsign field blur AND on form submit. Blur-triggered check catches most dupes before the operator finishes filling the form. Submit-triggered check is the final guard.
- **D-03:** Callsign validation is lenient — warn on empty or single-character input, but accept anything that looks like a callsign. DX stations have diverse formats; strict FCC/ITU validation would block valid entries. Submit is always allowed even with a validation warning.

### Layout & Navigation

- **D-04:** Single-page three-panel layout: QSO entry form at top (always accessible), stats bar in the middle (rate, score, band/mode counts — always visible), scrollable log table at the bottom. This works on both desktop and mobile without tab switching. No separate tabs or pages for logging vs viewing.
- **D-05:** QSO editing is inline — click a row in the log table to switch it to edit mode, save or cancel in place. No modal or separate edit page.

### Export

- **D-06:** Cabrillo export is a one-click button with no preview. Generates and downloads the file immediately. Operators export once before the ARRL submission deadline; verification happens by opening the downloaded file.

### Project Structure

- **D-07:** Single Go binary embeds the SvelteKit static build via `embed.FS` and serves both the REST API (`/api/*`) and the SPA (`/*` falling back to `index.html`). No nginx, no separate static file server, no Docker — one binary, one systemd unit.

### the agent's Discretion

- Points calculation table (which modes are 1pt vs 2pt) — hardcoded as described in the planning doc: CW, RTTY, FT8, FT4, PSK31, MFSK, JT65, JT9, OLIVIA, DOMINO = 2pts; SSB, FM, AM = 1pt.
- Keyboard shortcuts for the form (Tab order, Ctrl+Enter to submit) — standard web form behavior, no custom keybinding system.
- Rate meter time windows — last 10 minutes, last 1 hour, and overall session. Show current rate prominently, peak rate as secondary.
- Error handling — toast or inline notification for API errors. Network errors show "Could not reach server" with retry hint. Validation errors appear under the relevant field.
- Pagination for the log table — offset-based with 50 QSOs per page. Load more on scroll or page button.

## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Field Day Rules & Format
- `.planning/PROJECT.md` — Project context, constraints, core value, key decisions (Go + SvelteKit + SQLite)
- `.planning/REQUIREMENTS.md` — Phase 1 requirements: QSO-01–QSO-04, DUPE-01–DUPE-03, SCOR-01–SCOR-03, EXPR-01–EXPR-02
- `.planning/ROADMAP.md` § Phase 1 — Scope anchor and success criteria

### Original Planning Doc (detailed specs)
- `/home/jeremy/field-day-logger-plan.md` — Full system architecture, API design, database schema, frontend component list, Cabrillo format reference, points calculation logic

### External Specs
- ARRL Field Day Cabrillo format (referenced in planning doc Appendix A) — fixed-width QSO line format, required headers, bonus claim syntax

## Existing Code Insights

### Reusable Assets
- None — greenfield project. No existing components, utilities, or patterns to reuse.

### Established Patterns
- None yet — Phase 1 establishes the foundational patterns (Go project structure, SvelteKit SPA conventions, API design, database access layer).

### Integration Points
- SQLite database at project root (schema defined in planning doc §4.4)
- REST API at `/api/*` (endpoints defined in planning doc §4.5)
- SvelteKit static build embedded via Go's `embed.FS`
- Systemd service for deployment on Raspberry Pi

## Specific Ideas

From the planning document:
- QSO entry form should follow the exact layout shown in Appendix B: callsign input, band dropdown, mode dropdown, exchange field, with Tab navigation between fields and Ctrl+Enter to submit
- Keyboard-only workflow: ~4 seconds per QSO with practice
- Rate meter should display last 1h and 10m windows updating on each QSO
- Score display should show raw points, multiplier, bonus, and estimated total
- Band dropdown options: 160M, 80M, 40M, 20M, 15M, 10M, 6M, 2M, 70CM
- Mode dropdown options: CW, SSB, FM, RTTY, FT8, FT4, PSK31
- Recent QSO display shows: timestamp, callsign, band, mode, exchange, points

## Deferred Ideas

None — discussion stayed within Phase 1 scope. Following ideas are already in the roadmap backlog for later phases:
- Multi-user WebSocket sync (Phase 2)
- Offline IndexedDB buffer (Phase 3)
- Mobile-responsive layout (Phase 3)
- Dark mode (Phase 3)
- Bonus points tracker (Phase 4)
- Audio alerts (Phase 4)

---

*Phase: 1-Core Logger*
*Context gathered: 2026-05-29*
