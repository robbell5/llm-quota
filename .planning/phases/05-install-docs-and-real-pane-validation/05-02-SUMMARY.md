---
phase: 05-install-docs-and-real-pane-validation
plan: 02
subsystem: validation
tags: [human-uat, tmux, claude-statusline, install]
requires:
  - phase: 05-01
    provides: README install, setup, run, and troubleshooting instructions for human validation.
  - phase: 04-quota-display-and-responsive-rendering
    provides: Responsive width and terminal color behavior requiring real-pane UAT.
provides:
  - Approved real tmux-pane validation record for Phase 5.
  - Corrected Claude statusline cache-writer setup that preserves existing Claude settings and statusline behavior.
  - Debug records for install-path and Claude setup issues discovered during real validation.
affects: [release-validation, claude-setup, tui-footer-hints]
tech-stack:
  added: []
  patterns: [statusline-wrapper-cache-writer, symlink-preserving-json-write, human-uat-record]
key-files:
  created:
    - .planning/phases/05-install-docs-and-real-pane-validation/05-HUMAN-UAT.md
    - .planning/debug/install-command-not-found.md
    - .planning/debug/claude-hook-not-installing.md
  modified:
    - README.md
    - cmd/llm-quota/main.go
    - cmd/llm-quota/main_test.go
    - internal/install/claude_hook.go
    - internal/install/claude_hook_test.go
    - internal/tui/model.go
    - internal/tui/view.go
    - internal/tui/view_test.go
key-decisions:
  - "Claude quota cache writing now wraps Claude statusLine.command because real validation showed rate_limits are available there, not in PostToolUse hook input."
  - "Installer writes through symlinked Claude settings targets so dotfiles-managed settings remain symlinked."
  - "The TUI tracks installed Claude setup state so a missing cache after setup asks the user to open Claude instead of reinstalling."
patterns-established:
  - "Real-pane validation issues are recorded as resolved debug artifacts and folded into the Phase UAT record before approval."
  - "Local build documentation must pair `go build ./cmd/llm-quota` with `./llm-quota`, not bare PATH lookup."
requirements-completed: [DOC-01, DOC-02]
duration: 1h42min
completed: 2026-05-20
---

# Phase 05 Plan 02: Real-Pane Validation Summary

**Approved real tmux-pane validation with corrected Claude statusline cache setup and local build instructions**

## Performance

- **Duration:** 1h42min including checkpoint debugging and human re-validation
- **Started:** 2026-05-20T12:17:43Z
- **Completed:** 2026-05-20T13:59:25Z
- **Tasks:** 2
- **Files modified:** 12

## Accomplishments

- Created the Phase 5 human UAT artifact carrying forward Phase 4 real terminal checks and adding install/setup/troubleshooting checks.
- Recorded the approved real tmux-pane validation response: "It is working now - approved".
- Corrected README/UAT local build instructions so `go build ./cmd/llm-quota` is followed by `./llm-quota`, while `go install` uses bare `llm-quota`.
- Fixed real Claude setup discovered during validation by moving quota capture to a managed statusline wrapper, preserving symlinked settings files, preserving any existing statusline command, and removing the old managed tool hook.
- Updated footer hint behavior so an installed-but-empty Claude cache prompts `Claude: open Claude` rather than reinstall guidance.

## Task Commits

Each task/checkpoint step was committed atomically:

1. **Task 1: Create Phase 5 human validation artifact** - `3117c19` (docs)
2. **Checkpoint state: pause for real-pane validation** - `5b05baa` (docs)
3. **Task 2: Record approved validation and fix checkpoint-discovered setup bugs** - `e85b639` (fix)

## Files Created/Modified

- `.planning/phases/05-install-docs-and-real-pane-validation/05-HUMAN-UAT.md` - Manual real-pane validation checklist and approved results record.
- `.planning/debug/install-command-not-found.md` - Resolved debug record for local build command-path confusion.
- `.planning/debug/claude-hook-not-installing.md` - Resolved debug record for Claude setup, symlink preservation, statusline wrapping, and footer hint fixes.
- `README.md` - Corrected install/local-build setup instructions and Claude statusline cache-writer explanation.
- `cmd/llm-quota/main.go` - Added statusline cache-writer command, installed-state detection, and model wiring.
- `cmd/llm-quota/main_test.go` - Added installed statusline detection and source-backed model state coverage.
- `internal/install/claude_hook.go` - Switched installer to managed statusline wrapping, symlink-preserving writes, old managed hook cleanup, and statusline cache writing.
- `internal/install/claude_hook_test.go` - Added regression coverage for symlinks, statusline wrapping, old hook cleanup, and cache writer passthrough.
- `internal/tui/model.go` - Added installed Claude setup state to the model.
- `internal/tui/view.go` - Refined Claude missing-cache footer hint selection.
- `internal/tui/view_test.go` - Added footer coverage for installed-but-missing Claude cache.

## Decisions Made

