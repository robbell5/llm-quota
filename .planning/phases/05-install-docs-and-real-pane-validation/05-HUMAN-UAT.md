---
status: pending
phase: 05-install-docs-and-real-pane-validation
source: [05-01-PLAN.md, 05-02-PLAN.md, 04-HUMAN-UAT.md]
started: 2026-05-20T12:17:43Z
updated: 2026-05-20T12:17:43Z
---

# Phase 05 Human UAT

## How to run

Use a real dedicated tmux pane and the real `llm-quota` binary. Follow the README install/setup path:

1. Install with `go install github.com/rob/llm-quota/cmd/llm-quota@latest`, or smoke-check the local repo with `go build ./cmd/llm-quota`.
2. Run `llm-quota install-claude-hook` for explicit standalone Claude hook setup.
3. Open Claude so the app-owned hook can write `~/.cache/llm-quota/claude.json`.
4. Open Codex so local rollout data appears under `~/.codex/sessions`.
5. Start `llm-quota` in the tmux pane and validate the checklist below.

Do not use screenshots as the only evidence. Do not add or rely on demo mode, fixture mode, public validation mode, network fallback, or Rob's custom statusline.

## Tests

### 1. README install smoke

expected: README install instructions are usable through `go install github.com/rob/llm-quota/cmd/llm-quota@latest` or the local `go build ./cmd/llm-quota` smoke path without requiring broader PATH mutation.
result: pending

### 2. Standalone Claude hook setup

expected: `llm-quota install-claude-hook` installs or updates only the app-owned Claude hook/cache writer and preserves unrelated Claude configuration.
result: pending

### 3. Claude missing-data troubleshooting hint

expected: When Claude cache data is unavailable, the visible recovery path matches README and TUI footer copy: `Claude: run install-claude-hook`, followed by opening Claude so `~/.cache/llm-quota/claude.json` can be written.
result: pending

### 4. Claude stale-data troubleshooting hint

expected: Stale Claude data remains visible with an age hint such as `Claude data 2h old; open Claude`, and opening Claude plus refresh produces the recovery path described in README.
result: pending

### 5. Codex missing or stale-data troubleshooting hint

expected: Codex recovery matches README and TUI footer copy: `Codex: open Codex`, with local rollout data under `~/.codex/sessions`.
result: pending

### 6. Default refresh cadence

expected: The TUI uses the default 30-second refresh cadence while running in the foreground.
result: pending

### 7. Manual refresh key

expected: Pressing `r` refreshes immediately without disrupting the running pane.
result: pending

### 8. q quit

expected: Pressing `q` quits cleanly.
result: pending

### 9. Ctrl-C quit

expected: Pressing `Ctrl-C` quits cleanly.
result: pending

### 10. Responsive tmux widths

expected: In the real tmux pane, widths 50, 49, 30, and 29 remain readable without wrapping; below 30 columns progress bars are omitted while percent/reset status remains visible.
result: pending

### 11. Terminal color perception

expected: Low/yellow/red quota urgency colors are perceptible in the actual terminal: low usage appears green, 60-84% appears yellow, and 85%+ appears red without alert badges or warning words.
result: pending

## Results

Awaiting human validation in a real tmux pane.

## Summary

total: 11
passed: 0
issues: 0
pending: 11
skipped: 0
blocked: 0

## Gaps

None recorded yet.
