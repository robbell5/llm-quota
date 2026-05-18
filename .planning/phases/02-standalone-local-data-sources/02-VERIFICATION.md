---
phase: 02-standalone-local-data-sources
verified: 2026-05-18T15:33:12Z
status: passed
score: 22/22 must-haves verified
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: 17/19
  gaps_closed:
    - "User can install or update only the llm-quota-owned Claude hook without overwriting unrelated Claude configuration."
    - "User can see Claude quota data after the hook writes ~/.cache/llm-quota/claude.json and Codex quota data from the newest local rollout JSONL."
  gaps_remaining: []
  regressions: []
---

# Phase 2: Standalone Local Data Sources Verification Report

**Phase Goal:** User can choose whether to install the app-owned Claude hook and the app can read both local quota sources without relying on Rob's statusline.
**Verified:** 2026-05-18T15:33:12Z
**Status:** passed
**Re-verification:** Yes -- after gap closure

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | User is asked for permission before `llm-quota` installs or updates its Claude hook/cache writer. | ✓ VERIFIED | `cmd/llm-quota/main.go:121-146` prints consent text, installs only on `y`/`yes`, records decline otherwise, and then starts TUI; `cmd/llm-quota/main_test.go:50-122` covers decline/accept ordering. |
| 2 | User can decline Claude hook installation and still run the TUI with clear Claude placeholder rows. | ✓ VERIFIED | Decline path records then starts TUI in `main.go:143-146`; startup rows and setup hints are rendered in `internal/tui/view.go:57-104`; tests assert all four rows and footer behavior. |
| 3 | User can install or update only the `llm-quota`-owned Claude hook without overwriting unrelated Claude configuration. | ✓ VERIFIED | Gap closed: `internal/install/claude_hook.go:197-212` requires the `llm_quota_marker` and emits nested `hooks: [{type: command, command: ...}]` with no top-level command; tests preserve unrelated and markerless hooks and update old managed entries. |
| 4 | User can see Claude quota data after the hook writes `~/.cache/llm-quota/claude.json` and Codex quota data from the newest local rollout JSONL. | ✓ VERIFIED | Gap closed: `RunClaudeHookCacheWriter` converts hook stdin into `five_hour`/`seven_day`/`written_at` and writes atomically; `TestRunClaudeHookCacheWriterWritesReaderCompatibleCache` verifies `sources.NewClaudeReader(cachePath).Fetch(...)` reads generated cache; `CodexReader` scans newest usable local rollout. |
| 5 | Maintainer can verify Claude and Codex parser behavior with synthetic fixtures for valid, missing, malformed, stale, null, and no-usable-event cases. | ✓ VERIFIED | `internal/sources/claude_test.go` covers valid/missing/malformed/incomplete/stale and unreadable-cache read errors; `internal/sources/codex_test.go` covers malformed/null/no-usable/fallback/newest usable cases. |
| 6 | D-09: Claude cache parsing returns two windows only when both cache windows are valid. | ✓ VERIFIED | `claude.go:34-48` validates before returning both windows; test case `missing seven day rejects all` returns no windows with malformed category. |
| 7 | D-11: Old but valid Claude cache data returns windows with stale metadata instead of an error. | ✓ VERIFIED | `claude.go:38-48` computes stale age and returns windows; `claude_test.go:69-98` verifies stale windows. |
| 8 | D-12: Claude source failures are categorized for concise setup/rendering hints. | ✓ VERIFIED | `claude.go:20-36` maps missing/read/malformed into `SourceError`; unreadable-cache coverage exists at `claude_test.go:128-146`. |
| 9 | D-10: Codex parsing skips malformed, unrelated, and null rate-limit lines while searching for usable data. | ✓ VERIFIED | `codex.go:97-135` skips bad/non-matching lines; fixtures include unrelated, malformed, and null lines before usable data. |
| 10 | D-13: A newer unusable rollout file does not prevent fallback to an older usable rollout file. | ✓ VERIFIED | `codex.go:36-44` continues across candidates; `codex_test.go:52-89` verifies fallback. |
| 11 | D-14/D-15: Codex discovery scans all rollout JSONL files and orders them by file modification time. | ✓ VERIFIED | `codex.go:54-86` recursively walks rollout files and sorts by `ModTime()` descending; tests use `os.Chtimes`. |
| 12 | D-16: Codex `plan_type` is preserved as optional source metadata. | ✓ VERIFIED | `codex.go:159-166` stores `plan_type`; tests assert metadata values. |
| 13 | D-02: A first-launch decline is remembered so normal launches stop prompting repeatedly. | ✓ VERIFIED | `RecordClaudeHookDeclined` and `ClaudeHookDeclined` persist state; `TestClaudeHookDeclineStateIsRemembered` verifies read/write. |
| 14 | D-05: Only explicitly managed `llm-quota` hook entries count as app-owned. | ✓ VERIFIED | `isManagedHook` requires `llm_quota_marker`; tests verify markerless `llm-quota` hook is preserved/ignored and similarly named unowned hooks remain unchanged. |
| 15 | D-06/D-07/D-08: Hook installation preserves unrelated config, backs up only before changes, and is idempotent. | ✓ VERIFIED | `installManagedHook`, `backupFile`, and deep-equality unchanged checks are covered by installer tests. |
| 16 | D-01: The CLI supports exactly no-arg TUI launch and explicit `install-claude-hook` setup. | ✓ VERIFIED | `run` handles no args, `install-claude-hook`, and hook-internal writer only; unknown args return exit 2. Broad setup aliases are absent. |
| 17 | D-03: First-launch permission prompt happens before Bubble Tea alt-screen startup. | ✓ VERIFIED | `offerFirstLaunchInstall` runs before `StartTUI`; tests assert event order `install,tui` and `decline,tui`. |
| 18 | D-04: Declining install still launches the TUI with Claude placeholder rows and setup hints. | ✓ VERIFIED | Decline test reaches TUI; render tests assert placeholder rows and `Claude: run install-claude-hook` wide footer. |
| 19 | Existing Phase 1 quit and width behavior still passes. | ✓ VERIFIED | `go test ./...` passed, including `internal/tui` tests for widths 50, 49, and 29. |
| 20 | D-05/D-06/D-08 and CLD-03: the installed `llm-quota` managed Claude hook is a runnable Claude Code command hook entry and unrelated hooks remain unchanged. | ✓ VERIFIED | `managedHook` now uses nested command-hook shape at `claude_hook.go:201-212`; `assertManagedCommandHookShape` fails on top-level `command`; focused tests passed. |
| 21 | CLD-04: the installed hook command writes a ClaudeReader-readable cache with `five_hour`, `seven_day`, and `written_at` fields. | ✓ VERIFIED | `RunClaudeHookCacheWriter` validates both windows and writes `claudeHookCache`; install tests assert exact top-level keys and parse through `NewClaudeReader`. |
| 22 | Phase 2 verification gaps are closed without adding broad setup aliases beyond the hook-internal cache-writer command. | ✓ VERIFIED | `main.go:41-54` only adds `claude-hook-cache-writer` and `install-claude-hook`; invalid writer args return usage/exit 2; no `install`, `setup`, or `help` alias exists. |

