# Coding Conventions

**Analysis Date:** 2026-05-21

## Naming Patterns

**Files:**
- Use standard Go package files with short, domain-specific names in package directories: `internal/sources/claude.go`, `internal/sources/codex.go`, `internal/tui/update.go`, `internal/tui/view.go`, `internal/install/claude_hook.go`.
- Place command entrypoints under `cmd/<binary>/`: `cmd/llm-quota/main.go`.
- Place package tests next to implementation files with `_test.go` suffix: `internal/sources/claude_test.go`, `internal/tui/view_test.go`, `cmd/llm-quota/main_test.go`.
- Use underscore-separated filenames only for multi-word Go files where the domain name is clearer that way: `internal/install/claude_hook.go`, `internal/install/claude_hook_test.go`.

**Functions:**
- Use Go mixedCaps for exported and unexported functions: `NewClaudeReader` in `internal/sources/claude.go`, `sourceBackedModel` in `cmd/llm-quota/main.go`, `renderProgressBar` in `internal/tui/view.go`.
- Use constructor functions named `New<Type>` for concrete readers and models: `NewClaudeReader` in `internal/sources/claude.go`, `NewCodexReader` in `internal/sources/codex.go`, `NewModel` in `internal/tui/model.go`.
- Use `With...` option functions for configurable model construction: `WithReaders`, `WithClock`, `WithRefreshEvery`, and `WithClaudeHookInstalled` in `internal/tui/model.go`.
- Use `Test<Behavior>` for tests and subtest names for variants: `TestClaudeFetch` in `internal/sources/claude_test.go`, `TestRefresh` in `internal/tui/update_test.go`.

**Variables:**
- Use concise lowerCamelCase names that match domain concepts: `cachePath`, `sessionsRoot`, `writtenAt`, `staleAge`, and `refreshEvery` in `internal/sources/claude.go`, `internal/sources/codex.go`, and `internal/tui/model.go`.
- Use `got` and `want` for direct assertions, especially inside helpers: `assertWindows` in `internal/sources/claude_test.go`, `assertRenderedLineWidths` in `internal/tui/view_test.go`.
- Use `tt` or `tc` for table-driven test cases: `TestClaudeFetch` in `internal/sources/claude_test.go`, `TestRunClaudeHookCacheWriterCommandRejectsMissingOrExtraArgs` in `cmd/llm-quota/main_test.go`.
- Use package-level style/color variables for reusable UI state, not repeated inline styles: `shellStyle`, `titleStyle`, `missingStyle`, and `footerStyle` in `internal/tui/view.go`; color tokens in `internal/tui/colors.go`.

**Types:**
- Use exported domain types for cross-package contracts: `Window`, `Product`, `WindowKind`, `SourceError`, and `ErrorCategory` in `internal/sources/window.go`.
- Use unexported struct types for parser and JSON implementation details: `claudeCache`, `claudeCacheWindow` in `internal/sources/claude.go`; `codexEvent`, `codexPayload`, `codexRateLimits` in `internal/sources/codex.go`.
- Use small dependency structs at boundaries where tests inject behavior: `appStreams` and `appDeps` in `cmd/llm-quota/main.go`.
- Use message structs for Bubble Tea update flow, keeping them unexported: `refreshRequestedMsg`, `tickMsg`, `refreshMsg`, and `sourceRefreshResult` in `internal/tui/update.go`.

## Code Style

**Formatting:**
- Use `gofmt` / `go fmt ./...`; no formatter wrapper is present. This is documented in `AGENTS.md` and reflected by standard import grouping in `cmd/llm-quota/main.go`, `internal/tui/update.go`, and `internal/install/claude_hook.go`.
- Use tabs and standard Go alignment as emitted by `gofmt`; keep composite literals readable with one field per line for domain structs such as `sources.Window` in `internal/tui/view_test.go`.
- Use explicit octal file modes for local file writes: `0o600`, `0o700` in `internal/install/claude_hook.go`, `internal/install/claude_hook_test.go`, and `internal/sources/codex_test.go`.

**Linting:**
- No lint config is present. Not detected: `.golangci.yml`, `.golangci.yaml`, `golangci.toml`, `.editorconfig`.
- Use `go vet ./...` as the baseline static check; this is documented in `AGENTS.md`.
- Do not introduce a broad lint framework unless the repository adds a committed config alongside `go.mod`.

## Import Organization

**Order:**
1. Standard library imports first: `encoding/json`, `errors`, `os`, `path/filepath`, `time` in `internal/sources/codex.go`.
2. External imports next: `tea "charm.land/bubbletea/v2"`, `golang.org/x/sync/errgroup`, `charm.land/bubbles/v2/progress`, `charm.land/lipgloss/v2` in `internal/tui/update.go` and `internal/tui/view.go`.
3. Internal module imports last: `github.com/robbell5/llm-quota/internal/sources`, `github.com/robbell5/llm-quota/internal/install`, `github.com/robbell5/llm-quota/internal/tui` in `cmd/llm-quota/main.go`.

**Path Aliases:**
- Use `tea` alias for Bubble Tea v2 imports: `tea "charm.land/bubbletea/v2"` in `cmd/llm-quota/main.go` and `internal/tui/update.go`.
- No Go module path aliases are configured; internal imports use the full module path from `go.mod`: `github.com/robbell5/llm-quota/internal/...`.
- Use Charm v2 import paths only: `charm.land/bubbletea/v2`, `charm.land/bubbles/v2/progress`, and `charm.land/lipgloss/v2` in `go.mod`, `internal/tui/update.go`, and `internal/tui/view.go`.

