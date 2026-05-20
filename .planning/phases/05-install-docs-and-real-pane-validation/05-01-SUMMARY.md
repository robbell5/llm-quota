---
phase: 05-install-docs-and-real-pane-validation
plan: 01
subsystem: docs
tags: [readme, install, troubleshooting, tmux]
requires:
  - phase: 04-quota-display-and-responsive-rendering
    provides: Footer hints, responsive rendering behavior, and final quota row copy for documentation.
provides:
  - README install, setup, run, keys, tmux-pane, troubleshooting, and scope documentation.
  - Local-only recovery guidance matching the TUI footer hints.
affects: [phase-05, release-docs, human-uat]
tech-stack:
  added: []
  patterns: [local-only documentation, footer-hint troubleshooting]
key-files:
  created: [README.md]
  modified: []
key-decisions:
  - "README documents only the public no-arg launch and install-claude-hook setup paths; the internal cache-writer stays undocumented."
  - "Troubleshooting guidance mirrors footer hints and avoids raw source error category names."
patterns-established:
  - "User-facing recovery docs should name visible footer hints first, then the local action to refresh data."
requirements-completed: [DOC-01, DOC-02]
duration: 5min
completed: 2026-05-20
---

# Phase 05 Plan 01: README Install and Troubleshooting Summary

**README install and recovery guide for the local-only Claude/Codex tmux-pane quota monitor**

## Performance

- **Duration:** 5 min
- **Started:** 2026-05-20T12:11:00Z
- **Completed:** 2026-05-20T12:16:43Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments

- Created `README.md` with the required install, Claude setup, tmux run, keys, troubleshooting, and scope sections.
- Documented `go install github.com/rob/llm-quota/cmd/llm-quota@latest`, the local `go build ./cmd/llm-quota` smoke path, and `llm-quota install-claude-hook`.
- Mapped `Claude: run install-claude-hook`, `Claude: open Claude`, `Codex: open Codex`, and stale-data footer hints to concrete local recovery actions.
- Preserved the v1 boundary: no network/OAuth fallback, Keychain reads, statusline integration, daemon, alerts, forecasting, demo mode, or fixture mode.

## Task Commits

Each task was completed atomically:

1. **Task 1: Create README quickstart and operating guide** - `3062b3e` (docs)
2. **Task 2: Verify README against command surface and documentation boundaries** - no code commit; verification-only task with no file changes after checks passed

## Files Created/Modified

- `README.md` - Primary user-facing install, setup, run, key, pane, troubleshooting, and scope documentation.
- `.planning/phases/05-install-docs-and-real-pane-validation/05-01-SUMMARY.md` - Execution record for Plan 05-01.

## Decisions Made

- README documents only the public no-argument `llm-quota` launch and `llm-quota install-claude-hook` setup command.
- README describes the internal hook/cache writer by behavior, not as a user-facing command.
- Troubleshooting copy stays aligned with visible TUI footer hints instead of internal source error categories.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## Known Stubs

None.

## User Setup Required

None - no external service configuration required by this plan.

## Verification

- `test -f README.md && grep -q 'go install github.com/rob/llm-quota/cmd/llm-quota@latest' README.md && grep -q 'llm-quota install-claude-hook' README.md && grep -q 'Claude: run install-claude-hook' README.md && grep -q 'Codex: open Codex' README.md` — passed.
- `go test ./... -count=1` — passed.
- `! grep -E 'ErrorMissing|ErrorMalformed|ErrorRead|ErrorNoUsableEvent|missing_local_data|read_error|no_usable_event' README.md` — passed.
- `grep -q '^## Scope' README.md && grep -q 'statusline' README.md && grep -q 'network' README.md` — passed.

## Next Phase Readiness

Plan 05-02 can now use the README as the source for human real-pane install, setup, run, and troubleshooting validation.

## Self-Check: PASSED

- Found `README.md`.
- Found `.planning/phases/05-install-docs-and-real-pane-validation/05-01-SUMMARY.md`.
- Found task commit `3062b3e`.

---

*Phase: 05-install-docs-and-real-pane-validation*
*Completed: 2026-05-20*
