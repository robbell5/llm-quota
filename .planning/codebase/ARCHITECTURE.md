<!-- refreshed: 2026-05-21 -->
# Architecture

**Analysis Date:** 2026-05-21

## System Overview

```text
+-------------------------------------------------------------+
|                  Foreground CLI / TUI App                   |
|                  `cmd/llm-quota/main.go`                    |
+-------------------+-------------------+---------------------+
| Commands          | First-launch setup| TUI startup         |
| `cmd/llm-quota`   | `internal/install`| `internal/tui`      |
+---------+---------+---------+---------+----------+----------+
          |                   |                    |
          v                   v                    v
+-------------------------------------------------------------+
|                       Domain Sources                        |
|                       `internal/sources`                    |
| `ClaudeReader` reads cache JSON; `CodexReader` scans JSONL  |
+---------------------+-------------------+-------------------+
                      |                   |
                      v                   v
+-------------------------------------------------------------+
|                   Local Filesystem State                    |
| `~/.cache/llm-quota/claude.json`                            |
| `~/.cache/llm-quota/state.json`                             |
| `~/.claude/settings.json`                                   |
| `~/.codex/sessions/**/rollout-*.jsonl`                      |
+-------------------------------------------------------------+
```

## Component Responsibilities

| Component | Responsibility | File |
|-----------|----------------|------|
| CLI entrypoint | Parses subcommands, prompts for first-launch Claude setup, constructs readers, starts Bubble Tea. | `cmd/llm-quota/main.go` |
| Dependency injection harness | Allows tests to replace filesystem paths, hook installer functions, Codex root discovery, and TUI startup. | `cmd/llm-quota/main.go` |
| TUI model | Owns Bubble Tea state: terminal size, refresh cadence, reader interfaces, source windows, typed source errors, and Claude hook installation state. | `internal/tui/model.go` |
| TUI update loop | Handles keypresses, resize messages, refresh scheduling, concurrent source fetches, last-known-good merge behavior, and stale marking. | `internal/tui/update.go` |
| TUI renderer | Renders title, four fixed quota rows, responsive progress bars, reset countdowns, placeholders, and recovery footer hints. | `internal/tui/view.go` |
| TUI theme | Centralizes Lip Gloss colors for renderer styles. | `internal/tui/colors.go` |
| Source domain model | Defines products, rolling window kinds, unified quota window data, metadata, and typed source errors. | `internal/sources/window.go` |
| Claude source reader | Reads hook-produced Claude cache JSON and converts it to two `Window` values. | `internal/sources/claude.go` |
| Codex source reader | Walks local Codex session rollouts, selects the newest usable rate-limit event, and converts it to two `Window` values. | `internal/sources/codex.go` |
| Claude setup installer | Installs/uninstalls the managed Claude statusline cache writer, records first-launch decline state, writes cache JSON atomically, and preserves unrelated Claude config. | `internal/install/claude_hook.go` |
| Tool dependency anchors | Keeps tool-only imports in the module graph behind a build tag. | `tools/tools.go` |

## Pattern Overview

**Overall:** Small layered Go CLI using Bubble Tea Model-Update-View plus local-file source adapters.

**Key Characteristics:**
- Keep process entry and user prompts in `cmd/llm-quota/main.go`; do not put CLI argument handling inside `internal/tui` or `internal/sources`.
- Keep UI state mutation in `internal/tui/update.go`; background work returns Bubble Tea messages and does not mutate `Model` directly.
- Keep source parsing in `internal/sources`; the TUI consumes only `sources.Window` and `sources.SourceError`.
- Keep Claude configuration writes in `internal/install`; `internal/sources` reads cache data and does not modify Claude config.
- Use constructor-injected paths and interfaces for testable filesystem behavior in `cmd/llm-quota/main.go`, `internal/tui/model.go`, `internal/sources/claude.go`, and `internal/sources/codex.go`.

## Layers

