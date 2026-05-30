---
phase: 02-multi-user-real-time
plan: 00
subsystem: testing
tags: [vitest, svelte5, jsdom, testing-library, websocket-mock]

# Dependency graph
requires:
  - phase: 01-core-logger
    provides: "SvelteKit frontend shell, api.js, qso.svelte.js store, existing component library"
provides:
  - "Vitest test infrastructure with jsdom environment and Svelte 5 plugin"
  - "StationConfig.test.js scaffold — import test for StationConfig.svelte (Plan 02-01)"
  - "ws.test.js scaffold — WebSocket connect/disconnect tests with mocked WebSocket (Plan 02-03)"
  - "OperatorSelector.test.js scaffold — localStorage-backed operator input tests (Plan 02-03)"
  - "Minimal component/module placeholders for StationConfig, ws.js, OperatorSelector"
affects: ["02-01-station-config", "02-02-websocket-hub", "02-03-frontend-websocket"]

# Tech tracking
tech-stack:
  added: ["vitest@4.1.7", "@testing-library/svelte@5.3.1", "jsdom@29.1.1"]
  patterns:
    - "vitest vi.mock() hoisting for Svelte component mocks"
    - "jsdom resolve.conditions: ['browser'] for Svelte 5 client/server resolution"
    - "@testing-library/svelte render + cleanup pattern in vitest"
    - "WebSocket global stub via vi.stubGlobal('WebSocket', MockWebSocket)"

key-files:
  created:
    - frontend/vitest.config.ts
    - frontend/src/lib/components/StationConfig.test.js
    - frontend/src/lib/components/StationConfig.svelte
    - frontend/src/lib/ws.test.js
    - frontend/src/lib/ws.js
    - frontend/src/lib/components/OperatorSelector.test.js
    - frontend/src/lib/components/OperatorSelector.svelte
  modified:
    - frontend/package.json
    - frontend/package-lock.json

key-decisions:
  - "Created minimal Svelte component placeholders for test import resolution — vitest 4.x vi.mock() hoisting does not prevent Vite's import-analysis from rejecting missing files"
  - "Added resolve.conditions: ['browser'] to vitest.config.ts for Svelte 5 client-side module resolution in jsdom"
  - "Used afterEach(cleanup) from @testing-library/svelte to prevent DOM accumulation across test cases"
  - "Used vi.fn(function() {}) instead of arrow function for WebSocket constructor mock"

patterns-established:
  - "Svelte 5 test pattern: render component with @testing-library/svelte, query with screen.getByRole, interact with fireEvent"
  - "WebSocket test pattern: stub Global WebSocket, verify new WebSocket(url) call with correct protocol derivation"
  - "localStorage test pattern: vi.stubGlobal('localStorage', mock), verify getItem/setItem calls on mount/change"

requirements-completed:
  - CONF-01
  - SYNC-02
  - CONF-02

# Metrics
duration: 6 min
completed: 2026-05-30
---

# Phase 2 Plan 0: Wave 0 Test Scaffolding Summary

**Vitest test infrastructure for StationConfig, WebSocket client, and OperatorSelector with Svelte 5 + jsdom — 8 passing tests across 3 test files, clearing the path for TDD-driven Plans 02-01 through 02-03.**

## Performance

- **Duration:** 6 min
- **Started:** 2026-05-30T03:40:47Z
- **Completed:** 2026-05-30T03:46:48Z
- **Tasks:** 4
- **Files created:** 7
- **Files modified:** 2

## Accomplishments

- Installed vitest, @testing-library/svelte, and jsdom as devDependencies in the frontend package
- Created vitest.config.ts with Svelte 5 plugin, jsdom environment, browser resolve condition, and `$lib` path alias
- Scaffolded StationConfig.test.js with 1 passing import test (1 skipped, awaiting Plan 02-01 component)
- Scaffolded ws.test.js with 3 passing tests: module import, initial wsConnected state, and WebSocket URL construction (1 skipped, awaiting Plan 02-03 handler)
- Scaffolded OperatorSelector.test.js with 4 passing tests: module import, localStorage read on mount, labeled input rendering, and localStorage write on change

## Task Commits

Each task was committed atomically:

1. **Task 0: Install vitest + test deps and create vitest.config.ts** - `de7eb4e` (chore)
2. **Task 1: Scaffold StationConfig.test.js** - `d91df7d` (test)
3. **Task 2: Scaffold ws.test.js** - `295d75e` (test)
4. **Task 3: Scaffold OperatorSelector.test.js** - `fa9ecfb` (test)

## Files Created/Modified

- `frontend/vitest.config.ts` — Vitest configuration with jsdom, Svelte 5 plugin, browser resolve condition, `$lib` alias
- `frontend/src/lib/components/StationConfig.test.js` — Import test + skipped render test for StationConfig component
- `frontend/src/lib/components/StationConfig.svelte` — Minimal component placeholder (Plan 02-01 will replace)
- `frontend/src/lib/ws.test.js` — WebSocket module tests with mocked global WebSocket
- `frontend/src/lib/ws.js` — Minimal WebSocket client module placeholder (Plan 02-03 will replace)
- `frontend/src/lib/components/OperatorSelector.test.js` — localStorage-backed operator input tests
- `frontend/src/lib/components/OperatorSelector.svelte` — Minimal component placeholder (Plan 02-03 will replace)
- `frontend/package.json` — Added vitest, @testing-library/svelte, jsdom to devDependencies
- `frontend/package-lock.json` — Updated with new dependency tree

## Verification Summary

| Test File | Passing | Skipped | Total | Exit Code |
|-----------|---------|---------|-------|-----------|
| StationConfig.test.js | 1 | 1 | 2 | 0 |
| ws.test.js | 3 | 1 | 4 | 0 |
| OperatorSelector.test.js | 4 | 0 | 4 | 0 |

