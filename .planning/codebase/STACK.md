# Technology Stack

**Analysis Date:** 2026-05-21

## Languages

**Primary:**
- Go 1.26.3 - Main application, TUI, local source readers, Claude setup installer, and tests in `cmd/llm-quota/main.go`, `internal/tui/update.go`, `internal/tui/view.go`, `internal/sources/claude.go`, `internal/sources/codex.go`, and `internal/install/claude_hook.go`. The module declares `go 1.26.3` in `go.mod`, and the local toolchain reports `go version go1.26.3 darwin/arm64`.

**Secondary:**
- Markdown - User and planning documentation in `README.md`, `AGENTS.md`, `.planning/PROJECT.md`, `.planning/STATE.md`, and `docs/superpowers/specs/2026-05-16-llm-quota-tui-design.md`.
- Shell command strings - Claude statusline passthrough execution uses `sh -c` through `os/exec` in `internal/install/claude_hook.go`; no standalone shell scripts are committed.

## Runtime

**Environment:**
- Go CLI binary - The runtime entry point is `cmd/llm-quota/main.go`, which starts a foreground Bubble Tea program with `tea.NewProgram(model).Run()`.
- Terminal/tmux pane - The UI is designed as an always-running foreground terminal view, documented in `README.md` and implemented by `internal/tui/model.go`, `internal/tui/update.go`, and `internal/tui/view.go`.
- Local filesystem - Runtime data comes from `~/.cache/llm-quota/claude.json`, `~/.cache/llm-quota/state.json`, `~/.claude/settings.json`, and `~/.codex/sessions`, with default paths constructed in `cmd/llm-quota/main.go`.

**Package Manager:**
- Go modules - `go.mod` defines module `github.com/robbell5/llm-quota`.
- Lockfile: present - `go.sum` pins transitive module checksums.

## Frameworks

**Core:**
- Bubble Tea v2 (`charm.land/bubbletea/v2` v2.0.6) - TUI event loop, messages, commands, and program runner in `cmd/llm-quota/main.go` and `internal/tui/update.go`.
- Bubbles progress (`charm.land/bubbles/v2` v2.1.0, package `progress`) - Responsive quota progress bars in `internal/tui/view.go`.
- Lip Gloss v2 (`charm.land/lipgloss/v2` v2.0.3) - Styling, terminal width measurement, colors, and layout in `internal/tui/view.go` and `internal/tui/colors.go`.

**Testing:**
- Go `testing` package - Unit and render tests in `cmd/llm-quota/main_test.go`, `internal/install/claude_hook_test.go`, `internal/sources/claude_test.go`, `internal/sources/codex_test.go`, `internal/tui/update_test.go`, and `internal/tui/view_test.go`.
- Go stdlib test helpers - Tests use `t.TempDir`, `os.WriteFile`, `filepath`, `regexp`, and deterministic clocks in `cmd/llm-quota/main_test.go`, `internal/install/claude_hook_test.go`, `internal/sources/claude_test.go`, `internal/sources/codex_test.go`, and `internal/tui/view_test.go`.

**Build/Dev:**
- Go toolchain - Build, format, vet, test, and install use native Go commands documented in `README.md` and `.planning/research/STACK.md`.
- `golang.org/x/sync/errgroup` v0.20.0 - Concurrent Claude and Codex refreshes inside the Bubble Tea refresh command in `internal/tui/update.go`.
- `tools/tools.go` - Build-tagged tool import anchors for `charm.land/bubbles/v2/progress`, `charm.land/lipgloss/v2`, and `golang.org/x/sync/errgroup`.

## Key Dependencies