- Use Claude `statusLine.command` for quota cache capture because real validation showed `rate_limits` are delivered to statusline stdin, not `PostToolUse` hook stdin.
- Preserve existing user statusline behavior by running the original statusline command as a passthrough after writing the llm-quota cache.
- Resolve symlink targets before atomic JSON replacement so dotfiles-managed Claude settings remain linked.
- Treat installed-but-not-yet-populated Claude cache as an "open Claude" recovery state.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed local build instructions that invoked the wrong command name**
- **Found during:** Task 2 real-pane validation checkpoint
- **Issue:** `go build ./cmd/llm-quota` creates `./llm-quota`; the README/UAT path then used bare `llm-quota`, causing shell command lookup failure when the binary was not installed on PATH.
- **Fix:** Split installed and local smoke paths in README/UAT and documented `./llm-quota` for local builds.
- **Files modified:** `README.md`, `.planning/phases/05-install-docs-and-real-pane-validation/05-HUMAN-UAT.md`, `.planning/debug/install-command-not-found.md`
- **Verification:** `go build ./cmd/llm-quota && test -x ./llm-quota` passed before cleanup.
- **Committed in:** `e85b639`

**2. [Rule 1 - Bug] Preserved symlinked Claude settings during install**
- **Found during:** Task 2 real-pane validation checkpoint
- **Issue:** Atomic rename at `~/.claude/settings.json` replaced a symlinked settings file with a regular file.
- **Fix:** Resolve symlink targets before writing JSON atomically so the target file changes while the symlink remains intact.
- **Files modified:** `internal/install/claude_hook.go`, `internal/install/claude_hook_test.go`, `.planning/debug/claude-hook-not-installing.md`
- **Verification:** `go test ./... -count=1` passed, including symlink preservation coverage.
- **Committed in:** `e85b639`

**3. [Rule 2 - Missing Critical] Switched Claude quota capture to statusline wrapper**
- **Found during:** Task 2 real-pane validation checkpoint
- **Issue:** The managed `PostToolUse` hook did not receive `rate_limits`, so no Claude quota cache was written despite setup reporting installed.
- **Fix:** Install a managed `statusLine.command` wrapper that writes the cache from statusline stdin and passes the same input to any existing statusline command.
- **Files modified:** `cmd/llm-quota/main.go`, `cmd/llm-quota/main_test.go`, `internal/install/claude_hook.go`, `internal/install/claude_hook_test.go`, `README.md`, `.planning/debug/claude-hook-not-installing.md`
- **Verification:** `go test ./... -count=1` and `go build ./cmd/llm-quota` passed; human validation confirmed setup working.
- **Committed in:** `e85b639`

**4. [Rule 1 - Bug] Fixed misleading Claude footer after setup**
- **Found during:** Task 2 real-pane validation checkpoint
- **Issue:** TUI showed `Claude: run install-claude-hook` whenever the Claude cache was missing, even when setup was already installed and the user only needed to open Claude.
- **Fix:** Track installed Claude setup state in the TUI model and render `Claude: open Claude` for installed-but-empty cache.
- **Files modified:** `cmd/llm-quota/main.go`, `cmd/llm-quota/main_test.go`, `internal/tui/model.go`, `internal/tui/view.go`, `internal/tui/view_test.go`
- **Verification:** `go test ./... -count=1` passed.
- **Committed in:** `e85b639`

---

**Total deviations:** 4 auto-fixed (3 bugs, 1 missing critical functionality)
**Impact on plan:** All fixes were required to make DOC-01/DOC-02 true during real validation; no demo, fixture, network, OAuth, daemon, or statusline integration feature was added beyond the required app-owned Claude cache setup.

## Issues Encountered

- The first real validation attempt exposed install-command confusion and Claude setup bugs. Both were debugged, fixed, tested, documented, and re-validated before recording approval.

## Known Stubs

None. Stub-pattern scan only found test failure-message text containing "placeholder" in `internal/tui/view_test.go`; this is diagnostic copy, not a product stub.

## Threat Flags

| Flag | File | Description |
|------|------|-------------|
| threat_flag: local-command-wrapper | `internal/install/claude_hook.go` | Managed Claude `statusLine.command` now wraps and executes an existing local statusline command after writing the quota cache. |
| threat_flag: symlinked-settings-write | `internal/install/claude_hook.go` | Installer resolves a symlinked Claude settings path and writes to the target file to preserve dotfiles-managed settings. |

## User Setup Required

None remaining for Phase 5. The human checkpoint was approved after real local validation.

## Verification

- `go test ./... -count=1` — passed.
- `go build ./cmd/llm-quota` — passed.
- `test -x ./llm-quota` — passed after build; generated binary was removed before completion.
- `grep -q '^status: passed' .planning/phases/05-install-docs-and-real-pane-validation/05-HUMAN-UAT.md` — passed.
- `! grep -q '^result: pending' .planning/phases/05-install-docs-and-real-pane-validation/05-HUMAN-UAT.md` — passed.

## Next Phase Readiness

Phase 5 is ready for phase-level verification and closure. The final release docs and real-pane validation evidence are present, with checkpoint-discovered setup fixes covered by tests.

## Self-Check: PASSED

- Found `.planning/phases/05-install-docs-and-real-pane-validation/05-HUMAN-UAT.md`.
- Found `.planning/phases/05-install-docs-and-real-pane-validation/05-02-SUMMARY.md`.
- Found `.planning/debug/install-command-not-found.md`.
- Found `.planning/debug/claude-hook-not-installing.md`.
- Found task commit `3117c19`.
- Found checkpoint fix/approval commit `e85b639`.

---

*Phase: 05-install-docs-and-real-pane-validation*
*Completed: 2026-05-20*