**Command Layer:**
- Purpose: Parse supported commands, run setup commands, prompt for Claude hook installation, resolve default paths, and launch Bubble Tea.
- Location: `cmd/llm-quota/main.go`
- Contains: `main`, `run`, subcommand runners, first-launch prompt logic, default path helpers, hook-installed detection, and `startTUI`.
- Depends on: `internal/install`, `internal/sources`, `internal/tui`, `charm.land/bubbletea/v2`, and Go stdlib.
- Used by: End users invoking `llm-quota`, tests in `cmd/llm-quota/main_test.go`, and managed Claude statusline command entries.

**TUI Layer:**
- Purpose: Present quota state in a small terminal pane and manage refresh lifecycle.
- Location: `internal/tui`
- Contains: Bubble Tea `Model`, options, update handlers, refresh commands, render functions, colors, and responsive row layout.
- Depends on: `internal/sources`, `charm.land/bubbletea/v2`, `charm.land/bubbles/v2/progress`, `charm.land/lipgloss/v2`, and `golang.org/x/sync/errgroup`.
- Used by: `cmd/llm-quota/main.go` through `tui.NewModel` and `tea.NewProgram`.

**Source Layer:**
- Purpose: Convert local Claude and Codex data into a shared quota window model.
- Location: `internal/sources`
- Contains: `Window`, `SourceError`, `ClaudeReader`, `CodexReader`, source-specific JSON structs, validation, and parsing helpers.
- Depends on: Go stdlib filesystem, JSON, path, sorting, string, error, and time packages.
- Used by: `internal/tui` for refreshes and `cmd/llm-quota/main.go` for source construction.

**Install Layer:**
- Purpose: Manage the app-owned Claude statusline cache writer and cache-writing command behavior.
- Location: `internal/install`
- Contains: `ClaudeHookPaths`, `InstallResult`, installer/uninstaller functions, decline-state persistence, managed command builders, cache writer runners, atomic JSON writes, symlink resolution, and shell quoting.
- Depends on: Go stdlib filesystem, JSON, process execution, reflection, string, and time packages.
- Used by: `cmd/llm-quota/main.go` and tests in `internal/install/claude_hook_test.go`.

**Documentation and Planning Layer:**
- Purpose: Store user-facing documentation and GSD planning context.
- Location: `README.md`, `AGENTS.md`, `.planning`, `docs/superpowers/specs`
- Contains: Install/run guidance in `README.md`, project constraints in `AGENTS.md`, planning artifacts under `.planning`, and design research under `docs/superpowers/specs/2026-05-16-llm-quota-tui-design.md`.
- Depends on: Not applicable.
- Used by: Contributors and GSD planning commands.

## Data Flow

### Primary TUI Startup And Refresh Path

1. `main` delegates to `run` and passes CLI args (`cmd/llm-quota/main.go:36`, `cmd/llm-quota/main.go:40`).
2. No-argument startup calls `offerFirstLaunchInstall`, then builds a model through `sourceBackedModel` (`cmd/llm-quota/main.go:65`, `cmd/llm-quota/main.go:258`).
3. `sourceBackedModel` resolves Claude paths and Codex sessions root, checks whether the Claude hook is installed, constructs `sources.NewClaudeReader` and `sources.NewCodexReader`, and returns `tui.NewModel` with `tui.WithReaders` (`cmd/llm-quota/main.go:258`, `internal/sources/claude.go:15`, `internal/sources/codex.go:18`, `internal/tui/model.go:30`).
4. `startTUI` starts `tea.NewProgram(model)` (`cmd/llm-quota/main.go:382`).
5. `Model.Init` batches an immediate refresh request with the next 30-second tick (`internal/tui/update.go:28`).
6. `Model.Update` handles `refreshRequestedMsg`, sets `refreshing`, and returns `refreshCmd` (`internal/tui/update.go:32`, `internal/tui/update.go:77`).
7. `refreshCmd` fetches Claude and Codex concurrently with `errgroup.Group` and returns a single `refreshMsg` (`internal/tui/update.go:77`).
8. `fetchSource` normalizes missing readers and reader errors into `sources.SourceError` (`internal/tui/update.go:104`).
9. `mergeRefresh` stores successful windows, preserves last-known-good windows for failed products, clears resolved errors, and marks old data stale (`internal/tui/update.go:139`).
10. `Model.View` delegates to `render`, which renders the fixed four-row quota screen (`internal/tui/update.go:173`, `internal/tui/view.go:51`).

