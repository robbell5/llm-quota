---
phase: 07-row-alignment-claude-sonnet-limit-and-source-freshness
plan: "01"
subsystem: ui
tags: [go, bubble-tea, claude, quota-rows, rendering]
requires: []
provides:
  - Optional Claude Sonnet weekly quota parsing and cache writing
  - Persistent Sonnet 7d quota row placeholder
  - Fixed-column quota row rendering across normal and compact widths
affects: [phase-07, internal-tui, internal-sources, claude-cache]
tech-stack:
  added: []
  patterns:
    - Optional source window parsing does not invalidate required windows
    - Quota rows use ordered row specs and fixed right columns
key-files:
  created:
    - .planning/phases/07-row-alignment-claude-sonnet-limit-and-source-freshness/07-01-SUMMARY.md
  modified:
    - internal/sources/window.go
    - internal/sources/claude.go
    - internal/sources/claude_test.go
    - internal/install/claude_hook.go
    - internal/install/claude_hook_test.go
    - internal/tui/view.go
    - internal/tui/view_test.go
key-decisions:
  - "Use canonical sonnet_seven_day as the stored Claude cache key while accepting sonnet_weekly as a compatibility alias."
  - "Render Sonnet 7d persistently so missing optional data is visible instead of silently omitting the row."
patterns-established:
  - "Optional provider windows can be added after required windows without failing the whole source fetch when optional data is absent or malformed."
  - "Quota row rendering is driven by ordered row specs with fixed percent and reset columns."
requirements-completed: [CLD-05, CLD-06, POL-01, POL-03]
duration: 53min
completed: 2026-05-22
---

# Phase 07: Row Alignment, Claude Sonnet Limit, and Source Freshness Summary

**Claude Sonnet weekly quota support with a persistent Sonnet 7d row and fixed-column quota row rendering**

## Performance

- **Duration:** 53 min
- **Started:** 2026-05-21T23:52:36Z
- **Completed:** 2026-05-22T00:46:29Z
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments

- Added shared `WindowSonnetSevenDay` support and optional Claude cache parsing for `sonnet_seven_day` and `sonnet_weekly`.
- Updated the Claude statusline cache writer to emit the canonical `sonnet_seven_day` key only when valid optional Sonnet data exists.
- Reworked quota row rendering around ordered row specs so Claude rows render as `Claude 5h`, `Claude 7d`, `Sonnet 7d`, followed by Codex rows.
- Added fixed percent/reset columns, compact label variants, width-safety tests, and persistent Sonnet placeholder coverage.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add optional Claude Sonnet weekly source support** - `f67fe3a` (feat)
2. **Task 2: Refactor quota rows to fixed columns and persistent Sonnet placeholder** - `764c788` (feat)

**Plan metadata:** This summary and tracking commit complete the plan metadata.

## Files Created/Modified

- `internal/sources/window.go` - Defines `WindowSonnetSevenDay`.
- `internal/sources/claude.go` - Parses optional Sonnet weekly Claude cache fields without rejecting required windows.
- `internal/sources/claude_test.go` - Covers valid, absent, alias, and malformed optional Sonnet data.
- `internal/install/claude_hook.go` - Accepts optional Sonnet weekly statusline data and writes canonical cache JSON.
- `internal/install/claude_hook_test.go` - Verifies cache writing with and without optional Sonnet weekly data.
- `internal/tui/view.go` - Renders ordered row specs, persistent Sonnet placeholder rows, and fixed quota row columns.
- `internal/tui/view_test.go` - Verifies Sonnet placeholder/data rows, normal alignment, compact behavior, and width safety.

## Decisions Made

- Store only `sonnet_seven_day` in the managed cache to keep the file shape canonical.
- Accept `sonnet_weekly` as an input/cache compatibility alias so older or alternate producers can still be read.
- Keep optional Sonnet parsing tolerant: malformed optional data is skipped while required Claude windows remain strict.
- Prefer text readability over progress bar preservation at very narrow widths.

## Deviations from Plan

None - plan executed as written.

## Issues Encountered

- The executor did not return a final completion signal after committing both code tasks. The orchestrator used the execute-phase fallback: verified commits, ran tests, and created this summary/tracking artifact directly.

## Verification

- `go test ./internal/sources ./internal/install ./internal/tui` passed.
- `go test ./...` passed.
- Spot-check confirmed `07-01` commits exist and no summary self-check failure marker is present.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Plan 02 can now build freshness lines on top of the stable row order and the persistent `Sonnet 7d` row established here.

---
*Phase: 07-row-alignment-claude-sonnet-limit-and-source-freshness*
*Completed: 2026-05-22*
