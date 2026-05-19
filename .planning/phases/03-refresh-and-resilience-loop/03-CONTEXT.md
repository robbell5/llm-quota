# Phase 3: Refresh and Resilience Loop - Context

**Gathered:** 2026-05-19
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 3 delivers the running refresh and resilience behavior for `llm-quota`: the
TUI should refresh local Claude and Codex source data immediately on startup,
automatically every 30 seconds, and manually when the user presses `r`, while
preserving useful last-known-good rows when later source reads fail. This phase
also locks stale-data policy and model/update tests for refresh merge behavior.
It does not implement final quota bars, threshold colors, polished stale/error
copy, or narrow-pane row layouts; those remain Phase 4 responsibilities.

</domain>

<decisions>
## Implementation Decisions

### Stale Warning Policy

- **D-01:** Treat displayed quota data as stale when its captured time is more
  than one hour old.
- **D-02:** Use the same one-hour stale threshold for Claude and Codex in the
  TUI model so freshness has one user-facing meaning.
- **D-03:** Valid stale data should still be shown with warning state. Do not
  replace stale-but-valid quota values with placeholders.
- **D-04:** In Phase 3, stale handling should be represented in model state and
  tests. Phase 4 owns the final visible row/footer copy and styling for stale
  warnings.

### Last-Known-Good Failure State

- **D-05:** Preserve last-known-good data independently per source. A failed
  Claude refresh must not clear old Claude rows, and it must not prevent a
  successful Codex refresh from updating Codex rows, or vice versa.
- **D-06:** If the initial refresh fails before any source has last-known-good
  data, keep the existing placeholder rows and hints rather than exiting or
  introducing a full-screen error state.
- **D-07:** Store typed source errors in the TUI state so later rendering can map
  missing, malformed, no-usable-event, and read errors to concise hints.
- **D-08:** In Phase 3, temporary source failure warnings should be proven in
  state and tests. Phase 4 owns how prominent those warnings appear visually.

### Resize Refresh Semantics

- **D-09:** A terminal resize should rerender the current model only. It should
  not trigger new source file reads.
- **D-10:** Interpret the resize-related refresh requirement as layout refresh:
  resize immediately recalculates terminal dimensions and view output, while data
  refresh remains owned by startup, timer ticks, and manual `r`.
- **D-11:** Resize should not reset, delay, or otherwise affect the 30-second
  refresh timer schedule.
- **D-12:** Phase 3 tests should assert that `tea.WindowSizeMsg` stores width and
  height and returns no source refresh command.

### Refresh Interaction Feel

- **D-13:** Run an initial source refresh immediately from Bubble Tea `Init`, then
  continue with the 30-second periodic refresh cadence.
- **D-14:** Pressing `r` should trigger an immediate refresh without resetting or
  disrupting the next scheduled refresh tick.
- **D-15:** If `r` is pressed while a refresh is already running, coalesce or
  ignore the duplicate refresh rather than starting concurrent duplicate source
  reads.
- **D-16:** Do not add visible `refreshing...` status or last-updated copy in
  Phase 3. Keep Phase 3 focused on loop/state behavior; Phase 4 owns final
  display details.

### Agent Discretion

No decisions were explicitly delegated with "you decide." Downstream agents may
choose exact type names, message names, fake reader shapes, command helper names,
and test fixture names as long as the locked decisions above are respected.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Scope And Requirements

- `.planning/ROADMAP.md` -- Defines Phase 3 goal, success criteria, dependency
  status, and phase boundary.
- `.planning/REQUIREMENTS.md` -- Maps Phase 3 to `SRC-04`, `SRC-05`, `TUI-02`,
  `TUI-03`, and `TEST-03`.
- `.planning/PROJECT.md` -- Defines core value, local-only constraints,
  30-second refresh expectation, failure tolerance, and out-of-scope network,
  daemon, statusline, alerting, and history behavior.