### Claude Cache Setup And Read Path

1. `run` routes `install-claude-hook` to `runInstallClaudeHook` (`cmd/llm-quota/main.go:48`, `cmd/llm-quota/main.go:80`).
2. `InstallClaudeHook` reads `~/.claude/settings.json`, installs a managed `statusLine.command`, removes older managed `PostToolUse` entries, backs up changed config, and writes JSON atomically (`internal/install/claude_hook.go:37`).
3. The managed statusline command invokes `llm-quota claude-statusline-cache-writer --cache <path>` built by `ManagedStatusLineCommand` (`internal/install/claude_hook.go:250`).
4. The hidden writer command routes through `runClaudeStatusLineCacheWriter` and `install.RunClaudeStatusLineCacheWriter` (`cmd/llm-quota/main.go:121`, `internal/install/claude_hook.go:269`).
5. `writeClaudeCache` validates Claude rate-limit payloads and writes `~/.cache/llm-quota/claude.json` atomically (`internal/install/claude_hook.go:286`, `internal/install/claude_hook.go:404`).
6. `ClaudeReader.Fetch` reads the cache JSON and returns Claude five-hour and seven-day `Window` values (`internal/sources/claude.go:19`).

### Codex Rollout Read Path

1. `sourceBackedModel` constructs `sources.NewCodexReader` with `~/.codex/sessions` (`cmd/llm-quota/main.go:258`, `cmd/llm-quota/main.go:281`, `internal/sources/codex.go:18`).
2. `CodexReader.Fetch` checks the root exists, calls `rolloutCandidates`, and scans candidates newest-first (`internal/sources/codex.go:22`, `internal/sources/codex.go:54`).
3. `rolloutCandidates` walks `~/.codex/sessions`, keeps files named `rollout-*.jsonl`, and sorts by modification time descending (`internal/sources/codex.go:54`).
4. `windowsFromCodexRollout` reads each candidate file, splits lines, and keeps the last usable token-count rate-limit event in that file (`internal/sources/codex.go:91`).
5. `parseCodexRateLimitLine` accepts only `event_msg` payloads with `payload.type == "token_count"` and valid primary/secondary windows (`internal/sources/codex.go:122`).
6. `codexRateLimits.windows` returns Codex five-hour and seven-day `Window` values with optional `plan_type` metadata (`internal/sources/codex.go:154`).

### First-Launch Consent Path

1. No-argument `run` calls `offerFirstLaunchInstall` before model construction (`cmd/llm-quota/main.go:65`, `cmd/llm-quota/main.go:157`).
2. `offerFirstLaunchInstall` resolves paths, skips prompting if installed, and checks the decline state through `install.ClaudeHookDeclined` (`cmd/llm-quota/main.go:157`, `internal/install/claude_hook.go:138`).
3. Accepted prompts call `InstallClaudeHook`; declined prompts call `RecordClaudeHookDeclined` and then continue to TUI startup (`cmd/llm-quota/main.go:190`, `cmd/llm-quota/main.go:202`, `internal/install/claude_hook.go:125`).

**State Management:**
- Runtime UI state lives in `tui.Model` fields in `internal/tui/model.go`.
- Last-known-good quota state is kept in memory at `Model.windows` and merged by `mergeRefresh` in `internal/tui/update.go`.
- Source failures are stored per product in `Model.errors` using `sources.SourceError` from `internal/sources/window.go`.
- Persistent user-decline state is written to `~/.cache/llm-quota/state.json` through `internal/install/claude_hook.go`.
- Claude quota cache state is written to `~/.cache/llm-quota/claude.json` through `internal/install/claude_hook.go` and read by `internal/sources/claude.go`.

## Key Abstractions

