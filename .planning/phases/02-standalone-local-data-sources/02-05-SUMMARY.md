---
phase: 02-standalone-local-data-sources
plan: 05
subsystem: claude-hook-cache-writer
tags: [go, cli, claude, hooks, local-data]

requires:
  - phase: 02-standalone-local-data-sources
    provides: [Claude hook installer, Claude source contract]
provides:
  - Runnable Claude Code command-hook entry for the managed llm-quota hook
  - Hook-internal cache writer command that emits ClaudeReader-compatible JSON
  - First-launch upgrade path for old managed top-level-command hook entries
  - Regression coverage for malformed hook input preserving existing cache
affects: [phase-03-refresh-loop, phase-04-rendering, phase-05-install-docs]

tech-stack:
  added: []
  patterns:
    - Nested Claude command-hook settings entry with app-owned marker
    - Atomic cache writer that validates hook stdin before replacing cache files

key-files:
  created: []
  modified:
    - internal/install/claude_hook.go
    - internal/install/claude_hook_test.go
    - cmd/llm-quota/main.go
    - cmd/llm-quota/main_test.go

key-decisions:
  - "The installed Claude hook invokes llm-quota claude-hook-cache-writer --cache <path> instead of copying raw stdin."
  - "The hook-internal writer command is the only new CLI dispatch path; broader setup aliases remain out of scope."
  - "Malformed or trailing hook stdin is rejected before atomic cache replacement."

patterns-established:
  - "Install tests assert nested command-hook shape and shell-quoted cache paths."
  - "Cache writer tests prove generated files are readable through sources.NewClaudeReader."
  - "CLI tests exercise hook-internal dispatch without starting Bubble Tea."

requirements-completed: [CLD-03, CLD-04, TEST-01]

duration: 1 session
completed: 2026-05-18
---

# Phase 02 Plan 05: Claude Hook Gap Closure Summary

**Runnable managed Claude hook plus an atomic cache writer for ClaudeReader data**

## Performance

- **Duration:** 1 session
- **Completed:** 2026-05-18
- **Tasks:** 3 completed
- **Files modified:** 4

## Accomplishments

- Updated the managed Claude hook entry to the nested command-hook shape Claude Code can execute.
- Replaced the old raw `cat >` cache command with `llm-quota claude-hook-cache-writer --cache <path>`.
- Added `RunClaudeHookCacheWriter` to convert Claude hook stdin into `five_hour`, `seven_day`, and `written_at` cache JSON.
- Added CLI dispatch for the hook-internal cache writer without starting the TUI.
- Tightened first-launch detection so old managed hook entries are offered an upgrade.
- Added regression coverage that malformed trailing stdin does not overwrite an existing cache.

## Files Created/Modified

- `internal/install/claude_hook.go` - Emits nested command-hook settings and writes validated cache JSON atomically.
- `internal/install/claude_hook_test.go` - Covers hook shape, old managed-entry replacement, reader-compatible cache output, and malformed-input safety.
- `cmd/llm-quota/main.go` - Dispatches `claude-hook-cache-writer --cache <path>` with stdin and current time.
- `cmd/llm-quota/main_test.go` - Covers cache-writer command success, invalid argument handling, and no-TUI startup behavior.

## Decisions Made

- Kept app-owned hook detection compatible with old managed entries by preserving `name` and `llm_quota_marker` matching.
- Treated old app-owned top-level `command` hooks as not-current during first-launch detection so setup can upgrade them.
- Required both Claude rate-limit windows before writing cache data so partial hook payloads cannot replace usable data.
- Supported both top-level `rate_limits` and `payload.rate_limits` hook payload shapes.
- Rejected trailing JSON values or trailing malformed data before writing the cache.

## Deviations from Plan

None - the plan scope was executed without adding broader setup aliases.

## Issues Encountered

- Code review found that `json.Decoder.Decode` accepted one valid JSON object followed by trailing malformed data. Added a failing regression test and fixed the writer to reject trailing data before writing.

## Known Stubs

None for this plan.

## Verification

- `go test ./internal/install ./cmd/llm-quota -run 'TestInstallClaudeHook|TestClaudeHookCacheWriter|TestRunClaudeHookCacheWriter'` - PASS
- `go test ./internal/install ./cmd/llm-quota` - PASS
- `go test ./...` - PASS
- `grep "cat >" internal/install/*.go` - PASS, no matches
- `grep "claude-hook-cache-writer"` - PASS, found installer command and CLI dispatch coverage

## User Setup Required

None.

## Next Phase Readiness

Phase 2 Claude hook setup now produces a runnable local cache producer. Later refresh/render phases can rely on `ClaudeReader` receiving the normalized cache contract.

## Self-Check: PASSED

- Verified focused hook tests and full Go suite pass.
- Verified the direct raw cache write command is absent.
- Verified implementation and tests cover both failed Phase 2 truths.

---

*Phase: 02-standalone-local-data-sources*
*Completed: 2026-05-18*
