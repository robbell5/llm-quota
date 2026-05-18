---
phase: 02-standalone-local-data-sources
reviewed: 2026-05-18T13:54:08Z
depth: standard
files_reviewed: 11
files_reviewed_list:
  - cmd/llm-quota/main.go
  - cmd/llm-quota/main_test.go
  - internal/install/claude_hook.go
  - internal/install/claude_hook_test.go
  - internal/sources/claude.go
  - internal/sources/claude_test.go
  - internal/sources/codex.go
  - internal/sources/codex_test.go
  - internal/sources/window.go
  - internal/tui/view.go
  - internal/tui/view_test.go
findings:
  critical: 2
  warning: 2
  info: 0
  total: 4
status: issues_found
---

# Phase 2: Code Review Report

**Reviewed:** 2026-05-18T13:54:08Z
**Depth:** standard
**Files Reviewed:** 11
**Status:** issues_found

## Summary

Reviewed the CLI setup flow, Claude hook installer, Claude/Codex local source readers, shared source types, and startup placeholder rendering. The source readers are mostly defensive, but the Claude installer currently cannot produce the cache contract that the Claude reader accepts, and its installed hook entry is not in Claude Code's command-hook shape. There are also two robustness issues around Codex rollout ordering and non-interactive first launch behavior.

## Critical Issues

### CR-01: BLOCKER - Installed Claude hook writes the wrong cache format and is not atomic

**File:** `internal/install/claude_hook.go:200-210`

**Issue:** `managedHookCommand` installs `mkdir -p ... && cat > <cachePath>`, which copies Claude's raw hook stdin directly into `~/.cache/llm-quota/claude.json`. The Claude reader only accepts a top-level cache with `five_hour`, `seven_day`, and `written_at` (`internal/sources/claude.go:51-55`), so the setup flow can install successfully while still leaving the TUI with a malformed Claude source. This also violates the documented atomic cache write requirement because shell redirection truncates/writes the final cache file in place rather than using tmpfile plus rename.

**Fix:** Install a hook command that runs an app-owned cache writer which reads Claude hook stdin, extracts `rate_limits.five_hour` and `rate_limits.seven_day`, and writes the documented cache through an atomic temp-file rename. For example:

```go
func managedHookCommand(cachePath string) string {
	return "llm-quota claude-hook-cache-writer --cache " + shellQuote(cachePath)
}
```

Then implement that writer so it emits exactly the `ClaudeReader` cache contract and uses the same atomic write discipline as the installer.

### CR-02: BLOCKER - Hook entry shape is not a Claude Code command hook

**File:** `internal/install/claude_hook.go:200-206`

**Issue:** The installer appends a hook entry with top-level `command` and marker fields. Claude Code hook entries require a matcher entry containing a `hooks` array with command hook objects, so this installed entry is not a runnable command hook. A user can accept the prompt, receive “installed llm-quota Claude hook”, and still never get a cache producer.

**Fix:** Preserve the explicit app-owned marker, but put the command in the actual command-hook structure and update detection to look inside managed entries. For example:

```go
func managedHook(cachePath string) map[string]any {
	return map[string]any{
		"name":             managedHookName,
		"llm_quota_marker": managedHookMarker,
		"matcher":          "*",
		"hooks": []any{
			map[string]any{
				"type":    "command",
				"command": managedHookCommand(cachePath),
			},
		},
	}
}
```

Add an installer test that asserts the written config uses the command-hook shape Claude will execute.

## Warnings

### WR-01: WARNING - Codex rollout selection is nondeterministic when mtimes tie

**File:** `internal/sources/codex.go:84-86`

**Issue:** Rollout candidates are sorted only by modification time with `sort.Slice`. If two rollout files have the same mtime, which is plausible when files are copied, restored, or created within filesystem timestamp precision, the ordering is unspecified and the reader can pick different usable quota data across runs.

**Fix:** Add a deterministic tie-breaker after the mtime comparison:

```go
sort.Slice(candidates, func(i, j int) bool {
	if candidates[i].modTime.Equal(candidates[j].modTime) {
		return candidates[i].path > candidates[j].path
	}
	return candidates[i].modTime.After(candidates[j].modTime)
})
```

### WR-02: WARNING - Closed stdin records a permanent hook decline without user input

**File:** `cmd/llm-quota/main.go:110-130`

**Issue:** `ReadString('\n')` treats `io.EOF` as a usable empty answer. If the app is launched in a non-interactive context with stdin already closed, the code falls through to the decline branch and records `claude_hook_declined` permanently even though the user never answered the prompt.

**Fix:** Treat empty EOF as “no answer available” and skip recording a decline, or gate the prompt on an interactive stdin check. For example:

```go
answer, err := bufio.NewReader(streams.Stdin).ReadString('\n')
if err != nil {
	if err == io.EOF && strings.TrimSpace(answer) == "" {
		return 0, false
	}
	if err != io.EOF {
		fmt.Fprintf(streams.Stderr, "llm-quota: %v\n", err)
		return 1, true
	}
}
```

---

_Reviewed: 2026-05-18T13:54:08Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: standard_