**Critical:**
- `charm.land/bubbletea/v2` v2.0.6 - Required by `cmd/llm-quota/main.go` and `internal/tui/update.go` for `tea.Model`, `tea.Cmd`, `tea.Tick`, `tea.Batch`, `tea.KeyPressMsg`, `tea.WindowSizeMsg`, `tea.View`, and `tea.NewProgram`.
- `charm.land/bubbles/v2` v2.1.0 - Required by `internal/tui/view.go` for progress bar rendering with `progress.New`, width configuration, color configuration, and `ViewAs`.
- `charm.land/lipgloss/v2` v2.0.3 - Required by `internal/tui/view.go` and `internal/tui/colors.go` for styles, colors, and display-width measurement.
- `golang.org/x/sync` v0.20.0 - Required by `internal/tui/update.go` for `errgroup.Group` to fetch Claude and Codex sources concurrently.

**Infrastructure:**
- Go stdlib `encoding/json` - Parses Claude cache, Claude settings, Claude hook input, Codex rollout events, and local state in `cmd/llm-quota/main.go`, `internal/install/claude_hook.go`, `internal/sources/claude.go`, and `internal/sources/codex.go`.
- Go stdlib `os`, `path/filepath`, and `io/fs`-adjacent filesystem APIs - Read, walk, create, back up, and atomically replace local files in `cmd/llm-quota/main.go`, `internal/install/claude_hook.go`, and `internal/sources/codex.go`.
- Go stdlib `os/exec` - Runs an existing Claude statusline passthrough command after cache writing in `internal/install/claude_hook.go`.
- Go stdlib `time` - Computes refresh ticks, reset countdowns, stale windows, backup suffixes, and deterministic test times in `internal/tui/update.go`, `internal/tui/view.go`, `internal/sources/claude.go`, `internal/sources/codex.go`, and `internal/install/claude_hook.go`.

## Configuration

**Environment:**
- No `.env` files detected in the repository root or shallow subdirectories; no environment-variable configuration is implemented in `cmd/llm-quota/main.go`.
- Default runtime paths come from `os.UserHomeDir()` and `os.Executable()` in `cmd/llm-quota/main.go`.
- Claude setup state is stored in `~/.cache/llm-quota/state.json`, with read/write behavior in `internal/install/claude_hook.go`.
- Claude quota cache is stored in `~/.cache/llm-quota/claude.json`, read by `internal/sources/claude.go` and written by `internal/install/claude_hook.go`.
- Claude managed setup modifies only the app-owned `statusLine` entry in `~/.claude/settings.json`, implemented in `internal/install/claude_hook.go`.
- Codex quota data is discovered under `~/.codex/sessions` by `cmd/llm-quota/main.go` and scanned by `internal/sources/codex.go`.

**Build:**
- `go.mod` - Module path, Go version, direct Charm dependencies, and `golang.org/x/sync`.
- `go.sum` - Module checksum lockfile.
- `.gitignore` - Ignores the local built binary `/llm-quota`.
- `README.md` - Documents Homebrew HEAD install, Go install, local build, Claude setup, and runtime operation.
- `Formula/` - Directory present but contains no formula file.
- No Makefile, Dockerfile, GitHub Actions workflow, lint config, or formatter config detected in the repository.

## Platform Requirements

**Development:**
- Go 1.26.3 - Required by `go.mod`; local verification used `go version`.
- Run `go test ./...` as the primary correctness gate for parser, installer, update, and rendering behavior.
- Run `go test -race ./...` before release because refresh work uses goroutines in `internal/tui/update.go`.
- Run `go vet ./...` and `go fmt ./...` for baseline static checks and formatting; no separate lint or format framework is configured.
- A terminal capable of running Bubble Tea full-screen output is required to exercise `cmd/llm-quota/main.go`.

**Production:**
- Local Go binary installed through `go install github.com/robbell5/llm-quota/cmd/llm-quota@latest`, local `go build ./cmd/llm-quota`, or the README-documented Homebrew HEAD path in `README.md`.
- Runtime requires local Claude and Codex data only; no network service, database, daemon, or OAuth provider is configured.
- Claude rows require the app-owned Claude statusline cache writer installed via `llm-quota install-claude-hook`, implemented in `cmd/llm-quota/main.go` and `internal/install/claude_hook.go`.
- Codex rows require local Codex rollout JSONL files under `~/.codex/sessions`, read by `internal/sources/codex.go`.

---

*Stack analysis: 2026-05-21*
