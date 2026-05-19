<!-- markdownlint-disable MD013 MD060 -->

# Phase 4 Pattern Map

## Existing Analog Files

| File | Role | Reuse Pattern |
|------|------|---------------|
| `internal/tui/view.go` | Renderer | Keep row order, shell padding constants, ANSI-width-safe layout helpers, and renderer-only source mapping. |
| `internal/tui/view_test.go` | Render tests | Strip ANSI with `ansiEscapeRE`, use fixed clocks and synthetic `sources.Window` values, and assert line widths. |
| `internal/tui/colors.go` | Palette | Add Catppuccin semantic colors here; do not inline hex values throughout render code. |
| `internal/tui/model.go` | Render state | Read package-local `windows`, `errors`, `now`, and stale fields; do not move source parsing into rendering. |
| `internal/sources/window.go` | Data contract | Match rows by `Product` and `WindowKind`; use `UsedPercent`, `ResetsAt`, `Stale`, and `StaleAge` directly. |

## Key Contracts

From `internal/sources/window.go`:

```go
type Product string

const (
    ProductClaude Product = "claude"
    ProductCodex  Product = "codex"
)

type WindowKind string

const (
    WindowFiveHour WindowKind = "five_hour"
    WindowSevenDay WindowKind = "seven_day"
)

type Window struct {
    Product     Product
    Kind        WindowKind
    Label       string
    UsedPercent float64
    ResetsAt    time.Time
    CapturedAt  time.Time
    Stale       bool
    StaleAge    time.Duration
    Metadata    Metadata
}
```

From `internal/tui/view_test.go`:

```go
var ansiEscapeRE = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

func assertRenderedLineWidths(t *testing.T, rendered string, maxWidth int) {
    for lineNumber, line := range strings.Split(strings.TrimSuffix(rendered, "\n"), "\n") {
        plain := ansiEscapeRE.ReplaceAllString(line, "")
        if width := lipgloss.Width(plain); width > maxWidth {
            t.Fatalf("line %d width = %d, want <= %d: %q", lineNumber+1, width, maxWidth, plain)
        }
    }
}
```

## Data Flow

1. `cmd/llm-quota/main.go` injects Claude and Codex readers.
2. `internal/tui/update.go` fetches both readers and stores normalized windows/errors in `Model`.
3. `internal/tui/view.go` selects rows by `Product` + `WindowKind`, renders data rows or missing-data rows, and derives footer hints from model state.

## Constraints For Executors

- Same-wave file conflicts are not possible in Phase 4 because all rendering work touches `internal/tui/view.go` and `internal/tui/view_test.go`; use sequential plans.
- Keep tests synthetic and package-local. Do not read real `~/.claude`, `~/.codex`, or environment files.
- Preserve Phase 3 refresh and source semantics. Rendering work must not change `Update` behavior unless a compile fix is genuinely required.
