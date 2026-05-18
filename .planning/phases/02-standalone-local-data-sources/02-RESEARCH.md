# Phase 2: Standalone Local Data Sources - Research

**Researched:** 2026-05-18
**Status:** Ready for planning
**Confidence:** HIGH for Go package boundaries and fixture strategy; MEDIUM for private Claude/Codex payload stability.

## Executive Summary

Phase 2 should add three small, testable seams without changing the Phase 1 TUI runtime model more than necessary:

1. `internal/sources` defines normalized quota windows, typed source errors, and local-only Claude/Codex readers.
2. `internal/install` owns the Claude hook/cache-writer installation policy, including first-launch decline memory and idempotent explicit install.
3. `cmd/llm-quota/main.go` remains the thin edge that dispatches `install-claude-hook`, runs a pre-alt-screen first-launch prompt, wires real default paths, and starts `tea.NewProgram` only after setup decisions are complete.

Do not add network fallback, OAuth, Keychain reads, fsnotify, a broad CLI framework, or statusline runtime coupling.

## Requirements and Decision Map

| Source | Items | Planning Implication |
|--------|-------|----------------------|
| ROADMAP | Phase 2 goal and success criteria | Plans must prove permissioned setup, safe hook install/update, Claude cache parsing, Codex rollout parsing, and fixture coverage. |
| REQUIREMENTS.md | CLD-01, CLD-02, CLD-03, CLD-04, SRC-01, SRC-02, SRC-03, TEST-01, TEST-02 | Every plan must list the relevant IDs in frontmatter. |
| CONTEXT.md | D-01 through D-16 | Task actions must reference the decision IDs they implement. |
| Phase 1 summaries | Existing TUI spine and render patterns | Preserve Charm v2 APIs, existing quit/resize behavior, width-safe render tests, and thin `main.go` style. |

## Recommended Architecture

### `internal/sources`

Create a normalized source layer that is independent of Bubble Tea:

- `internal/sources/window.go`
  - `type Product string` with `ProductClaude` and `ProductCodex`.
  - `type WindowKind string` with `WindowFiveHour` and `WindowSevenDay`.
  - `type Window struct` carrying product, kind, label, used percent, reset time, captured time, stale flag/age, and optional metadata.
  - `type SourceError` or equivalent typed error with categories: `missing`, `malformed`, `no_usable_event`, and `read_error`.
- `internal/sources/claude.go`
  - `NewClaudeReader(cachePath string)`.
  - `Fetch(now time.Time) ([]Window, error)`.
  - Read exactly one cache JSON file.
  - Treat `five_hour`, `seven_day`, and `written_at` as the all-or-nothing contract per D-09.
  - Return windows plus stale metadata for old valid caches per D-11.
- `internal/sources/codex.go`
  - `NewCodexReader(sessionsRoot string)`.
  - `Fetch(now time.Time) ([]Window, error)`.
  - Walk all rollout JSONL files under the sessions root per D-14.
  - Sort by modification time descending per D-15.
  - For each file, read all lines, skip malformed/unrelated/null-limit lines per D-10, and return the newest file containing a usable token-count event per D-13.
  - Preserve `plan_type` as optional metadata per D-16.

### `internal/install`

Create a narrowly scoped installer package:

- `internal/install/claude_hook.go`
  - Manage only Claude hook entries with an explicit `llm-quota` marker/name per D-05.
  - Preserve unrelated hook entries when the JSON shape is understood per D-06.
  - Create a timestamped backup only when a write is needed per D-07.
  - Update existing managed entries in place and report changed vs unchanged per D-08.
  - Store first-launch decline state so normal launches do not repeatedly prompt per D-02.

The installer should accept explicit config/state/cache paths so tests use `t.TempDir()` and never touch real `~/.claude` or `~/.cache`.

### `cmd/llm-quota/main.go`

Keep `main.go` as the only user-facing command edge:

- No args: run the first-launch setup check before `tea.NewProgram(...).Run()` per D-03.
- `install-claude-hook`: run explicit install/update and exit without starting the TUI per D-01.
- Any other arg: preserve the current plain unknown-argument failure behavior.

## Data Contracts

