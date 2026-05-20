---
status: testing
phase: 02-standalone-local-data-sources
source: [.planning/phases/02-standalone-local-data-sources/02-01-SUMMARY.md, .planning/phases/02-standalone-local-data-sources/02-02-SUMMARY.md, .planning/phases/02-standalone-local-data-sources/02-03-SUMMARY.md, .planning/phases/02-standalone-local-data-sources/02-04-SUMMARY.md, .planning/phases/02-standalone-local-data-sources/02-05-SUMMARY.md]
started: 2026-05-18T15:05:39Z
updated: 2026-05-18T15:05:39Z
---

## Current Test
<!-- OVERWRITE each test - shows where we are -->

number: 1
name: Cold Start Smoke Test
expected: |
  Kill any running llm-quota process. Clear only ephemeral state if present, then start the application from scratch. The command starts without errors, the first-launch setup prompt or TUI appears as appropriate, and no startup panic or unexpected filesystem error is shown.
awaiting: user response

## Tests

### 1. Cold Start Smoke Test
expected: Kill any running llm-quota process. Clear only ephemeral state if present, then start the application from scratch. The command starts without errors, the first-launch setup prompt or TUI appears as appropriate, and no startup panic or unexpected filesystem error is shown.
result: [pending]

### 2. Claude Cache Reader Handles Local Cache States
expected: With app-owned Claude cache data present, llm-quota can read both Claude quota windows; with missing, malformed, incomplete, unreadable, or stale cache data, it reports a safe source state instead of crashing or reading real home-directory data during tests.
result: [pending]

### 3. Codex Rollout Reader Finds Usable Local Quota Data
expected: With local Codex rollout JSONL data available, llm-quota can surface both Codex quota windows from the newest usable rollout; malformed, null, unrelated, or incomplete rollout events are skipped without crashing, and older usable rollout files can still be used.
result: [pending]

### 4. Claude Hook Installation Is Safe and Idempotent
expected: Running the Claude hook installer adds or updates only the llm-quota-managed hook, preserves unrelated Claude settings, creates a backup only when changing settings, and repeated installs do not create unnecessary changes.
result: [pending]

### 5. First-Launch Hook Consent Flow Works
expected: Launching llm-quota before the hook is installed asks for permission before changing Claude settings; accepting installs or upgrades the managed hook before the TUI starts, while declining records the choice and continues without repeatedly prompting.
result: [pending]

### 6. Setup Hint Is Visible in the TUI
expected: When local source data is missing, the startup screen shows readable placeholder rows and a footer hint that points to the explicit install-claude-hook command at wider widths while staying compact at narrow widths.
result: [pending]

### 7. Claude Hook Cache Writer Produces Reader-Compatible Data
expected: When Claude Code invokes the managed hook, llm-quota writes validated cache JSON containing both Claude quota windows; malformed or trailing hook input is rejected without overwriting an existing usable cache.
result: [pending]

## Summary

total: 7
passed: 0
issues: 0
pending: 7
skipped: 0
blocked: 0

## Gaps

[none yet]
