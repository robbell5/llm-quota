# Codebase Concerns

**Analysis Date:** 2026-05-21

## Tech Debt

**Private local source formats are treated as product contracts:**
- Issue: Claude quota data depends on the observed Claude statusline stdin shape and Codex quota data depends on observed `rollout-*.jsonl` `token_count` events. These are private local files/events, not stable public APIs.
- Files: `internal/install/claude_hook.go`, `internal/sources/claude.go`, `internal/sources/codex.go`, `internal/sources/claude_test.go`, `internal/sources/codex_test.go`
- Impact: Upstream Claude or Codex format drift can silently degrade the TUI to placeholders or stale last-known-good values even when quota data exists locally.
- Fix approach: Keep parser logic isolated in `internal/sources/` and `internal/install/`; add synthetic fixtures for each newly observed shape before broadening parsers; preserve typed `SourceError` categories so `internal/tui/view.go` can render actionable hints instead of raw parser errors.

**Installer/config mutation is concentrated in large files:**
- Issue: `internal/install/claude_hook.go` owns config parsing, statusline wrapping, old hook cleanup, cache writing, shell quoting, backups, symlink handling, and atomic writes in one 472-line file; `cmd/llm-quota/main.go` owns CLI routing, first-launch prompts, path discovery, source wiring, and hook-installed detection in one 386-line file.
- Files: `internal/install/claude_hook.go`, `cmd/llm-quota/main.go`, `internal/install/claude_hook_test.go`, `cmd/llm-quota/main_test.go`
- Impact: Future installer changes can accidentally affect unrelated behavior such as uninstall restoration, symlink preservation, or first-launch TUI startup.
- Fix approach: Keep new install behavior covered first in `internal/install/claude_hook_test.go`; if these files grow, split by responsibility into config read/write, statusline command construction, cache writer, and CLI command handling without changing public behavior.

**Release/install evidence is split between docs and local repo state:**
- Issue: `README.md` documents `brew install --HEAD robbell5/tap/llm-quota`, while the local `Formula/` directory is empty and no `.github/workflows/` CI/release pipeline is detected in this repo.
- Files: `README.md`, `Formula/`, `.github/workflows/`, `go.mod`
- Impact: The documented Homebrew path cannot be validated from this repository alone, and contributors have no committed CI signal that `go test ./...`, `go vet ./...`, or packaging checks run before release.
- Fix approach: Add or link the actual tap/release evidence from `README.md`; add a minimal CI workflow under `.github/workflows/` that runs `go test ./...` and `go vet ./...`; add release/tap checks where the package is maintained.

**Ignored local build artifact is present in the working tree:**
- Issue: A built executable exists at repo root and is ignored by `.gitignore`.
- Files: `llm-quota`, `.gitignore`, `README.md`
- Impact: Local smoke tests are convenient, but stale local binaries can diverge from source and confuse manual validation if the binary is run without rebuilding.
- Fix approach: Continue ignoring `llm-quota`; before manual UAT, rebuild with `go build ./cmd/llm-quota` and run `./llm-quota` only as the fresh local artifact documented in `README.md`.

## Known Bugs

**Current refresh errors are hidden when last-known-good rows still exist and are not stale:**
- Symptoms: After a successful source refresh, a later source failure preserves the prior rows in `Model.windows` and stores the error in `Model.errors`, but `footerRecoveryHints` only shows missing-source or stale-source hints. A fresh-looking last-known-good row can remain visible with no immediate indication that the latest refresh failed.
- Files: `internal/tui/update.go`, `internal/tui/view.go`, `internal/tui/update_test.go`, `internal/tui/view_test.go`, `.planning/milestones/v1.0-MILESTONE-AUDIT.md`
- Trigger: Fetch valid Claude or Codex windows, then make the same source unreadable/malformed before the stale threshold elapses.
- Workaround: Wait until data becomes stale and the stale footer hint appears, open the affected local tool, or press `r` after fixing the local source.

**README scope contradicts the implemented Claude statusline setup:**
- Symptoms: `README.md` explains that setup registers a Claude `statusLine.command`, but the final scope paragraph also says v1 does not use statusline integration.
- Files: `README.md`, `internal/install/claude_hook.go`, `internal/install/claude_hook_test.go`
- Trigger: A user reads the setup section and scope section together.
- Workaround: Treat the setup section as authoritative: Claude quota capture is implemented by wrapping Claude `statusLine.command`; the TUI does not provide a separate Claude statusline display.

**Codex rows do not reflect OpenCode/OpenAI usage stored outside Codex rollout files:**
- Symptoms: GPT/OpenAI usage through OpenCode does not update Codex quota rows unless Codex-style rollout `token_count` events with non-null `rate_limits` are written under `~/.codex/sessions`.
- Files: `internal/sources/codex.go`, `README.md`, `.planning/debug/codex-usage-limits-not-updat.md`
- Trigger: Use OpenCode or another OpenAI client instead of Codex CLI, then expect the Codex quota rows to move.
- Workaround: Open Codex locally so Codex writes rollout quota data; v1 has no OpenCode database reader and no network quota fallback.

