---
phase: 02-standalone-local-data-sources
plan: 04
subsystem: cli-ui
tags: [go, bubble-tea, cli, setup, tui]

requires:
  - phase: 02-standalone-local-data-sources
    provides: [Claude source contract, Codex source reader, Claude hook installer]
provides:
  - Testable command dispatch for no-arg TUI launch and explicit Claude hook install
  - Pre-TUI first-launch Claude hook consent prompt with decline handling
  - Readable placeholder/footer setup hints for missing local data
affects: [phase-03-refresh-loop, phase-04-rendering, phase-05-install-docs]

tech-stack:
  added: []
  patterns:
    - Dependency-injected command edge for CLI dispatch tests
    - Synthetic-path command tests that do not touch real home-directory data

key-files:
  created: [cmd/llm-quota/main_test.go]
  modified: [cmd/llm-quota/main.go, internal/tui/view.go, internal/tui/view_test.go]

key-decisions:
  - "Command dispatch remains intentionally narrow: only no-arg TUI launch and install-claude-hook are supported in Phase 2."
  - "First-launch setup consent is handled before Bubble Tea startup and uses injected dependencies in tests to avoid real Claude config mutation."
  - "The wide footer uses the exact install-claude-hook command hint while width 50 and narrower keep the compact footer."

patterns-established:
  - "Command edge tests inject streams, installer functions, decline-state functions, and TUI startup."
  - "CLI smoke checks that need real exit codes should use a compiled temporary binary rather than go run."

requirements-completed: [CLD-01, CLD-02, CLD-03, SRC-03]

duration: 4 min
completed: 2026-05-18
---

# Phase 02 Plan 04: Command Setup and Placeholder Hint Summary

**Pre-TUI Claude hook consent flow with explicit install-claude-hook dispatch and readable missing-data setup hints**

## Performance

- **Duration:** 4 min
- **Started:** 2026-05-18T13:44:00Z
- **Completed:** 2026-05-18T13:48:54Z
- **Tasks:** 3 completed
- **Files modified:** 4

## Accomplishments

- Refactored `cmd/llm-quota` so command dispatch can be tested without entering Bubble Tea.
- Added first-launch Claude hook consent handling before TUI startup, including accept, decline, and explicit install command coverage.
- Updated the startup footer to point users at `install-claude-hook` while preserving compact no-wrap behavior for width 50 and narrower.
- Verified the full Go suite and unknown-argument smoke behavior with a compiled temporary binary.

## Task Commits

Each TDD task was committed atomically:

1. **Task 1 RED: command setup tests** - `905b576` (test)
2. **Task 1 GREEN: command setup flow** - `b38c6c5` (feat)
3. **Task 2 RED: setup hint render test** - `9c0d29f` (test)
4. **Task 2 GREEN: setup hint footer** - `e535201` (feat)
5. **Task 3: integration verification** - no code commit; verification-only task with no repository changes

**Plan metadata:** final docs commit

## Files Created/Modified

- `cmd/llm-quota/main.go` - Adds testable command dispatch, first-launch prompt handling, explicit hook install command, and default local paths.
- `cmd/llm-quota/main_test.go` - Covers explicit install dispatch, accepted and declined first-launch prompts, and unknown-argument behavior without real home-directory mutation.
- `internal/tui/view.go` - Updates wide footer setup hint to the explicit install command.
- `internal/tui/view_test.go` - Asserts the new wide hint, absence of inactive refresh-key copy, and existing width guards.

## Decisions Made

- Kept command modes to exactly no-arg launch and `install-claude-hook`, matching D-01 and avoiding out-of-scope aliases.
- Used dependency injection at the command edge instead of invoking Bubble Tea or real Claude paths in tests.
- Treated `go run` exit status behavior as unsuitable for exit-code smoke verification and used a compiled temporary binary for the real exit-code check.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- `go run ./cmd/llm-quota --definitely-unknown` printed the expected error but returned the Go tool's normalized failure status instead of preserving the program's exit code 2. Re-ran the smoke check against a compiled temporary binary, which returned status 2 and the original plain error.

## Known Stubs

| File | Line | Reason |
|------|------|--------|
| `internal/tui/view.go` | 57-95 | Startup rows intentionally render placeholder values until Phase 3/4 wire refresh state and real quota rows. This does not block the plan because readable placeholder rows are the intended Phase 2 output. |

## Verification

- `go test ./cmd/llm-quota` — PASS
- `go test ./internal/tui -run TestRenderStartupScreen` — PASS
- `go test ./...` — PASS
- Compiled temporary binary unknown-argument smoke — PASS (`exit 2`, `llm-quota: unknown argument: --definitely-unknown`)
- Test source path scan for `~/.claude` / `~/.codex` literals — PASS (no matches)

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Phase 2 command-edge setup and local-source foundation are complete. The next phase can wire refresh behavior and source fetches into the running TUI without changing the setup prompt contract.

## Self-Check: PASSED

- Verified key files exist on disk.
- Verified task commits exist in git history.
- Verification commands listed above passed before state updates.

---

*Phase: 02-standalone-local-data-sources*
*Completed: 2026-05-18*
