---
phase: 06-i-think-we-are-missing-an-uninstaller
plan: 01
subsystem: cli
tags: [go, claude-settings, uninstall, tdd]
requires:
  - phase: 05-install-docs-and-real-pane-validation
    provides: Validated managed Claude statusline wrapper, symlink-preserving settings writes, and legacy managed hook cleanup behavior.
provides:
  - Safe `llm-quota uninstall-claude-hook` command.
  - Marker-scoped Claude settings uninstaller that restores wrapped statusline commands.
  - Regression tests for preserving unrelated Claude config during uninstall.
affects: [claude-setup, cli, documentation, release-validation]
tech-stack:
  added: []
  patterns: [marker-scoped-settings-mutation, dependency-injected-cli-command, red-green-tdd]
key-files:
  created:
    - .planning/phases/06-i-think-we-are-missing-an-uninstaller/06-01-SUMMARY.md
  modified:
    - internal/install/claude_hook.go
    - internal/install/claude_hook_test.go
    - cmd/llm-quota/main.go
    - cmd/llm-quota/main_test.go
key-decisions:
  - "Uninstall removes only llm-quota-marked Claude settings and does not delete app cache or state files."
  - "Wrapped statusline passthrough commands are restored as plain Claude statusLine commands during uninstall."
patterns-established:
  - "Public setup commands use appDeps injection and print result messages plus backup paths without starting Bubble Tea."
  - "Claude settings uninstall tests cover both current statusline ownership and legacy PostToolUse cleanup."
requirements-completed: [CLD-03]
duration: 3 min
completed: 2026-05-21
---

# Phase 06 Plan 01: Safe Claude Hook Uninstaller Summary

**Safe Claude setup uninstaller with statusline restoration, legacy hook cleanup, and CLI dispatch coverage**

## Performance

- **Duration:** 3 min
- **Started:** 2026-05-21T18:36:42Z
- **Completed:** 2026-05-21T18:39:44Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments

- Added `UninstallClaudeHook` with marker-scoped removal for app-owned Claude settings.
- Restored wrapped user statusline commands when `llm_quota_passthrough` is present.
- Removed old managed `PostToolUse` entries while preserving unrelated hooks and markerless config.
- Wired the public `llm-quota uninstall-claude-hook` command through dependency injection.
- Covered installer and CLI behavior with RED/GREEN TDD commits and focused Go tests.

## Task Commits

Each task was committed atomically:

1. **Task 1 RED: Add uninstall primitive tests** - `9128999` (test)
2. **Task 1 GREEN: Implement uninstall primitive** - `f0cab7d` (feat)
3. **Task 2 RED: Add uninstall CLI tests** - `71c8e45` (test)
4. **Task 2 GREEN: Wire uninstall CLI command** - `818b60c` (feat)

**Plan metadata:** created in final docs commit after state and roadmap updates

## Files Created/Modified

- `internal/install/claude_hook.go` - Adds marker-scoped Claude setup uninstall logic.
- `internal/install/claude_hook_test.go` - Adds uninstall coverage for statusline restoration, statusline removal, legacy hook cleanup, unmanaged config preservation, and symlink preservation.
- `cmd/llm-quota/main.go` - Adds `uninstall-claude-hook` dispatch and default dependency wiring.
- `cmd/llm-quota/main_test.go` - Adds CLI command tests for uninstall success and extra-argument rejection.
- `.planning/phases/06-i-think-we-are-missing-an-uninstaller/06-01-SUMMARY.md` - Records plan outcome.

## Decisions Made

- Uninstall removes only entries marked with `llm_quota_marker == "llm-quota"`, matching the ownership boundary used by install.
- Uninstall restores a non-empty `llm_quota_passthrough` as a plain command statusline instead of deleting the user's previous statusline behavior.
- Uninstall leaves cache and state files untouched because this plan only removes Claude settings integration.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## Known Stubs

None. Stub-pattern scan found no placeholder production paths in files changed by this plan.

## User Setup Required

None - no external service configuration required.

## Verification

- `go test ./internal/install -run 'TestUninstallClaudeHook' -count=1` — passed.
- `go test ./cmd/llm-quota -run 'TestRunUninstallClaudeHook|TestRunUnknownArgumentPreservesErrorAndExitCode' -count=1` — passed.
- `go test ./internal/install ./cmd/llm-quota -count=1` — passed.
- `go test ./... -count=1` — passed.

## Next Phase Readiness

Ready for Plan 06-02 to document uninstall usage and perform real-local uninstall/reinstall validation.

## Self-Check: PASSED

- Found `internal/install/claude_hook.go`.
- Found `internal/install/claude_hook_test.go`.
- Found `cmd/llm-quota/main.go`.
- Found `cmd/llm-quota/main_test.go`.
- Found `.planning/phases/06-i-think-we-are-missing-an-uninstaller/06-01-SUMMARY.md`.
- Found task commit `9128999`.
- Found task commit `f0cab7d`.
- Found task commit `71c8e45`.
- Found task commit `818b60c`.

---

*Phase: 06-i-think-we-are-missing-an-uninstaller*
*Completed: 2026-05-21*
