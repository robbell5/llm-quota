# Milestones

## v1.0 MVP (Shipped: 2026-05-21)

**Phases completed:** 6 phases, 15 plans, 34 tasks

**Archive:**

- Roadmap: [milestones/v1.0-ROADMAP.md](milestones/v1.0-ROADMAP.md)
- Requirements: [milestones/v1.0-REQUIREMENTS.md](milestones/v1.0-REQUIREMENTS.md)
- Audit: [milestones/v1.0-MILESTONE-AUDIT.md](milestones/v1.0-MILESTONE-AUDIT.md)

**Audit result:** requirement-complete with warning-level tech debt.

**Known deferred items at close:** 7 open artifact records acknowledged as deferred; see `STATE.md` Deferred Items.

**Key accomplishments:**

- Go/Bubble Tea command spine with clean quit behavior and pinned v2 dependencies
- Future-shaped startup screen with placeholder quota rows and width-aware key hints
- Path-injected Claude cache reader with normalized quota windows, stale metadata, and typed source errors
- Codex quota extraction from local rollout JSONL with tolerant event scanning, older-file fallback, and normalized 5h/7d windows
- Safe llm-quota-owned Claude hook installation with idempotent updates, backups, and remembered first-launch declines
- Pre-TUI Claude hook consent flow with explicit install-claude-hook dispatch and readable missing-data setup hints
- Runnable managed Claude hook plus an atomic cache writer for ClaudeReader data
- Bubble Tea refresh loop with injected source readers, per-source last-known-good merge, coalesced manual refresh, and one-hour stale state
- Real Claude/Codex reader startup wiring with minimal source-backed quota rows that preserve Phase 3 copy limits
- Four Claude/Codex quota rows with static Bubbles progress bars, threshold urgency colors, and two-part reset countdowns
- Responsive Claude/Codex quota rows with safe recovery footers for missing and stale local data
- README install and recovery guide for the local-only Claude/Codex tmux-pane quota monitor
- Approved real tmux-pane validation with corrected Claude statusline cache setup and local build instructions
- Safe Claude setup uninstaller with statusline restoration, legacy hook cleanup, and CLI dispatch coverage
- Claude setup uninstall instructions with approved real-local install -> uninstall -> reinstall validation

---
