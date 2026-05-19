---
phase: 03-refresh-and-resilience-loop
verified: 2026-05-19T15:46:17Z
status: human_needed
score: 15/16 must-haves verified; 1 deferred to Phase 4
overrides_applied: 0
deferred:
  - truth: "User sees stale-data warnings when displayed quota data is older than the accepted freshness threshold."
    addressed_in: "Phase 4"
    evidence: "Phase 4 success criterion: 'User sees helpful placeholder rows and footer hints when source data is missing, malformed, stale, or temporarily unavailable.' Phase 3 plans explicitly limited stale handling to model/renderable data and deferred visible warning copy."
human_verification:
  - test: "Run the TUI in a real terminal/tmux pane and leave it running across a refresh interval."
    expected: "The app remains running and available quota rows refresh after the default 30-second cadence."
    why_human: "Automated tests verify Bubble Tea command scheduling, but real-time terminal behavior requires runtime observation."
  - test: "While the TUI is running in a real terminal, press r between scheduled refreshes."
    expected: "Quota data refreshes immediately and the next scheduled refresh still occurs on its cadence without visible disruption."
    why_human: "Automated tests verify command semantics and coalescing; the live keypress/user-flow needs terminal verification."
---

# Phase 3: Refresh and Resilience Loop Verification Report

**Phase Goal:** User can leave the TUI running and trust that refreshes update available data without blanking useful rows during temporary source failures.  
**Verified:** 2026-05-19T15:46:17Z  
**Status:** human_needed  
**Re-verification:** No -- initial verification

## Goal Achievement

Phase 3 is implemented at the code and test level for refresh scheduling, manual refresh, per-source last-known-good merge, typed error state, stale model state, real reader wiring, and minimal source-backed rows.

One roadmap success criterion asks for visible stale-data warnings. The actual Phase 3 plans and code intentionally keep visible stale warning copy out of Phase 3; stale-but-valid data is marked in model state and remains renderable, while warning presentation is explicitly covered by Phase 4. That item is recorded as deferred, not as an actionable Phase 3 gap.

