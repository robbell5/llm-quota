# Requirements: llm-quota

**Defined:** 2026-05-21
**Milestone:** v1.1 UI Polish and Small Features
**Core Value:** Rob can glance at one tmux pane and immediately know how close Claude Code and Codex are to their 5-hour and 7-day limits.

## v1.1 Requirements

Requirements for the UI polish and small-features milestone. Each maps to one roadmap phase.

### Layout Polish

- [ ] **CLD-05**: User can see Claude Sonnet-only weekly quota usage as an additional Claude row when local Claude quota data exposes that limit.
- [ ] **CLD-06**: User sees a clear placeholder or omission behavior for the Claude Sonnet-only weekly row when local Claude quota data does not expose that limit.
- [ ] **POL-01**: User can read percent and reset countdown values in a cleanly aligned right column even when reset text uses different-width values such as `0h` and `21h`.
- [ ] **POL-02**: User can see one last-refreshed date/time line under the Claude rows and one last-refreshed date/time line under the Codex rows.
- [ ] **POL-03**: User can still read aligned rows and source freshness lines at normal, narrow, and very narrow pane widths without broken wrapping.
- [ ] **POL-04**: User can see a concise source-level hint when a refresh fails but last-known-good rows are still being displayed.

### Display Preferences

- [ ] **CFG-01**: User can choose the current segmented progress bar style or a solid progress bar style.
- [ ] **CFG-02**: User can hide Claude rows when only Codex quota is relevant.
- [ ] **CFG-03**: User can hide Codex rows when only Claude quota is relevant.
- [ ] **CFG-04**: User gets a clear fallback or validation error if display preferences would hide every source.

### Refresh Feedback

- [ ] **ANIM-01**: User sees quota bars fill from empty to the current usage level on initial load.
- [ ] **ANIM-02**: User sees quota bars fill from empty to the current usage level after pressing `r`.
- [ ] **ANIM-03**: User sees quota bars animate from previous usage to refreshed usage after an automatic refresh.
- [ ] **ANIM-04**: User can resize or refresh while animation is active without stale values, row wrapping, or broken layout.

### Documentation and Verification

- [ ] **DOC-01**: User can discover the solid bar and provider visibility options from README or command help.
- [ ] **TEST-01**: Maintainer can verify right-column alignment, source freshness rows, solid bars, and provider visibility through deterministic render tests.
- [ ] **TEST-02**: Maintainer can verify refresh animation behavior with injected time or tick messages without relying on wall-clock sleeps.
- [ ] **UAT-01**: Maintainer can validate the polished view in a real tmux pane at small, normal, and narrow widths.

## Future Requirements

Deferred to a later milestone unless they become necessary while implementing v1.1.

### Preferences

- **CFG-05**: User can persist display preferences in a dedicated config file instead of only startup flags or environment variables.
- **CFG-06**: User can toggle provider visibility from inside the running TUI.
- **CFG-07**: User can configure refresh interval after the default cadence continues to prove useful.

### Polish

- **POL-05**: User can choose additional color themes or a high-contrast mode.
- **POL-06**: User can set a terminal title for easier pane/window identification.
- **POL-07**: User can see richer source diagnostics beyond the compact source-level freshness lines.

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| New quota sources or OpenCode/OpenAI support | This milestone is UI polish for the shipped Claude/Codex source model. |
| Network fallback for Claude or Codex | The project remains local-file-only at steady state. |
| Alerts or notifications | v1.1 should improve glanceability, not add background attention mechanisms. |
| Historical charts or usage graphs | The product remains a current-status tmux pane. |
| Forecasting or burn-rate projection | Reset countdowns and current usage remain the trusted signal. |
| General per-model breakdowns | Only the explicitly requested Claude Sonnet-only weekly cap is in scope. |
| Full interactive settings UI | Startup preferences are enough for this small milestone; runtime settings can be considered later. |
| Persistent preference migration system | Too heavy for a small display-preference milestone. |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

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

---

*Requirements defined: 2026-05-21*
*Last updated: 2026-05-21 after roadmap creation*