## Security Considerations

**Wrapped Claude statusline passthrough executes through the shell:**
- Risk: `RunClaudeStatusLineCacheWriter` runs the preserved passthrough command with `exec.Command("sh", "-c", passthrough)`. The passthrough comes from the user's existing Claude config, but any future change that builds this value from untrusted input would become command-injection risk.
- Files: `internal/install/claude_hook.go`, `internal/install/claude_hook_test.go`, `cmd/llm-quota/main.go`
- Current mitigation: App-generated executable and cache paths are shell-quoted by `shellQuote`; tests cover paths with spaces/apostrophes and passthrough preservation in `internal/install/claude_hook_test.go`.
- Recommendations: Keep passthrough sourced only from existing user-owned Claude config; never interpolate hook input or quota JSON into shell commands; prefer argv-based execution if Claude statusline config later supports structured command arguments.

**Local private source data must stay out of logs and fixtures:**
- Risk: Claude cache payloads and Codex rollout files can contain local session-adjacent metadata. Copying real files into tests or logging raw parse input can expose private usage/session data.
- Files: `internal/sources/claude.go`, `internal/sources/codex.go`, `internal/install/claude_hook.go`, `internal/sources/claude_test.go`, `internal/sources/codex_test.go`, `README.md`
- Current mitigation: Source readers return typed errors without printing raw file contents; tests use synthetic payloads; `README.md` states the TUI does not print private Claude or Codex payloads.
- Recommendations: Keep all new parser fixtures synthetic; log only source, category, and concise recovery hints; do not add debug output that includes raw JSON/JSONL lines.

**Symlinked Claude settings are followed for writes:**
- Risk: `resolveWritePath` preserves a symlinked `~/.claude/settings.json` by writing through to its target. This supports dotfiles-managed configs, but a malicious or unexpected symlink target would be modified by install/uninstall.
- Files: `internal/install/claude_hook.go`, `internal/install/claude_hook_test.go`
- Current mitigation: The installer runs as the current user, backs up existing configs, and has regression tests for symlink preservation.
- Recommendations: Do not run `llm-quota install-claude-hook` with elevated privileges; consider warning when the resolved target is outside the user's home directory if install scope broadens beyond a personal tool.

## Performance Bottlenecks

**Codex refresh walks and sorts the full sessions tree every 30 seconds:**
- Problem: `CodexReader.Fetch` recursively scans `sessionsRoot`, collects every `rollout-*.jsonl` candidate, sorts by modification time, and then reads candidates until one has usable rate limits.
- Files: `internal/sources/codex.go`, `internal/tui/update.go`, `internal/tui/model.go`
- Cause: The reader is intentionally stateless and simple; refresh uses the same full discovery path for periodic and manual refreshes.
- Improvement path: Keep the current implementation for small trees; add bounded discovery, recent-directory pruning, or a remembered newest usable path if `~/.codex/sessions` grows large enough to make refresh visibly slow.

**Selected rollout files are read entirely into memory:**
- Problem: `windowsFromCodexRollout` uses `os.ReadFile` and splits the full JSONL file into strings.
- Files: `internal/sources/codex.go`, `internal/sources/codex_test.go`, `AGENTS.md`, `.planning/research/PITFALLS.md`
- Cause: Avoiding default `bufio.Scanner` limits is deliberate, and current rollout files are expected to be small.
- Improvement path: If rollout files become large, parse from the end of the file or stream with an explicitly increased buffer while preserving the "last usable event wins" behavior.

## Fragile Areas

**Claude config install/uninstall is high impact:**
- Files: `internal/install/claude_hook.go`, `internal/install/claude_hook_test.go`, `cmd/llm-quota/main.go`, `cmd/llm-quota/main_test.go`
- Why fragile: Install/uninstall modifies a user-owned config file, wraps a pre-existing statusline command, removes old app-owned hook entries, preserves unrelated config, writes backups, and supports symlinked settings files.
- Safe modification: Add tests before changing installer behavior; verify idempotent install, uninstall restoration, markerless hook preservation, backup creation, symlink preservation, and quoted path handling.
- Test coverage: Strong unit coverage exists in `internal/install/claude_hook_test.go` and `cmd/llm-quota/main_test.go`; no committed CI workflow enforces those tests.

**Width-sensitive rendering can regress with small text changes:**
- Files: `internal/tui/view.go`, `internal/tui/view_test.go`, `internal/tui/colors.go`
- Why fragile: Row layout budgets depend on Lip Gloss cell widths, Bubbles progress output, labels, reset text, footer hints, and terminal widths around 50, 49, 30, 29, and 20 columns.
- Safe modification: Update render tests with every label, hint, or progress-bar change; keep `assertRenderedLineWidths` coverage for the narrow pane widths the product targets.
- Test coverage: `internal/tui/view_test.go` covers normal, compact, narrow, missing, stale, and threshold states; tests are not golden-file snapshots and may miss subtle visual polish changes outside string/width assertions.