Human verification remains required for live terminal/tmux behavior because this phase is about an always-running TUI and real-time refresh/key handling.

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | User sees quota data refresh automatically every 30 seconds while the TUI keeps running. | ✓ VERIFIED + human spot-check needed | `internal/tui/model.go:32` sets `30 * time.Second`; `internal/tui/update.go:28-30` batches immediate refresh and tick from `Init`; `internal/tui/update.go:54-55` schedules the next tick. `TestInitRequestsRefreshAndSchedulesTick` and `TestTickSchedulesRefreshAndNextTick` cover command behavior. |
| 2 | User can press `r` to refresh immediately without disrupting the next scheduled refresh. | ✓ VERIFIED + human spot-check needed | `internal/tui/update.go:34-43` returns only `requestRefreshCmd()` for `r`; scheduled ticks are only created in `Init` and `tickMsg`. `TestRefresh/manual r requests refresh when idle without scheduling tick` covers this. |
| 3 | Pressing `r` while a refresh is already running does not start duplicate source reads. | ✓ VERIFIED | `internal/tui/update.go:38-41` and `48-53` coalesce while `refreshing` is true. `TestRefresh/duplicate manual r while refreshing` and `refresh request starts one refresh and coalesces duplicate requests` verify one fetch per source. |
| 4 | User continues seeing last-known-good rows when a later refresh fails for Claude, Codex, or both. | ✓ VERIFIED | `mergeRefresh` stores errors and `continue`s without replacing existing windows on failed source results (`internal/tui/update.go:147-154`). This preserves existing windows independently for each source. Covered by `TestRefresh/last-known-good is preserved per product after later source failure`. |
| 5 | Maintainer can verify refresh merge behavior preserves last-known-good data after source failures. | ✓ VERIFIED | Focused and full test runs passed: `go test ./internal/tui -run 'TestRefresh|TestUpdateQuits|TestUpdateStoresWindowSize|TestInit|TestTick|TestRender'`; `go test ./...`. |
| 6 | Accepted windows older than one hour are marked stale but remain visible model data. | ✓ VERIFIED | `staleAfter` defaults to `time.Hour` in `internal/tui/model.go:33`; `markStale` sets `Stale`/`StaleAge` while keeping windows in the model (`internal/tui/update.go:153,158-170`). Tests cover stale flags and renderability. |
| 7 | User sees stale-data warnings when displayed quota data is older than the accepted freshness threshold. | ↪ DEFERRED | Code intentionally does **not** render visible stale copy in Phase 3 (`view_test.go:105-109`, `update_test.go:275-293`). Phase 4 explicitly covers stale placeholder/footer hints. |
| 8 | Typed source errors are stored in TUI state for missing, malformed, no-usable-event, and read failures. | ✓ VERIFIED | `Model.errors` stores `sources.SourceError`; `normalizeSourceError` preserves source errors and maps generic errors to `ErrorRead` (`internal/tui/update.go:123-137`). Tests cover all categories (`update_test.go:201-245`). |
| 9 | Temporary source failure warning state is represented in model tests while visual prominence remains deferred. | ✓ VERIFIED | Per-source errors are retained in `Model.errors` and tests assert error categories. Phase 3 rendering forbids warning/status copy; Phase 4 owns visual prominence. |
| 10 | `WindowSizeMsg` stores width/height and returns no source refresh command. | ✓ VERIFIED | `internal/tui/update.go:44-47` stores dimensions and returns nil. `TestUpdateStoresWindowSize` verifies width/height and nil command. |
| 11 | Resize is layout rerender only and does not reset, delay, or otherwise affect the 30-second refresh timer schedule. | ✓ VERIFIED | Resize handler returns nil and does not call `requestRefreshCmd` or `tickCmd` (`internal/tui/update.go:44-47`). |
| 12 | Real no-arg TUI startup wires Claude and Codex readers so Init can refresh local data. | ✓ VERIFIED | `cmd/llm-quota/main.go:204-217` constructs `sources.NewClaudeReader(paths.CachePath)` and `sources.NewCodexReader(codexSessionsRoot)`, then passes both via `tui.WithReaders`. `TestRunNoArgStartupConstructsSourceBackedModelWithoutStartingRealTUI` verifies source-backed model construction. |
| 13 | Valid stale data remains renderable model data instead of reverting to placeholders. | ✓ VERIFIED | `findWindow` ignores `Stale` and renders data rows for present windows (`internal/tui/view.go:79-82,113-121`). `TestRenderSourceBackedRowsWithoutPhaseFourCopy` asserts stale Claude data renders as `42%` and not as a placeholder. |
| 14 | If no successful refresh exists, existing placeholder rows and hints remain available. | ✓ VERIFIED | Missing windows render placeholder rows in `internal/tui/view.go:85-107`; `TestRenderStartupScreen` verifies the four rows, em dash placeholders, and missing local data hint. |
| 15 | Phase 3 does not add visible refreshing status, last-updated copy, final stale text, or visible `r` refresh hint. | ✓ VERIFIED | `view_test.go:42-44` and `105-109` assert forbidden copy is absent. `view.go` footer contains quit/setup text only. |
| 16 | Refresh commands fetch Claude and Codex concurrently using `errgroup`. | ✓ VERIFIED | `internal/tui/update.go:89-98` uses `errgroup.Group` and two `group.Go` calls for Claude and Codex source fetches. |

**Score:** 15/16 truths verified; 1 deferred to Phase 4.

### Deferred Items

Items not yet met but explicitly addressed in later milestone phases.

