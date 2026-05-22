# Roadmap: llm-quota

## Overview

The v1.1 roadmap improves the shipped local-only quota pane without changing the data-source architecture. The milestone focuses on glanceability: clean right-side alignment, an added Claude Sonnet-only weekly limit row, per-source refreshed date/time lines, display preferences for one-provider use, optional solid bars, and refresh animation that makes updates visible without turning the pane into a noisy dashboard.

## Milestones

- [x] **v1.0 MVP** - Phases 1-6 shipped 2026-05-21. Full archive: [milestones/v1.0-ROADMAP.md](milestones/v1.0-ROADMAP.md)
- [ ] **v1.1 UI Polish and Small Features** - Phases 7-9 improve display polish, small preferences, refresh feedback, and real-pane validation.

## Phases

**Phase Numbering:**

- Integer phases continue from the previous milestone.
- v1.1 starts at Phase 7 because v1.0 shipped Phases 1-6.

- [ ] **Phase 7: Row Alignment, Claude Sonnet Limit, and Source Freshness** - Add the Claude Sonnet-only weekly row, clean up quota row layout, and add one last-refreshed date/time line per source group.
- [ ] **Phase 8: Display Preferences** - Add solid-bar rendering and Claude-only or Codex-only display controls.
- [ ] **Phase 9: Refresh Animation and Polish Validation** - Animate refresh transitions and verify the polished pane in tests and real tmux layouts.

## Phase Details

### Phase 7: Row Alignment, Claude Sonnet Limit, and Source Freshness

**Goal**: User can read quota rows, including the Claude Sonnet-only weekly limit when available, with a clean right column and source-level freshness information in compact tmux panes.
**Depends on**: v1.0 MVP
**Requirements**: CLD-05, CLD-06, POL-01, POL-02, POL-03, POL-04
**Success Criteria** (what must be TRUE):

1. Claude quota rendering includes a Sonnet-only weekly row/bar when the local Claude cache exposes that limit.
2. Claude quota rendering handles missing Sonnet-only weekly data with a clear placeholder or omission behavior.
3. Percent and reset countdown text line up cleanly even when one row shows values like `0h 54m` and another shows `21h 1m`.
4. Claude rows render one source-level last-refreshed date/time line under the Claude window rows.
5. Codex rows render one source-level last-refreshed date/time line under the Codex window rows.
6. Normal, narrow, and very narrow layouts remain readable without wrapping or incoherent gaps.
7. A refresh failure can surface a concise source-level hint while preserving last-known-good quota rows.

**Plans**: 2 plans

Plans:
**Wave 1**

- [x] 07-01-PLAN.md - Add the Claude Sonnet-only weekly row and refactor quota row layout widths, right-column alignment, and compact reset text rendering.

**Wave 2** *(blocked on Wave 1 completion)*

- [ ] 07-02-PLAN.md - Add source freshness rows and immediate last-known-good refresh-failure hints.

**UI hint**: yes

### Phase 8: Display Preferences

**Goal**: User can tailor the pane for personal use by choosing bar style and hiding unused providers without changing local source behavior.
**Depends on**: Phase 7
**Requirements**: CFG-01, CFG-02, CFG-03, CFG-04, DOC-01
**Success Criteria** (what must be TRUE):

1. User can run the TUI with the existing segmented bar style or a new solid bar style.
2. User can run a Claude-only view with Codex rows omitted.
3. User can run a Codex-only view with Claude rows omitted.
4. The app rejects or safely falls back from preferences that would hide both sources.
5. README or command help documents the new display options concisely.

**Plans**: 2 plans

Plans:
**Wave 1**

- [ ] 08-01-PLAN.md - Add the display preference model, CLI/help wiring, provider visibility filtering, and validation.

**Wave 2** *(blocked on Wave 1 completion)*

- [ ] 08-02-PLAN.md - Add solid-bar rendering, provider-specific layout behavior, and docs/help updates.

**UI hint**: yes

### Phase 9: Refresh Animation and Polish Validation

**Goal**: User can see refreshes happen through subtle bar animation while tests and real-pane checks prove the polished layout remains stable.
**Depends on**: Phase 8
**Requirements**: ANIM-01, ANIM-02, ANIM-03, ANIM-04, TEST-01, TEST-02, UAT-01
**Success Criteria** (what must be TRUE):

1. Initial load animates visible bars from empty to the current quota usage.
2. Manual refresh via `r` animates visible bars from empty to the refreshed usage.
3. Automatic refresh animates visible bars from their prior value to the refreshed value.
4. Hidden providers, missing data, stale data, and resize events do not produce broken animation states.
5. Deterministic tests cover alignment, freshness rows, provider visibility, bar modes, and animation state transitions.
6. Real tmux-pane validation records the polished view at small, normal, and narrow widths.

**Plans**: 2 plans

Plans:
**Wave 1**

- [ ] 09-01-PLAN.md - Implement refresh animation state, timing messages, and deterministic transition tests.

**Wave 2** *(blocked on Wave 1 completion)*

- [ ] 09-02-PLAN.md - Broaden render tests and record real tmux-pane validation for the polished display.

**UI hint**: yes

## Progress

**Execution Order:**
Phases execute in numeric order: 7 -> 8 -> 9

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. Foreground TUI Foundation | v1.0 | 2/2 | Complete | 2026-05-16 |
| 2. Standalone Local Data Sources | v1.0 | 5/5 | Complete | 2026-05-18 |
| 3. Refresh and Resilience Loop | v1.0 | 2/2 | Complete | 2026-05-19 |
| 4. Quota Display and Responsive Rendering | v1.0 | 2/2 | Complete | 2026-05-19 |
| 5. Install, Docs, and Real-Pane Validation | v1.0 | 2/2 | Complete | 2026-05-20 |
| 6. Claude Setup Uninstaller | v1.0 | 2/2 | Complete | 2026-05-21 |
| 7. Row Alignment, Claude Sonnet Limit, and Source Freshness | v1.1 | 1/2 | In Progress|  |
| 8. Display Preferences | v1.1 | 0/2 | Planned | - |
| 9. Refresh Animation and Polish Validation | v1.1 | 0/2 | Planned | - |

## Coverage

| Requirement | Phase | Status |
|-------------|-------|--------|
| CLD-05 | Phase 7 | Pending |
| CLD-06 | Phase 7 | Pending |
| POL-01 | Phase 7 | Pending |
| POL-02 | Phase 7 | Pending |
| POL-03 | Phase 7 | Pending |
| POL-04 | Phase 7 | Pending |
| CFG-01 | Phase 8 | Pending |
| CFG-02 | Phase 8 | Pending |
| CFG-03 | Phase 8 | Pending |
| CFG-04 | Phase 8 | Pending |
| DOC-01 | Phase 8 | Pending |
| ANIM-01 | Phase 9 | Pending |
| ANIM-02 | Phase 9 | Pending |
| ANIM-03 | Phase 9 | Pending |
| ANIM-04 | Phase 9 | Pending |
| TEST-01 | Phase 9 | Pending |
| TEST-02 | Phase 9 | Pending |
| UAT-01 | Phase 9 | Pending |

**Coverage:**

- v1.1 requirements: 18 total
- Mapped to phases: 18
- Unmapped: 0

## Next

Run `$gsd-discuss-phase 7` to clarify the row alignment and source freshness implementation, or `$gsd-plan-phase 7` to create the executable plan directly.
