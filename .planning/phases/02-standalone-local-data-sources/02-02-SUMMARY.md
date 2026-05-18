---
phase: 02-standalone-local-data-sources
plan: 02
subsystem: local-sources
tags: [go, codex, jsonl, local-files, testing]

requires:
  - phase: 02-standalone-local-data-sources
    provides: [normalized source contracts, SourceError categories, Window metadata]
provides:
  - Codex rollout JSONL reader with newest-usable-file selection
  - Synthetic Codex parser tests for malformed, null, fallback, and missing usable events
  - Codex plan_type preservation in normalized window metadata
affects: [phase-03-refresh, phase-04-rendering, codex-source-hints]

tech-stack:
  added: []
  patterns: [path-injected local source readers, typed source errors, mtime-ordered rollout discovery]

key-files:
  created: [internal/sources/codex.go, internal/sources/codex_test.go]
  modified: [internal/sources/claude_test.go]

key-decisions:
  - "Codex rollout discovery scans every rollout JSONL under the injected sessions root and selects by file modification time."
  - "Codex parsing skips malformed, unrelated, null, and structurally incomplete events instead of surfacing raw local payload data."
  - "Codex plan_type is carried as optional Window metadata for later footer rendering without coupling the TUI to Codex JSON."

patterns-established:
  - "Source readers accept explicit paths so tests never touch real home-directory data."
  - "Private local source parsers return typed SourceError categories and do not log or render directly."

requirements-completed: [SRC-01, SRC-02, TEST-02]

duration: 3 min
completed: 2026-05-18
---

# Phase 02 Plan 02: Local Codex Rollout Reader Summary

**Codex quota extraction from local rollout JSONL with tolerant event scanning, older-file fallback, and normalized 5h/7d windows**

## Performance

- **Duration:** 3 min
- **Started:** 2026-05-18T13:37:10Z
- **Completed:** 2026-05-18T13:40:27Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments

- Added fixture-driven Codex tests that build synthetic sessions trees, control rollout modification times, and never read real `~/.codex` data.
- Implemented `CodexReader` with recursive rollout discovery, mtime ordering, malformed/null-event tolerance, and older usable rollout fallback.
- Returned normalized Codex 5-hour and 7-day `Window` values with optional `plan_type` metadata and typed source errors.

## Task Commits

Each task was committed atomically:

1. **Task 1: Write Codex rollout reader tests** - `ddf841f` (test)
2. **Task 2: Implement Codex rollout reader** - `e8bfd14` (feat)

**Plan metadata:** committed separately after state and roadmap updates.

_Note: TDD tasks produced RED then GREEN commits._

## Files Created/Modified

- `internal/sources/codex.go` - Implements the local Codex rollout reader and JSONL parsing logic.
- `internal/sources/codex_test.go` - Covers newest usable rollout selection, null/malformed event skipping, older rollout fallback, no usable event errors, and metadata preservation.
- `internal/sources/claude_test.go` - Extends shared window assertions to compare metadata so Codex tests prove `plan_type` preservation.

## Decisions Made

- Codex rollout discovery scans every matching rollout file under the injected sessions root and sorts by `ModTime()` rather than private filename timestamp conventions.
- Malformed, unrelated, null, and structurally incomplete Codex events are skipped so local file noise does not crash or poison the source.
- `plan_type` is preserved as optional normalized metadata, keeping source-specific details out of downstream TUI parsing.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Added metadata comparison to shared window assertions**
- **Found during:** Task 2 (Implement Codex rollout reader)
- **Issue:** The shared test helper compared core window fields but not metadata, so the `plan_type` expectations would not fail if metadata was dropped.
- **Fix:** Extended the helper to compare `Window.Metadata` while keeping existing Claude tests unchanged.
- **Files modified:** `internal/sources/claude_test.go`
- **Verification:** `go test ./internal/sources -run TestCodex` and `go test ./...`
- **Committed in:** `e8bfd14`

---

**Total deviations:** 1 auto-fixed (1 missing critical)
**Impact on plan:** The auto-fix was necessary to make the planned metadata assertion meaningful. No scope creep.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Known Stubs

None.

## Next Phase Readiness

The normalized Codex reader is ready for Phase 3 refresh behavior and Phase 4 rendering. Phase 2 can proceed to Plan 02-04 to wire setup behavior and placeholder hints at the command edge.

## Self-Check: PASSED

- Verified created files exist: `internal/sources/codex.go`, `internal/sources/codex_test.go`
- Verified task commits exist: `ddf841f`, `e8bfd14`
- Verified commands passed: `go test ./internal/sources -run TestCodex`, `go test ./...`

---

*Phase: 02-standalone-local-data-sources*
*Completed: 2026-05-18*
