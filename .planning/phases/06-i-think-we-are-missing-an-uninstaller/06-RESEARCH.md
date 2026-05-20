# Phase 6: Uninstaller Research

**Researched:** 2026-05-20
**Status:** Ready for planning

## Question

What needs to be known to plan an `llm-quota` uninstaller well?

## Relevant Existing Behavior

- The install command surface currently accepts `install-claude-hook`, plus internal cache-writer commands, in `cmd/llm-quota/main.go`.
- Claude setup is owned by `internal/install/claude_hook.go` through a managed `statusLine` wrapper marked with `llm_quota_marker: "llm-quota"`.
- The installer preserves unrelated Claude configuration, wraps an existing `statusLine.command` as `llm_quota_passthrough`, removes older managed `PostToolUse` entries, writes backups on changed existing config files, and writes through symlinked `~/.claude/settings.json` targets.
- Phase 5 real validation showed Claude quota capture must use the statusline wrapper, so uninstall must restore any passthrough command rather than simply deleting `statusLine` blindly.

## Implementation Findings

### Public command surface

Add a narrow public command:

```text
llm-quota uninstall-claude-hook
```

Keep the existing no-arg TUI launch, `install-claude-hook`, and internal cache-writer commands unchanged. Unknown args should continue returning exit code 2.

### Safe uninstall semantics

The uninstaller should remove only app-owned Claude integration:

- If `statusLine.llm_quota_marker == "llm-quota"` and `llm_quota_passthrough` is non-empty, restore `statusLine` to `{ "type": "command", "command": <passthrough> }`.
- If `statusLine.llm_quota_marker == "llm-quota"` and `llm_quota_passthrough` is empty, remove `statusLine`.
- Remove managed `PostToolUse` hook entries with `llm_quota_marker == "llm-quota"` for compatibility with older installs.
- Preserve markerless or unrelated hooks/statusline values.
- Preserve symlinked settings by reusing the existing atomic JSON writer path.
- Report unchanged when no app-owned integration is present.

### Cache and state files

Do not delete `~/.cache/llm-quota/claude.json` or `~/.cache/llm-quota/state.json` in this phase. The phase target is uninstalling the Claude config integration, not wiping local cache files. This avoids destructive data removal and keeps the command reversible by rerunning install.

## Test Strategy

Use existing table/unit test style:

- `internal/install/claude_hook_test.go` covers statusline passthrough restore, statusline removal, old managed hook cleanup, unrelated config preservation, idempotent unchanged result, backup-on-change, and symlink preservation.
- `cmd/llm-quota/main_test.go` covers `uninstall-claude-hook` dispatch, no TUI startup, result messages, and invalid extra args.
- `README.md` docs are grep-verified for install and uninstall command copy.

## Security / Safety Considerations

- Trust boundary: untrusted local Claude settings JSON is read, transformed, backed up, and rewritten.
- Main risk: deleting or overwriting unrelated Claude settings. Mitigation is marker-based ownership checks plus explicit tests for unrelated hooks/statusline preservation.
- Main command risk: passthrough statusline command is local user configuration; uninstall only restores it as data and does not execute it.
- No network, OAuth, Keychain, daemon, or credential reads are needed.

## Recommended Plan Shape

1. Implement and test the installer-package uninstall primitive plus CLI command dispatch.
2. Document uninstall usage and add release validation checks that prove uninstall/reinstall remains safe and reversible.
