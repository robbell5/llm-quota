---
phase: 03-refresh-and-resilience-loop
plan: 02
subsystem: cli-ui
tags: [go, bubble-tea, local-sources, rendering, tdd]

requires:
  - phase: 03-refresh-and-resilience-loop
    provides: [refresh state machine, source reader injection]
provides:
  - Real no-arg TUI startup wiring for Claude cache and Codex sessions readers
  - Minimal source-backed row rendering for available quota windows
  - Regression coverage for command-edge reader wiring and Phase 3 copy boundaries
affects: [phase-04-quota-display, phase-05-install-docs]

tech-stack:
  added: []
  patterns:
    - Command edge constructs real source readers while tests inject paths and TUI startup
    - Renderer consumes normalized model windows without parsing source files

key-files:
  created: []
  modified:
    - cmd/llm-quota/main.go
    - cmd/llm-quota/main_test.go
    - internal/tui/view.go
    - internal/tui/view_test.go

key-decisions:
  - "Real local source paths remain owned by cmd/llm-quota/main.go, with the TUI receiving only injected readers."
  - "Phase 3 rendering shows simple percent/reset text for available windows while leaving progress bars, threshold styling, and visible stale/status copy to Phase 4."

patterns-established:
  - "No-arg CLI startup can be tested by capturing the constructed tui.Model through an injected StartTUI seam."
  - "Rows are selected by normalized Product and WindowKind values rather than source-specific JSON details."

requirements-completed: [SRC-04, SRC-05, TUI-02, TUI-03]

duration: 4 min
completed: 2026-05-19
---

# Phase 03 Plan 02: Source Wiring and Minimal Row Rendering Summary

**Real Claude/Codex reader startup wiring with minimal source-backed quota rows that preserve Phase 3 copy limits**

## Performance

- **Duration:** 4 min
- **Started:** 2026-05-19T15:33:38Z
- **Completed:** 2026-05-19T15:38:28Z
- **Tasks:** 3 completed
- **Files modified:** 4

## Accomplishments

- Wired no-arg `llm-quota` startup to construct a source-backed `tui.Model` using the Claude cache path and Codex sessions root.
- Added command-edge regression coverage that captures the model through an injected TUI seam instead of starting a real Bubble Tea program.
- Updated row rendering to show simple percentage and reset values for available normalized windows.
- Preserved placeholder fallback rows and guarded against Phase 4-only visible copy such as refreshing status, last-updated text, stale copy, and refresh footer hints.

## Task Commits

Each task was committed atomically:

1. **Task 1 RED: startup reader wiring test** - `7515a41` (test)
2. **Task 1 GREEN: source-backed TUI startup** - `9eda288` (feat)
3. **Task 2 RED: source-backed render test** - `02b5d1f` (test)
4. **Task 2 GREEN: minimal quota row rendering** - `9852b9e` (feat)
5. **Task 3: phase verification** - no code commit; verification-only task with no repository changes

**Plan metadata:** committed separately after state and roadmap updates.

## Files Created/Modified

- `cmd/llm-quota/main.go` - Imports source readers, builds a source-backed model from command-edge defaults, and passes it into Bubble Tea startup.
- `cmd/llm-quota/main_test.go` - Adds no-real-TUI startup coverage with injected paths and updates command tests for the model-passing seam.
- `internal/tui/view.go` - Renders normalized windows as simple percent/reset rows while preserving placeholder behavior for missing data.
- `internal/tui/view_test.go` - Covers source-backed rows, stale-but-valid display as data, forbidden Phase 3 copy, and existing width guards.

## Decisions Made

- Kept real default path construction in `cmd/llm-quota/main.go` rather than moving home-directory knowledge into `internal/tui` or `internal/sources`.
- Kept Phase 3 rendering intentionally minimal: percentage and countdown-compatible reset text only, with bars and polished warnings deferred.
- Used synthetic temp paths and injected startup seams for tests so no test reads real Claude or Codex data.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## Known Stubs

None for this plan. Placeholder rows remain intentional fallback behavior for missing source windows.

## Verification

- `go test ./cmd/llm-quota -run 'TestRun|TestStart|TestInstall|TestClaudeHook'` — PASS
- `go test ./internal/tui -run 'TestRender'` — PASS
- `go test ./internal/tui` — PASS
- `go fmt ./cmd/llm-quota ./internal/tui` — PASS
- `go test ./cmd/llm-quota ./internal/tui` — PASS
- `go test ./...` — PASS
- `rg '~/\.claude|~/\.codex|os\.UserHomeDir\(' cmd internal --glob '*_test.go'` — PASS, no real home-directory source reads in tests

## Threat Flags

| Flag | File | Description |
|------|------|-------------|
| threat_flag: local-file-defaults | `cmd/llm-quota/main.go` | No-arg startup now passes real user-local Claude cache and Codex sessions paths to source readers at runtime; tests use injected temp paths. |

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Phase 3 refresh behavior and source wiring are complete. Phase 4 can add final progress bars, thresholds, warning copy, and responsive row polish on top of normalized model windows.

## Self-Check: PASSED

- Verified key modified files and this summary exist on disk.
- Verified task commits `7515a41`, `9eda288`, `02b5d1f`, and `9852b9e` exist in git history.
- Verification commands listed above passed before state updates.

---

*Phase: 03-refresh-and-resilience-loop*
*Completed: 2026-05-19*