| # | Item | Addressed In | Evidence |
|---|------|--------------|----------|
| 1 | Visible stale-data warning copy for old displayed data | Phase 4 | Phase 4 success criterion includes helpful placeholder rows and footer hints when source data is "stale"; Phase 3 plans explicitly defer final stale warning copy/styling. |

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/tui/model.go` | Injected source readers, refresh defaults, stale threshold, per-source windows/errors, in-flight state | ✓ VERIFIED | Exists and passed `gsd-sdk query verify.artifacts`; contains `SourceReader`, `WithReaders`, `30 * time.Second`, and `time.Hour`. |
| `internal/tui/update.go` | Refresh request/tick/manual/resize/merge handling | ✓ VERIFIED | Exists and passed artifact checks; contains `refreshMsg`, `errgroup.Group`, tick handling, resize handling, stale marking, and per-source merge behavior. |
| `internal/tui/update_test.go` | Regression tests for startup, tick, manual refresh, coalescing, stale marking, LKG merge | ✓ VERIFIED | Exists and passed artifact checks; focused test suite passed. |
| `cmd/llm-quota/main.go` | Real Claude cache and Codex sessions readers wired into `tui.NewModel` | ✓ VERIFIED | Exists and passed artifact checks; constructs both real readers and passes them through `tui.WithReaders`. |
| `cmd/llm-quota/main_test.go` | Command-edge regression test proving no-real-TUI source-backed startup | ✓ VERIFIED | Exists and passed artifact checks; `TestRunNoArgStartupConstructsSourceBackedModelWithoutStartingRealTUI` captures model through injected `StartTUI`. |
| `internal/tui/view.go` | Minimal source-backed row rendering within Phase 3 copy boundaries | ✓ VERIFIED | Exists and passed artifact checks; `renderRows(m Model, ...)` reads model windows and renders percentages/reset text or placeholders. |
| `internal/tui/view_test.go` | Render regression tests for source-backed rows and forbidden copy | ✓ VERIFIED | Exists and passed artifact checks; verifies `42%`, stale-but-valid rendering, forbidden copy absence, and width guards. |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `internal/tui/update.go` | `internal/sources/window.go` | `SourceReader` returns `[]sources.Window` and `sources.SourceError` | ✓ WIRED | `gsd-sdk query verify.key-links` passed; code imports `internal/sources` and stores normalized windows/errors. |
| `internal/tui/update.go` | `golang.org/x/sync/errgroup` | Parallel Claude/Codex fetches inside refresh command | ✓ WIRED | `errgroup.Group` and two `group.Go` calls found in `refreshCmd`. |
| `cmd/llm-quota/main.go` | `internal/sources/claude.go` | Default cache path wired to `sources.NewClaudeReader` | ✓ WIRED | `sourceBackedModel` constructs Claude reader from hook cache path. |
| `cmd/llm-quota/main.go` | `internal/sources/codex.go` | Default sessions root wired to `sources.NewCodexReader` | ✓ WIRED | `sourceBackedModel` constructs Codex reader from `.codex/sessions` root. |
| `internal/tui/view.go` | `internal/tui/model.go` | Rows read normalized windows from `Model` state | ✓ WIRED | `renderRows` calls `findWindow(m, product, kind)` and renders returned windows. |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|--------------------|--------|
| `cmd/llm-quota/main.go` -> `internal/tui/model.go` | `Model.claudeReader`, `Model.codexReader` | `sources.NewClaudeReader(paths.CachePath)` and `sources.NewCodexReader(codexSessionsRoot)` | Yes -- real source reader constructors are used at runtime; tests inject temp paths/seams | ✓ FLOWING |
| `internal/tui/update.go` | `refreshMsg.results[].windows` | `SourceReader.Fetch(now)` for both readers in `refreshCmd` | Yes -- successful fetch results are merged into `Model.windows`; source errors preserve existing windows | ✓ FLOWING |
| `internal/tui/view.go` | `Model.windows[product]` | `mergeRefresh` writes normalized source windows; tests also construct model windows directly | Yes -- `findWindow` selects present normalized windows and renders percent/reset values | ✓ FLOWING |
| `internal/tui/view.go` placeholders | Missing row fallback | Empty `Model.windows` / first refresh failure | Intentional fallback, not dynamic source data | ✓ VERIFIED FALLBACK |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Focused TUI refresh/update/render behavior | `go test ./internal/tui -run 'TestRefresh|TestUpdateQuits|TestUpdateStoresWindowSize|TestInit|TestTick|TestRender'` | `ok github.com/rob/llm-quota/internal/tui 0.316s` | ✓ PASS |
| Command-edge source-backed startup behavior | `go test ./cmd/llm-quota -run 'TestRun|TestStart|TestInstall|TestClaudeHook'` | `ok github.com/rob/llm-quota/cmd/llm-quota 0.225s` | ✓ PASS |
| Full repository test suite | `go test ./...` | All packages passed | ✓ PASS |
| Live terminal cadence/key behavior | Not run automatically | Requires running foreground TUI and observing timer/key behavior | ? HUMAN |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|----------|
| SRC-04 | 03-01, 03-02 | User continues seeing last-known-good rows when a later refresh fails for one source. | ✓ SATISFIED | `mergeRefresh` preserves existing windows on per-source error; render reads existing windows. Test covers Claude failure while Codex updates. Logic also preserves both products if both results are errors. |
| SRC-05 | 03-01, 03-02 | User sees stale-data warnings when displayed quota data is older than freshness threshold. | ↪ PARTIAL/DEFERRED | Stale model state and stale-but-valid rendering are implemented and tested. Visible warning copy/styling is intentionally absent in Phase 3 and explicitly covered by Phase 4. |
| TUI-02 | 03-01, 03-02 | User sees quota data refresh automatically every 30 seconds. | ✓ SATISFIED + human runtime check | Default interval is 30 seconds; `Init` schedules immediate refresh and tick; tick schedules next tick and refresh. Runtime terminal observation remains required. |
| TUI-03 | 03-01, 03-02 | User can press `r` to refresh immediately without disrupting the next scheduled refresh. | ✓ SATISFIED + human runtime check | `r` requests refresh without tick creation/cancellation; duplicate refresh requests coalesce. Runtime terminal observation remains required. |
| TEST-03 | 03-01 | Maintainer can verify refresh merge behavior preserves last-known-good data after source failures. | ✓ SATISFIED | `TestRefresh/last-known-good is preserved per product after later source failure` exists and passed. |

No additional Phase 3 requirement IDs were found in `.planning/REQUIREMENTS.md` beyond SRC-04, SRC-05, TUI-02, TUI-03, and TEST-03.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| `internal/tui/view.go` | 63 | Slice literal matched generic hardcoded-empty scan | ℹ️ Info | Legitimate fixed row metadata for the four quota rows; not a stub. |
| `internal/tui/update.go` | 84 | Slice literal matched generic hardcoded-empty scan | ℹ️ Info | Legitimate fixed result slots for Claude/Codex concurrent fetches; not a stub. |
| `*_test.go` files | multiple | Fixture slice literals / placeholder wording in assertions | ℹ️ Info | Test fixtures and expected placeholder fallback assertions; not production stubs. |

No blocker TODO/FIXME/placeholder implementation, empty handler, ignored fetch result, or hardcoded user-visible empty data was found in the Phase 3 modified production files.

### Human Verification Required

#### 1. Live 30-second refresh cadence

**Test:** Run `llm-quota` in a real terminal/tmux pane and leave it running across at least one refresh interval.  
**Expected:** The TUI stays active and refreshes available quota rows after the default 30-second cadence.  
**Why human:** Tests verify the Bubble Tea command/tick contract without sleeping; actual terminal timing requires observation.

#### 2. Live manual `r` refresh behavior

**Test:** While the TUI is running, press `r` between scheduled refreshes.  
**Expected:** Data refreshes immediately, duplicate in-flight reads are not started, and the next scheduled refresh still happens normally.  
**Why human:** Tests verify message semantics and coalescing, but the user-facing keypress flow requires terminal observation.

### Gaps Summary

No actionable Phase 3 code gaps were found.

The only unmet roadmap truth is visible stale-data warning copy. This is not implemented in Phase 3 and is intentionally deferred by the Phase 3 plans to Phase 4, whose success criteria explicitly cover stale placeholder/footer hints.

---

_Verified: 2026-05-19T15:46:17Z_  
_Verifier: the agent (gsd-verifier)_
