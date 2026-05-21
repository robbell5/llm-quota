# Codebase Structure

**Analysis Date:** 2026-05-21

## Directory Layout

```text
llm-quota/
|-- AGENTS.md                         # Project instructions and GSD context
|-- README.md                         # User-facing install, setup, run, and troubleshooting docs
|-- go.mod                            # Go module and direct dependency declarations
|-- go.sum                            # Go module checksums
|-- .gitignore                        # Ignores local built binary `llm-quota`
|-- cmd/
|   `-- llm-quota/
|       |-- main.go                   # Executable entrypoint, CLI routing, path defaults, TUI launch
|       `-- main_test.go              # CLI and startup behavior tests
|-- internal/
|   |-- install/
|   |   |-- claude_hook.go            # Claude statusline setup, cache writers, atomic JSON writes
|   |   `-- claude_hook_test.go       # Installer and cache writer tests
|   |-- sources/
|   |   |-- window.go                 # Shared source domain types and typed errors
|   |   |-- claude.go                 # Claude cache reader
|   |   |-- codex.go                  # Codex rollout reader
|   |   |-- claude_test.go            # Claude reader tests
|   |   `-- codex_test.go             # Codex reader tests
|   `-- tui/
|       |-- model.go                  # Bubble Tea model and options
|       |-- update.go                 # Bubble Tea update loop and refresh commands
|       |-- view.go                   # Renderer and responsive row layout
|       |-- colors.go                 # Lip Gloss theme colors
|       |-- update_test.go            # Update and refresh behavior tests
|       `-- view_test.go              # Render and responsive layout tests
|-- tools/
|   `-- tools.go                      # Build-tagged tool dependency anchors
|-- docs/
|   `-- superpowers/specs/
|       `-- 2026-05-16-llm-quota-tui-design.md # Design specification
|-- Formula/                          # Homebrew formula directory, currently empty
`-- .planning/                        # GSD project state, research, milestones, and codebase maps
```

## Directory Purposes

**`cmd/llm-quota`:**
- Purpose: Contains the single executable package for the `llm-quota` binary.
- Contains: CLI argument routing, setup command runners, first-launch prompt flow, default path discovery, source-backed model construction, and Bubble Tea startup.
- Key files: `cmd/llm-quota/main.go`, `cmd/llm-quota/main_test.go`

**`internal/tui`:**
- Purpose: Contains all terminal UI state, update, rendering, and styling code.
- Contains: Bubble Tea model, reader interface, functional options, refresh commands, concurrent fetch orchestration, last-known-good merge behavior, responsive row renderer, footer hints, and color constants.
- Key files: `internal/tui/model.go`, `internal/tui/update.go`, `internal/tui/view.go`, `internal/tui/colors.go`, `internal/tui/update_test.go`, `internal/tui/view_test.go`

**`internal/sources`:**
- Purpose: Contains local data readers and the shared source domain model.
- Contains: Product/window enums, unified `Window` struct, `SourceError` categories, Claude cache parser, Codex rollout scanner/parser, source-specific validation structs, and parser tests.
- Key files: `internal/sources/window.go`, `internal/sources/claude.go`, `internal/sources/codex.go`, `internal/sources/claude_test.go`, `internal/sources/codex_test.go`

**`internal/install`:**
- Purpose: Contains Claude setup, uninstall, cache writer, state persistence, command construction, backup, symlink, and atomic write behavior.
- Contains: `ClaudeHookPaths`, `InstallResult`, managed statusline installer, first-launch decline state helpers, stdin cache writer runners, shell quoting, backup creation, and JSON write helpers.
- Key files: `internal/install/claude_hook.go`, `internal/install/claude_hook_test.go`

**`tools`:**
- Purpose: Pins tool/import-only dependencies that are needed by the project but not referenced directly by normal builds in every path.
- Contains: Build-tagged blank imports.
- Key files: `tools/tools.go`

**`docs`:**
- Purpose: Stores design/specification documents.
- Contains: Product design details and implementation research under `docs/superpowers/specs`.
- Key files: `docs/superpowers/specs/2026-05-16-llm-quota-tui-design.md`

