# Phase 4: Quota Display and Responsive Rendering - Context

**Gathered:** 2026-05-19T18:23:35Z
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 4 delivers the visible quota dashboard: all four Claude Code and Codex quota windows, urgency styling, progress bars, reset timing, missing/stale hints, and layouts that remain readable in narrow tmux panes.

This phase clarifies how to render existing source/model data. It does not add new data sources, notifications, history, forecasting, multi-account support, statusline integration, or one-shot output.

</domain>

<decisions>
## Implementation Decisions

### Responsive Row Layout

- **D-01:** At the normal target width around 50 columns, rows should still show the compact full quota picture: product/window label, progress bar, percent, and short reset text.
- **D-02:** Below 50 columns, shorten text before removing the bar. Abbreviate labels and reset copy first so the visual quota cue survives as long as it remains useful.
- **D-03:** At very narrow widths under 30 columns, omit the bar but keep the minimum useful row as short label, percent, and reset token when it fits.
- **D-04:** Reset countdowns should use two-part time from the design spec: `2h 14m` under 24 hours and `4d 06h` for multi-day windows. Use narrower forms only when required to avoid wrapping.

### Urgency Styling

- **D-05:** Threshold color should appear on both the progress bar and percent text so urgency remains visible even when bars become short.
- **D-06:** Use the design-spec thresholds exactly: green below 60%, yellow from 60% to below 85%, red at 85% and above.
- **D-07:** High-usage rows should stay calm: use color only. Do not add alert markers, badges, blinking, or extra warning words for red rows.
- **D-08:** Stale-but-valid rows should keep urgency color based on their last-known percent. Staleness is explained through hints, not by muting or overriding the quota signal.

### Missing And Stale Hints

- **D-09:** When a source has no usable data, rows should show placeholders with terse source-state copy, and the footer should carry the actionable next step.
- **D-10:** UI copy should prioritize action over diagnosis. Prefer concise hints like install hook, open Claude, or open Codex instead of exposing raw categories such as malformed or no usable event.
- **D-11:** Stale-but-valid data should render normally in rows and add a footer age note, such as `Claude data 2h old; open Claude`.
- **D-12:** If multiple footer hints apply, show the most actionable recovery/setup hint first and keep footer content to what fits. Do not rotate hints.

### Progress Bar Style

- **D-13:** Use `charm.land/bubbles/v2/progress` for progress bars. The current Bubbles v2 API supports static `ViewAs`, width setters, and Lip Gloss colors, which match this phase.
- **D-14:** Progress bars should be static only. Quota values update on refresh; animation would distract in an always-running tmux pane and make tests more brittle.
- **D-15:** Render a colored fill over a subtle Catppuccin surface track. Avoid blank tracks and high-contrast tracks.
- **D-16:** Render tests should strip ANSI and assert plain row content, widths, threshold cases, missing/stale states, and narrow layouts. Do not make tests brittle by asserting exact ANSI escape sequences.

### the agent's Discretion

No user decisions were delegated back to the agent. Planner discretion remains limited to implementation mechanics that satisfy the locked behavior above.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Scope And Requirements

- `.planning/PROJECT.md` — Defines the core value, active requirements, local-only constraint, failure tolerance, width target, and prior validated decisions.
- `.planning/REQUIREMENTS.md` — Defines Phase 4 requirement IDs: DISP-01 through DISP-06, TUI-05, TUI-06, and TEST-04.
- `.planning/ROADMAP.md` — Defines the Phase 4 goal, success criteria, phase boundary, and dependency on Phase 3.

### Design Spec

- `docs/superpowers/specs/2026-05-16-llm-quota-tui-design.md` — Defines the four-row dashboard concept, progress bar requirement, threshold cutoffs, reset countdown format, stale-data rendering, failure rendering, and sizing behavior.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- `internal/tui/view.go` — Already owns fixed four-row rendering, width fallback, row label order, missing placeholders, reset text, footer switching, and line-width-safe layout helpers.
- `internal/tui/colors.go` — Already has Catppuccin Mocha base, surface, text, subtext, blue, yellow, and red colors. Phase 4 likely needs a green color and explicit threshold styles.
- `internal/tui/model.go` — Already stores per-source windows, source errors, refresh state, injected clock, and stale threshold.
- `internal/tui/update.go` — Already preserves last-known-good windows per source, records source errors, marks stale data, and stores terminal width/height on resize.
- `internal/sources/window.go` — `Window` already carries product, window kind, used percent, reset time, captured time, stale status, stale age, and optional metadata.
- `internal/tui/view_test.go` — Already strips ANSI for render assertions and checks target/narrow widths, source-backed rows, placeholder rendering, and Phase 3 copy exclusions.

### Established Patterns

- Rendering is currently plain helper functions in `internal/tui/view.go`; keep Phase 4 changes local unless a small helper improves clarity.
- Source readers do not render. TUI rendering maps `SourceError` categories to user-facing placeholders and hints.
- Last-known-good data is preserved independently per source. A source error should not blank existing rows for that source if usable windows remain.
- Resize currently triggers relayout only, not source refresh. Preserve the existing refresh semantics unless a later phase changes them.
- Existing tests prefer deterministic clocks and ANSI-stripped width/content assertions.

### Integration Points

- Update `renderRows`, `renderDataRow`, `resetText`, and footer rendering in `internal/tui/view.go` to implement Phase 4 behavior.
- Add or adjust styles in `internal/tui/colors.go` for threshold colors and subtle tracks.
- Use `charm.land/bubbles/v2/progress` with static `ViewAs` and width setters; do not introduce animated progress commands.
- Extend `internal/tui/view_test.go` for all-green, mixed-threshold, missing-source, stale-source, and narrow-width states.

</code_context>

<specifics>
## Specific Ideas

- The target pane around 50 columns should remain genuinely glanceable, not just technically unwrapped.
- The UI should feel calm and non-alerting even when a row is red; alerts and notifications remain out of scope.
- Footer content should be useful but bounded. Prefer the next action over internal diagnostic detail.
- Plain-text render tests are preferred over exact ANSI assertions so Bubbles/Lip Gloss implementation details do not make tests fragile.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 4-Quota Display and Responsive Rendering*
*Context gathered: 2026-05-19T18:23:35Z*