- `.planning/phases/02-standalone-local-data-sources/02-CONTEXT.md` -- Carries
  forward source-reader contracts, typed source errors, stale Claude metadata,
  Codex rollout fallback behavior, and setup/source boundaries.
- `.planning/phases/01-foreground-tui-foundation/01-CONTEXT.md` -- Carries
  forward startup screen, placeholder rows, clean quit behavior, compact footer,
  alt-screen direction, and minimal command handling.

### Stack And Architecture

- `.planning/research/STACK.md` -- Locks Go, Bubble Tea v2, Lip Gloss v2, Bubbles
  progress, `tea.Tick`, `tea.Batch`, `tea.KeyPressMsg`, and `errgroup` guidance.
- `.planning/research/ARCHITECTURE.md` -- Defines the refresh flow, last-known-good
  merge policy, TUI/source boundary, pure rendering seam, and resize flow notes.

### Product Design

- `docs/superpowers/specs/2026-05-16-llm-quota-tui-design.md` -- Original design
  spec covering local-only runtime, four quota windows, refresh cadence, manual
  refresh, source placeholders, and fixture expectations.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- `internal/sources/window.go` defines normalized `Window`, `Product`,
  `WindowKind`, `Metadata`, typed `SourceError`, and source error categories.
  Phase 3 should use this source-independent contract in TUI refresh state.
- `internal/sources/claude.go` and `internal/sources/codex.go` already expose
  `Fetch(now time.Time) ([]Window, error)` readers suitable for injection into
  Bubble Tea refresh commands.
- `internal/tui/model.go` is intentionally small and currently stores only
  terminal size. Phase 3 should extend it with source readers, last-known-good
  windows, source errors, refresh timing, and deterministic clock/test seams.
- `internal/tui/update.go` already handles `q`, `ctrl+c`, `WindowSizeMsg`,
  `Init`, and `View`. Phase 3 should add `r`, initial refresh command, tick
  scheduling, refresh messages, and per-source merge logic here.
- `internal/tui/view.go` already renders placeholder rows and footer hints. Phase
  3 can keep visible rendering minimal while preparing model state for Phase 4.
- `cmd/llm-quota/main.go` owns real defaults and starts `tea.NewProgram` with
  `tui.NewModel()`. Phase 3 should keep real path/default wiring at this edge and
  keep tests path-injected.

### Established Patterns

- Use Bubble Tea v2 APIs consistently: `charm.land/bubbletea/v2`,
  `tea.KeyPressMsg`, `tea.WindowSizeMsg`, `tea.Cmd`, and `View() tea.View`.
- Keep filesystem reads out of rendering and out of direct model mutation. Source
  reads should happen inside commands that return typed messages to `Update`.
- Preserve source parsing and TUI rendering boundaries. The TUI should not parse
  Claude cache JSON or Codex rollout JSONL internals.
- Tests should use synthetic fixtures, fake readers, fixed times, and direct
  update/render helpers rather than real home-directory data or sleeping timers.

### Integration Points

- `tui.NewModel` likely needs injected readers and clock/timing dependencies for
  tests, while `main.go` supplies real Claude/Codex readers and default paths.
- `Model.Init` should batch immediate refresh and timer setup.
- `Update` should merge refresh results per source: success replaces windows and
  clears that source error; failure stores the error and keeps old windows.
- `tea.WindowSizeMsg` should continue to store width and height only.

</code_context>

<specifics>
## Specific Ideas

- A one-hour stale threshold is intentionally shared across sources for a simple
  user mental model.
- Phase 3 should add `r` behavior but should not show the `r refresh` hint unless
  current rendering scope can do so without pulling in Phase 4 display polish.
- Coalescing duplicate manual refreshes is preferred over concurrent duplicate
  reads, but exact implementation details are left to planning.

</specifics>

<deferred>
## Deferred Ideas

None -- discussion stayed within phase scope.

</deferred>

---

*Phase: 3-Refresh and Resilience Loop*
*Context gathered: 2026-05-19*
