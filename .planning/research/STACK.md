# Stack Research

**Domain:** Small always-running Go terminal quota/status dashboard
**Researched:** 2026-05-16
**Confidence:** HIGH for Go/Charm stack, MEDIUM for quota file shape durability

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.26.3 | Language, runtime, build toolchain | Current supported Go release as of 2026-05-16; stdlib covers local file IO, JSON/JSONL parsing, time math, tests, and installation without adding runtime weight. |
| Bubble Tea | `charm.land/bubbletea/v2` v2.0.6 | TUI event loop and terminal renderer | The project is exactly Model-Update-View shaped: periodic refresh commands, resize messages, `q`/`ctrl+c` exits, and one full-screen pane. v2 is current and should be used from the start instead of learning the older v1 API. |
| Bubbles progress | `charm.land/bubbles/v2/progress` v2.1.0 | Quota progress bars | Official Bubble Tea component for progress bars; handles width and rendering better than a hand-rolled bar while staying inside the Charm ecosystem. |
| Lip Gloss | `charm.land/lipgloss/v2` v2.0.3 | Styling, colors, and layout measurements | Current Charm styling library; provides cell-aware width measurement, color handling, joins, and simple style values for testable row rendering. |
| Go stdlib `encoding/json`, `os`, `path/filepath`, `time` | Go 1.26.3 | Source parsing and filesystem access | The sources are local JSON/JSONL files. Stdlib is sufficient and keeps v1 small; no database, HTTP client, config framework, or watcher library is needed. |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `golang.org/x/sync/errgroup` | v0.20.0 | Parallel Claude/Codex fetches | Use inside the Bubble Tea refresh command to fetch both sources concurrently and return one `refreshMsg`. |
| `testing` | Go 1.26.3 | Unit and golden tests | Use table-driven tests for source parsers and deterministic render tests with fixed `now`, fixed width, and fixture data. |
| `testing/fstest` | Go 1.26.3 | Optional parser tests over `fs.FS` | Use only if source readers are designed around `fs.FS`; otherwise `t.TempDir()` plus real files is simpler. |
| `regexp` | Go 1.26.3 | Test-only ANSI stripping helper | Use a tiny local helper in render tests if asserting styled output after Lip Gloss/Bubbles emit escape sequences. Do not add a dependency just for stripping ANSI. |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| `go test ./...` | Primary correctness gate | Covers parser fixtures, stale-data behavior, Bubble Tea update transitions, and golden render output. |
| `go test -race ./...` | Concurrency safety gate | Run before release because refresh uses goroutines. The model should only be mutated by `Update`; goroutines return data through messages. |
| `go vet ./...` | Baseline static checks | Cheap stdlib check; enough for v1 without adopting a large lint framework immediately. |
| `gofmt` / `go fmt ./...` | Formatting | Standard Go formatting; do not introduce a formatter wrapper. |

## Installation

```bash
# Module initialization
go mod init github.com/rob/llm-quota

# Core TUI dependencies
go get charm.land/bubbletea/v2@v2.0.6
go get charm.land/bubbles/v2@v2.1.0
go get charm.land/lipgloss/v2@v2.0.3

# Concurrency helper
go get golang.org/x/sync@v0.20.0

# Build/install
go install ./cmd/llm-quota
```

## Prescriptive Implementation Pattern

Use a tiny module with three internal packages/areas:

```text
cmd/llm-quota/main.go
internal/sources/window.go
internal/sources/claude.go
internal/sources/codex.go
internal/tui/model.go
internal/tui/update.go
internal/tui/view.go
internal/tui/colors.go
testdata/
```

### Bubble Tea v2 model

Use Bubble Tea v2 APIs, not v1 examples:

- Import `tea "charm.land/bubbletea/v2"`.
- `View()` returns `tea.View`, not `string`.
- Set `v.AltScreen = true` in `View()` instead of passing `tea.WithAltScreen()`.
- Handle `tea.KeyPressMsg` for `q`, `ctrl+c`, and `r`.
- Handle `tea.WindowSizeMsg` by storing `width` and `height` in the model.
- Use `tea.Tick(30*time.Second, ...)` for periodic refresh.
- Use `tea.Batch(refreshCmd, tickCmd)` from `Init()`.

