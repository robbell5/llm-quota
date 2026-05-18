---
phase: 02-standalone-local-data-sources
verified: 2026-05-18T13:57:43Z
status: gaps_found
score: 17/19 must-haves verified
overrides_applied: 0
gaps:
  - truth: "User can install or update only the llm-quota-owned Claude hook without overwriting unrelated Claude configuration."
    status: failed
    reason: "The installer preserves unrelated entries, but the managed entry is not a runnable Claude Code command hook because it writes a top-level command field instead of a matcher entry containing a hooks array with a command hook object."
    artifacts:
      - path: "internal/install/claude_hook.go"
        issue: "managedHook returns name/marker/matcher/command directly at lines 200-206; code review CR-02 identifies this as not the Claude Code command-hook shape."
      - path: "internal/install/claude_hook_test.go"
        issue: "Tests only assert that a command string points at the cache path, not that the installed JSON uses the runnable Claude Code command-hook shape."
    missing:
      - "Change managedHook to emit a matcher entry with hooks: [{type: command, command: ...}] while preserving the app-owned marker."
      - "Update detection and tests to inspect the nested command hook shape."
  - truth: "User can see Claude quota data after the hook writes ~/.cache/llm-quota/claude.json and Codex quota data from the newest local rollout JSONL."
    status: failed
    reason: "Codex local rollout parsing works, but the installed Claude hook/cache producer cannot produce the cache contract consumed by ClaudeReader. It copies raw hook stdin directly into claude.json and writes in place."
    artifacts:
      - path: "internal/install/claude_hook.go"
        issue: "managedHookCommand returns mkdir -p ... && cat > cachePath at lines 209-210, which stores raw hook input instead of top-level five_hour/seven_day/written_at JSON and is not atomic."
      - path: "internal/sources/claude.go"
        issue: "ClaudeReader only accepts five_hour, seven_day, and written_at at lines 51-55, so the current hook output is malformed for the reader."
      - path: ".planning/phases/02-standalone-local-data-sources/02-REVIEW.md"
        issue: "CR-01 independently classifies the wrong cache format and non-atomic write as BLOCKER."
    missing:
      - "Implement an app-owned Claude hook cache writer command that reads Claude hook stdin, extracts rate_limits.five_hour and rate_limits.seven_day, and writes the documented cache JSON contract."
      - "Use temp-file-plus-rename for cache writes."
      - "Add tests proving hook stdin is converted into a ClaudeReader-readable cache file."
human_verification: []
---

# Phase 2: Standalone Local Data Sources Verification Report

