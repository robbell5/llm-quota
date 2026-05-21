# Phase 7 Pattern Map

**Phase:** 07 - Row Alignment, Claude Sonnet Limit, and Source Freshness
**Created:** 2026-05-21
**Status:** Complete

## Files To Modify

| File | Role | Closest Existing Pattern | Notes |
|------|------|--------------------------|-------|
| `internal/sources/window.go` | Shared quota window model | Existing `WindowKind` constants | Add one optional Claude-only kind without changing the `Window` struct. |
| `internal/sources/claude.go` | Claude cache reader | Required `five_hour` / `seven_day` validation, `claudeCacheWindow.window` | Keep required windows strict; parse optional Sonnet weekly data only when valid. |
| `internal/sources/claude_test.go` | Claude reader tests | Table-driven `TestClaudeFetch` with `assertWindows` | Add cases for present, absent, and malformed optional Sonnet cache fields. |
| `internal/install/claude_hook.go` | Claude cache writer | `writeClaudeCache`, `claudeHookRateLimits.validate` | Preserve required rate limit validation and write canonical optional `sonnet_seven_day` when recognized optional data is present. |
| `internal/install/claude_hook_test.go` | Cache writer tests | Existing cache writer tests and JSON map assertions | Add statusline input cases that prove optional Sonnet writing and backwards-compatible absence. |
| `internal/tui/view.go` | TUI row/freshness renderer | `renderRows`, `renderDataRow`, `footerRecoveryHints`, `appendHintWithinWidth` | Refactor fixed row specs and layout budgets before adding source freshness lines. |
| `internal/tui/view_test.go` | Render tests | ANSI stripping, line-width assertions at fixed widths | Add deterministic rows for Sonnet placeholder/real data, alignment, freshness, and errors. |
| `internal/tui/update_test.go` | Refresh state tests | Last-known-good preserved per product | Extend coverage for visible current-error state on freshness lines. |

## Existing Patterns

### Shared Window Kind Pattern

`internal/sources/window.go` already centralizes product and window identifiers:

```go
type WindowKind string

const (
	WindowFiveHour WindowKind = "five_hour"
	WindowSevenDay WindowKind = "seven_day"
)
```

Use the same pattern for `WindowSonnetSevenDay = "sonnet_seven_day"`. Do not encode Sonnet as metadata or a raw string in the TUI.

### Claude Reader Pattern

`internal/sources/claude.go` currently validates the two required windows before constructing `Window` values:

```go
return []Window{
	cache.FiveHour.window(WindowFiveHour, "Claude 5h", writtenAt, stale, staleAge),
	cache.SevenDay.window(WindowSevenDay, "Claude 7d", writtenAt, stale, staleAge),
}, nil
```

Keep `five_hour`, `seven_day`, and `written_at` required. Add optional parsing after required validation so malformed optional Sonnet data cannot reject the existing required windows.

### Claude Cache Writer Pattern

`internal/install/claude_hook.go` writes a typed cache shape:

```go
cache := claudeHookCache{
	FiveHour:  rateLimits.FiveHour,
	SevenDay:  rateLimits.SevenDay,
	WrittenAt: &writtenAt,
}
```

Extend this with a canonical optional `SonnetSevenDay` field. Accept known incoming optional field names behind one helper and write only the canonical cache key.

### Responsive Row Rendering Pattern

`internal/tui/view.go` currently owns all row rendering and already has constants for label widths and minimum bar width:

```go
const (
	fullRowLabelWidth  = 9
	shortRowLabelWidth = 5
	minProgressWidth   = 6
)
```

Keep rendering in `view.go`. Refactor around a row spec list and computed fixed columns instead of computing bar width from each row's rendered percent/reset text.

### Last-Known-Good Error Pattern

`internal/tui/update.go` preserves existing rows when one source fails:

```go
if result.err.Category != "" {
	m.errors[result.product] = result.err
	continue
}
```

Freshness/error rendering should consume `m.errors[product]` and `m.windows[product]`; update logic does not need a new state field.

## Data Flow

1. Claude statusline input is decoded by `writeClaudeCache`.
2. `writeClaudeCache` writes required `five_hour`, required `seven_day`, optional canonical `sonnet_seven_day`, and `written_at`.
3. `ClaudeReader.Fetch` reads required windows and appends `WindowSonnetSevenDay` only if the optional cache field is present and valid.
4. `renderRows` declares Claude rows as `Claude 5h`, `Claude 7d`, `Sonnet 7d`, then Claude freshness; Codex rows as `Codex 5h`, `Codex 7d`, then Codex freshness.
5. Source freshness lines derive time/status from each product's windows and current source error.

## Planning Constraints

- Missing Sonnet data must show a row-level placeholder, not remove the row.
- Row percent and reset columns must be fixed at normal width.
- Compact layouts may abbreviate labels and reset text, but percent remains visible.
- Footer recovery hints stay for first-run missing data only; once rows exist, stale/current refresh status belongs on source freshness lines.
- Do not expose raw source error categories in rendered output.