**`sources.Window`:**
- Purpose: Shared display model for Claude and Codex quota windows.
- Examples: `internal/sources/window.go`, `internal/sources/claude.go`, `internal/sources/codex.go`, `internal/tui/view.go`
- Pattern: Source-specific readers convert private local data shapes to one stable UI-facing struct.

**`sources.SourceError`:**
- Purpose: Preserve source name and normalized category without exposing raw parser details to the renderer.
- Examples: `internal/sources/window.go`, `internal/tui/update.go`, `internal/tui/view.go`
- Pattern: Reader errors are normalized at the TUI boundary and drive footer hints.

**`tui.SourceReader`:**
- Purpose: Define the narrow fetch contract used by refresh commands.
- Examples: `internal/tui/model.go`, `internal/sources/claude.go`, `internal/sources/codex.go`, `internal/tui/update_test.go`
- Pattern: Interface is owned by the consumer package; concrete readers remain in `internal/sources`.

**`tui.Model` plus `Option`:**
- Purpose: Encapsulate Bubble Tea state and allow tests to inject readers, clocks, refresh interval, and hook-installed state.
- Examples: `internal/tui/model.go`, `internal/tui/update_test.go`, `internal/tui/view_test.go`
- Pattern: Functional options keep defaults local to `NewModel`.

**`appDeps` and `appStreams`:**
- Purpose: Make CLI behavior testable without touching real home-directory files or starting a real TUI.
- Examples: `cmd/llm-quota/main.go`, `cmd/llm-quota/main_test.go`
- Pattern: Main package uses explicit dependency structs instead of globals.

**`install.ClaudeHookPaths`:**
- Purpose: Pass all Claude setup paths as a single value and keep path discovery outside `internal/install`.
- Examples: `internal/install/claude_hook.go`, `cmd/llm-quota/main.go`, `internal/install/claude_hook_test.go`
- Pattern: Installer accepts paths from caller; tests supply temp paths.

## Entry Points

**Default TUI:**
- Location: `cmd/llm-quota/main.go`
- Triggers: `llm-quota` with no arguments.
- Responsibilities: Offer first-launch Claude setup, construct source-backed `tui.Model`, and run Bubble Tea.

**Install Claude Hook:**
- Location: `cmd/llm-quota/main.go`
- Triggers: `llm-quota install-claude-hook`.
- Responsibilities: Resolve default paths and call `install.InstallClaudeHook`.

**Uninstall Claude Hook:**
- Location: `cmd/llm-quota/main.go`
- Triggers: `llm-quota uninstall-claude-hook`.
- Responsibilities: Resolve default paths and call `install.UninstallClaudeHook`.

**Claude Statusline Cache Writer:**
- Location: `cmd/llm-quota/main.go`, `internal/install/claude_hook.go`
- Triggers: Managed Claude `statusLine.command` built by `install.ManagedStatusLineCommand`.
- Responsibilities: Read Claude statusline JSON from stdin, write local cache when rate limits are present, and optionally run a passthrough statusline command.

**Legacy Claude Hook Cache Writer:**
- Location: `cmd/llm-quota/main.go`, `internal/install/claude_hook.go`
- Triggers: `llm-quota claude-hook-cache-writer --cache <path>`.
- Responsibilities: Read Claude hook JSON from stdin and require rate limits before writing local cache.

**Package Tests:**
- Location: `cmd/llm-quota/main_test.go`, `internal/install/claude_hook_test.go`, `internal/sources/*_test.go`, `internal/tui/*_test.go`
- Triggers: `go test ./...`.
- Responsibilities: Validate CLI routing, installer behavior, source parsing, update transitions, and rendering.

## Architectural Constraints

