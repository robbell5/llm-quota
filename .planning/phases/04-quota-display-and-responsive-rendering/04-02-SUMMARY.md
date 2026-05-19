---
phase: 04-quota-display-and-responsive-rendering
plan: 02
subsystem: ui
tags: [go, bubble-tea, bubbles-progress, lip-gloss, rendering, responsive, tdd]

requires:
  - phase: 04-quota-display-and-responsive-rendering
    provides: [final quota data rows, threshold progress bars, reset countdown tokens]
provides:
  - Missing-source footer recovery hints mapped from typed source errors
  - Stale-source footer hints that preserve last-known quota rows
  - Responsive row degradation for widths 50, 49, 30, 29, and 20
affects: [phase-05-install-docs, tmux-pane-validation, troubleshooting-copy]

tech-stack:
  added: []
  patterns:
    - Footer hint priority maps model errors and stale windows to bounded user-facing copy
    - Render tests use synthetic models plus ANSI-stripped width assertions at breakpoint widths

key-files:
  created: []
  modified:
    - internal/tui/view.go
    - internal/tui/view_test.go

key-decisions:
  - "Footer recovery copy is selected from typed model state and never renders raw SourceError category strings."
  - "Responsive breakpoints are implemented against inner render width so rows stay within the shell padding budget."
  - "Baseline footer includes r refresh only when it fits and does not displace missing or stale recovery hints."

patterns-established:
  - "Missing-source tests populate Model.errors directly and assert fixed actionable footer hints."
  - "Stale rows remain normal data rows while footer copy carries source age and recovery action."

requirements-completed: [DISP-01, DISP-02, DISP-03, DISP-04, DISP-05, DISP-06, TUI-05, TUI-06, TEST-04]

duration: 3 min
completed: 2026-05-19
---

# Phase 04 Plan 02: Missing/Stale Footer Hints and Responsive Layout Summary

**Responsive Claude/Codex quota rows with safe recovery footers for missing and stale local data**

## Performance

- **Duration:** 3 min
- **Started:** 2026-05-19T22:03:23Z
- **Completed:** 2026-05-19T22:07:04Z
- **Tasks:** 2 completed
- **Files modified:** 2

## Accomplishments

- Added RED render coverage for missing Claude/Codex source hints, stale source age hints, raw category suppression, and narrow breakpoint line widths.
- Implemented footer priority for missing Claude setup, missing Codex recovery, stale Claude/Codex data, and baseline quit/refresh copy.
- Updated row rendering so full, compact, and very narrow layouts preserve useful percent/reset status without wrapping.

## Task Commits

Each task was committed atomically:

1. **Task 1 RED: missing/stale and responsive render coverage** - `30d4025` (test)
2. **Task 2 GREEN: footer priority and responsive degradation** - `4cb2552` (feat)

**Plan metadata:** committed separately after state and roadmap updates.

## Files Created/Modified

- `internal/tui/view_test.go` - Adds missing-source, stale-source, raw-category suppression, and breakpoint width tests.
- `internal/tui/view.go` - Maps source/stale states to bounded footer hints and degrades rows across full, compact, and very narrow widths.

## Decisions Made

- Footer recovery copy is selected from typed model state and never renders raw `SourceError.Category` strings.
- Breakpoint branches use inner render width so shell padding cannot push ANSI-stripped lines over the terminal width.
- Baseline `r refresh` copy appears only when it fits and only after higher-priority recovery hints are absent.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Removed duplicate Bubbles progress percentage from rows**
- **Found during:** Task 2 (Implement footer priority and responsive row degradation)
- **Issue:** The static Bubbles progress component rendered its own percentage, producing duplicate percent text in data rows and wasting narrow layout space.
- **Fix:** Configured the progress component with `progress.WithoutPercentage()` and kept the explicit threshold-colored percent column as the single visible percentage.
- **Files modified:** `internal/tui/view.go`
- **Verification:** `go test ./internal/tui -run TestRender -count=1` and `go test ./... -count=1` passed.
- **Committed in:** `4cb2552`

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** The fix was required for row correctness and narrow-width readability; it did not change the planned UI scope.

## Issues Encountered

- RED tests failed as expected before implementation because stale footer hints and responsive breakpoint behavior were not yet implemented.

## Known Stubs

None. Existing missing-data placeholder rows are intentional fallback UI and are now paired with actionable footer hints when source errors are known.

## Threat Flags

None. The security-relevant surfaces introduced by this plan were already covered by the plan threat model: source-error-to-footer copy, terminal width row assembly, and stale data display.

## Verification

- `go test ./internal/tui -run 'TestRender(MissingAndStaleFooterHints|ResponsiveQuotaLayouts|SourceBackedRows)' -count=1` — PASS after Task 2
- `go test ./internal/tui -run TestRender -count=1` — PASS
- `go test ./internal/tui -count=1` — PASS
- `go test ./... -count=1` — PASS

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Phase 4 rendering is complete and ready for Phase 5 install, documentation, troubleshooting, and real tmux-pane validation.

## Self-Check: PASSED

- Verified modified files exist on disk.
- Verified task commits `30d4025` and `4cb2552` exist in git history.
- Verification commands listed above passed before state updates.

---

*Phase: 04-quota-display-and-responsive-rendering*
*Completed: 2026-05-19*
