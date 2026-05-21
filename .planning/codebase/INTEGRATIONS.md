# External Integrations

**Analysis Date:** 2026-05-21

## APIs & External Services

**Claude Code local statusline integration:**
- Claude Code - Supplies rate-limit data through the JSON payload sent to the configured Claude `statusLine.command`; the app installs a managed command into `~/.claude/settings.json`.
  - SDK/Client: Go stdlib only; setup and cache writing are implemented in `internal/install/claude_hook.go`, and command routing is in `cmd/llm-quota/main.go`.
  - Auth: Not applicable; the app does not read Claude credentials, OAuth tokens, or macOS Keychain entries.
  - Local config: `~/.claude/settings.json` is read and updated by `internal/install/claude_hook.go`.
  - Cache output: `~/.cache/llm-quota/claude.json` is written by `internal/install/claude_hook.go` and read by `internal/sources/claude.go`.
  - Commands: `llm-quota install-claude-hook`, `llm-quota uninstall-claude-hook`, `llm-quota claude-statusline-cache-writer --cache <path> [--passthrough <command>]`, and legacy `llm-quota claude-hook-cache-writer --cache <path>` are dispatched in `cmd/llm-quota/main.go`.

**Codex local session files:**
- Codex CLI - Supplies quota data through local session rollout JSONL files under `~/.codex/sessions`.
  - SDK/Client: Go stdlib only; `internal/sources/codex.go` uses `filepath.WalkDir`, `os.ReadFile`, `encoding/json`, and line splitting.
  - Auth: Not applicable; the app does not read Codex credentials or call Codex network APIs.
  - Local data: `internal/sources/codex.go` scans `rollout-*.jsonl` files and parses `event_msg` payloads with `type: "token_count"` and non-null `rate_limits`.
  - Default path: `cmd/llm-quota/main.go` constructs `~/.codex/sessions` from `os.UserHomeDir()`.

**Network APIs:**
- Not detected - No `net/http`, OAuth client, Anthropic API client, Codex API client, database driver, telemetry exporter, or webhook HTTP server is imported in application code under `cmd/` or `internal/`.
- Explicitly out of scope - `.planning/PROJECT.md`, `README.md`, and `docs/superpowers/specs/2026-05-16-llm-quota-tui-design.md` state that the app avoids network fallback, OAuth, Keychain reads, daemons, and credential handling.

## Data Storage

**Databases:**
- Not detected - No SQL, embedded database, ORM, migration tool, or database driver appears in `go.mod`, `go.sum`, `cmd/`, or `internal/`.
  - Connection: Not applicable.
  - Client: Not applicable.

**File Storage:**
- Local filesystem only - Source readers and setup state operate on user-local files through Go stdlib filesystem APIs.
- Claude cache file - `~/.cache/llm-quota/claude.json` stores `five_hour`, `seven_day`, and `written_at` fields; written by `internal/install/claude_hook.go` and read by `internal/sources/claude.go`.
- Claude setup state - `~/.cache/llm-quota/state.json` stores `claude_hook_declined`; read and written by `internal/install/claude_hook.go`.
- Claude settings - `~/.claude/settings.json` stores the app-owned managed `statusLine` command and marker fields; read, backed up, and atomically updated by `internal/install/claude_hook.go`.
- Claude settings backups - Changed installs and uninstalls write `settings.json.llm-quota-backup-<timestamp>` next to the Claude config path from `internal/install/claude_hook.go`.
- Codex rollout files - `~/.codex/sessions/**/rollout-*.jsonl` are scanned by `internal/sources/codex.go`; newest usable rollout event wins.

**Caching:**
- Claude cache - `internal/install/claude_hook.go` writes the app-owned Claude quota cache using temp file plus rename, and `internal/sources/claude.go` marks data stale after one hour.
- In-memory last-known-good data - `internal/tui/update.go` preserves prior successful source windows when a later refresh returns a source error.
- Codex cache - None owned by this app; Codex data is read directly from Codex-owned local rollout files in `internal/sources/codex.go`.

## Authentication & Identity

