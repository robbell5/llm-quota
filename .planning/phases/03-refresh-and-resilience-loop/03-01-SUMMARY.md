---
phase: 03-refresh-and-resilience-loop
plan: 01
subsystem: tui
tags: [go, bubble-tea, refresh-loop, resilience, testing]

requires:
  - phase: 02-standalone-local-data-sources
    provides: normalized Claude and Codex source readers with typed errors
provides:
  - Tested Bubble Tea refresh state machine for startup, tick, and manual refresh
  - Per-source last-known-good merge policy with typed source error state
  - Shared one-hour stale marking in TUI model state
affects: [phase-03-plan-02, phase-04-rendering, source-error-hints]

tech-stack:
  added: []
  patterns:
    - Injected TUI source readers and clock seams for deterministic refresh tests
    - Bubble Tea commands return typed refresh request, tick, and refresh result messages
    - Source reads happen in commands and model mutation remains inside Update

key-files:
  created:
    - .planning/phases/03-refresh-and-resilience-loop/03-01-SUMMARY.md
  modified:
    - internal/tui/model.go
    - internal/tui/update.go
    - internal/tui/update_test.go

key-decisions:
  - "Manual refresh requests do not alter scheduled tick cadence; tick handling owns tick rescheduling."
  - "Refresh results merge independently by source so one failed reader cannot blank another source's data."
  - "Stale state is stored in model data without introducing Phase 4 warning copy or styling."

patterns-established:
  - "Use Model options for source readers, clock, and refresh interval test seams."
  - "Use refreshRequestedMsg to coalesce refresh starts before invoking source readers."

requirements-completed: [SRC-04, SRC-05, TUI-02, TUI-03, TEST-03]

duration: 4 min
completed: 2026-05-19
---

# Phase 03 Plan 01: Refresh State Machine Summary

**Bubble Tea refresh loop with injected source readers, per-source last-known-good merge, coalesced manual refresh, and one-hour stale state**

## Performance

- **Duration:** 4 min
- **Started:** 2026-05-19T15:25:24Z
- **Completed:** 2026-05-19T15:29:58Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- Added deterministic fake-reader tests covering startup refresh, tick scheduling, manual `r`, duplicate refresh coalescing, resize semantics, source failures, and stale marking.
- Implemented model options for injected readers, deterministic clocks, refresh interval defaults, last-known-good windows, typed source errors, and in-flight refresh state.
- Added Bubble Tea refresh messages and commands that fetch Claude and Codex concurrently while preserving model mutation inside `Update`.

## TDD Execution

- **RED:** `843955b` added failing refresh state tests; `go test ./internal/tui -run 'TestRefresh|TestUpdateStoresWindowSize|TestInit'` failed because the model refresh fields, options, and message types did not exist yet.
- **GREEN:** `a169117` implemented the refresh state machine; targeted TUI tests, `go test ./internal/tui`, and `go test ./...` passed.
- **REFACTOR:** No separate refactor commit was needed.

## Task Commits

Each task was committed atomically:

1. **Task 1: Specify refresh state behavior with fake-reader tests** - `843955b` (test)
2. **Task 2: Implement refresh commands, merge policy, stale marking, and resize semantics** - `a169117` (feat)

**Plan metadata:** committed separately after state and roadmap updates.

## Files Created/Modified

- `internal/tui/model.go` - Adds the source reader interface, injected model options, refresh timing defaults, stale threshold, last-known-good windows, typed errors, and refresh in-flight state.
- `internal/tui/update.go` - Adds refresh request, tick, and result messages plus concurrent source fetching and per-source merge behavior.
- `internal/tui/update_test.go` - Adds fake-reader regression tests for refresh lifecycle, merge resilience, stale state, typed errors, and resize behavior.
- `.planning/phases/03-refresh-and-resilience-loop/03-01-SUMMARY.md` - Documents plan execution results.

## Decisions Made

- Manual `r` returns a refresh request only; it does not create, reset, or cancel scheduled ticks.
- Tick messages batch a refresh request with the next tick command so cadence continues independently from manual refreshes.
- The TUI model enforces one-hour stale state on accepted windows from both sources while leaving final warning copy to Phase 4.

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Verification

- `go test ./internal/tui -run 'TestRefresh|TestUpdateQuits|TestUpdateStoresWindowSize|TestInit'` passed.
- `go test ./internal/tui -run 'TestRefresh|TestUpdateStoresWindowSize|TestInit'` passed after implementation; the same command failed during RED before Task 2.
- `go test ./internal/tui` passed.
- `go test ./...` passed.

## Next Phase Readiness

Plan 03-02 can wire real Claude and Codex readers into `NewModel` and begin rendering source-backed rows while reusing the refresh state, typed errors, and stale markers established here.

## Self-Check: PASSED

- Found `internal/tui/model.go`, `internal/tui/update.go`, and `internal/tui/update_test.go`.
- Found task commits `843955b` and `a169117` in git history.
- Verified plan-level commands passed: `go test ./internal/tui` and `go test ./...`.

---

*Phase: 03-refresh-and-resilience-loop*
*Completed: 2026-05-19*
