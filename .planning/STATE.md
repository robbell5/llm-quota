# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-05-16)

**Core value:** Rob can glance at one tmux pane and immediately know how close Claude Code and Codex are to their 5-hour and 7-day limits.
**Current focus:** Phase 1: Foreground TUI Foundation

## Current Position

Phase: 1 of 5 (Foreground TUI Foundation)
Plan: TBD of TBD in current phase
Status: Ready to plan
Last activity: 2026-05-16 — Roadmap created from v1 requirements, research, architecture guidance, and corrected standalone Claude hook design.

Progress: [░░░░░░░░░░] 0%

## Performance Metrics

**Velocity:**

- Total plans completed: 0
- Average duration: -
- Total execution time: 0.0 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 1. Foreground TUI Foundation | 0/TBD | - | - |
| 2. Standalone Local Data Sources | 0/TBD | - | - |
| 3. Refresh and Resilience Loop | 0/TBD | - | - |
| 4. Quota Display and Responsive Rendering | 0/TBD | - | - |
| 5. Install, Docs, and Real-Pane Validation | 0/TBD | - | - |

**Recent Trend:**

- Last 5 plans: none yet
- Trend: Not started

*Updated after each plan completion*

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table. Recent decisions affecting current work:

- Use Go with Bubble Tea, Bubbles, Lip Gloss, and `golang.org/x/sync/errgroup`.
- Use local files only at steady state; no network, OAuth, Keychain, daemon, or statusline runtime dependency.
- Claude quota capture must be standalone through an app-owned hook/cache writer installed only after permission.
- Preserve last-known-good data and render placeholders/hints instead of crashing or blanking the screen.

### Pending Todos

None yet.

### Blockers/Concerns

- Codex rollout JSONL and Claude hook cache shapes are private/local contracts, so source parsing needs tolerant fixture coverage.
- Actual tmux ergonomics for alt-screen, 30-second cadence, and footer density still need manual validation in the final phase.

## Deferred Items

Items acknowledged and carried forward from previous milestone close:

| Category | Item | Status | Deferred At |
|----------|------|--------|-------------|
| *(none)* | | | |

## Session Continuity

Last session: 2026-05-16
Stopped at: Roadmap and initial state created; next step is phase planning for Phase 1.
Resume file: None
