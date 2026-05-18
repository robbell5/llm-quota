---
phase: 02-standalone-local-data-sources
plan: 03
subsystem: install
tags: [go, claude-hook, local-files, tdd]

requires:
  - phase: 01-foreground-tui-foundation
    provides: foreground Go command and TUI spine
  - phase: 02-standalone-local-data-sources
    provides: normalized Claude source contract
provides:
  - Safe idempotent Claude hook installer policy
  - App-owned decline state helpers for first-launch setup
  - Synthetic installer tests for preservation, backups, idempotence, and declines
affects: [phase-02-cli-wiring, phase-05-install-docs]

tech-stack:
  added: []
  patterns: [path-injected local file mutation, temp-file-plus-rename writes, table-driven filesystem tests]

key-files:
  created:
    - internal/install/claude_hook.go
    - internal/install/claude_hook_test.go
  modified: []

key-decisions:
  - "Claude hook ownership is explicit: only llm-quota named or marked entries are app-owned."
  - "Claude config writes use sibling backups only when the config changes, followed by temp-file-plus-rename replacement."
  - "First-launch hook declines are stored in an app-owned state file supplied by the command edge."

patterns-established:
  - "Installer functions accept explicit paths so tests never touch real Claude config or cache files."
  - "Managed Claude hook entries are updated in place while unrelated entries are preserved."

requirements-completed: [CLD-01, CLD-02, CLD-03]

duration: 3 min
completed: 2026-05-18
---

# Phase 02 Plan 03: Claude Hook Installer Summary

**Safe llm-quota-owned Claude hook installation with idempotent updates, backups, and remembered first-launch declines**

## Performance

- **Duration:** 3 min
- **Started:** 2026-05-18T13:30:05Z
- **Completed:** 2026-05-18T13:33:53Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments

- Added TDD safety coverage for explicit `llm-quota` ownership, unrelated hook preservation, backup behavior, idempotence, and remembered declines.
- Implemented `internal/install` with exported install and decline-state helpers for later CLI wiring.
- Kept all steady-state behavior local-file-only and path-injected so tests use synthetic temp-directory fixtures.

## Task Commits

Each task was committed atomically:

1. **Task 1: Write Claude hook installer safety tests** - `4dd7c39` (test)
2. **Task 2: Implement safe hook install/update policy** - `9567df7` (feat)

**Plan metadata:** pending final metadata commit

_Note: This plan used the required RED → GREEN flow: the test commit failed before implementation, then the feature commit made the tests pass._

## Files Created/Modified

- `internal/install/claude_hook.go` - Exports hook install/update and decline-state functions with safe file mutation.
- `internal/install/claude_hook_test.go` - Verifies synthetic Claude config preservation, managed-entry updates, backup policy, idempotence, and decline persistence.

## Decisions Made

- Claude hook ownership is explicit: only entries named or marked `llm-quota` are app-owned, preventing accidental mutation of unrelated user hooks.
- Existing Claude settings are backed up only before a changed write; idempotent re-runs report unchanged and create no backup.
- Decline state is stored through path-injected app-owned JSON so first-launch setup can suppress repeated prompts without touching real home-directory data in tests.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Verification

- `go test ./internal/install` — passed
- `go test ./...` — passed

## Self-Check: PASSED

- Found `internal/install/claude_hook.go`
- Found `internal/install/claude_hook_test.go`
- Found `.planning/phases/02-standalone-local-data-sources/02-03-SUMMARY.md`
- Found task commit `4dd7c39`
- Found task commit `9567df7`

## Next Phase Readiness

Ready for Phase 02 Plan 04 CLI wiring to call `InstallClaudeHook`, `RecordClaudeHookDeclined`, and `ClaudeHookDeclined` from the command edge.

---
*Phase: 02-standalone-local-data-sources*
*Completed: 2026-05-18*