**Refresh state relies on Bubble Tea message discipline:**
- Files: `internal/tui/update.go`, `internal/tui/update_test.go`, `internal/tui/model.go`
- Why fragile: Source fetches run concurrently in a command, but model state must only change when a `refreshMsg` returns through `Update`.
- Safe modification: Keep readers side-effect-free except local file reads; never mutate `Model` from goroutines; preserve duplicate refresh coalescing when adding more refresh triggers.
- Test coverage: `internal/tui/update_test.go` covers coalescing, tick scheduling, stale marking, source errors, and last-known-good preservation; `go test -race ./...` is recommended but not automated in repo metadata.

## Scaling Limits

**Provider/window set is hard-coded to four rows:**
- Current capacity: The TUI displays Claude 5-hour, Claude 7-day, Codex 5-hour, and Codex 7-day rows.
- Limit: Adding more providers or plan windows requires touching source models, hard-coded row labels, rendering budgets, footer hints, and tests.
- Scaling path: Keep `sources.Window` as the shared model, but introduce a declarative row list/config only when a fifth row/provider is actually needed.

**Codex session history is unbounded from the app's perspective:**
- Current capacity: Refresh remains cheap while `~/.codex/sessions` contains a modest number of rollout files.
- Limit: Very large session trees can make every 30-second refresh spend noticeable time walking and sorting stale files.
- Scaling path: Cache candidate metadata in `CodexReader`, prune by date/modtime, or read only recent Codex session directories after measuring real-world latency.

## Dependencies at Risk

**Claude and Codex private file/event shapes:**
- Risk: The app depends on Claude statusline `rate_limits` input and Codex rollout `payload.rate_limits` fields remaining compatible.
- Impact: Quota rows can become missing or stale without a compile-time failure.
- Migration plan: Treat parser changes as compatibility updates in `internal/sources/` and `internal/install/`; add fixtures before changing behavior; keep user-visible fallback hints in `internal/tui/view.go`.

**Charm v2 APIs are pinned but central to the TUI shape:**
- Risk: `charm.land/bubbletea/v2`, `charm.land/bubbles/v2`, and `charm.land/lipgloss/v2` drive `tea.View`, message types, progress rendering, and width measurement.
- Impact: Dependency upgrades can affect rendering widths, key messages, or alt-screen behavior.
- Migration plan: Run `go test ./...` and `go test -race ./...` after upgrades; specifically inspect `internal/tui/view_test.go` width assertions and `internal/tui/update_test.go` message assertions.

## Missing Critical Features

**Repository-local CI is absent:**
- Problem: No `.github/workflows/` files are detected, so test/vet/race gates are not encoded in this repository.
- Blocks: Contributors cannot rely on committed automation to catch parser, installer, or rendering regressions before merge.

**Immediate current-error visibility is missing for non-stale last-known-good data:**
- Problem: The model stores latest source errors while preserving last-known-good rows, but the footer does not surface current errors until rows are missing or stale.
- Blocks: A user can mistake preserved data for freshly refreshed data during the first hour after a source starts failing.

**No bounded Codex scan policy exists:**
- Problem: Codex discovery scans all nested sessions every refresh.
- Blocks: Very large local Codex histories can eventually require a performance fix before the always-running pane remains unobtrusive.

## Test Coverage Gaps

**Current-error rendering with last-known-good rows:**
- What's not tested: A render case where `Model.windows` has current-looking rows and `Model.errors` contains a later source failure.
- Files: `internal/tui/view.go`, `internal/tui/view_test.go`, `internal/tui/update_test.go`
- Risk: Refresh failures can stay invisible until stale state, preserving the known bug.
- Priority: High

**Codex filesystem edge cases and large-file behavior:**
- What's not tested: Missing root, unreadable root/file handling, very large JSONL files, equal modification times, and a newest candidate whose validation errors are ignored in favor of older usable data.
- Files: `internal/sources/codex.go`, `internal/sources/codex_test.go`
- Risk: Local filesystem or upstream rollout changes can turn into placeholders or slower refreshes without targeted regression tests.
- Priority: Medium

**Race detector and packaging checks are manual:**
- What's not tested: `go test -race ./...`, Homebrew install path validation, and release/tap checks are not enforced by repository-local CI.
- Files: `go.mod`, `README.md`, `.github/workflows/`, `Formula/`
- Risk: Concurrency, packaging, or documented install regressions can pass local unit tests if maintainers do not run the extra checks manually.
- Priority: Medium

---

*Concerns audit: 2026-05-21*
