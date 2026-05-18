# Roadmap: llm-quota

## Overview

The v1 roadmap delivers a small, standalone, local-file-backed Go/Bubble Tea TUI for Claude Code and Codex quota windows. The phases move from a compiling foreground TUI spine, through standalone Claude hook setup and local source readers, into refresh semantics, responsive rendering, and final install/documentation validation. The roadmap keeps the corrected premise explicit: Claude quota capture is owned by `llm-quota` through a prompted hook/cache writer and does not depend on Rob's custom statusline.

## Phases

**Phase Numbering:**

- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Foreground TUI Foundation** - Create the compiling Go/Bubble Tea app spine with clean foreground runtime and quit behavior.
- [x] **Phase 2: Standalone Local Data Sources** - Add the app-owned Claude hook/cache setup and defensive Claude/Codex source readers. (completed 2026-05-18)
- [ ] **Phase 3: Refresh and Resilience Loop** - Make the running TUI refresh automatically and manually while preserving last-known-good data.
- [ ] **Phase 4: Quota Display and Responsive Rendering** - Render all quota windows with thresholds, placeholders, footer hints, and narrow-pane layouts.
- [ ] **Phase 5: Install, Docs, and Real-Pane Validation** - Wire real defaults and document setup/troubleshooting for a usable tmux-pane release.

## Phase Details

### Phase 1: Foreground TUI Foundation

**Goal**: User can start and stop a minimal foreground `llm-quota` TUI with a pinned, coherent Go/Bubble Tea stack.
**Depends on**: Nothing (first phase)
**Requirements**: TUI-01, TUI-04
**Success Criteria** (what must be TRUE):

  1. User can run `llm-quota` and see a stable foreground TUI screen instead of a one-shot command.
  2. User can exit cleanly with `q` without leaving terminal output in a broken state.
  3. User can exit cleanly with `Ctrl-C` without a panic or hung process.

**Plans**: 2 plans
Plans:
**Wave 1**

- [x] 01-01-PLAN.md — Pin the Go module, command entrypoint, and Bubble Tea quit/update spine.

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 01-02-PLAN.md — Render the Phase 1 startup screen and verify launch/quit behavior.

**UI hint**: yes

### Phase 2: Standalone Local Data Sources

**Goal**: User can choose whether to install the app-owned Claude hook and the app can read both local quota sources without relying on Rob's statusline.
**Depends on**: Phase 1
**Requirements**: CLD-01, CLD-02, CLD-03, CLD-04, SRC-01, SRC-02, SRC-03, TEST-01, TEST-02
**Success Criteria** (what must be TRUE):

  1. User is asked for permission before `llm-quota` installs or updates its Claude hook/cache writer.
  2. User can decline Claude hook installation and still run the TUI with clear Claude placeholder rows.
  3. User can install or update only the `llm-quota`-owned Claude hook without overwriting unrelated Claude configuration.
  4. User can see Claude quota data after the hook writes `~/.cache/llm-quota/claude.json` and Codex quota data from the newest local rollout JSONL.
  5. Maintainer can verify Claude and Codex parser behavior with synthetic fixtures for valid, missing, malformed, stale, null, and no-usable-event cases.

**Plans**: 4 plans
Plans:
**Wave 1**

- [x] 02-01-PLAN.md — Create normalized source contracts and the Claude cache reader.
- [x] 02-03-PLAN.md — Create the safe, idempotent Claude hook installer package.

**Wave 2** *(blocked on Wave 1 source contract completion)*

- [x] 02-02-PLAN.md — Implement the local Codex rollout JSONL reader.

**Wave 3** *(blocked on source readers and installer completion)*

- [x] 02-04-PLAN.md — Wire setup behavior into the command edge and preserve placeholder hints.
**UI hint**: yes

### Phase 3: Refresh and Resilience Loop

**Goal**: User can leave the TUI running and trust that refreshes update available data without blanking useful rows during temporary source failures.
**Depends on**: Phase 2
**Requirements**: SRC-04, SRC-05, TUI-02, TUI-03, TEST-03
**Success Criteria** (what must be TRUE):

  1. User sees quota data refresh automatically every 30 seconds while the TUI keeps running.
  2. User can press `r` to refresh immediately without disrupting the next scheduled refresh.
  3. User continues seeing last-known-good rows when a later refresh fails for Claude, Codex, or both.
  4. User sees stale-data warnings when displayed quota data is older than the accepted freshness threshold.
  5. Maintainer can verify refresh merge behavior preserves last-known-good data after source failures.

**Plans**: TBD
**UI hint**: yes

### Phase 4: Quota Display and Responsive Rendering

**Goal**: User can glance at one tmux pane and understand all four Claude/Codex quota windows, including urgency, reset timing, missing data, and narrow layouts.
**Depends on**: Phase 3
**Requirements**: DISP-01, DISP-02, DISP-03, DISP-04, DISP-05, DISP-06, TUI-05, TUI-06, TEST-04
**Success Criteria** (what must be TRUE):

  1. User can see Claude Code 5-hour, Claude Code 7-day, Codex 5-hour, and Codex 7-day quota rows in the TUI.
  2. User can see percent used, a colored progress bar, and reset countdown for each available quota window.
  3. User can interpret quota urgency from green, yellow, and red threshold styling.
  4. User can resize the tmux pane and still read useful quota status, including very narrow panes where bars are omitted.
  5. User sees helpful placeholder rows and footer hints when source data is missing, malformed, stale, or temporarily unavailable.

**Plans**: TBD
**UI hint**: yes

### Phase 5: Install, Docs, and Real-Pane Validation

**Goal**: User can install the binary, complete standalone Claude hook setup, troubleshoot missing data, and validate the TUI in the intended tmux-pane environment.
**Depends on**: Phase 4
**Requirements**: DOC-01, DOC-02
**Success Criteria** (what must be TRUE):

  1. User can install the `llm-quota` binary from documented instructions.
  2. User can complete Claude hook setup from documented instructions without needing Rob's custom statusline.
  3. User can troubleshoot missing Claude or Codex data using the same placeholder hints shown in the TUI.
  4. User can run the finished app in a dedicated tmux pane and confirm the default cadence, quit keys, and responsive layout feel usable.

**Plans**: TBD
**UI hint**: yes

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foreground TUI Foundation | 2/2 | Complete | 2026-05-16 |
| 2. Standalone Local Data Sources | 4/4 | Complete   | 2026-05-18 |
| 3. Refresh and Resilience Loop | 0/TBD | Not started | - |
| 4. Quota Display and Responsive Rendering | 0/TBD | Not started | - |
| 5. Install, Docs, and Real-Pane Validation | 0/TBD | Not started | - |
