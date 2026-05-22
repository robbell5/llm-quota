# llm-quota

## What This Is

`llm-quota` is a tiny terminal UI that shows current Claude Code and Codex subscription quota usage in one always-running screen. It is built for Rob as a dedicated tmux-pane tool that refreshes automatically and avoids network calls by reading local usage data.

The shipped v1 product shows all four primary rolling subscription windows: Claude Code 5-hour, Claude Code 7-day, Codex 5-hour, and Codex 7-day. Each row shows percent used, a colored progress bar, and the reset countdown. It also includes standalone Claude cache setup, safe uninstall/reinstall behavior, local Codex rollout parsing, last-known-good refresh behavior, and responsive rendering for narrow panes.

## Core Value

Rob can glance at one tmux pane and immediately know how close Claude Code and Codex are to their 5-hour and 7-day limits.

## Current State

**Shipped version:** v1.0 MVP on 2026-05-21

v1.0 is requirement-complete and archived. The milestone audit status is `tech_debt`: all 27 v1 requirements and all 6 phases are satisfied, with warning-level follow-up items recorded for later prioritization.

Phase 7 of v1.1 is complete: the quota view now has fixed right-column alignment, a persistent Claude Sonnet weekly row/placeholder, source freshness lines under Claude and Codex groups, and concise source-level `refresh failed` status for preserved last-known-good rows.

**Archive:**

- [v1.0 roadmap archive](milestones/v1.0-ROADMAP.md)
- [v1.0 requirements archive](milestones/v1.0-REQUIREMENTS.md)
- [v1.0 audit archive](milestones/v1.0-MILESTONE-AUDIT.md)

## Current Milestone: v1.1 UI Polish and Small Features

**Goal:** Make the shipped quota pane more polished, better aligned, and easier to tailor for one-provider use while preserving the local-only foreground TUI model.

**Target features:**

- Clean right-side alignment for percent and reset countdown text, including mixed-width values such as `0h` and `21h`.
- A Claude Sonnet-only weekly limit row/bar when local Claude quota data exposes that limit.
- One source-level last-refreshed date/time line under the Claude rows and one under the Codex rows.
- Optional solid progress bars in addition to the current segmented bar look.
- Provider visibility controls so users can run Claude-only or Codex-only views without placeholder clutter.
- Refresh animation that fills from empty on startup/manual refresh and animates from previous value to new value on automatic refresh.
- Small glanceability polish, including immediate refresh-failure hints when last-known-good rows remain visible.

## Requirements

### Validated

- All 27 v1.0 requirements are complete and archived in [milestones/v1.0-REQUIREMENTS.md](milestones/v1.0-REQUIREMENTS.md).
- Phase 02 validated Codex quota data parsing from local rollout JSONL files using synthetic fixtures.
- Phase 02 validated first-launch permission prompting and app-owned Claude setup installation without mutating unrelated Claude settings.
- Phase 03 validated automatic refresh, manual refresh, stale state, and per-source last-known-good preservation.
- Phase 04 validated all four quota rows, threshold progress bars, reset countdowns, missing/stale hints, and responsive layouts.
- Phase 05 validated install/setup/troubleshooting documentation and real tmux-pane operation.
- Phase 06 validated safe Claude setup uninstall/reinstall behavior that preserves unrelated Claude settings and cache/state files.
- Phase 07 validated row alignment, Claude Sonnet-only weekly quota support, source freshness lines, compact/very narrow width behavior, and source-level refresh-failure hints.

### Active

- Add small display preferences for solid bars and Claude-only or Codex-only views.
- Add refresh animations that communicate updates without distracting from the tmux-pane use case.
- Update tests and documentation so the new display behavior is deterministic and discoverable.

### Out of Scope

- Usage history or graphing -- the goal is a glanceable current-status pane.
- Forecasting or projections -- v1 reports current usage and reset times only.
- Alerts or notifications -- thresholds are visual only.
- Multi-account support -- this is for Rob's local machine and current accounts.
- General per-model breakdowns beyond the explicit Claude Sonnet-only weekly limit -- the target view stays focused on subscription windows and one known Claude-specific cap.
- A daemon or background service -- the tool is a foreground Bubble Tea program.
- A one-shot mode -- the tool always runs in its TUI event loop.
- Network fallback for Claude or Codex -- local data keeps the tool small and low-friction.
- Reading Claude credentials or macOS Keychain data -- avoids credential prompts and secret handling.
- Depending on Rob's custom Claude statusline script -- the app installs and removes its own managed cache writer.