## Error Handling

**Patterns:**
- Return errors instead of logging or exiting from package code. Source readers in `internal/sources/claude.go` and `internal/sources/codex.go` return `SourceError` values; installers in `internal/install/claude_hook.go` return `InstallResult, error`.
- Normalize external or unexpected reader errors at the TUI boundary with `normalizeSourceError` in `internal/tui/update.go`; preserve typed `SourceError` categories when provided.
- Wrap source-specific filesystem and parse errors with categories from `internal/sources/window.go`: `ErrorMissing`, `ErrorMalformed`, `ErrorNoUsableEvent`, and `ErrorRead`.
- Convert CLI errors to stderr and exit codes only in `cmd/llm-quota/main.go`. Use exit code `2` for invalid arguments and `1` for runtime failures.
- Preserve last-known-good UI data on per-source refresh failures. `mergeRefresh` in `internal/tui/update.go` stores errors without deleting existing `windows` for that product.
- Use `errors.Is` for sentinel filesystem checks and `errors.As` for typed source errors: `internal/sources/claude.go`, `internal/sources/codex.go`, and `internal/tui/update.go`.
- Include actionable context when decoding user/tool input: `fmt.Errorf("decode Claude hook input: %w", err)` in `internal/install/claude_hook.go`.

## Logging

**Framework:** console

**Patterns:**
- There is no logging framework. Package code in `internal/sources/*.go`, `internal/tui/*.go`, and `internal/install/claude_hook.go` does not log.
- CLI output uses `fmt.Fprintln` and `fmt.Fprintf` against injected streams in `cmd/llm-quota/main.go`.
- Renderable recovery hints are user-facing strings in `internal/tui/view.go`, not log lines. Do not expose raw error categories such as `read_error` or `no_usable_event` in the TUI.

## Comments

**When to Comment:**
- Comments are sparse. Prefer self-explanatory names and small functions over explanatory comments, following `internal/sources/claude.go`, `internal/sources/codex.go`, `internal/tui/update.go`, and `internal/tui/view.go`.
- Add comments only when code needs operational context that names cannot carry. Existing code generally uses none.
- Do not add noisy comments to table-driven tests; use descriptive test names and assertion messages as in `internal/install/claude_hook_test.go`.

**JSDoc/TSDoc:**
- Not applicable. This is a Go module.
- Go doc comments are not used consistently for exported identifiers yet. If adding exported API intended for external consumers, add standard Go doc comments in the same file as the exported type or function.

## Function Design

**Size:** Keep functions focused on one boundary or transformation.
- Parsing and validation stay close to data structures: `claudeCache.validate` and `claudeCacheWindow.validate` in `internal/sources/claude.go`; `codexRateLimits.validate` and `codexRateLimitWindow.validate` in `internal/sources/codex.go`.
- UI render helpers split by concern: `renderRows`, `renderDataRow`, `renderFooter`, `footerRecoveryHints`, and `appendHintWithinWidth` in `internal/tui/view.go`.
- CLI dispatch is centralized in `run` in `cmd/llm-quota/main.go`; command-specific behavior lives in helpers such as `runInstallClaudeHook`, `runUninstallClaudeHook`, and `runClaudeStatusLineCacheWriter`.

**Parameters:** Pass dependencies explicitly at boundaries.
- Use constructor path arguments for filesystem readers: `NewClaudeReader(cachePath string)` in `internal/sources/claude.go`, `NewCodexReader(sessionsRoot string)` in `internal/sources/codex.go`.
- Pass `now time.Time` or clock functions into read/update paths for deterministic tests: `Fetch(now time.Time)` in `internal/sources/window.go`, `WithClock` in `internal/tui/model.go`, `RunClaudeHookCacheWriter(..., now time.Time)` in `internal/install/claude_hook.go`.
- Use dependency structs for CLI tests rather than package globals: `appStreams` and `appDeps` in `cmd/llm-quota/main.go`.

**Return Values:** Return data plus typed errors.
- Source readers return `([]sources.Window, error)` and use an empty/nil window slice on error: `internal/sources/claude.go`, `internal/sources/codex.go`.
- Install operations return `InstallResult` with `Changed`, `BackupPath`, and `Message` plus `error`: `internal/install/claude_hook.go`.
- Bubble Tea updates return `(tea.Model, tea.Cmd)` and keep mutation inside `Update`: `internal/tui/update.go`.

## Module Design

**Exports:** Export only cross-package contracts and constructors.
- `internal/sources/window.go` exports shared domain types used by `internal/tui` and `cmd/llm-quota`.
- `internal/sources/claude.go` and `internal/sources/codex.go` export reader constructors and `Fetch` methods; JSON struct types remain unexported.
- `internal/tui/model.go` exports `Model`, `Option`, `SourceReader`, `NewModel`, and options needed by `cmd/llm-quota/main.go`; message and render internals stay unexported in `internal/tui/update.go` and `internal/tui/view.go`.
- `internal/install/claude_hook.go` exports install/cache-writer functions used by `cmd/llm-quota/main.go`; helper functions such as `readClaudeConfig`, `writeJSONAtomic`, and `shellQuote` remain package-private.

**Barrel Files:** Not used.
- There are no aggregate export files. Add new code to the package that owns the behavior instead of creating a barrel file.
- Keep project-local commands in `cmd/llm-quota/main.go`, source parsing in `internal/sources/`, TUI logic in `internal/tui/`, and Claude setup/cache writer behavior in `internal/install/`.

---

*Convention analysis: 2026-05-21*
