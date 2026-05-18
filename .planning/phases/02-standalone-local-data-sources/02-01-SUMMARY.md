---
phase: 02-standalone-local-data-sources
plan: 01
subsystem: data-sources
tags: [go, local-files, claude, parsing, testing]

requires:
  - phase: 01-foreground-tui-foundation
    provides: Go module, command spine, and test conventions
provides:
  - Normalized quota window contract for Claude and Codex source readers
  - Typed source error categories for missing, malformed, no-usable-event, and read failures
  - Local-only Claude cache reader for the app-owned cache file shape
  - Synthetic fixture coverage for valid, missing, malformed, incomplete, and stale Claude cache data
affects: [phase-02-plan-02, phase-03-refresh, phase-04-rendering]

tech-stack:
  added: []
  patterns:
    - Path-injected source readers so tests never touch real home-directory data
    - All-or-nothing source parsing for required two-window quota contracts
    - Stale local data returned with metadata instead of treated as a hard source failure

key-files:
  created:
    - internal/sources/window.go
    - internal/sources/claude.go
    - internal/sources/claude_test.go
  modified: []

key-decisions:
  - "Claude cache parsing rejects partial two-window data rather than returning partial rows."
  - "Old but valid Claude cache data is returned with stale metadata so the TUI can warn without blanking values."
  - "Source errors expose typed categories while keeping source parsing independent from TUI rendering."

patterns-established:
  - "Source readers accept explicit paths and perform no home-directory lookup themselves."
  - "Window data is normalized around product, window kind, label, percent, reset time, capture time, stale flag, and metadata."

requirements-completed: [CLD-04, SRC-03, TEST-01]

duration: 2 min
completed: 2026-05-18
---

# Phase 02 Plan 01: Normalized Source Contract and Claude Cache Reader Summary

**Path-injected Claude cache reader with normalized quota windows, stale metadata, and typed source errors**

## Performance

- **Duration:** 2 min
- **Started:** 2026-05-18T13:23:34Z
- **Completed:** 2026-05-18T13:26:22Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments

- Added the shared `internal/sources` contract for products, quota window kinds, normalized windows, metadata, and typed source errors.
- Added table-driven Claude cache reader tests using only `t.TempDir()` synthetic fixtures.
- Implemented the Claude cache reader for valid, missing, malformed, incomplete, read-error, and stale-cache behavior.

## TDD Execution

- **RED:** `be4b837` added the source contract and failing Claude reader expectations; `go test ./internal/sources -run TestClaude` failed because `NewClaudeReader` did not exist yet.
- **GREEN:** `35841a5` added the minimal Claude reader implementation; `go test ./internal/sources -run TestClaude` and `go test ./...` passed.
- **REFACTOR:** No separate refactor commit was needed.

## Task Commits

Each task was committed atomically:

1. **Task 1: Define source contracts and Claude parser expectations** - `be4b837` (test)
2. **Task 2: Implement Claude cache reader** - `35841a5` (feat)

**Plan metadata:** pending final docs commit

## Files Created/Modified

- `internal/sources/window.go` - Normalized source types and typed source error categories.
- `internal/sources/claude.go` - Local-only Claude cache reader with path injection and defensive parsing.
- `internal/sources/claude_test.go` - Synthetic fixture tests for Claude cache behavior.

## Decisions Made

- Claude cache parsing is all-or-nothing for the required five-hour and seven-day windows, matching D-09.
- Stale cache data remains usable and carries stale metadata instead of becoming an error, matching D-11.
- Source errors use typed categories that downstream TUI code can map to concise placeholder/footer hints, matching D-12.

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 02-02 can reuse `Window`, `Metadata`, and `SourceError` for the Codex rollout reader.
- Plan 02-04 can map typed Claude source errors to placeholder rows and footer hints without parsing Claude-specific JSON details.

## Self-Check: PASSED

- Found `internal/sources/window.go`, `internal/sources/claude.go`, and `internal/sources/claude_test.go`.
- Found task commits `be4b837` and `35841a5` in git history.

---

*Phase: 02-standalone-local-data-sources*
*Completed: 2026-05-18*
