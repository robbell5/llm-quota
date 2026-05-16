# llm-quota

## What This Is

`llm-quota` is a tiny terminal UI that shows current Claude Code and Codex subscription quota usage in one always-running screen. It is built for Rob as a dedicated tmux-pane tool that refreshes automatically and avoids network calls by reading local usage data.

The v1 product shows all four rolling subscription windows: Claude Code 5-hour, Claude Code 7-day, Codex 5-hour, and Codex 7-day. Each row shows percent used, a colored progress bar, and the reset countdown.

## Core Value

Rob can glance at one tmux pane and immediately know how close Claude Code and Codex are to their 5-hour and 7-day limits.

## Requirements

### Validated

(None yet -- ship to validate)

### Active

- [ ] Show Claude Code 5-hour quota usage with percent, progress bar, and reset countdown.
- [ ] Show Claude Code 7-day quota usage with percent, progress bar, and reset countdown.
- [ ] Show Codex 5-hour quota usage with percent, progress bar, and reset countdown.
- [ ] Show Codex 7-day quota usage with percent, progress bar, and reset countdown.
- [ ] Refresh quota data automatically every 30 seconds while running.
- [ ] Refresh immediately when the user presses `r` or the terminal pane is resized.
- [ ] Exit cleanly on `q` or `Ctrl-C`.
- [ ] Read Codex quota data from the most recent local rollout JSONL file.
- [ ] Read Claude quota data from a local cache file written by the existing Claude statusline script.
- [ ] Keep rendering last-known-good data when a source temporarily fails.
- [ ] Render helpful placeholder rows and footer hints when source data is missing or malformed.
- [ ] Adapt layout for narrow tmux panes, including dropping bars at very narrow widths.
- [ ] Provide tests for source parsing, stale data handling, and rendered output.

### Out of Scope

- Usage history or graphing -- the goal is a glanceable current-status pane.
- Forecasting or projections -- v1 only reports current usage and reset times.
- Alerts or notifications -- thresholds are visual only.
- Multi-account support -- this is for Rob's local machine and current accounts.
- Per-model breakdowns -- v1 tracks product-level subscription windows only.
- A daemon or background service -- the tool is a foreground Bubble Tea program.
- A one-shot mode -- the tool always runs in its TUI event loop.
- Network fallback for Claude or Codex -- local data keeps the tool small and low-friction.
- Folding this view into the existing statusline -- this is a standalone tmux-pane tool.

## Context

The project is based on the design spec at `docs/superpowers/specs/2026-05-16-llm-quota-tui-design.md`. Both data sources are local: Codex exposes quota data in session rollout JSONL files, while Claude quota data will be written to `~/.cache/llm-quota/claude.json` by a small additive extension to the existing Claude statusline script.

The implementation is intentionally small. Go and Bubble Tea are chosen because Rob wants to learn Bubble Tea and because an always-running terminal pane fits Bubble Tea's event loop model. The Go project owns the TUI and source readers; the Claude statusline extension lives separately in the dotfiles repo.

The TUI should never crash because quota data is missing, stale, or malformed. Old data with a clear warning is more useful than a blank display, and missing first-run data should produce readable placeholder rows plus actionable footer hints.

## Constraints

- **Tech stack**: Use Go with Bubble Tea, Bubbles, Lip Gloss, and `golang.org/x/sync/errgroup` -- this supports the learning goal and keeps runtime dependencies focused.
- **Data access**: Use local files only at steady state -- avoids OAuth, Keychain prompts, platform-specific credential reads, and network dependencies.
- **Runtime model**: Always-running foreground TUI -- intended to live in a dedicated tmux pane.
- **Refresh behavior**: Refresh every 30 seconds and on explicit user action or resize -- keeps quota information current without creating a distracting loop.
- **Display footprint**: Fit comfortably in a small terminal pane -- the view should work around 50 columns and degrade below that.
- **Failure tolerance**: Source errors must not crash the program -- render placeholders or last-known-good data with hints.
- **Repository split**: Claude statusline changes belong in the dotfiles repo, not this Go project -- commit and review them separately.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Use Go and Bubble Tea | Rob wants to learn Bubble Tea, and the app is naturally event-loop driven. | -- Pending |
| Use local data sources only | Keeps the tool small and avoids credential prompts, OAuth fallback paths, and network dependencies. | -- Pending |
| Read Claude data from a statusline-written cache | Claude credentials are in macOS Keychain, while the statusline already receives rate-limit data. | -- Pending |
| Read Codex data from the newest rollout JSONL | Codex writes quota data locally during interactive sessions. | -- Pending |
| Keep last-known-good data on refresh failure | An old number with a warning is more useful than blanking the display. | -- Pending |
| Start with Bubble Tea alt-screen | Cleaner for a dedicated tmux pane, with a known spike to revisit if scrollback matters. | -- Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition**:

1. Requirements invalidated? -> Move to Out of Scope with reason
2. Requirements validated? -> Move to Validated with phase reference
3. New requirements emerged? -> Add to Active
4. Decisions to log? -> Add to Key Decisions
5. "What This Is" still accurate? -> Update if drifted

**After each milestone**:

1. Full review of all sections
2. Core Value check -- still the right priority?
3. Audit Out of Scope -- reasons still valid?
4. Update Context with current state

---

*Last updated: 2026-05-16 after initialization*