**Phase Goal:** User can choose whether to install the app-owned Claude hook and the app can read both local quota sources without relying on Rob's statusline.
**Verified:** 2026-05-18T13:57:43Z
**Status:** gaps_found
**Re-verification:** No -- initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | User is asked for permission before `llm-quota` installs or updates its Claude hook/cache writer. | ✓ VERIFIED | `cmd/llm-quota/main.go:106-116` prints consent text and only calls `InstallClaudeHook` on yes; `cmd/llm-quota/main_test.go:82-117` verifies accept-before-TUI ordering. |
| 2 | User can decline Claude hook installation and still run the TUI with clear Claude placeholder rows. | ✓ VERIFIED | `cmd/llm-quota/main.go:128-131` records decline and continues; `main_test.go:45-80` verifies decline then TUI; `internal/tui/view.go:57-95` renders all four placeholder rows. |
| 3 | User can install or update only the `llm-quota`-owned Claude hook without overwriting unrelated Claude configuration. | ✗ FAILED | Preservation and idempotence work, but the installed entry is not a runnable Claude Code command hook: `internal/install/claude_hook.go:200-206` emits a top-level `command` instead of nested `hooks: [{type, command}]`; confirmed by review CR-02. |
| 4 | User can see Claude quota data after the hook writes `~/.cache/llm-quota/claude.json` and Codex quota data from the newest local rollout JSONL. | ✗ FAILED | Codex reader works, but Claude hook writer does not: `managedHookCommand` at `internal/install/claude_hook.go:209-210` copies raw hook stdin via `cat >`, while `ClaudeReader` requires `five_hour`, `seven_day`, and `written_at` at `internal/sources/claude.go:51-55`; review CR-01 marks this blocker. |
| 5 | Maintainer can verify Claude and Codex parser behavior with synthetic fixtures for valid, missing, malformed, stale, null, and no-usable-event cases. | ✓ VERIFIED | `internal/sources/claude_test.go` covers valid/missing/malformed/missing-seven-day/stale cache; `internal/sources/codex_test.go` covers usable rollout, null/malformed skips, fallback, and no usable rollout; `go test ./...` passed. |
| 6 | D-09: Claude cache parsing returns two windows only when both cache windows are valid. | ✓ VERIFIED | `claude.go:34-48` validates both windows before returning exactly two windows; tests include `missing seven day rejects all`. |
| 7 | D-11: Old but valid Claude cache data returns windows with stale metadata instead of an error. | ✓ VERIFIED | `claude.go:38-48` marks stale data and returns windows; `claude_test.go:69-98` verifies stale cache returns windows. |
| 8 | D-12: Claude source failures are categorized for concise setup/rendering hints. | ✓ VERIFIED | `claude.go:20-36` maps missing/read/malformed errors to `SourceError` categories from `window.go:33-46`. |
| 9 | D-10: Codex parsing skips malformed, unrelated, and null rate-limit lines while searching for usable data. | ✓ VERIFIED | `codex.go:97-135` skips invalid lines/events; `codex_test.go:15-20` includes unrelated, malformed, and null lines before usable data. |
| 10 | D-13: A newer unusable rollout file does not prevent fallback to an older usable rollout file. | ✓ VERIFIED | `codex.go:36-44` continues across candidates; `codex_test.go:52-89` verifies fallback. |
| 11 | D-14/D-15: Codex discovery scans all rollout JSONL files and orders them by file modification time. | ✓ VERIFIED | `codex.go:54-86` walks recursively and sorts by `ModTime()` descending; tests set mtimes with `os.Chtimes`. Warning: tied mtimes are nondeterministic (WR-01), but this does not block the phase goal. |
| 12 | D-16: Codex `plan_type` is preserved as optional source metadata. | ✓ VERIFIED | `codex.go:159-162` stores plan_type in `Metadata`; `codex_test.go:38,47,77,86` assert metadata. |
| 13 | D-02: A first-launch decline is remembered so normal launches stop prompting repeatedly. | ✓ VERIFIED | `RecordClaudeHookDeclined` and `ClaudeHookDeclined` persist/read `claude_hook_declined`; installer test `TestClaudeHookDeclineStateIsRemembered` verifies this. Warning: closed stdin records decline without input (WR-02). |
| 14 | D-05: Only explicitly managed `llm-quota` hook entries count as app-owned. | ✓ VERIFIED | `isManagedHook` checks `name` or `llm_quota_marker`; `TestInstallClaudeHookUpdatesOnlyExplicitlyManagedLLMQuotaEntry` preserves a similar but unowned hook. |
| 15 | D-06/D-07/D-08: Hook installation preserves unrelated config, backs up only before changes, and is idempotent. | ✓ VERIFIED | `installManagedHook`, `backupFile`, and deep-equality unchanged check are covered by `TestInstallClaudeHookPreservesUnrelatedHooksAndCreatesBackupOnlyOnChange`. |
| 16 | D-01: The CLI supports exactly no-arg TUI launch and explicit `install-claude-hook` setup. | ✓ VERIFIED | `run` accepts no args or exact `install-claude-hook`; unknown args return exit 2; tests cover explicit install and unknown argument. |
| 17 | D-03: First-launch permission prompt happens before Bubble Tea alt-screen startup. | ✓ VERIFIED | `offerFirstLaunchInstall` runs before `StartTUI` in `run`; tests record `install,tui` and `decline,tui` event order. |
| 18 | D-04: Declining install still launches the TUI with Claude placeholder rows and setup hints. | ✓ VERIFIED | Decline test proves TUI starts; `view.go:17-18` and `view_test.go:34-56` verify setup hint copy and compact footer behavior. |
| 19 | Existing Phase 1 quit and width behavior still passes. | ✓ VERIFIED | `go test ./...` passed, including `internal/tui` width guard tests for 50, 49, and 29 columns. |

