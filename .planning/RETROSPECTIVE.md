# Project Retrospective

*A living document updated after each milestone. Lessons feed forward into future planning.*

## Milestone: v1.0 - MVP

**Shipped:** 2026-05-21
**Phases:** 6 | **Plans:** 15 | **Sessions:** not tracked

### What Was Built

- A foreground Go/Bubble Tea TUI that displays Claude Code and Codex 5-hour and 7-day quota windows.
- Local-only Claude and Codex source readers, including app-owned Claude cache setup and Codex rollout JSONL parsing.
- Refresh behavior with automatic 30-second updates, manual `r` refresh, stale state, and per-source last-known-good preservation.
- Responsive quota row rendering with threshold colors, progress bars, reset countdowns, placeholders, and footer hints.
- Install, setup, troubleshooting, real tmux-pane validation, and safe Claude setup uninstall/reinstall documentation.

### What Worked

- Keeping source readers path-injected made parser and setup tests independent from real home-directory data.
- Typed source errors let the TUI render concise recovery copy without exposing parser internals.
- The phase structure caught a missing uninstaller before the milestone closed, avoiding an irreversible setup story.
- Real tmux-pane UAT corrected assumptions about Claude quota capture and install instructions.

### What Was Inefficient

- Phase 1 lacked a verification artifact until milestone audit time, which created avoidable closeout work.
- Some UAT and verification artifacts still carry stale `partial` or `human_needed` statuses even though later phases closed the practical gaps.
- README wording around "statusline integration" drifted after the managed Claude statusline cache writer became the validated implementation path.

### Patterns Established

- Keep local source contracts normalized behind `sources.Window` so the TUI does not know private Claude/Codex JSON shapes.
- Preserve user configuration with marker-scoped ownership, backups only on content changes, and symlink-target-safe writes.
- Treat responsive terminal widths as testable behavior, including ANSI-stripped width checks.
- Record real-local validation results without committing private user configuration contents.

### Key Lessons

1. Verification artifacts should be created when a phase completes, not reconstructed during milestone audit.
2. Local-only integrations still need explicit uninstall and rollback paths when they modify user tool configuration.
3. Private local data formats need tolerant parsers plus clear documentation about what tools actually emit quota-window data.
4. "No statusline integration" was too imprecise once the app-owned Claude cache writer moved into the statusline command path.

### Cost Observations

- Model mix: not tracked.
- Sessions: not tracked.
- Notable: audit-driven closure surfaced planning-document drift more than product-code risk.

---

## Cross-Milestone Trends

### Process Evolution

| Milestone | Sessions | Phases | Key Change |
|-----------|----------|--------|------------|
| v1.0 | not tracked | 6 | Established phase plans, verification, UAT, and milestone archival for the project. |

### Cumulative Quality

| Milestone | Tests | Coverage | Zero-Dep Additions |
|-----------|-------|----------|-------------------|
| v1.0 | `go test ./... -count=1` passed at audit close | Parser, refresh, rendering, install, uninstall, and command-edge behavior | Local-only source readers and setup flow without network/OAuth dependencies |

### Top Lessons (Verified Across Milestones)

1. Keep data-source parsing, UI rendering, and command setup concerns separated so each can be tested without real user files.
2. Treat install and uninstall as one product surface for any tool that modifies another application's config.
