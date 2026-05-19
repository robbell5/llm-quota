# Phase 3 Pattern Map

## Target Files And Closest Analogs

| Target | Role | Closest Existing Pattern |
|--------|------|--------------------------|
| `internal/tui/model.go` | Bubble Tea state and injected dependencies | Current minimal `Model` with width/height state |
| `internal/tui/update.go` | MVU update loop, commands, ticks, refresh merge | Current key/resize handling plus Bubble Tea v2 stack guidance |
| `internal/tui/update_test.go` | Deterministic update/command tests | Existing table-driven quit and resize tests |
| `internal/tui/view.go` | Minimal source-backed row rendering | Existing placeholder rows and footer width behavior |
| `internal/tui/view_test.go` | ANSI-stripped render assertions | Existing startup screen and width guard tests |
| `cmd/llm-quota/main.go` | Real dependency wiring at command edge | Existing dependency-injected CLI setup and `startTUI` |
| `cmd/llm-quota/main_test.go` | CLI and TUI startup wiring tests | Existing injected `StartTUI` command-edge tests |

## Key Existing Code Excerpts

### Bubble Tea Update Shape

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyPressMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        }
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
    }

    return m, nil
}
```

### Source Reader Contracts

```go
func (r ClaudeReader) Fetch(now time.Time) ([]Window, error)
func (r CodexReader) Fetch(now time.Time) ([]Window, error)
```

### Normalized Window State

```go
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

### Real Program Startup Edge

```go
func startTUI() error {
    program := tea.NewProgram(tui.NewModel())
    _, err := program.Run()
    return err
}
```

## Implementation Notes For Executors

- Preserve `View() tea.View` and `v.AltScreen = true`; do not switch to Bubble Tea
  v1 examples.
- Keep tests synthetic and injected; do not read real user home directories.
- Keep `cmd/llm-quota/main.go` thin: compute default paths, construct readers,
  and call `tui.NewModel(...)`.
- Keep refresh behavior in `internal/tui/update.go`; keep source parsing in
  `internal/sources`.