## Context

The project is based on the design spec at `docs/superpowers/specs/2026-05-16-llm-quota-tui-design.md`. Both data sources are local: Codex exposes quota data in session rollout JSONL files, while Claude quota data is written to `~/.cache/llm-quota/claude.json` by an app-owned statusline cache writer installed by `llm-quota`.

The implementation is intentionally small. Go and Bubble Tea are chosen because Rob wants to learn Bubble Tea and because an always-running terminal pane fits Bubble Tea's event loop model. The Go project owns the TUI, source readers, Claude setup flow, and Claude setup uninstaller.

The TUI should never crash because quota data is missing, stale, or malformed. Old data with a clear warning is more useful than a blank display, and missing first-run data should produce readable placeholder rows plus actionable footer hints.

## Constraints

- **Tech stack**: Use Go with Bubble Tea, Bubbles, Lip Gloss, and `golang.org/x/sync/errgroup` -- this supports the learning goal and keeps runtime dependencies focused.
- **Data access**: Use local files only at steady state -- avoids OAuth, Keychain prompts, platform-specific credential reads, and network dependencies.
- **Runtime model**: Always-running foreground TUI -- intended to live in a dedicated tmux pane.
- **Refresh behavior**: Refresh every 30 seconds and on explicit user action or resize -- keeps quota information current without creating a distracting loop.
- **Display footprint**: Fit comfortably in a small terminal pane -- the view should work around 50 columns and degrade below that.
- **Failure tolerance**: Source errors must not crash the program -- render placeholders or last-known-good data with hints.
- **Standalone setup**: Installing or first launching the TUI should prompt for permission to install the Claude hook so a new user can get the cache producer and viewer set up together.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Use Go and Bubble Tea | Rob wants to learn Bubble Tea, and the app is naturally event-loop driven. | Good; shipped v1.0 |
| Use local data sources only | Keeps the tool small and avoids credential prompts, OAuth fallback paths, and network dependencies. | Good; shipped v1.0 |
| Install an app-owned Claude hook/cache writer | The TUI must work for users without Rob's custom statusline while still avoiding OAuth, Keychain reads, and network calls. | Validated in Phase 02 and Phase 05 |
| Prompt for Claude hook installation during setup or first launch | A new user should be able to install the viewer and required Claude cache producer in one flow. | Validated in Phase 02 |
| Read Codex data from the newest rollout JSONL | Codex writes quota data locally during interactive sessions. | Validated in Phase 02 |
| Keep last-known-good data on refresh failure | An old number with a warning is more useful than blanking the display. | Validated in Phase 03; footer hint polish remains tech debt |
| Use calm threshold colors instead of alerts | The pane should stay glanceable without noisy badges or warning words. | Validated in Phase 04 |
| Start with Bubble Tea alt-screen | Cleaner for a dedicated tmux pane. | Good for v1.0 real-pane validation |
| Use Claude statusline wrapper for quota cache capture | Real validation showed quota `rate_limits` are available to the statusline command, not the earlier PostToolUse hook path. | Validated in Phase 05 |
| Preserve symlinked Claude settings during install | Rob's settings are dotfiles-managed; installer writes must update the target without replacing the symlink. | Validated in Phase 05 |
| Provide a safe Claude setup uninstaller | Users need a reversible setup flow that removes only app-owned configuration and leaves local cache/state files intact. | Validated in Phase 06 |
| Keep v1.1 focused on UI polish and small preferences | The v1.0 data-source model works; this milestone should make the pane easier to read and tailor without adding provider complexity. | In progress; Phase 7 complete |
| Add Claude Sonnet-only weekly limit as a targeted row | Rob specifically wants the known Claude weekly Sonnet cap visible, while broad per-model breakdowns remain out of scope. | Validated in Phase 07 |

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

*Last updated: 2026-05-22 after completing Phase 07*
