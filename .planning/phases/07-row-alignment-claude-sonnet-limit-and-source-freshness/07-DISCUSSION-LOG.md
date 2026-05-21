# Phase 7: Row Alignment, Claude Sonnet Limit, and Source Freshness - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md - this log preserves the alternatives considered.

**Date:** 2026-05-21
**Phase:** 7-Row Alignment, Claude Sonnet Limit, and Source Freshness
**Areas discussed:** Sonnet-only weekly row behavior, Right-column alignment and compact reset text, Source-level freshness line format, Refresh-failure hint behavior

---

## Sonnet-Only Weekly Row Behavior

| Question | Options Presented | User's Choice |
|----------|-------------------|---------------|
| When Claude local data does not expose the Sonnet-only weekly limit, should the TUI omit the Sonnet row, show a placeholder row, or let the agent decide? | Omit the Sonnet row; Show a placeholder row; You decide | Show a placeholder row |
| What should the Sonnet row label be? | Claude Sonnet 7d; Sonnet 7d; Cl Sonnet / Sonnet in compact layouts | Sonnet 7d |
| Where should the Sonnet row appear in the Claude group order? | Claude 5h, Claude 7d, Sonnet 7d; Claude 5h, Sonnet 7d, Claude 7d; You decide | Claude 5h, Claude 7d, Sonnet 7d |
| When the Claude cache includes a Sonnet limit, how strict should parsing be? | Optional field only; Require all three Claude rows once schema supports Sonnet; Show placeholder on malformed Sonnet only | Optional field only |

**Notes:** Sonnet should be a stable visual row but an optional parser/cache field, preserving existing Claude 5h and 7d behavior.

---

## Right-Column Alignment and Compact Reset Text

| Question | Options Presented | User's Choice |
|----------|-------------------|---------------|
| For normal-width panes, should the right side be treated as a fixed-width column containing both percent and reset countdown? | Fixed percent + fixed reset columns; Only align reset text; You decide | Fixed percent + fixed reset columns |
| How compact should reset countdowns become when space is tight? | Normal: `21h 1m`; compact: `21h`; very narrow: omit reset; Always show minutes when under 24h; Use terse tokens like `21h1m` | Normal can include minutes; compact drops to hours; very narrow may omit reset |
| At widths below the normal layout, should bars remain visible as long as possible? | Keep bars until they would crowd columns; Drop bars earlier for perfect text alignment; You decide | Keep bars until they would crowd columns |
| How should the Sonnet placeholder row participate in alignment? | Use same columns as real rows; Simpler missing-data text; You decide | Use same columns as real rows |

**Notes:** Text readability takes priority over bar visibility in tight widths.

---

## Source-Level Freshness Line Format

| Question | Options Presented | User's Choice |
|----------|-------------------|---------------|
| Where should freshness lines appear? | Under each source group; One combined freshness footer; Only show freshness when stale/erroring | Under each source group |
| What should the freshness line say at normal width? | `Claude updated 2:14 PM`; `Claude last refreshed at 2:14 PM`; `Updated 2:14 PM` under the group | `Claude updated 2:14 PM` |
| Should freshness use absolute time, relative age, or both? | Absolute time only; Relative age only; Both when wide | Both when wide |
| How should the freshness line behave in compact and very narrow layouts? | Compact: `Claude 2:14 PM`; very narrow: `Cl 2:14`; Compact: `Updated 2:14 PM`; very narrow: omit; You decide | Compact: `Claude 2:14 PM`; very narrow: `Cl 2:14` |

**Notes:** Freshness should remain tied to the source group and keep source identity even when compact.

---

## Refresh-Failure Hint Behavior

| Question | Options Presented | User's Choice |
|----------|-------------------|---------------|
| When a refresh fails but old rows are still shown, where should the current-error hint appear? | On the source freshness line; In the global footer; Both when width allows | On the source freshness line |
| How strong should the wording be? | `refresh failed`; `using saved data`; `open Claude` / `open Codex` | `refresh failed` |
| If both stale age and current refresh failure apply, which signal should win on the freshness line? | Show both when width allows; Show stale age first, omit failure in compact widths; Show refresh failed first, omit age in compact widths | Show both when width allows |
| Should the old footer hints still exist after source-level error text is added? | Keep footer only for missing first-run data; Keep footer for all source errors too; You decide | Keep footer only for missing first-run data |

**Notes:** Source lines should become the primary place for current and stale source status once rows exist.

---

## The Agent's Discretion

- Choose exact compact breakpoints and internal field naming during implementation, as long as the locked visible behavior is preserved.

## Deferred Ideas

None.