**Score:** 17/19 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/sources/window.go` | Shared quota window and typed source error contract | ✓ VERIFIED | Exists, substantive types/constants at lines 5-58; used by Claude and Codex readers. |
| `internal/sources/claude.go` | Claude cache reader for documented cache shape | ✓ VERIFIED | Reads injected cache path, validates both windows, categorizes errors, returns stale metadata. |
| `internal/sources/claude_test.go` | Synthetic Claude cache behavior tests | ✓ VERIFIED | Uses `t.TempDir()` and fixture JSON; covers valid, missing, malformed, incomplete, and stale cases. |
| `internal/sources/codex.go` | Codex rollout JSONL reader | ✓ VERIFIED | Walks injected sessions root, sorts rollouts by mtime, skips noisy events, returns normalized windows. |
| `internal/sources/codex_test.go` | Synthetic Codex parser and rollout selection tests | ✓ VERIFIED | Uses temp sessions tree, `os.Chtimes`, null/malformed events, fallback, no-usable-event checks. |
| `internal/install/claude_hook.go` | Claude hook install/update and decline-state policy | ✗ FAILED | File is substantive and wired, but managed hook shape and cache writer command are blockers for functional setup/cache production. |
| `internal/install/claude_hook_test.go` | Synthetic config tests for hook safety | ⚠️ PARTIAL | Covers preservation/backups/idempotence/decline, but misses runnable Claude Code hook shape and cache-writer output contract. |
| `cmd/llm-quota/main.go` | CLI dispatch and pre-TUI setup prompt wiring | ✓ VERIFIED | Exact command dispatch, consent prompt, install/decline paths, and TUI startup wiring exist. |
| `cmd/llm-quota/main_test.go` | Command dispatch tests without Bubble Tea | ✓ VERIFIED | Tests install command, accept/decline prompt paths, and unknown args without real home mutation. |
| `internal/tui/view.go` | Specific placeholder/setup hint copy | ✓ VERIFIED | Renders placeholder rows and setup footer copy; no real data rendering expected in Phase 2. |
| `internal/tui/view_test.go` | Render regression tests for setup hints | ✓ VERIFIED | Asserts setup hint, no `r refresh` copy, and width constraints. |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `internal/sources/claude.go` | `internal/sources/window.go` | `Fetch` returns `[]Window` and typed `SourceError` values | ✓ WIRED | SDK key-link check passed; code uses `Window`, `ProductClaude`, and error categories. |
| `internal/sources/codex.go` | `internal/sources/window.go` | Fetch returns normalized Codex windows | ✓ WIRED | SDK key-link check passed; code returns `ProductCodex` windows. |
| `internal/install/claude_hook.go` | `cmd/llm-quota/main.go` | CLI calls installer functions | ✓ WIRED | SDK key-link check passed; `main.go` wires `install.InstallClaudeHook` and decline helpers. |
| `cmd/llm-quota/main.go` | `internal/install/claude_hook.go` | `install-claude-hook` command and prompt call installer | ✓ WIRED | SDK key-link check passed for command path. |
| `internal/tui/view.go` | Phase 2 D-04 | Placeholder rows remain readable after decline | ✓ WIRED | SDK literal pattern expected `Claude: install hook` and failed, but actual Phase 2 accepted copy is `Claude: run install-claude-hook` in `view.go:17`, verified by tests. |
| `internal/install/claude_hook.go` | `internal/sources/claude.go` | Installed hook/cache producer writes reader-compatible cache | ✗ NOT WIRED | Hook command writes raw stdin using `cat >`; no cache writer converts to `five_hour`/`seven_day`/`written_at`. |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| `internal/sources/claude.go` | `[]Window` | Injected cache JSON file | Yes, when file already matches documented cache contract | ✓ FLOWING |
| `internal/sources/codex.go` | `[]Window` | Injected Codex rollout JSONL files | Yes, from `payload.rate_limits` in local rollout events | ✓ FLOWING |
| `internal/install/claude_hook.go` | `~/.cache/llm-quota/claude.json` | Installed Claude hook command | No; command copies raw hook stdin and is not runnable hook shape | ✗ DISCONNECTED |
| `internal/tui/view.go` | Placeholder row labels/footer copy | Static startup view | Static by design for Phase 2 | ✓ VERIFIED |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Full Phase 2 automated suite | `go test ./...` | All packages passed: `cmd/llm-quota`, `internal/install`, `internal/sources`, `internal/tui` | ✓ PASS |
| Plan artifact verification | `gsd-sdk query verify.artifacts ...` for all four plans | 11/11 artifacts passed existence/substance checks | ✓ PASS |
| Plan key-link verification | `gsd-sdk query verify.key-links ...` for all four plans | 4/5 literal links passed; the failed literal footer pattern was manually verified under the updated accepted copy | ⚠️ PARTIAL |
| Hook cache producer compatibility | Code inspection of `managedHookCommand` and `ClaudeReader` contract | Current producer writes raw stdin; reader expects normalized cache JSON | ✗ FAIL |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|----------|
| CLD-01 | 02-03, 02-04 | User is prompted for permission to install the hook/cache writer during setup or first launch. | ✓ SATISFIED | Consent prompt before TUI in `main.go:106-116`; tests cover accept/decline. |
| CLD-02 | 02-03, 02-04 | User can decline hook installation and still run the TUI with clear Claude placeholder rows. | ✓ SATISFIED | Decline state and TUI start tested; placeholder rows render. |
| CLD-03 | 02-03, 02-04 | User can install/update only the `llm-quota`-owned Claude hook without overwriting unrelated config. | ✗ BLOCKED | Preservation passes, but installed entry is not a runnable Claude Code command hook, so functional hook installation is not achieved. |
| CLD-04 | 02-01 | User can get Claude quota data from `~/.cache/llm-quota/claude.json` after the hook has run. | ✗ BLOCKED | Reader can parse a valid cache, but installer's hook writes raw stdin in-place and cannot produce the reader contract. |
| SRC-01 | 02-02 | User can get Codex quota data from the most recent local Codex rollout JSONL file. | ✓ SATISFIED | `CodexReader` scans rollouts and tests verify newest usable selection. |
| SRC-02 | 02-02 | User sees Codex placeholder rows and concise hint when no usable Codex quota event exists. | ✓ SATISFIED | `ErrorNoUsableEvent` exists and startup placeholder rows/hints render; detailed source-specific rendering comes later. |
| SRC-03 | 02-01, 02-04 | User sees Claude placeholder rows and concise hook/setup hint when Claude cache is missing/malformed/unavailable. | ✓ SATISFIED | Claude source categorizes missing/malformed; TUI footer says `Claude: run install-claude-hook`. |
| TEST-01 | 02-01 | Maintainer can verify Claude cache parsing for valid, missing, malformed, and stale cache files without real home data. | ✓ SATISFIED | `claude_test.go` uses temp files and covers required cases. |
| TEST-02 | 02-02 | Maintainer can verify Codex rollout parsing for newest-file selection, null rate limits, malformed events, and missing usable events. | ✓ SATISFIED | `codex_test.go` covers these cases with synthetic files. |

No additional Phase 2 requirement IDs were found in `.planning/REQUIREMENTS.md` beyond the listed nine.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| `internal/install/claude_hook.go` | 205 | Top-level `command` in managed hook entry | 🛑 Blocker | Claude Code will not execute this as the required command hook shape. |
| `internal/install/claude_hook.go` | 210 | `cat >` direct cache write | 🛑 Blocker | Writes raw hook stdin, truncates the final cache file in place, and does not produce `ClaudeReader` cache JSON. |
| `internal/sources/codex.go` | 84-86 | Sort by mtime only | ⚠️ Warning | Tied rollout mtimes are nondeterministic; review WR-01. |
| `cmd/llm-quota/main.go` | 110-130 | Empty EOF treated as decline | ⚠️ Warning | Non-interactive launch can persist a decline without user input; review WR-02. |

### Human Verification Required

None for this phase decision. The blocking issues are observable in code and confirmed by the advisory review.

### Gaps Summary

Phase 2 does not achieve the standalone Claude setup/cache-producer portion of the goal. The source readers and placeholder startup behavior are mostly in place and tested, and Codex local parsing is functional. The blocker is the connection between setup and Claude data: the installer writes a non-runnable Claude Code hook entry and the command it installs cannot transform Claude hook input into the normalized cache that `ClaudeReader` consumes.

Fixing the two related Claude hook gaps should focus on `internal/install/claude_hook.go` and tests: install the correct command-hook JSON shape, add an app-owned cache writer, write the cache atomically, and prove the generated cache is readable by `NewClaudeReader(...).Fetch(...)`.

---

_Verified: 2026-05-18T13:57:43Z_
_Verifier: the agent (gsd-verifier)_