All three `npx vitest run` commands exit 0. All test files committed to repository.

## Decisions Made

- **Placeholder files over vi.mock-only approach:** vitest 4.x `vi.mock()` hoisting does not prevent Vite's import-analysis plugin from rejecting missing file paths. Created minimal `.svelte` and `.js` placeholder files that Plans 02-01/02-03 will replace with full implementations. This is a more robust pattern than relying on mock interception for file resolution.
- **Browser resolve condition:** Svelte 5 exports separate client/server modules. Without `resolve.conditions: ['browser']`, `@testing-library/svelte` imports the server-side Svelte module which lacks `mount()`. Added `conditions: ['browser']` to vitest.config.ts.
- **afterEach(cleanup) pattern:** Standard testing-library DOM cleanup prevents multiple `render()` calls from accumulating elements in the jsdom body across test cases.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] vitest 4.x vi.mock() does not prevent file path resolution failure**
- **Found during:** Task 1 (StationConfig.test.js)
- **Issue:** Plan expected `vi.mock('$lib/components/StationConfig.svelte', () => ({ default: {} }))` to allow importing a non-existent file. vitest 4.x hoists the mock but Vite's import-analysis rejects the path before the mock intercepts.
- **Fix:** Created minimal `StationConfig.svelte` placeholder file (empty component) so Vite can resolve the import path.
- **Files modified:** `frontend/src/lib/components/StationConfig.svelte` (created)
- **Committed in:** `d91df7d`

**2. [Rule 3 - Blocking] Same vi.mock() path resolution issue for ws.js**
- **Found during:** Task 2 (ws.test.js)
- **Issue:** Same as above — `vi.mock('$lib/ws.js', () => ({}))` couldn't prevent import resolution failure.
- **Fix:** Created minimal `ws.js` placeholder with connectWebSocket, disconnectWebSocket, wsConnected exports.
- **Files modified:** `frontend/src/lib/ws.js` (created)
- **Committed in:** `295d75e`

**3. [Rule 1 - Bug] Arrow function cannot be used as WebSocket constructor**
- **Found during:** Task 2 (ws.test.js)
- **Issue:** `vi.fn(() => mockWebSocket)` is an arrow function, which cannot be invoked with `new` in JavaScript. `connectWebSocket()` calls `new WebSocket(...)`, producing `TypeError: () => mockWebSocket is not a constructor`.
- **Fix:** Changed to `vi.fn(function() { return mockWebSocket; })` — regular function supports `new` invocation.
- **Files modified:** `frontend/src/lib/ws.test.js`
- **Committed in:** `295d75e`

**4. [Rule 3 - Blocking] Svelte 5 client/server module resolution in vitest with jsdom**
- **Found during:** Task 3 (OperatorSelector.test.js)
- **Issue:** `@testing-library/svelte` `render()` call failed with `lifecycle_function_unavailable: mount(...) is not available on the server`. Svelte 5's package.json exports `browser` condition for client-side module but vitest didn't set it.
- **Fix:** Added `resolve.conditions: ['browser']` to `vitest.config.ts` to force Svelte 5 client module resolution in the jsdom test environment.
- **Files modified:** `frontend/vitest.config.ts`
- **Committed in:** `fa9ecfb`

**5. [Rule 1 - Bug] Multiple render() calls accumulating in jsdom DOM**
- **Found during:** Task 3 (OperatorSelector.test.js)
- **Issue:** Each test called `render(OperatorSelector, {})` without cleanup, causing `screen.getByRole('textbox')` to find multiple elements as previous renders accumulated in the DOM.
- **Fix:** Added `afterEach(() => { cleanup(); })` to the describe block, standard testing-library pattern.
- **Files modified:** `frontend/src/lib/components/OperatorSelector.test.js`
- **Committed in:** `fa9ecfb`

---

**Total deviations:** 5 auto-fixed (2 Rule 1 bugs, 3 Rule 3 blocking)
**Impact on plan:** All fixes were necessary for test execution. Three placeholder files were created that downstream plans will replace — this is additive, not conflicting. No scope creep.

## Issues Encountered

- vitest 4.x differs from earlier versions in how it handles vi.mock() + missing files; the plan's approach was written for an earlier vitest version. Mitigated by creating placeholder files.
- Svelte 5's dual client/server module exports require explicit browser condition in vitest resolve config — not documented in current @testing-library/svelte setup guides.

## Known Stubs

| File | Line | Description |
|------|------|-------------|
| `frontend/src/lib/components/StationConfig.svelte` | 1-2 | Minimal placeholder component — Plan 02-01 will replace with full station configuration form |
| `frontend/src/lib/ws.js` | 1-12 | Minimal WebSocket client module placeholder — Plan 02-03 will replace with full WebSocket hub client |
| `frontend/src/lib/components/OperatorSelector.svelte` | 1-23 | Minimal operator selector placeholder — Plan 02-03 will replace with full component |

All three stubs are intentional: they provide the minimum contract needed for test files to pass import resolution. Downstream plans (02-01, 02-03) will replace them with full implementations.

## Next Phase Readiness

- All three test scaffold files are in place and passing, ready for Plans 02-01, 02-02, and 02-03 to implement against
- `vitest.config.ts` is fully configured for Svelte 5 component testing with jsdom
- The `vi.mock()` pattern is established but downstream plans should test against real implementations, not mocks (the placeholder files enable this)

---

## Self-Check: PASSED

All 7 created files exist on disk. All 4 task commits found in git history.

---

*Phase: 02-multi-user-real-time*
*Completed: 2026-05-30*
