---
phase: 02
slug: standalone-local-data-sources
status: verified
threats_open: 0
asvs_level: not provided
created: 2026-05-18
---

# Phase 02 - Security

Per-phase security contract: threat register, accepted risks, and audit trail.

## Trust Boundaries

<!-- markdownlint-disable MD013 -->

| Boundary | Description | Data Crossing |
| -------- | ----------- | ------------- |
| local cache file -> source reader | User-writable Claude cache JSON crosses into trusted normalized source state. | Local quota cache fields and timestamps. |
| local Codex rollout files -> parser | Private local JSONL session artifacts cross into normalized source state. | Local session quota events and plan metadata. |
| filesystem tree -> rollout discovery | Arbitrary local file names and modification times influence Codex candidate order. | File paths, file metadata, JSONL content. |
| llm-quota -> Claude config file | The app writes an explicitly managed hook entry to user-owned Claude configuration. | Claude settings JSON. |
| terminal consent -> file mutation | User permission controls whether hook install or update writes happen. | CLI input and installer decision state. |
| CLI args -> command dispatch | User-provided args select TUI, install, cache-writer, or error paths. | Local command-line arguments. |
| setup state -> TUI startup | Decline state suppresses repeated prompts without blocking launch. | App-owned decline state file. |
| Claude Code hook stdin -> cache writer | Private local hook payload crosses into the app-owned cache writer. | Hook JSON payload with quota limits. |
| cache writer -> Claude cache | The app writes local cache consumed later by ClaudeReader. | Normalized five_hour, seven_day, and written_at JSON. |

<!-- markdownlint-enable MD013 -->

## Threat Register

<!-- markdownlint-disable MD013 -->

| Threat ID | Category | Component | Disposition | Mitigation | Status | Evidence |
| --------- | -------- | --------- | ----------- | ---------- | ------ | -------- |
| T-02-01-01 | Tampering | `internal/sources/claude.go` | mitigate | Validate both required cache windows and reject partial cache data per D-09. | closed | `claude.go:34-35,57-72`; `claude_test.go:60-66` |
| T-02-01-02 | Information Disclosure | `internal/sources/claude.go` | mitigate | Return typed categories without embedding raw cache contents in error strings. | closed | `window.go:42-54`; `claude.go:31,35,57-90` |
| T-02-01-03 | Denial of Service | `internal/sources/claude.go` | mitigate | Convert malformed, missing, and read failures into errors; never panic on local file input. | closed | `claude.go:20-35`; `claude_test.go:50-57,128-145` |
| T-02-02-01 | Tampering | `internal/sources/codex.go` | mitigate | Require exact JSON event predicate and required primary/secondary window fields before returning data. | closed | `codex.go:122-135,170-199`; `codex_test.go:15-20,91-101` |
| T-02-02-02 | Information Disclosure | `internal/sources/codex.go` | mitigate | Do not include raw rollout payloads in returned errors. | closed | `codex.go:46,124-132`; `window.go:48-54` |
| T-02-02-03 | Denial of Service | `internal/sources/codex.go` | mitigate | Skip malformed lines and continue scanning; return categorized errors instead of panicking. | closed | `codex.go:97-117,122-132`; `codex_test.go:15-20,56-61,91-101` |
| T-02-03-01 | Tampering | `internal/install/claude_hook.go` | mitigate | Update only explicit `llm-quota` managed entries; preserve unrelated hooks. | closed | `claude_hook.go:153-165,197-212`; installer tests |
| T-02-03-02 | Repudiation | `internal/install/claude_hook.go` | mitigate | Return `InstallResult` with changed, unchanged, and backup path for user-visible reporting. | closed | `InstallResult`; CLI prints message and backup path |
| T-02-03-03 | Denial of Service | `internal/install/claude_hook.go` | mitigate | Use temp-file-plus-rename and pre-change backups to recover from interrupted writes. | closed | backup and temp-file rename in `claude_hook.go` |
| T-02-03-04 | Elevation of Privilege | `internal/install/claude_hook.go` | accept | Local user is already allowed to edit their Claude config; no privilege escalation is introduced. | closed | Accepted risk AR-02-01 |
| T-02-04-01 | Spoofing | `cmd/llm-quota/main.go` | accept | Local CLI invocation has no identity boundary; no authentication is required. | closed | Accepted risk AR-02-02 |
| T-02-04-02 | Tampering | `cmd/llm-quota/main.go` | mitigate | Dispatch only exact `install-claude-hook`; unknown args preserve explicit error path. | closed | exact dispatch and unknown-arg tests |
| T-02-04-03 | Repudiation | `cmd/llm-quota/main.go` | mitigate | Print changed/unchanged install result messages from `InstallResult`. | closed | install result messages printed in CLI |
| T-02-04-04 | Denial of Service | `internal/tui/view.go` | mitigate | Declined setup still starts TUI and renders placeholders; no source error can crash startup. | closed | decline path starts TUI; placeholder rendering tested |
| T-02-05-01 | Tampering | `internal/install/claude_hook.go` config mutation | mitigate | Preserve unrelated hooks, detect only explicit `llm-quota` ownership markers, and keep backup-before-change behavior. | closed | managed marker preservation and backups verified |
| T-02-05-02 | Tampering | `RunClaudeHookCacheWriter` cache output | mitigate | Validate both quota windows before writing and use temp-file-plus-rename via `writeJSONAtomic`. | closed | validates both windows before atomic write |
| T-02-05-03 | Information Disclosure | hook stdin parse errors | mitigate | Return concise errors without logging raw hook payload contents. | closed | concise writer errors; no raw payload logging |
| T-02-05-04 | Denial of Service | malformed hook stdin | mitigate | Fail the writer command without truncating existing cache because validation occurs before atomic rename. | closed | malformed stdin test preserves existing cache |

<!-- markdownlint-enable MD013 -->

Status: open or closed. Disposition: mitigate, accept, or transfer.

## Accepted Risks Log

<!-- markdownlint-disable MD013 -->

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
| ------- | ---------- | --------- | ----------- | ---- |
| AR-02-01 | T-02-03-04 | The app only writes files the local user can already edit; no new privilege boundary is introduced. | Phase plan disposition | 2026-05-18 |
| AR-02-02 | T-02-04-01 | Local CLI invocation has no identity boundary in this single-user terminal tool. | Phase plan disposition | 2026-05-18 |

<!-- markdownlint-enable MD013 -->

## Security Audit 2026-05-18

| Metric | Count |
| ------ | ----- |
| Threats found | 18 |
| Closed | 18 |
| Open | 0 |

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
| ---------- | ------------- | ------ | ---- | ------ |
| 2026-05-18 | 18 | 18 | 0 | gsd-security-auditor |

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-05-18
