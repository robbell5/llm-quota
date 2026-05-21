# Phase 7: Row Alignment, Claude Sonnet Limit, and Source Freshness - Context

**Gathered:** 2026-05-21
**Status:** Ready for planning

<domain>
## Phase Boundary

This phase improves the existing quota pane layout and Claude/Codex source status rendering. It adds a Claude Sonnet-only weekly row when local Claude quota data exposes that limit, keeps a stable placeholder row when it does not, aligns quota row percent/reset columns, adds one source-level freshness line under each provider group, and surfaces concise refresh-failure hints while preserving last-known-good rows.

This phase does not add provider visibility preferences, solid bars, refresh animation, new quota sources, alerts, history, forecasting, networking, OAuth, Keychain reads, or a daemon.

</domain>

<decisions>
## Implementation Decisions

### Sonnet-Only Weekly Row
- **D-01:** Show a `Sonnet 7d` placeholder row when Claude local data does not expose the Sonnet-only weekly limit. Do not omit the row.
- **D-02:** Use `Sonnet 7d` as the normal row label. Compact layouts may shorten the label to preserve readability.
- **D-03:** Order Claude rows as `Claude 5h`, `Claude 7d`, then `Sonnet 7d`.
- **D-04:** Treat the Sonnet weekly limit as an optional Claude cache field. Existing required `five_hour` and `seven_day` Claude windows remain valid even when Sonnet is absent. Add the Sonnet row as real data only when a known optional field is present and valid.

### Row Alignment and Compact Reset Text
- **D-05:** Use fixed percent and fixed reset columns in normal-width rows so mixed-width values such as `0h 54m` and `21h 1m` line up predictably.
- **D-06:** Normal reset text can include minutes. Compact layouts should drop reset text to hours. Very narrow layouts may omit reset text.
- **D-07:** Keep progress bars visible until they would crowd the fixed text columns. Text readability wins over bar visibility when space is tight.
- **D-08:** The Sonnet placeholder row should use the same aligned columns as real rows, reserving space for placeholder bar/marker, percent placeholder, and reset placeholder.

### Source Freshness Lines
- **D-09:** Place one source freshness line under each source group: Claude rows followed by Claude freshness, Codex rows followed by Codex freshness.
- **D-10:** At normal width, use text shaped like `Claude updated 2:14 PM` and `Codex updated 2:14 PM`.
- **D-11:** Show both absolute time and relative age when width allows, for example `Claude updated 2:14 PM (3m ago)`.
- **D-12:** Compact freshness text should keep source identity, for example `Claude 2:14 PM`. Very narrow layouts may use `Cl 2:14`.

### Refresh-Failure Hints
- **D-13:** When a refresh fails but last-known-good rows remain visible, show the current-error hint on the affected source freshness line.
- **D-14:** Use `refresh failed` as the concise wording for current source refresh failures.
- **D-15:** If both stale age and current refresh failure apply, show both when width allows, for example `Claude updated 2:14 PM (2h old, refresh failed)`.
- **D-16:** Keep footer recovery hints only for missing first-run data. Once rows exist, source freshness lines should carry current/stale source status to avoid duplicated warnings.

### The Agent's Discretion
- Compact label shortening and exact width breakpoints are implementation choices, as long as the visible behavior above holds at normal, narrow, and very narrow widths.
- The exact internal field name for the optional Claude Sonnet weekly cache value may be chosen during research/implementation based on observed local Claude `rate_limits` shape, but parser changes must preserve backward compatibility with existing `five_hour` and `seven_day` cache data.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Planning Scope
- `.planning/ROADMAP.md` - Phase 7 goal, success criteria, and planned split between row/layout work and freshness/error hint work.
- `.planning/REQUIREMENTS.md` - Requirements CLD-05, CLD-06, POL-01, POL-02, POL-03, and POL-04.
- `.planning/PROJECT.md` - Core product value, local-only constraints, and v1.1 milestone decisions.
- `.planning/STATE.md` - Current milestone state and known refresh-failure concern.

### Codebase Maps
- `.planning/codebase/CONVENTIONS.md` - Go naming, error handling, rendering, and test conventions.
- `.planning/codebase/STRUCTURE.md` - Where to add TUI, source reader, and test changes.
- `.planning/codebase/STACK.md` - Bubble Tea v2, Bubbles progress, Lip Gloss v2, and Go testing stack.
- `.planning/codebase/CONCERNS.md` - Known width-sensitive rendering fragility and current-error rendering gap.

### Product Design
- `docs/superpowers/specs/2026-05-16-llm-quota-tui-design.md` - Original v1 source and rendering design; use as background, with Phase 7 decisions taking precedence for v1.1 polish.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/sources/window.go`: Shared `sources.Window`, `Product`, `WindowKind`, `CapturedAt`, `Stale`, `StaleAge`, `Metadata`, and typed `SourceError` model. Phase 7 should extend this model conservatively for the Sonnet weekly window.
- `internal/sources/claude.go`: Claude cache reader currently requires `five_hour`, `seven_day`, and `written_at`; Sonnet parsing should be optional and should not invalidate existing rows when absent.
- `internal/install/claude_hook.go`: Claude statusline cache writer currently writes `five_hour`, `seven_day`, and `written_at` from local statusline `rate_limits`; Sonnet support likely starts here if the local payload exposes a model-specific weekly limit.
- `internal/tui/view.go`: Owns row ordering, labels, progress bars, reset text, footer hints, and width-sensitive rendering. This is the main integration point for alignment, placeholder rows, source freshness lines, and current-error display.
- `internal/tui/update.go`: `mergeRefresh` preserves last-known-good rows and records current source errors; freshness/error rendering can consume this existing state.
- `internal/tui/view_test.go` and `internal/tui/update_test.go`: Existing deterministic tests should be extended for alignment, placeholder Sonnet rows, source freshness lines, and current-error visibility.

### Established Patterns
- Source readers return typed errors and do not render, log, or exit.
- TUI rendering handles missing/stale data with readable placeholders and concise hints rather than crashing or blanking the display.
- Width-sensitive rendering is tested with explicit line-width assertions and deterministic clocks.
- The project prefers small package-local helpers over broad abstractions.

### Integration Points
- Add any new Sonnet window kind or metadata in `internal/sources/window.go`, then wire it through `internal/sources/claude.go`, `internal/install/claude_hook.go`, and `internal/tui/view.go`.
- Refactor row rendering in `internal/tui/view.go` around fixed label/bar/percent/reset column budgets while preserving responsive fallbacks.
- Use existing `Model.windows`, `Model.errors`, `CapturedAt`, `Stale`, and `StaleAge` fields for source freshness and current-error text.

</code_context>

<specifics>
## Specific Ideas

- Preferred normal source freshness shape: `Claude updated 2:14 PM (3m ago)` when width allows.
- Preferred compact freshness shape: `Claude 2:14 PM`.
- Preferred very narrow freshness shape: `Cl 2:14`.
- Preferred current-error suffix: `refresh failed`.
- Preferred stale and current-error combination: `2h old, refresh failed`.

</specifics>

<deferred>
## Deferred Ideas

None - discussion stayed within phase scope.

</deferred>

---

*Phase: 7-Row Alignment, Claude Sonnet Limit, and Source Freshness*
*Context gathered: 2026-05-21*
