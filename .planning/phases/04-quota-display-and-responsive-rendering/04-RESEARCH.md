<!-- markdownlint-disable MD013 -->

# Phase 4 Research: Quota Display and Responsive Rendering

**Phase:** 4 — Quota Display and Responsive Rendering
**Date:** 2026-05-19
**Status:** Complete

## Research Question

What does the executor need to know to implement the final v1 quota dashboard rendering without changing the already-validated local source and refresh architecture?

## Inputs Reviewed

- `.planning/PROJECT.md`
- `.planning/ROADMAP.md`
- `.planning/REQUIREMENTS.md`
- `.planning/phases/04-quota-display-and-responsive-rendering/04-CONTEXT.md`
- `.planning/phases/04-quota-display-and-responsive-rendering/04-UI-SPEC.md`
- `.planning/phases/03-refresh-and-resilience-loop/03-01-SUMMARY.md`
- `.planning/phases/03-refresh-and-resilience-loop/03-02-SUMMARY.md`
- `.planning/phases/02-standalone-local-data-sources/02-05-SUMMARY.md`
- `docs/superpowers/specs/2026-05-16-llm-quota-tui-design.md`
- `internal/tui/view.go`
- `internal/tui/view_test.go`
- `internal/tui/colors.go`
- `internal/tui/model.go`
- `internal/tui/update.go`
- `internal/sources/window.go`

## Standard Stack

- Keep Go 1.26.3 and Charm v2 imports already pinned in `go.mod`.
- Use `charm.land/bubbles/v2/progress` for static progress bars per D-13.
- Use `charm.land/lipgloss/v2` for colors, width measurement, padding, and ANSI-aware layout.
- Keep source readers local-only and untouched; Phase 4 renders existing `sources.Window` and `sources.SourceError` state.

## Implementation Findings

### Existing architecture to preserve

- `internal/tui/update.go` owns refresh cadence, manual `r`, resize storage, per-source last-known-good merge, and stale marking. Phase 4 must not introduce refresh-on-resize or new keybindings.
- `internal/tui/model.go` already exposes everything rendering needs inside the package: `windows`, `errors`, injected clock, and stale metadata.
- `internal/sources/window.go` already normalizes all four rows with `Product`, `WindowKind`, `UsedPercent`, `ResetsAt`, `Stale`, and `StaleAge`.
- `internal/tui/view.go` already renders the correct row order and has the width shell constants that all line-width tests should continue using.

### Bubbles progress v2 notes

- Import path is `charm.land/bubbles/v2/progress`.
- Static rendering should use `progress.Model.ViewAs(fraction)`; do not add animation commands or handle `progress.FrameMsg` in `Update`.
- Width is set with options or setters (`progress.WithWidth(width)` / `SetWidth(width)`), not by assigning a public `Width` field.
- Colors are `image/color.Color` values; `lipgloss.Color("#hex")` is valid for v2. Use threshold color for fill and `#313244` for the track.

### Rendering strategy

- Keep rendering concentrated in `internal/tui/view.go` and style constants in `internal/tui/colors.go`.
- Add small deterministic helpers: threshold color selection, clamped progress fraction, two-part reset text, age text, missing-source footer hint selection, and ANSI-width-safe row assembly.
- Use the UI-SPEC row breakpoints exactly:
  - `>= 50` columns: full labels, progress bar, percent, reset token.
  - `30-49` columns: abbreviated labels first, preserve bar while it remains useful.
  - `< 30` columns: no bars, keep short label, percent, and reset token when it fits.
- At width 50, remember `innerWidth` is `46` because shell horizontal padding is 4. Tests should assert final rendered lines, not just row helper output.

### Copy and hint strategy

- Do not expose raw `SourceError.Category` values such as `malformed`, `read_error`, or `no_usable_event`.
- Missing Claude rows should drive `Claude: run install-claude-hook` when Claude has no usable rows; otherwise prefer `Claude: open Claude` for stale/recovery wording.
- Missing Codex rows should drive `Codex: open Codex`.
- Stale-but-valid rows render normally with threshold color; staleness appears only in bounded footer copy such as `Claude data 2h old; open Claude`.
- If multiple hints apply, order by UI-SPEC priority and include only what fits within the current inner width.

## Common Pitfalls

- Do not regress Phase 3 source isolation: a Claude error must not blank Codex rows, and a Codex error must not blank Claude rows.
- Do not assert exact ANSI escape sequences. Existing tests strip ANSI and assert content/width; keep that pattern.
- Do not add alert words, badges, blinking, or row backgrounds for red usage.
- Do not invent zero-percent bars for missing data. Missing rows use missing-data copy and no quota bar.
- Do not round reset times to a single coarse unit in the normal target layout; use `2h 14m` and `4d 06h` style output.

## Recommended Plan Shape

1. Implement tested data-row rendering: threshold colors, Bubbles progress bars, four quota rows, and two-part reset tokens.
2. Implement tested resilience and responsive rendering: missing-source rows, stale footer hints, footer priority/fit, and width breakpoints.

## Validation Architecture

- `go test ./internal/tui -run TestRender` should cover normal, mixed-threshold, missing-source, stale-source, and narrow-width states.
- `go test ./internal/tui` should prove update-model seams still compile with render changes.
- `go test ./...` should remain the full phase gate.

## Research Complete

Phase 4 can proceed with the approved UI-SPEC and two sequential rendering plans. No external service setup is required.