### Claude Cache

Expected file: `~/.cache/llm-quota/claude.json`

```json
{
  "five_hour": {"used_percentage": 42.3, "resets_at": 1778942485},
  "seven_day": {"used_percentage": 85.7, "resets_at": 1779382265},
  "written_at": 1778940000
}
```

Parsing rules:

- `resets_at` and `written_at` are Unix seconds.
- `used_percentage` must be numeric.
- Missing either Claude window rejects the whole Claude source.
- Old-but-valid data returns windows with stale metadata rather than an error.

### Codex Rollout JSONL

Search root: `~/.codex/sessions`

Usable line predicate:

- `type == "event_msg"`
- `payload.type == "token_count"`
- `payload.rate_limits != null`

Rate-limit shape:

```json
{
  "limit_id": "codex",
  "primary": {"used_percent": 40.0, "window_minutes": 300, "resets_at": 1778942485},
  "secondary": {"used_percent": 18.0, "window_minutes": 10080, "resets_at": 1779382265},
  "plan_type": "prolite"
}
```

Parsing rules:

- `primary` maps to Codex 5-hour.
- `secondary` maps to Codex 7-day.
- Malformed individual lines do not poison the whole scan when another usable event exists.
- A newest rollout file with no usable event should fall back to older rollout files ordered by modification time.

## Testing Strategy

Use colocated Go tests and synthetic fixtures only.

| File | Required Tests |
|------|----------------|
| `internal/sources/claude_test.go` | valid cache returns two windows; missing file returns `missing`; malformed JSON returns `malformed`; missing one Claude window rejects all; old valid cache returns windows with stale metadata. |
| `internal/sources/codex_test.go` | newest usable rollout returns two windows; null `rate_limits` lines are skipped; malformed lines are skipped; newest unusable file falls back to older usable file; no usable event returns `no_usable_event`; `plan_type` is preserved in metadata. |
| `internal/install/claude_hook_test.go` | explicit managed marker is required for ownership; unrelated config entries are preserved; backup is created only when writing; existing managed hook updates idempotently; decline state suppresses repeated first-launch prompts. |
| `cmd/llm-quota/main_test.go` or install package tests | `install-claude-hook` dispatch does not enter the TUI; unknown arguments still exit plainly. |

Verification commands should include:

```bash
go test ./internal/sources ./internal/install ./cmd/llm-quota
go test ./...
```

## Security and Safety Considerations

Trust boundaries for Phase 2 are local but still important:

- Claude configuration JSON is user-owned local state. Only mutate managed entries, preserve unrelated entries, and backup before writes.
- Claude hook input is generated by Claude Code. The hook/cache writer must parse defensively and write only the small cache contract.
- Codex rollout files are private local session artifacts. Tests must never copy real session data, and parsing must avoid exposing payload contents in errors.
- Terminal prompts happen before alt-screen startup so consent is explicit and visible.

## Architectural Responsibility Map

| Concern | Owns It | Must Not Own It |
|---------|---------|-----------------|
| Real default home/cache/session paths | `cmd/llm-quota/main.go` | Tests, source readers as globals |
| Claude hook install/update policy | `internal/install` | `internal/tui`, `internal/sources` |
| Claude/Codex JSON parsing | `internal/sources` | `internal/tui`, `cmd/llm-quota/main.go` |
| Placeholder row rendering and hints | `internal/tui` | Source readers |
| Bubble Tea program lifecycle | `cmd/llm-quota/main.go`, `internal/tui` | Install/source parsing packages |

## Common Pitfalls

- Do not infer hook ownership from command text or quota behavior. Require an explicit managed marker/name.
- Do not prompt after entering alt-screen. Consent prompts must complete first.
- Do not read real `~/.claude`, `~/.codex`, or cache files in tests.
- Do not treat a stale Claude cache as a hard error.
- Do not return partial Claude source data if one of the two required windows is invalid.
- Do not stop Codex parsing at the newest file when it contains no usable event.
- Do not add broader `install`, `setup`, or help aliases in this phase.

## Discovery Level

Level 1: quick verification against existing project research and Phase 1 code patterns. No new external libraries are needed.