**Auth Provider:**
- None - `llm-quota` does not authenticate with external services, does not manage user identity, and does not use OAuth.
  - Implementation: Local file permissions and user-owned CLI config files only.

**Credential Handling:**
- No credential reads - `cmd/llm-quota/main.go`, `internal/install/claude_hook.go`, `internal/sources/claude.go`, and `internal/sources/codex.go` do not read `.credentials`, `.env`, Keychain, tokens, or API keys.
- No credential files detected - The repository scan found no `.env`, `*secret*`, or `*credential*` files under the shallow project tree.
- Claude setup command preserves unrelated settings - `internal/install/claude_hook.go` uses marker fields and only removes/replaces `llm-quota`-managed entries.

## Monitoring & Observability

**Error Tracking:**
- None - No Sentry, OpenTelemetry, Datadog, Honeycomb, Rollbar, or logging service appears in `go.mod` or application imports.

**Logs:**
- CLI stderr/stdout only - Command errors and setup messages are printed in `cmd/llm-quota/main.go`.
- TUI footer hints - Missing, stale, and unavailable local source states are surfaced through rendered footer hints in `internal/tui/view.go`.
- Source errors - Structured local source categories (`missing`, `malformed`, `no_usable_event`, `read_error`) are defined in `internal/sources/window.go` and merged into TUI state in `internal/tui/update.go`.
- No persistent application log file is configured in `cmd/`, `internal/`, or `README.md`.

## CI/CD & Deployment

**Hosting:**
- Not applicable - The app is a local terminal binary, not a hosted service.
- Homebrew distribution is documented in `README.md` as `brew install --HEAD robbell5/tap/llm-quota`, but this repository's `Formula/` directory is empty.
- Go install distribution is documented in `README.md` as `go install github.com/robbell5/llm-quota/cmd/llm-quota@latest`.
- Local build distribution is documented in `README.md` as `go build ./cmd/llm-quota`.

**CI Pipeline:**
- None detected - No `.github/workflows`, Makefile, CI YAML, release config, or task runner file is present in the repository.
- Local verification commands are documented in `.planning/research/STACK.md`: `go test ./...`, `go test -race ./...`, `go vet ./...`, and `go fmt ./...`.

## Environment Configuration

**Required env vars:**
- None - `cmd/llm-quota/main.go` does not require environment variables.
- Path discovery uses `os.UserHomeDir()` for `~/.claude/settings.json`, `~/.cache/llm-quota/state.json`, `~/.cache/llm-quota/claude.json`, and `~/.codex/sessions`.
- Executable discovery uses `os.Executable()` in `cmd/llm-quota/main.go` so the managed Claude statusline command can call the installed binary.

**Secrets location:**
- Not applicable - The application intentionally avoids secrets and credential stores.
- `.env` files are not present in the scanned repository.
- Claude credentials and Codex credentials are not read by application code; `.planning/PROJECT.md` and `README.md` keep network/OAuth/Keychain paths out of scope.

## Webhooks & Callbacks

**Incoming:**
- Claude statusline stdin payload - Claude invokes the managed `statusLine.command`; `cmd/llm-quota/main.go` routes `claude-statusline-cache-writer`, and `internal/install/claude_hook.go` reads stdin JSON, extracts `rate_limits`, writes the cache, and optionally passes stdin through to the previously configured statusline command.
- Legacy Claude hook stdin payload - `cmd/llm-quota/main.go` still routes `claude-hook-cache-writer`, and `internal/install/claude_hook.go` can write the same cache shape from stdin JSON, but current install logic uses managed statusline setup.
- No HTTP incoming webhooks - No HTTP server or route handler is present in `cmd/` or `internal/`.

**Outgoing:**
- Statusline passthrough command - `internal/install/claude_hook.go` optionally runs the previously configured Claude statusline command via `sh -c` after attempting cache writing.
- No outbound webhooks or API callbacks - No HTTP client, webhook sender, queue client, or network callback implementation appears in `cmd/` or `internal/`.

---

*Integration audit: 2026-05-21*