- **Threading:** Bubble Tea drives state transitions on the event loop in `internal/tui/update.go`; source reads run concurrently inside `refreshCmd` with `errgroup.Group` and return data by message.
- **Global state:** Renderer styles and colors are package-level values in `internal/tui/view.go` and `internal/tui/colors.go`; no mutable service singleton is used for source state.
- **Circular imports:** Not detected. The package graph is `cmd/llm-quota` -> `internal/install`, `internal/sources`, `internal/tui`; `internal/tui` -> `internal/sources`; `internal/install` and `internal/sources` do not import app packages.
- **Filesystem only:** Steady-state data access is local file reads from `internal/sources/claude.go` and `internal/sources/codex.go`; setup writes local Claude settings, app cache, and app state through `internal/install/claude_hook.go`.
- **Internal package boundary:** Application code belongs under `internal/*` unless it is the executable entrypoint under `cmd/llm-quota`.
- **Small fixed display:** Rendering is built around four fixed rows in `internal/tui/view.go`; do not introduce list/table abstractions for the current quota windows.

## Anti-Patterns

### Parsing Source-Specific JSON In The TUI

**What happens:** New parsing code is added to `internal/tui/view.go` or `internal/tui/update.go`.
**Why it's wrong:** The TUI layer is designed to consume `sources.Window` and `sources.SourceError`; source-specific JSON details belong in `internal/sources`.
**Do this instead:** Add parsing and validation to `internal/sources/claude.go` or `internal/sources/codex.go`, then return `sources.Window` values from the reader.

### Mutating Model From Goroutines

**What happens:** A goroutine writes directly to `Model.windows`, `Model.errors`, or `Model.refreshing`.
**Why it's wrong:** Bubble Tea state changes must happen through `Update` in `internal/tui/update.go`; direct goroutine mutation creates races and bypasses last-known-good merge rules.
**Do this instead:** Return a message from a `tea.Cmd`, then merge it in `Model.Update` and `mergeRefresh` in `internal/tui/update.go`.

### Reading Real Home Files In Tests

**What happens:** Tests rely on `~/.claude`, `~/.codex`, or `~/.cache/llm-quota`.
**Why it's wrong:** Existing tests use injected paths, readers, streams, and temp directories to avoid mutating user state.
**Do this instead:** Inject paths through `appDeps` in `cmd/llm-quota/main.go`, use source constructors in `internal/sources`, and use `t.TempDir()` in tests.

### Adding Broad Configuration Surface

**What happens:** A config framework or persistent settings package is added for behavior already derived from default local paths.
**Why it's wrong:** The product is a dedicated local-only TUI with minimal setup commands; extra configuration expands the architecture without an existing use path.
**Do this instead:** Keep path discovery in `cmd/llm-quota/main.go` and path-dependent filesystem operations behind constructors or explicit path structs.

## Error Handling

**Strategy:** Convert local-file and parse failures into typed source errors, preserve last-known-good display data, and render short recovery hints instead of crashing the TUI.

**Patterns:**
- Source readers return `sources.SourceError` with categories from `internal/sources/window.go`.
- `internal/tui/update.go` normalizes unknown errors to `ErrorRead`.
- `mergeRefresh` keeps existing product windows when a later refresh for that product fails.
- `internal/tui/view.go` renders placeholders and user-action hints rather than raw error category strings.
- CLI setup commands return process exit codes and prefix stderr messages with `llm-quota:` in `cmd/llm-quota/main.go`.
- Installer writes JSON atomically with temporary files and `os.Rename` in `internal/install/claude_hook.go`.

## Cross-Cutting Concerns

**Logging:** No logging framework is used. CLI commands write human-facing errors to stderr in `cmd/llm-quota/main.go`; TUI recovery hints are rendered in `internal/tui/view.go`.
**Validation:** Source payload validation lives beside parser structs in `internal/sources/claude.go`, `internal/sources/codex.go`, and `internal/install/claude_hook.go`.
**Authentication:** Not applicable. The app reads local files and does not perform OAuth, network calls, or Keychain reads.
**Privacy:** Private Claude and Codex payloads are not rendered; `internal/tui/view.go` shows percentages, reset countdowns, placeholders, and recovery hints.
**Atomic writes:** Claude config, app state, and cache writes go through `writeJSONAtomic` in `internal/install/claude_hook.go`.

---

*Architecture analysis: 2026-05-21*
