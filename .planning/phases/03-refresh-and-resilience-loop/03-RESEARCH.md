# Phase 3 Research: Refresh and Resilience Loop

**Phase:** 3 — Refresh and Resilience Loop  
**Researched:** 2026-05-19  
**Confidence:** HIGH for Bubble Tea refresh architecture; HIGH for local-source
merge policy; MEDIUM for exact visible row polish because final rendering remains
Phase 4 scope.

## Research Question

What does the planner need to know to make the running TUI refresh local Claude
and Codex data on startup, every 30 seconds, and on `r`, while preserving
last-known-good data through temporary source failures?

## Applicable Stack And Existing Contracts

- Keep Bubble Tea v2 APIs: `charm.land/bubbletea/v2`, `tea.Cmd`,
  `tea.Batch`, `tea.Tick`, `tea.KeyPressMsg`, and `tea.WindowSizeMsg`.
- `golang.org/x/sync/errgroup` is already pinned and should be used inside the
  refresh command to fetch Claude and Codex concurrently.
- Source readers already expose the shape needed by the TUI:
  `Fetch(now time.Time) ([]sources.Window, error)`.
- `sources.Window` already carries `Product`, `Kind`, `Label`, `UsedPercent`,
  `ResetsAt`, `CapturedAt`, `Stale`, `StaleAge`, and optional metadata.
- `sources.SourceError` already carries typed categories for missing,
  malformed, no-usable-event, and read errors.

## Recommended Architecture

### Model State

Extend `internal/tui.Model` with source-independent state:

- injected Claude and Codex readers;
- injected `now func() time.Time` test seam;
- `refreshEvery` defaulting to `30 * time.Second`;
- `staleAfter` defaulting to `time.Hour`;
- per-source last-known-good windows;
- per-source typed errors;
- `refreshing bool` to coalesce duplicate manual refreshes.

Use options such as `WithReaders`, `WithClock`, and `WithRefreshEvery` so tests
can inject fake readers and zero-duration/manual intervals without touching real
home-directory data. `cmd/llm-quota/main.go` should be the only place that wires
default Claude cache and Codex sessions paths into the TUI.

### Refresh Messages

Use typed messages instead of mutating model state from goroutines:

- `refreshRequestedMsg` starts a refresh if one is not already running.
- `refreshMsg` returns one result per source plus the fetch timestamp.
- `tickMsg` triggers a scheduled refresh and always schedules the next tick.

`Init` should batch an immediate refresh request and the first `tea.Tick(30s,
...)`. A manual `r` key should request refresh without touching the scheduled
tick. `tea.WindowSizeMsg` should store width and height only.

### Merge Policy

Each refresh result is merged independently per source:

```text
if source result succeeded:
    replace only that source's windows
    clear only that source's error
else:
    store only that source's error
    keep that source's existing windows unchanged
```

If the initial refresh fails before any windows exist, the model keeps empty
windows and the existing placeholder renderer continues to show source rows.

### Stale Policy

The source readers already mark stale Claude windows. Phase 3 should make the TUI
model enforce the shared one-hour stale threshold for both products so Claude and
Codex freshness have one meaning. The model should mark any accepted window stale
when `now.Sub(window.CapturedAt) > time.Hour`, clamp negative ages to zero, and
keep stale-but-valid windows in last-known-good state.

Final visible stale copy, colors, and footer warning priority remain Phase 4
scope; Phase 3 should prove stale state through model tests.

## Testing Strategy

Use table-driven tests under `internal/tui/update_test.go` with fake readers and
fixed times. Do not sleep. Tests should call commands directly when needed and
assert returned message types and merged model state.

Critical cases:

- `Init` returns a command that requests immediate refresh and schedules a tick.
- `tickMsg` returns both a refresh request and the next tick command.
- `r` requests refresh when idle.
- `r` returns no refresh command while `refreshing` is true.
- Claude failure after Claude success preserves Claude windows.
- Codex failure does not block successful Claude update, and vice versa.
- initial failures preserve empty windows and store typed errors.
- stale Codex windows older than one hour are marked stale and preserved.
- `tea.WindowSizeMsg` stores dimensions and returns no source refresh command.

Render tests in Phase 3 should stay minimal: existing placeholder layout must not
gain `refreshing...`, `last updated`, or visible refresh-hint copy. Basic data row
rendering may show percent/reset values when windows exist, but bars, threshold
colors, polished stale/error copy, and narrow-pane polish remain Phase 4.

## Common Pitfalls

- Do not start file reads from `View` or render helpers.
- Do not let one source error clear another source's successful data.
- Do not reset, delay, or recreate the scheduled tick cadence on manual `r`.
- Do not trigger source reads from `tea.WindowSizeMsg`.
- Do not read `~/.claude`, `~/.codex`, or real cache data in tests.
- Do not add network, OAuth, Keychain, daemon, fsnotify, alerts, or history.
- Do not introduce final Phase 4 visual copy such as `refreshing...`,
  `last updated`, stale age text, or a visible `r refresh` footer hint.

## Validation Architecture

Phase 3 verification should rely on `go test ./internal/tui ./cmd/llm-quota`,
then `go test ./...`. The important invariant is not elapsed wall-clock time; it
is command/message behavior and deterministic merge state. Tests should prove the
30-second interval by asserting configured defaults or tick command construction,
not by sleeping.

## Source Coverage Notes

- `SRC-04`: covered by independent last-known-good merge tests.
- `SRC-05`: covered by one-hour stale-state model tests for both source products.
- `TUI-02`: covered by `Init` and tick scheduling tests.
- `TUI-03`: covered by manual `r` tests and duplicate-refresh coalescing tests.
- `TEST-03`: covered by refresh merge behavior regression tests.

---

*Research complete for Phase 3 planning.*
