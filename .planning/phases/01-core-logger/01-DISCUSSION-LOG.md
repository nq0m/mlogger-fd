# Phase 1: Core Logger - Discussion Log

**Discussion date:** 2026-05-29
**Mode:** Auto (all gray areas auto-selected, recommended options auto-chosen)

## Gray Areas Discussed

### 1. QSO Form Behavior
**Question:** How should the QSO entry form behave during rapid logging?
**Options:**
- Auto-clear fields on submit, focus returns to callsign — fastest for rapid entry (Recommended)
- Keep fields populated after submit — easier to correct mistakes, slower for next entry
**[auto] Selected:** Auto-clear on submit (Recommended) — Field Day logging demands speed; operators correct mistakes via the log table editor.

### 2. Dupe Check Timing
**Question:** When should the dupe check fire?
**Options:**
- On callsign field blur and on form submit — catches most dupes, minimizes noise (Recommended)
- On every keystroke after 2+ characters — most responsive, more API calls
- Only on form submit — simplest, but can't prevent wasted typing
**[auto] Selected:** On blur + on submit (Recommended) — Catches dupes before the operator finishes the form, with submit as the final safety net.

### 3. Log & Stats Layout
**Question:** How should the log, rate meter, and score be laid out?
**Options:**
- Three-panel layout: entry form (top), stats bar (middle), log table (bottom, scrollable) — works on both desktop and mobile (Recommended)
- Tabbed interface: separate tabs for "Log Entry", "Recent QSOs", and "Stats" — simpler per-tab, requires switching
**[auto] Selected:** Three-panel layout (Recommended) — All critical info visible without tab switching; single-page design simplifies the Phase 1 SPA.

### 4. Cabrillo Export UX
**Question:** How should Cabrillo export work?
**Options:**
- One-click download button with no preview — fastest, operators export once before submission deadline (Recommended)
- Preview before download — shows formatted output, good for verification
**[auto] Selected:** One-click download (Recommended) — Operators verify by opening the downloaded file; preview adds complexity with minimal benefit for a once-per-event action.

### 5. QSO Editing
**Question:** How should QSO editing work?
**Options:**
- Inline edit in the log table — click row to edit, save/cancel inline (Recommended)
- Modal/overlay edit form — dedicated editing interface
**[auto] Selected:** Inline edit (Recommended) — Simpler implementation, faster for quick corrections, consistent with single-page approach.

### 6. Callsign Validation
**Question:** How strict should callsign validation be?
**Options:**
- Lenient — warn on obviously invalid (empty, too short), but accept anything that looks like a callsign — DX calls have many formats (Recommended)
- Strict — validate against FCC/ITU prefix rules, reject non-matching
**[auto] Selected:** Lenient validation (Recommended) — DX stations and special event calls don't follow standard formats; strict validation would block valid QSOs.

### 7. Project Structure
**Question:** How should the Go backend serve the SvelteKit frontend?
**Options:**
- Single binary serves API + embedded static files — simplest deployment, one binary to ship (Recommended)
- Separate API server + static file server (nginx) — more traditional, more config
**[auto] Selected:** Single binary with embedded files (Recommended) — Aligns with project constraint of single-binary deployment on Raspberry Pi.

## the agent's Discretion Areas

- Points calculation table: hardcoded per planning doc modes
- Keyboard shortcuts: standard web form Tab/Ctrl+Enter
- Rate meter windows: 10min, 1hour, session total
- Error display: toast/inline notifications
- Pagination: offset-based, 50 per page, load-more

## Deferred Ideas

None generated during this discussion — all ideas within Phase 1 scope were covered.

---

*Discussion logged: 2026-05-29*