**`.planning`:**
- Purpose: Stores GSD project artifacts and generated codebase maps.
- Contains: Project state, roadmap, requirements, research notes, milestone artifacts, debug notes, and `.planning/codebase`.
- Key files: `.planning/PROJECT.md`, `.planning/STATE.md`, `.planning/ROADMAP.md`, `.planning/REQUIREMENTS.md`, `.planning/codebase/ARCHITECTURE.md`, `.planning/codebase/STRUCTURE.md`

**`Formula`:**
- Purpose: Placeholder directory for Homebrew formula-related files.
- Contains: No files detected.
- Key files: Not detected.

## Key File Locations

**Entry Points:**
- `cmd/llm-quota/main.go`: Main executable entrypoint and all supported CLI subcommands.
- `cmd/llm-quota/main.go`: `llm-quota` no-argument TUI startup path through `run`, `offerFirstLaunchInstall`, `sourceBackedModel`, and `startTUI`.
- `cmd/llm-quota/main.go`: `install-claude-hook`, `uninstall-claude-hook`, `claude-statusline-cache-writer`, and `claude-hook-cache-writer` command routing.

**Configuration:**
- `go.mod`: Module path `github.com/robbell5/llm-quota`, Go version, and direct dependencies.
- `go.sum`: Dependency checksums.
- `.gitignore`: Ignores the local built binary `llm-quota`.
- `AGENTS.md`: Project instructions, constraints, stack guidance, and GSD workflow rules.

**Core Logic:**
- `internal/tui/model.go`: State shape and model options.
- `internal/tui/update.go`: Refresh lifecycle, Bubble Tea message handling, concurrent source fetching, stale marking, and last-known-good merge logic.
- `internal/tui/view.go`: Display rendering, responsive layouts, progress bars, reset text, and footer recovery hints.
- `internal/sources/window.go`: Shared source domain model.
- `internal/sources/claude.go`: Claude cache reader.
- `internal/sources/codex.go`: Codex rollout reader.
- `internal/install/claude_hook.go`: Claude setup/uninstall and cache writer implementation.

**Testing:**
- `cmd/llm-quota/main_test.go`: CLI routing, first-launch prompt, source-backed model startup, and hidden cache writer command tests.
- `internal/tui/update_test.go`: Bubble Tea update behavior, refresh coalescing, errors, stale marking, and tick scheduling tests.
- `internal/tui/view_test.go`: Renderer output, footer hints, responsive row layouts, progress thresholds, and line-width tests.
- `internal/sources/claude_test.go`: Claude cache parsing and source error tests.
- `internal/sources/codex_test.go`: Codex rollout selection and parser tests.
- `internal/install/claude_hook_test.go`: Claude config preservation, statusline wrapping, symlink behavior, uninstall behavior, and cache writer tests.

**Documentation:**
- `README.md`: Install, setup, run, keys, troubleshooting, and product scope.
- `docs/superpowers/specs/2026-05-16-llm-quota-tui-design.md`: TUI design specification.
- `.planning/research/STACK.md`: Stack research embedded in project context.

## Naming Conventions

**Files:**
- Use package-oriented lowercase Go filenames: `internal/tui/model.go`, `internal/tui/update.go`, `internal/tui/view.go`.
- Use source-specific names for reader implementations: `internal/sources/claude.go`, `internal/sources/codex.go`.
- Use behavior-specific test names colocated with package code: `internal/tui/view_test.go`, `internal/install/claude_hook_test.go`.
- Keep the executable package under `cmd/<binary>/main.go`: `cmd/llm-quota/main.go`.

**Directories:**
- Use Go's `cmd/<binary>` convention for executable packages: `cmd/llm-quota`.
- Use Go's `internal/<package>` convention for app-private packages: `internal/tui`, `internal/sources`, `internal/install`.
- Keep generated/planning documents under `.planning`, not under source packages.
- Keep design documents under `docs/superpowers/specs`.

