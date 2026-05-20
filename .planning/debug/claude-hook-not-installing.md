---
status: resolved
trigger: "Claude hook is not working after running the setup path; system settings.json does not show the hook registered and the hook script is not visible in the expected directory. TUI still shows `Claude: run install-claude-hook`."
created: 2026-05-20
updated: 2026-05-20
---

# Debug Session: claude-hook-not-installing

## Symptoms

- Expected behavior: Running `llm-quota install-claude-hook` or `./llm-quota install-claude-hook` installs/registers the app-owned Claude hook, creates any required hook script or executable target, and after Claude runs the TUI can read `~/.cache/llm-quota/claude.json`.
- Actual behavior: The TUI still shows `Claude: run install-claude-hook`; user reports no hook in system `settings.json` and no hook script in the expected directory.
- Error messages: No command error provided for the install command itself; visible TUI footer says `Claude: run install-claude-hook`.
- Timeline: Found during Phase 5 real-pane validation after local binary execution path was corrected.
- Reproduction: Build/run local binary, run the Claude hook install command, inspect expected Claude settings/hooks locations, then start TUI.

## Current Focus

- hypothesis: unknown
- test: inspect installer path resolution, settings mutation, script/executable creation expectations, and tests
- expecting: identify whether install command is a no-op, writes to a different settings file, fails silently, or installs a command reference rather than a script
- next_action: gather initial installer evidence
- reasoning_checkpoint:
- tdd_checkpoint:

## Evidence

- timestamp: 2026-05-20T12:55:00Z
  observation: Running the real installer returned `llm-quota Claude hook already installed`, and a targeted marker search found `llm_quota_marker`, `llm-quota`, and `claude-hook-cache-writer` in `~/.claude/settings.json`.
  implication: The hook is registered in user Claude settings as a command hook; no separate hook script file is expected by the current implementation.
- timestamp: 2026-05-20T12:55:00Z
  observation: `internal/tui/view.go` rendered `Claude: run install-claude-hook` whenever the Claude source error category was missing, without considering whether the hook was already installed.
  implication: The TUI footer conflated a missing cache file with hook absence, causing the observed misleading prompt after hook registration.
- timestamp: 2026-05-20T13:45:00Z
  observation: `~/.claude/settings.json` had been a symlink to `~/dotfiles/claude/.claude/settings.json`, but the installer wrote via atomic rename at the symlink path, replacing the symlink with a regular file.
  implication: The installer broke dotfiles-managed Claude settings and wrote the hook into the replacement file instead of the source-controlled target.
- timestamp: 2026-05-20T13:52:00Z
  observation: Claude project transcripts showed repeated non-blocking hook errors: `llm-quota: missing rate_limits` from the `PostToolUse` hook command. The user's existing statusline script receives `.rate_limits.five_hour` and `.rate_limits.seven_day` from statusline stdin.
  implication: Claude quota data is available to the statusline command, not to `PostToolUse` hook input. The cache producer must wrap statusline, not install a quota-reading tool hook.

## Eliminated

## Resolution

- root_cause: Three bugs combined. First, the installer wrote with atomic rename directly at `~/.claude/settings.json`, which replaced symlinked dotfiles-managed settings with a regular file. Second, the TUI had no installed-hook state and mapped every missing Claude cache file to `Claude: run install-claude-hook`, making a registered-but-not-yet-populated hook look uninstalled. Third, the cache producer was installed as a `PostToolUse` hook, but Claude quota `rate_limits` are delivered to statusline stdin, not tool hook stdin.
- fix: Repaired the local symlink, copied the settings into the dotfiles target, and changed atomic JSON writes to resolve symlinked file paths before renaming. Replaced the managed `PostToolUse` hook with a managed `statusLine.command` wrapper that writes `~/.cache/llm-quota/claude.json` from statusline stdin and passes the same input through to any existing statusline command. Added regression tests for symlink preservation, statusline wrapping, managed hook cleanup, and statusline cache writing. Added installed-hook state to the TUI model and updated docs/UAT.
- verification: `go test ./... -count=1` passed; `go build ./cmd/llm-quota && test -x ./llm-quota` passed; `ls -l ~/.claude/settings.json` confirmed the symlink is restored; targeted grep confirmed the settings now contain `claude-statusline-cache-writer` and no `claude-hook-cache-writer` entry.
- files_changed: cmd/llm-quota/main.go; cmd/llm-quota/main_test.go; internal/install/claude_hook.go; internal/install/claude_hook_test.go; internal/tui/model.go; internal/tui/view.go; internal/tui/view_test.go; README.md; .planning/phases/05-install-docs-and-real-pane-validation/05-HUMAN-UAT.md; .planning/debug/claude-hook-not-installing.md