The model should own last-known-good windows and source errors. Failed refreshes
must update error state but keep previous successful data.

### Source readers

Keep source readers independent of Bubble Tea:

- `ClaudeSource.Fetch(now time.Time) ([]Window, error)` reads one cache JSON file.
- First launch, `llm-quota install`, or `llm-quota install-claude-hook` prompts
  for permission to install the Claude hook/cache writer that produces that file.
- `CodexSource.Fetch(now time.Time) ([]Window, error)` scans local rollout files.
- Constructors accept paths so tests never touch `~/.claude` or `~/.codex`.
- Return source-specific errors; do not log, exit, or render from source code.

For Codex rollout discovery, prefer `filepath.WalkDir` over `filepath.Glob`.
Go's `filepath.Glob` does not provide shell-style recursive `**` semantics, so
`WalkDir` is less surprising for `~/.codex/sessions/YYYY/MM/DD/...`.

For JSONL parsing, read the selected rollout file with `os.ReadFile` and split
lines. The design says files are typically under 5 MB, so whole-file parsing is
simpler and avoids `bufio.Scanner` token-size edge cases.

### Rendering

Use Lip Gloss for labels, footer text, dividers, and width measurement. Use
`progress.Model.ViewAs(usedPct / 100)` for bars. Recompute progress width during
render from the current terminal width rather than storing derived widths in the
model.

Responsive rules for v1:

- At `width >= 50`, render product, window, bar, percentage, and reset text.
- At `30 <= width < 50`, shorten labels and shrink or omit secondary footer text.
- At `width < 30`, drop bars and show only label, percent, and reset/placeholder.

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| Bubble Tea v2 | Bubble Tea v1 (`github.com/charmbracelet/bubbletea`) | Only for maintaining an older app already pinned to v1. New code should use v2 imports and message types. |
| Bubbles `progress` | Hand-rolled Unicode bar | Use hand-rolled only if `progress` proves hard to snapshot-test or style exactly. Start with Bubbles because it is official and width-aware. |
| Lip Gloss layout | Manual ANSI strings and padding | Use manual strings only for unstyled fallback rows. Lip Gloss avoids cell-width bugs and keeps colors centralized. |
| `tea.Tick` polling | `fsnotify` file watching | Use `fsnotify` only if latency becomes important. A 30-second quota dashboard does not need filesystem event complexity. |
| Stdlib JSON parsing | `gjson`, `jsoniter`, or generated parsers | Use alternatives only for very large or unstable schemas. These files are small and the shapes are known. |
| `go test` golden files | Snapshot-testing framework | Use a framework only if golden update workflow becomes painful. Stdlib tests are enough for v1. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| Network or OAuth fallback for Claude/Codex | Explicitly out of scope for v1; adds credentials, prompts, platform-specific behavior, and failure modes that conflict with a tiny local dashboard. | Render local-data placeholders and footer hints. |
| macOS Keychain reads from the Go TUI | Can prompt, is platform-specific, and couples this repo to Claude credential storage. | Read the hook-written Claude cache file only. |
| Broad CLI/config framework | v1 only needs the TUI plus a tiny setup command for Claude hook installation. | Hand-roll minimal argument handling unless setup grows beyond install/help. |
| `fsnotify` | File watching is unnecessary for 30-second refresh and can be noisy across editor/atomic writes. | `tea.Tick` plus `r` refresh. |
| Database or embedded store | There is no history or persistence requirement. | Last-known-good state in memory. |
| `bubble-table`, `viewport`, list components | The UI is four fixed rows, not an interactive table or scrollable view. | Plain Lip Gloss rows plus Bubbles progress. |
| `bufio.Scanner` with defaults for JSONL | Default token limit can surprise if a rollout event grows. | `os.ReadFile` plus line splitting for small files. |
| Old Charm import paths | v2 moved to `charm.land/...`; old paths teach the wrong API and `View()` shape. | `charm.land/bubbletea/v2`, `charm.land/bubbles/v2`, `charm.land/lipgloss/v2`. |