**Packages and Types:**
- Package names are short lowercase nouns: `tui`, `sources`, `install`, `main`.
- Export constructors and public data types needed across packages: `sources.NewClaudeReader`, `sources.NewCodexReader`, `sources.Window`, `install.InstallClaudeHook`, `tui.NewModel`.
- Keep source-specific JSON payload structs unexported: `claudeCache`, `codexRateLimits`, `claudeHookPayload`.
- Use `Product` and `WindowKind` typed string constants from `internal/sources/window.go` instead of raw product/window strings in TUI logic.

## Where to Add New Code

**New CLI Command:**
- Primary code: Add routing and command runner to `cmd/llm-quota/main.go`.
- Tests: Add command behavior tests to `cmd/llm-quota/main_test.go`.
- Rules: Keep command side effects behind `appDeps` or explicit streams when tests need isolation.

**New TUI State Or Behavior:**
- Primary code: Add fields/options to `internal/tui/model.go` and message/update handling to `internal/tui/update.go`.
- Rendering: Add display logic to `internal/tui/view.go` and shared colors to `internal/tui/colors.go`.
- Tests: Add transition tests to `internal/tui/update_test.go` and render tests to `internal/tui/view_test.go`.
- Rules: Mutate `Model` only in `Update` or helpers called from `Update`.

**New Source Reader:**
- Primary code: Add reader implementation under `internal/sources/<source>.go`.
- Shared model: Reuse `sources.Window`, `sources.Product`, `sources.WindowKind`, and `sources.SourceError` from `internal/sources/window.go`.
- TUI integration: Inject the reader through `tui.WithReaders` or an added model option in `internal/tui/model.go`.
- Tests: Add parser and source error tests under `internal/sources/<source>_test.go`.
- Rules: Source readers read local data and return typed errors; they do not render, log, or exit.

**New Claude Setup Behavior:**
- Primary code: Add setup, command-building, or writer behavior to `internal/install/claude_hook.go`.
- CLI wiring: Route user-facing commands through `cmd/llm-quota/main.go`.
- Tests: Add installer and writer tests to `internal/install/claude_hook_test.go`; add command routing tests to `cmd/llm-quota/main_test.go` when CLI behavior changes.
- Rules: Preserve unrelated Claude config and use `writeJSONAtomic` for JSON writes.

**New Rendering Component/Module:**
- Implementation: Keep four-row quota rendering in `internal/tui/view.go` unless a separate helper file in `internal/tui` removes meaningful complexity.
- Tests: Add width and content assertions to `internal/tui/view_test.go`.
- Rules: Respect 50-column and narrower layouts; use Lip Gloss width measurement for rendered text.

**Utilities:**
- Shared source helpers: Add to `internal/sources` only when used by multiple source readers.
- Shared TUI helpers: Add to `internal/tui` only when used by update/render logic.
- Shared install helpers: Add to `internal/install` when they support Claude setup, cache writing, or local state writes.
- Avoid broad `internal/util` packages; place helpers with the package that owns the behavior.

**Documentation Updates:**
- User-facing usage: Update `README.md`.
- GSD project context: Update `.planning/*` through GSD workflows.
- Design/spec updates: Add or update files under `docs/superpowers/specs`.

## Special Directories

**`.planning`:**
- Purpose: GSD project state, requirements, roadmap, research, debug notes, milestone artifacts, and generated codebase maps.
- Generated: Yes
- Committed: Yes

**`.planning/codebase`:**
- Purpose: Generated codebase intelligence consumed by GSD planning and execution commands.
- Generated: Yes
- Committed: Yes

**`Formula`:**
- Purpose: Homebrew formula workspace.
- Generated: No
- Committed: Yes

**`tools`:**
- Purpose: Build-tagged dependency anchors for tool/import-only packages.
- Generated: No
- Committed: Yes

**`docs/superpowers/specs`:**
- Purpose: Product and design specification documents.
- Generated: No
- Committed: Yes

**Local binary `llm-quota`:**
- Purpose: Developer-built executable from `go build ./cmd/llm-quota`.
- Generated: Yes
- Committed: No, ignored by `.gitignore`.

---

*Structure analysis: 2026-05-21*
