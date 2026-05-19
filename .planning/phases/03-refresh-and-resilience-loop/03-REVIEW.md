---
phase: 03-refresh-and-resilience-loop
reviewed: 2026-05-19T15:43:34Z
depth: standard
files_reviewed: 7
files_reviewed_list:
  - cmd/llm-quota/main.go
  - cmd/llm-quota/main_test.go
  - internal/tui/model.go
  - internal/tui/update.go
  - internal/tui/update_test.go
  - internal/tui/view.go
  - internal/tui/view_test.go
findings:
  critical: 2
  warning: 1
  info: 0
  total: 3
status: issues_found
---

# Phase 03: Code Review Report

**Reviewed:** 2026-05-19T15:43:34Z
**Depth:** standard
**Files Reviewed:** 7
**Status:** issues_found

## Summary

Reviewed the listed command startup, TUI model/update loop, rendering, and tests. The refresh loop preserves last-known-good data and coalesces duplicate refresh requests, but the current UI omits a core quota signal and misreports expired reset times. Startup also records a persistent hook-decline state when stdin reaches EOF without an explicit user refusal.

## Critical Issues

### CR-01: BLOCKER - Quota rows never render progress bars

**File:** `internal/tui/view.go:123-148`
**Issue:** `renderDataRow` renders only label, percentage, and reset text. The project contract says each quota row shows percent used, a colored progress bar, and reset countdown; the Bubbles progress dependency is present but unused. This means the main TUI screen cannot deliver the glanceable quota visualization required for v1.
**Fix:** Add a width-aware progress bar to data rows at the full-width breakpoint and keep the narrow fallback for very small panes. For example:

```go
bar := progress.New(
	progress.WithDefaultGradient(),
	progress.WithWidth(barWidth),
)
barText := bar.ViewAs(clamp(window.UsedPercent/100, 0, 1))

return fmt.Sprintf(
	"%s  %s  %s  reset %s",
	labelStyle.Render(fmt.Sprintf("%-9s", fullLabel)),
	barText,
	percent,
	reset,
)
```

### CR-02: BLOCKER - Expired reset times are displayed as one minute remaining

**File:** `internal/tui/view.go:156-172`
**Issue:** `resetText` clamps negative durations to zero, then forces any sub-minute value to at least `1m`. A reset time that is already reached or in the past is therefore rendered as `1m`, which falsely tells the user there is still time remaining.
**Fix:** Return an explicit zero/now value before applying the minimum one-minute display to positive durations:

```go
remaining := resetsAt.Sub(now)
if remaining <= 0 {
	return "0m"
}
minutes := int(math.Ceil(remaining.Minutes()))
if minutes < 1 {
	minutes = 1
}
return fmt.Sprintf("%dm", minutes)
```

## Warnings

### WR-01: WARNING - EOF during first-launch prompt is persisted as a user decline

**File:** `cmd/llm-quota/main.go:133-154`
**Issue:** `ReadString('\n')` treats `io.EOF` as a valid empty answer, then the non-yes path records `RecordClaudeHookDeclined`. If the app starts with closed/non-interactive stdin, or input is accidentally exhausted, it permanently suppresses the Claude hook setup prompt even though the user never declined.
**Fix:** Only record a decline after an explicit non-empty answer. For EOF with no input, skip recording and continue to the TUI so the prompt can be offered on a later interactive launch:

```go
answer, err := bufio.NewReader(streams.Stdin).ReadString('\n')
if err != nil && err != io.EOF {
	fmt.Fprintf(streams.Stderr, "llm-quota: %v\n", err)
	return 1, true
}
answer = strings.TrimSpace(answer)
if isYes(answer) {
	// install hook
}
if answer == "" && err == io.EOF {
	return 0, false
}
if err := deps.RecordClaudeHookDeclined(paths.StatePath); err != nil {
	fmt.Fprintf(streams.Stderr, "llm-quota: could not record Claude hook decline: %v\n", err)
}
```

---

_Reviewed: 2026-05-19T15:43:34Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: standard_