## Stack Patterns by Variant

**For v1 as specified:**

- Use Bubble Tea v2 + Bubbles progress + Lip Gloss + stdlib readers.
- Include a first-launch/setup prompt that asks before installing the Claude
  hook/cache writer.
- Use local files only and no config surface beyond source path constructors.
- Because the product is a dedicated tmux-pane monitor with four fixed rows.

**If progress bars become hard to test or style:**

- Replace only `internal/tui/bar.go` with a hand-rolled bar.
- Keep Bubble Tea and Lip Gloss; do not change architecture.
- Because the progress component is isolated behind row rendering.

**If source file shapes drift:**

- Add parser fixtures and tolerant optional fields.
- Keep the shared `Window` model stable.
- Because the TUI should not know source-specific JSON details.

## Version Compatibility

| Package A | Compatible With | Notes |
|-----------|-----------------|-------|
| `charm.land/bubbletea/v2` v2.0.6 | `charm.land/bubbles/v2` v2.1.0 | Both use v2 Bubble Tea message types such as `tea.KeyPressMsg`. |
| `charm.land/bubbletea/v2` v2.0.6 | `charm.land/lipgloss/v2` v2.0.3 | Bubble Tea v2 expects `View() tea.View`; terminal feature flags such as alt-screen are view fields. |
| `charm.land/bubbles/v2/progress` v2.1.0 | `charm.land/lipgloss/v2` v2.0.3 | Progress colors are `color.Color` values; use `lipgloss.Color("#hex")`, not v1 string color fields. |
| Go 1.26.3 | `golang.org/x/sync` v0.20.0 | `errgroup.Group` is stable for the two-source refresh command. |

## Confidence Assessment

| Recommendation | Confidence | Reason |
|----------------|------------|--------|
| Use Go 1.26.3 | HIGH | Verified against official Go release history on 2026-05-16. |
| Use Bubble Tea v2 import path and APIs | HIGH | Verified through Context7 and official GitHub release docs; v2.0.6 is latest from module version listing. |
| Use Bubbles progress | HIGH | Verified through Context7; current v2 package provides progress component with width setters and `ViewAs`. |
| Use Lip Gloss v2 | HIGH | Verified through Context7; v2 import path and pure style/value behavior fit testable rendering. |
| Use stdlib source parsing | HIGH | Data is local JSON/JSONL and small; no requirement needs third-party parsing. |
| Codex/Claude local file shapes | MEDIUM | Shapes are confirmed by local design spec, but upstream tools may change their private file/cache formats. Keep parser tests and placeholder rendering. |

## Sources

- Context7 `/charmbracelet/bubbletea` — v2 Model-View-Update, `tea.Tick`, `tea.WindowSizeMsg`, `tea.KeyPressMsg`, declarative `tea.View`, alt-screen field.
- Context7 `/charmbracelet/bubbles` — v2 progress component, import path, width setters, color API, v2 compatibility notes.
- Context7 `/charmbracelet/lipgloss` — v2 styling, width/height measurement, joins, color handling, pure `Style` values.
- Context7 `/websites/pkg_go_dev_golang_org_x_sync` — `errgroup.Group`, `Go`, `Wait`, `WithContext` behavior.
- Official Go release history — Go 1.26.3 released 2026-05-07 and supported as current stable release.
- GitHub releases for `charmbracelet/bubbletea` — v2.0.6 latest, released 2026-04-16; v2 introduced `charm.land` import paths and declarative views.
- `go list -m -versions` — verified current module versions: Bubble Tea v2.0.6, Bubbles v2.1.0, Lip Gloss v2.0.3, x/sync v0.20.0.
- Project context `.planning/PROJECT.md` and design spec `docs/superpowers/specs/2026-05-16-llm-quota-tui-design.md` — v1 scope, local-only data sources, no network/OAuth fallback.

---

*Stack research for: llm-quota*
*Researched: 2026-05-16*
