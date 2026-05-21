---
status: resolved
trigger: "codex usage limits not updating when using GPT models via OpenCode (same subscription)"
created: 2026-05-21T21:36:17Z
updated: 2026-05-21T21:55:00Z
---

# Debug Session: codex-usage-limits-not-updat

## Symptoms

- Expected behavior: Codex 5-hour and/or 7-day usage rows should increase after GPT usage through OpenCode on the same Codex subscription.
- Actual behavior: Codex usage limits do not change after GPT usage through OpenCode.
- Error messages: No visible errors; values simply do not update.
- Timeline: Unknown or newly noticed while using GPT models through OpenCode.
- Reproduction: Run GPT-model requests through OpenCode, then refresh or wait for llm-quota.

## Current Focus

- hypothesis: llm-quota only reads Codex CLI rollout `rate_limits`; OpenCode GPT usage is stored separately as token/cost history and does not include quota window percent/reset data.
- test: confirmed through source trace and OpenCode metadata-only DB schema/key queries
- expecting: no code-level quota update is possible from OpenCode local state without a data source that records subscription rate-limit windows
- next_action: no code fix applied; keep session as resolved root-cause with documented limitation
- reasoning_checkpoint: complete
- tdd_checkpoint: not started

## Evidence

- timestamp: 2026-05-21T21:40:00Z
  observation: Runtime construction in `cmd/llm-quota/main.go` wires Codex data exclusively through `sources.NewCodexReader(codexSessionsRoot)`, and `defaultCodexSessionsRoot()` resolves only `~/.codex/sessions`.
  supports: Codex quota rows are sourced from Codex CLI session rollout files, not OpenCode local state.
- timestamp: 2026-05-21T21:42:00Z
  observation: `internal/sources/codex.go` recursively scans files named `rollout-*.jsonl` and only accepts JSONL events where `type == "event_msg"`, `payload.type == "token_count"`, and `payload.rate_limits` is non-null.
  supports: GPT usage through OpenCode cannot affect displayed Codex percentages unless OpenCode also writes Codex-style rollout token-count events with non-null `rate_limits` under `~/.codex/sessions`.
- timestamp: 2026-05-21T21:46:00Z
  observation: `opencode debug paths` reports OpenCode stores data under `~/.local/share/opencode`; `opencode db path` reports `~/.local/share/opencode/opencode.db`, separate from `~/.codex/sessions`.
  supports: OpenCode GPT usage is persisted outside the tree scanned by `CodexReader`.
- timestamp: 2026-05-21T21:48:00Z
  observation: Metadata-only OpenCode DB inspection showed OpenCode session messages for provider `openai` and model `gpt-5.5`; part data contains `tokens` and `cost` fields, while provider metadata keys are limited to OpenAI item/reasoning/phase identifiers and do not include `rate_limits`, reset times, or subscription window percentages.
  supports: OpenCode records usage history, but not the quota window data shape required by `llm-quota`'s Codex rows.
- timestamp: 2026-05-21T21:51:00Z
  observation: README and design spec both document Codex quota data as coming from `~/.codex/sessions` rollout JSONL and instruct users to open Codex locally for fresh Codex data.
  supports: Current implementation matches the documented v1 local-only Codex data source, but that source does not cover OpenCode usage.
- timestamp: 2026-05-21T21:53:00Z
  observation: `go test ./...` passed for all packages.
  supports: No regression in the existing Codex source behavior while investigating.

## Eliminated

- OpenCode GPT usage is being written into `~/.codex/sessions` but skipped by parsing: eliminated. The OpenCode data path is a separate SQLite database under `~/.local/share/opencode`, not Codex rollout JSONL.
- `CodexReader` is reading OpenCode data but expecting the wrong JSON shape: eliminated. `CodexReader` has no OpenCode DB path or SQL access; it only walks rollout JSONL files.
- Existing source tests are failing due to a parser regression: eliminated. `go test ./...` passes.

## Resolution

- root_cause: The Codex rows update only from Codex CLI rollout `rate_limits` under `~/.codex/sessions`; OpenCode GPT usage is stored in OpenCode's separate SQLite database as token/cost history and does not persist subscription quota window percentages or reset timestamps.
- fix: No code fix applied because the local OpenCode data available on disk does not contain the quota-window data needed to correctly update Codex 5-hour/7-day subscription percentages. Adding token totals would be an estimate, not the same metric.
- verification: Source trace of `cmd/llm-quota/main.go` and `internal/sources/codex.go`; metadata-only OpenCode DB schema/key queries; `go test ./...` passed.
- files_changed: `.planning/debug/codex-usage-limits-not-updat.md`
