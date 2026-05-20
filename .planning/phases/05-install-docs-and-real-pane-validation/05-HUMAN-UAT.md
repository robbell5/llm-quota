---
status: passed
phase: 05-install-docs-and-real-pane-validation
source: [05-01-PLAN.md, 05-02-PLAN.md, 04-HUMAN-UAT.md]
started: 2026-05-20T12:17:43Z
updated: 2026-05-20T14:05:00Z
---

# Phase 05 Human UAT

## How to run

Use a real dedicated tmux pane and the real `llm-quota` binary. Follow the README install/setup path:

1. Choose one path. Installed path: run `go install github.com/rob/llm-quota/cmd/llm-quota@latest`, then `llm-quota install-claude-hook`.
2. Local smoke path: from the repo root, run `go build ./cmd/llm-quota`, then `./llm-quota install-claude-hook`. If `./llm-quota` is missing, run the build command again before continuing.
3. Open Claude so the app-owned statusline cache writer can write `~/.cache/llm-quota/claude.json`.
4. Open Codex so local rollout data appears under `~/.codex/sessions`.
5. Start `llm-quota` in the tmux pane after `go install`, or `./llm-quota` from the repo root after the local build path, and validate the checklist below.

Do not use screenshots as the only evidence. Do not add or rely on demo mode, fixture mode, public validation mode, network fallback, or Rob's custom statusline.

## Tests

### 1. README install smoke

expected: README install instructions are usable through `go install github.com/rob/llm-quota/cmd/llm-quota@latest` or the local `go build ./cmd/llm-quota` plus `./llm-quota` smoke path without requiring broader PATH mutation.
result: passed

### 2. Standalone Claude hook setup

expected: The matching setup command installs or updates only the app-owned Claude statusline cache writer in Claude settings, preserves unrelated Claude configuration and any existing statusline command, preserves a symlinked `~/.claude/settings.json`, and does not require a separate hook script file.
result: passed

### 3. Claude missing-data troubleshooting hint

expected: When Claude cache data is unavailable before setup, the visible recovery path matches README and TUI footer copy: `Claude: run install-claude-hook`; after setup, opening Claude lets the statusline cache writer populate `~/.cache/llm-quota/claude.json`.
result: passed

### 4. Claude stale-data troubleshooting hint

expected: Stale Claude data remains visible with an age hint such as `Claude data 2h old; open Claude`, and opening Claude plus refresh produces the recovery path described in README.
result: passed

### 5. Codex missing or stale-data troubleshooting hint

expected: Codex recovery matches README and TUI footer copy: `Codex: open Codex`, with local rollout data under `~/.codex/sessions`.
result: passed

### 6. Default refresh cadence

expected: The TUI uses the default 30-second refresh cadence while running in the foreground.
result: passed

### 7. Manual refresh key

expected: Pressing `r` refreshes immediately without disrupting the running pane.
result: passed

### 8. q quit

expected: Pressing `q` quits cleanly.
result: passed

### 9. Ctrl-C quit

expected: Pressing `Ctrl-C` quits cleanly.
result: passed

### 10. Responsive tmux widths

expected: In the real tmux pane, widths 50, 49, 30, and 29 remain readable without wrapping; below 30 columns progress bars are omitted while percent/reset status remains visible.
result: passed

### 11. Terminal color perception

expected: Low/yellow/red quota urgency colors are perceptible in the actual terminal: low usage appears green, 60-84% appears yellow, and 85%+ appears red without alert badges or warning words.
result: passed

## Results

Human checkpoint approved: "It is working now - approved".

Validation notes:

- README and UAT local-build instructions were corrected to distinguish installed `llm-quota` usage from local `./llm-quota` usage.
- Claude setup was validated after preserving the symlinked `~/.claude/settings.json` target and switching the cache writer to a managed `statusLine.command` wrapper.
- The installed Claude settings contain `claude-statusline-cache-writer` and no stale `claude-hook-cache-writer` registration.
- `~/.cache/llm-quota/claude.json` exists after opening Claude.
- `./llm-quota install-claude-hook` reports the Claude hook/cache writer is already installed.

## Summary

total: 11
passed: 11
issues: 0
pending: 0
skipped: 0
blocked: 0

## Gaps

None recorded yet.

## Issues

None remaining after approved validation. During checkpoint debugging, two setup problems were found and fixed before approval: the local build path needed `./llm-quota`, and Claude quota capture needed a statusline wrapper that preserves symlinked settings and existing statusline behavior.