**Score:** 22/22 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/sources/window.go` | Shared quota window and typed source error contract | ✓ VERIFIED | Exists and is used by Claude/Codex readers. |
| `internal/sources/claude.go` | Claude cache reader for documented cache shape | ✓ VERIFIED | Validates `five_hour`, `seven_day`, and `written_at`; returns stale metadata and typed errors. |
| `internal/sources/claude_test.go` | Synthetic Claude cache behavior tests | ✓ VERIFIED | Covers valid, missing, malformed, incomplete, stale, and unreadable cache cases. |
| `internal/sources/codex.go` | Codex rollout JSONL reader | ✓ VERIFIED | Recursively discovers rollouts, sorts by mtime, skips noisy events, returns normalized windows. |
| `internal/sources/codex_test.go` | Synthetic Codex parser and rollout selection tests | ✓ VERIFIED | Covers newest usable data, null/malformed skipping, fallback, no usable event, and metadata. |
| `internal/install/claude_hook.go` | Nested Claude command-hook shape plus app-owned cache writer | ✓ VERIFIED | Substantive, wired from CLI, uses marker-required ownership, nested command hooks, cache writer, and atomic writes. |
| `internal/install/claude_hook_test.go` | Synthetic tests for hook safety and cache writer output | ✓ VERIFIED | Tests nested hook shape, old managed entry replacement, executable path persistence, quoted cache path, reader-compatible cache, and malformed-input safety. |
| `cmd/llm-quota/main.go` | CLI dispatch and pre-TUI setup prompt wiring | ✓ VERIFIED | Dispatches no-arg TUI, `install-claude-hook`, and hook-internal writer; rejects unknown/invalid args. |
| `cmd/llm-quota/main_test.go` | Command dispatch tests without Bubble Tea | ✓ VERIFIED | Covers install, prompt accept/decline, old hook upgrade prompt, quoted cache detection, writer dispatch, invalid args, and unknown args. |
| `internal/tui/view.go` | Specific placeholder/setup hint copy | ✓ VERIFIED | Renders four placeholder rows and install hint; no real data rendering expected until later phases. |
| `internal/tui/view_test.go` | Render regression tests for setup hints | ✓ VERIFIED | Asserts setup hint, no refresh hint, and width guards. |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `internal/sources/claude.go` | `internal/sources/window.go` | `Fetch` returns `[]Window` and typed `SourceError` values | ✓ WIRED | SDK key-link check passed. |
| `internal/sources/codex.go` | `internal/sources/window.go` | Fetch returns normalized Codex 5h and 7d windows | ✓ WIRED | SDK key-link check passed. |
| `internal/install/claude_hook.go` | `cmd/llm-quota/main.go` | Installer/check functions consumed by CLI wiring | ✓ WIRED | SDK key-link check passed; command edge imports `internal/install`. |
| `cmd/llm-quota/main.go` | `internal/install/claude_hook.go` | `install-claude-hook` command and first-launch prompt call installer | ✓ WIRED | SDK key-link check passed. |
| `internal/tui/view.go` | Phase 2 D-04 | Placeholder rows remain readable after declined setup | ✓ WIRED | Literal SDK pattern expected older copy `Claude: install hook`; actual accepted copy is `Claude: run install-claude-hook`, verified by code and tests. |
| `internal/install/claude_hook.go` | Claude Code settings `hooks.PostToolUse[].hooks[]` | `managedHook` returns matcher entry with nested command hook object | ✓ WIRED | Manual check verified `hooks` array with `{type: command, command: ...}` at `claude_hook.go:201-212`; SDK regex did not match across lines. |
| `internal/install/claude_hook.go` | `internal/sources/claude.go` | Cache writer emits cache JSON consumed by `NewClaudeReader(...).Fetch(...)` | ✓ WIRED | Manual check verified writer struct tags and tests invoking `sources.NewClaudeReader(cachePath).Fetch`; SDK regex did not match across lines. |
| `cmd/llm-quota/main.go` | `internal/install/claude_hook.go` | `claude-hook-cache-writer --cache <path>` dispatch invokes installer cache writer | ✓ WIRED | SDK key-link check passed; `main.go:86-95` invokes `install.RunClaudeHookCacheWriter`. |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| `internal/install/claude_hook.go` | `~/.cache/llm-quota/claude.json` | Claude Code hook stdin via installed command | Yes | ✓ FLOWING -- `RunClaudeHookCacheWriter` reads real hook payload fields, validates both windows, and writes normalized JSON atomically. |
| `internal/sources/claude.go` | `[]Window` | App-owned cache JSON file | Yes | ✓ FLOWING -- generated cache is parsed by `ClaudeReader` in tests. |
| `internal/sources/codex.go` | `[]Window` | Local Codex rollout JSONL files | Yes | ✓ FLOWING -- parser extracts token-count `rate_limits` from newest usable rollout. |
| `cmd/llm-quota/main.go` | Hook install decision | Terminal stdin prompt / explicit command | Yes | ✓ FLOWING -- prompt answer controls install/decline; explicit command bypasses TUI. |
| `internal/tui/view.go` | Placeholder row labels/footer copy | Static startup view | Static by Phase 2 design | ✓ VERIFIED |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Focused closure tests | `go test ./internal/install ./cmd/llm-quota -run 'TestInstallClaudeHook|TestClaudeHookCacheWriter|TestRunClaudeHookCacheWriter'` | PASS for both packages | ✓ PASS |
| Full Go tests | `go test ./...` | PASS for `cmd/llm-quota`, `internal/install`, `internal/sources`, and `internal/tui` | ✓ PASS |
| Vet gate | `go vet ./...` | Exit 0, no findings printed | ✓ PASS |
| Race gate | `go test -race ./...` | PASS for all packages | ✓ PASS |
| Hook cache writer smoke | `printf synthetic rate_limits | go run ./cmd/llm-quota claude-hook-cache-writer --cache <temp cache>` then `go test ./internal/sources -run TestClaudeFetch` | Command wrote a non-empty cache file under temp dir; source tests passed | ✓ PASS |
| Plan artifact verification | `gsd-sdk query verify.artifacts` for all five plans | 15/15 artifacts passed existence/substance checks | ✓ PASS |
| Plan key-link verification | `gsd-sdk query verify.key-links` for all five plans plus manual multiline checks | Literal SDK checks passed 5/8; 3 false negatives manually verified in code/tests | ✓ PASS |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|----------|
| CLD-01 | 02-03, 02-04 | User is prompted for permission to install the hook/cache writer during setup or first launch. | ✓ SATISFIED | Consent prompt in `main.go`; accept/decline tests cover ordering. |
| CLD-02 | 02-03, 02-04 | User can decline hook installation and still run the TUI with clear Claude placeholder rows. | ✓ SATISFIED | Decline state plus TUI start is tested; placeholder rows render. |
| CLD-03 | 02-03, 02-04, 02-05 | User can install/update only the `llm-quota`-owned Claude hook without overwriting unrelated config. | ✓ SATISFIED | Marker-required ownership, preservation tests, nested runnable command-hook shape, and old-entry upgrade are implemented. |
| CLD-04 | 02-01, 02-05 | User can get Claude quota data from `~/.cache/llm-quota/claude.json` after the hook has run. | ✓ SATISFIED | Hook writer produces reader-compatible cache and `ClaudeReader` parses it in tests. |
| SRC-01 | 02-02 | User can get Codex quota data from the most recent local Codex rollout JSONL file. | ✓ SATISFIED | `CodexReader` scans and returns newest usable rollout data. |
| SRC-02 | 02-02 | User sees Codex placeholder rows and concise hint when no usable Codex quota event exists. | ✓ SATISFIED | `ErrorNoUsableEvent` and startup placeholder rows/hints exist. |
| SRC-03 | 02-01, 02-04 | User sees Claude placeholder rows and concise hook/setup hint when Claude cache is missing/malformed/unavailable. | ✓ SATISFIED | Claude reader categorizes missing/malformed/read errors; TUI footer says `Claude: run install-claude-hook`. |
| TEST-01 | 02-01, 02-05 | Maintainer can verify Claude cache parsing for valid, missing, malformed, and stale cache files without touching real home-directory data. | ✓ SATISFIED | Synthetic `t.TempDir()` tests cover reader cases plus generated cache compatibility and unreadable cache behavior. |
| TEST-02 | 02-02 | Maintainer can verify Codex rollout parsing for newest-file selection, null rate limits, malformed events, and missing usable events. | ✓ SATISFIED | Synthetic Codex tests cover these cases. |

No additional Phase 2 requirement IDs were found in `.planning/REQUIREMENTS.md` beyond the listed nine.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| `cmd/llm-quota/main_test.go` | 129 | `cat > old-cache.json` | ℹ️ Info | Test fixture for upgrading the previous broken managed entry; not an installed command in implementation. |
| `internal/install/claude_hook.go` | 187 | `return []any{}, nil` | ℹ️ Info | Empty slice default for missing hook event; not user-visible stub data. |
| `internal/sources/claude.go`, `internal/sources/codex.go` | 45, 164 | `return []Window{...}` | ℹ️ Info | Real parsed data construction, not hardcoded empty output. |

No blocker anti-patterns were found. No `TODO`, `FIXME`, `PLACEHOLDER`, or implementation `cat >` matches were found in Go implementation files.

### Human Verification Required

None for the Phase 2 decision. Actual tmux-pane/user install documentation validation is explicitly later Phase 5 scope; the Phase 2 hook shape, cache production, local parsers, and command wiring are verifiable in code and tests.

### Gaps Summary

No blocking gaps remain. The previous managed-hook shape and cache-writer gaps are closed with code evidence and focused regression tests. The phase goal is achieved: the user can choose whether to install the app-owned Claude hook, the installer preserves unrelated Claude configuration, the installed command writes a ClaudeReader-compatible cache atomically, and Codex local rollout parsing is implemented with synthetic coverage.

---

_Verified: 2026-05-18T15:33:12Z_
_Verifier: the agent (gsd-verifier)_
