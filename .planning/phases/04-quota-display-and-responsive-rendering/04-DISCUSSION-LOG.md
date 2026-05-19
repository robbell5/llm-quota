# Phase 4: Quota Display and Responsive Rendering - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md; this log preserves the alternatives considered.

**Date:** 2026-05-19T18:23:35Z
**Phase:** 4-Quota Display and Responsive Rendering
**Areas discussed:** Responsive row layout, Urgency styling, Missing/stale hints, Progress bar style

---

## Responsive Row Layout

| Question | Options Presented | User's Choice |
|----------|-------------------|---------------|
| At the normal target width around 50 columns, how much should each row show? | Compact full row; No bar at 50; Wide only bars | Compact full row |
| Below 50 columns, what should the first layout tradeoff be? | Shorten text first; Shrink bar first; Drop reset first | Shorten text first |
| At very narrow widths under 30 columns, what is the minimum useful row? | Label percent reset; Label percent only; Percent reset only | Label percent reset |
| How precise should reset countdowns be in the row text? | Two-part time; Single token; Adaptive precision | Two-part time |

**Notes:** Preserve the full quota signal at the target width. Narrow layouts should trade text length before dropping the bar.

---

## Urgency Styling

| Question | Options Presented | User's Choice |
|----------|-------------------|---------------|
| Where should threshold color appear for normal data rows? | Bar and percent; Bar only; Whole row accent | Bar and percent |
| Should the threshold cutoffs stay exactly as the design spec says? | Spec cutoffs; More conservative; You decide | Spec cutoffs |
| How should red/high-usage rows behave visually? | Color only; Add marker; Bold percent | Color only |
| When a row is stale but still has usable data, should urgency color remain based on the percent? | Keep urgency color; Mute stale rows; Stale overrides color | Keep urgency color |

**Notes:** Red is a visual state, not an alert system. Stale data keeps its quota urgency signal.

---

## Missing/Stale Hints

| Question | Options Presented | User's Choice |
|----------|-------------------|---------------|
| Where should source problem details live when a source has no usable data? | Rows plus footer; Footer only; Rows only | Rows plus footer |
| How specific should missing/malformed/no-event wording be in the tiny UI? | Action over diagnosis; Show category names; Detailed diagnostics | Action over diagnosis |
| How should stale-but-valid data be signaled? | Footer age note; Inline stale marker; Both row and footer | Footer age note |
| If multiple footer hints apply at once, what should be prioritized? | Most actionable first; Stale first; Rotate all hints | Most actionable first |

**Notes:** Rows can show placeholders and terse state. Footer should prioritize the next useful action.

---

## Progress Bar Style

| Question | Options Presented | User's Choice |
|----------|-------------------|---------------|
| Which progress bar implementation should Phase 4 lock in? | Use Bubbles progress; Hand-rolled bar; Start Bubbles, allow swap | Use Bubbles progress |
| Should the progress bar be static or animated? | Static only; Animate changes; You decide | Static only |
| How should the unfilled part of the bar look? | Subtle track; Blank space; High-contrast track | Subtle track |
| How should tests treat progress bars? | Golden plain text; Assert ANSI colors; Helper-level tests | Golden plain text |

**Notes:** Context7 check confirmed Bubbles v2 progress supports static `ViewAs`, width setters, and Lip Gloss colors.

---

## the agent's Discretion

None.

## Deferred Ideas

None.
