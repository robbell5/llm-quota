# Project Research Summary

**Project:** llm-quota
**Domain:** Tiny local-file-backed Go/Bubble Tea terminal quota dashboard
**Researched:** 2026-05-16
**Confidence:** HIGH

## Executive Summary

`llm-quota` is a deliberately small always-running terminal dashboard for a
dedicated tmux pane. Expert implementations of this kind of tool are boring on
purpose: a foreground event loop, fixed rows, local file reads, deterministic
rendering, terse keyboard controls, and resilient failure states. The product's
value is not breadth; it is letting Rob glance at one pane and compare Claude
Code and Codex 5-hour and 7-day quota windows immediately.

The recommended approach is a Go 1.26.3 module using Bubble Tea v2, Bubbles
progress, Lip Gloss, stdlib JSON/file handling, and a small `errgroup`-based
refresh command. Keep three hard boundaries: source readers normalize unofficial
local files into shared quota windows, the Bubble Tea model owns refresh cadence
and last-known-good state, and the renderer remains pure enough for fixed-time,
fixed-width golden tests. Do not add daemon behavior, network fallback,
credential access, history, alerts, settings, or multi-account support in v1.

The central risk is treating local Claude/Codex data as stable and then letting
source errors corrupt the UI. Both data sources are local observations rather
than durable public APIs, so implementation should start with fixture-driven
tolerant parsers and continue with per-source last-known-good merge tests before
visual polish. Layout risk is secondary but still core: the tool must remain
readable in a narrow tmux pane, with progress bars dropped below very narrow
widths rather than wrapping.

## Key Findings

### Recommended Stack

Use the current Go and Charm stack from the start. The project maps directly to
Bubble Tea's Model-Update-View shape: periodic ticks, manual refresh, resize
messages, quit keys, and one fixed full-screen view. The stack should stay small;
stdlib parsing and filesystem APIs are enough because v1 reads local JSON and
JSONL files only.

**Core technologies:**

- Go 1.26.3: language and build toolchain — current supported release with
  enough stdlib support for file I/O, JSON parsing, time math, and tests.
- Bubble Tea v2 (`charm.land/bubbletea/v2` v2.0.6): TUI event loop — exact fit
  for periodic refresh commands, key handling, resize handling, and foreground
  terminal rendering.
- Bubbles progress (`charm.land/bubbles/v2/progress` v2.1.0): progress bars —
  official width-aware component for static quota bars.
- Lip Gloss v2 (`charm.land/lipgloss/v2` v2.0.3): styling and layout — provides
  cell-aware width measurement and centralized colors/layout helpers.
- Go stdlib `encoding/json`, `os`, `path/filepath`, `time`: local source parsing
  — avoids unnecessary parsing/config/runtime dependencies.
- `golang.org/x/sync/errgroup` v0.20.0: parallel refresh helper — fetches
  Claude and Codex independently while returning one refresh message.

Critical version requirement: pick one Charm major version before coding. The
research recommends v2 imports and APIs consistently; mixing v1 examples with v2
types is an early compile-churn trap.

### Expected Features

The v1 feature boundary should be strict. The product succeeds if it shows four
quota windows clearly, refreshes without fuss, tolerates missing data, and exits
predictably. Features that introduce accounts, auth, network calls, background
processes, history, alerts, or broad configuration should be rejected for v1.

**Must have (table stakes):**

- Four fixed quota rows: Claude 5h, Claude 7d, Codex 5h, Codex 7d.
- Percent-used readout, colored progress bar, and reset countdown per row.
- Local Codex rollout JSONL reader and local Claude cache reader.
- 30-second automatic refresh plus `r` manual refresh.
- Clean quit on `q` and `Ctrl-C`.
- Last-known-good preservation after source failures.
- Placeholder rows and footer hints for missing, malformed, stale, or first-run
  data.
- Responsive tmux-pane layout, including dropping bars at very narrow widths.
- Tests for source parsing, stale/failure behavior, update transitions, and
  rendered output.

**Should have (only if cheap):**

- Minimal source freshness hints in the footer.
- Passive Codex plan display when available.
- README troubleshooting notes for missing local data and how to refresh source
  files.

**Defer (v1.x or v2+):**

- Configurable refresh interval and normal-screen mode until real use proves the
  hardcoded defaults wrong.
- Terminal title update and richer freshness display as optional polish.
- Multi-account support, history, forecasting, alerts, network providers,
  per-model breakdowns, daemon mode, one-shot mode, settings UI, mouse support,
  sorting/filtering, and statusline integration.

### Architecture Approach

Build a foreground Bubble Tea program with explicit seams rather than a generic
dashboard framework. Source readers should be Bubble Tea-free and path-injected;
the TUI model should depend on source interfaces and merge per-source results;
the view should render from model state, width, and `now` without filesystem
access. `main.go` should remain thin, owning default paths and program startup
only.

**Major components:**

1. `cmd/llm-quota/main.go` — wires default paths and starts the Bubble Tea
   program; contains no parsing, rendering, or policy logic.
2. `internal/sources/window.go` — defines the shared normalized `Window` shape
   and source identifiers.
3. `internal/sources/claude.go` — reads and validates the statusline-written
   Claude cache file only.
4. `internal/sources/codex.go` — finds the newest Codex rollout JSONL and
   extracts the last usable rate-limit event.
5. `internal/tui/model.go` and `update.go` — hold durable UI state, typed
   messages, refresh commands, tick scheduling, key handling, resize handling,
   and last-known-good merge policy.
6. `internal/tui/view.go` and `colors.go` — render rows, bars, countdowns,
   placeholders, and footer hints using deterministic inputs.
7. `testdata/` — synthetic source fixtures and golden render snapshots; never
   real home-directory data.

### Critical Pitfalls

1. **Treating local quota files as stable APIs** — use tolerant parsers, validate
   only the fields required for the four windows, skip null Codex rate-limit
   events, and cover shape drift with fixtures.
2. **Losing last-known-good data on refresh failure** — store Claude and Codex
   results separately, merge successes per source, and preserve previous windows
   when a source fails.
3. **Blocking or racing the Bubble Tea update loop** — keep filesystem work
   inside commands that return typed messages; never mutate the model from
   goroutines.
4. **Getting narrow-pane layout wrong** — define breakpoints, use Lip Gloss
   width helpers, and test full, compact, and barless layouts.
5. **Brittle or meaningless render tests** — inject fixed time and width, use
   synthetic models, and strip ANSI unless color escapes are intentionally under
   test.
6. **Scope creep into network/history/alerts/accounts** — add explicit non-goal
   checks to every phase and require a separate spec before expanding v1.

## Implications for Roadmap

Based on research, suggested phase structure:

### Phase 1: Module, Stack Pinning, and Domain Model

**Rationale:** The earliest risk is API drift between Charm v1 examples and v2
docs. Pinning Go/Charm versions first prevents churn and creates a compiling
spine for later work.

**Delivers:** Go module, pinned dependencies, minimal Bubble Tea app with quit
behavior, shared `sources.Window` model, formatting helpers, and initial tests
for countdown/percentage/threshold behavior.

**Addresses:** stack setup, clean quit, reset countdown formatting, shared four
window vocabulary.

**Avoids:** mixed Charm major-version APIs and inconsistent source-specific
labels leaking into the TUI.

### Phase 2: Local Source Readers and Fixtures

**Rationale:** The dashboard cannot be trusted until the unofficial local data
sources are normalized defensively. This should come before view work so the UI
does not encode source-specific assumptions.

**Delivers:** Claude cache reader, Codex rollout reader, fixture suite for valid,
missing, malformed, stale, null-rate-limit, no-usable-event, and swapped/missing
window cases.

**Addresses:** local Claude data, local Codex data, missing-data placeholders,
source parsing tests.

**Avoids:** treating local files as stable APIs, reading real home directories in
tests, and silently swapping 5-hour/7-day windows.

### Phase 3: Bubble Tea Model, Refresh, and Last-Known-Good Policy

**Rationale:** Once readers can produce normalized results, the update layer must
prove it can refresh repeatedly without racing, blocking, or blanking good data.
This is the core product behavior behind the always-running pane.

**Delivers:** model state split by source, refresh command fetching sources in
parallel, 30-second tick scheduling, `r` manual refresh, resize message handling,
per-source errors, and tests for success/failure merge sequences.

**Uses:** Bubble Tea v2 commands/messages, `tea.Tick`, `tea.Batch`, and
`errgroup` for source fetches.

**Implements:** TUI model/update architecture and last-known-good retention.

**Avoids:** source errors clearing useful rows, direct filesystem reads inside
`Update`, and goroutine races mutating model state.

### Phase 4: Renderer, Responsive Layout, and Footer Hints

**Rationale:** Rendering depends on the model semantics from Phase 3. The UI must
be built with width breakpoints and deterministic tests from the start because
narrow tmux panes are not a polish case; they are the target environment.

**Delivers:** four fixed rows, percent text, static colored bars, reset
countdowns, placeholders, stale/error hints, minimal key hints, optional
freshness/plan footer metadata, and golden render tests at representative widths.

**Uses:** Lip Gloss v2, Bubbles progress, fixed `now`, fixed terminal width, and
ANSI-stripped golden snapshots where appropriate.

**Avoids:** wrapped bars, overlong footers, brittle timestamped snapshots, and
weak render assertions that miss layout regressions.

### Phase 5: Main Wiring, Documentation, and Manual Tmux Validation

**Rationale:** Real paths and user-facing guidance should be added after the
tested core exists. The external Claude statusline writer belongs outside this
repo, so this phase should document the cache contract and validate the local
experience without taking ownership of dotfiles changes.

**Delivers:** `main.go` default paths, home-directory handling, startup error
handling, README install/troubleshooting notes, manual tmux-pane validation, and
a checklist for separately verifying the Claude statusline cache writer.

**Addresses:** install/use workflow, first-run missing data, source hints,
30-second cadence feel, alt-screen behavior, and real-pane readability.

**Avoids:** repository-boundary confusion, network/Keychain fallback, and adding
configuration before defaults are validated.

### Phase Ordering Rationale

- Pin dependencies before implementation because Charm v1/v2 API confusion is a
  known early failure mode.
- Build parsers before the TUI because all visible correctness depends on
  trustworthy normalized `Window` values and source errors.
- Build model/update before rendering so last-known-good and refresh semantics
  are testable without ANSI/layout noise.
- Build rendering with golden tests before real-path wiring so the narrow-pane
  product requirement is verified against synthetic, deterministic states.
- Leave documentation, real defaults, and manual tmux validation until the core
  is tested, while keeping the dotfiles statusline extension as a separate repo
  concern.

### Research Flags

Phases likely needing deeper research during planning:

- **Phase 2:** MEDIUM research need. Codex rollout and Claude cache shapes are
  local/private contracts; planning should inspect current observed fixtures and
  update parser acceptance criteria without copying real data into the repo.
- **Phase 5:** LOW-to-MEDIUM research need. Manual validation may need a short
  spike on Bubble Tea alt-screen vs normal-screen behavior inside Rob's tmux
  workflow.

Phases with standard patterns (skip research-phase):

- **Phase 1:** Standard Go module setup and Charm dependency pinning are already
  well documented by the stack research.
- **Phase 3:** Bubble Tea command/message patterns are documented and the
  architecture research gives clear test seams.
- **Phase 4:** Lip Gloss/Bubbles progress layout patterns are documented; spend
  effort on implementation tests rather than more research.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Go version, Charm v2 packages, Bubbles progress, Lip Gloss, and `x/sync` were checked against official/module sources. |
| Features | HIGH | v1 scope is strongly anchored in `.planning/PROJECT.md` and the design spec; broader competitor landscape is less relevant. |
| Architecture | HIGH | Component boundaries and data flow follow standard Bubble Tea MVU patterns and match the product constraints. |
| Pitfalls | HIGH | Most pitfalls are derived from concrete Bubble Tea, terminal layout, Go parsing, and project-scope risks; source file durability remains medium confidence. |

**Overall confidence:** HIGH

### Gaps to Address

- **Private source file durability:** Codex rollout JSONL and Claude cache shapes
  are not stable public APIs. Handle with fixtures, tolerant parsing, and clear
  source errors rather than broader integrations.
- **Actual tmux ergonomics:** Alt-screen, 30-second refresh cadence, and footer
  density should be validated manually after the core TUI works.
- **Claude statusline cache writer:** The Go repo depends on a cache file
  produced by dotfiles. Keep the cache contract documented here, but implement
  and review the writer separately.
- **Bubbles progress snapshot behavior:** If progress rendering proves awkward to
  test or too wide for the pane, isolate it behind row rendering and replace only
  the bar implementation with a hand-rolled static bar.

## Sources

### Primary (HIGH confidence)

- `.planning/PROJECT.md` — project goals, active requirements, constraints,
  non-goals, and repository split.
- `docs/superpowers/specs/2026-05-16-llm-quota-tui-design.md` — source shapes,
  UX behavior, failure modes, and visual design intent.
- Context7 `/charmbracelet/bubbletea` — MVU, commands, ticks, window-size
  messages, key messages, v2 imports and view shape.
- Context7 `/charmbracelet/bubbles` — progress component, v2 compatibility, and
  static progress rendering.
- Context7 `/charmbracelet/lipgloss` — styling, cell-aware width measurement,
  joining/layout helpers, v2 color/import APIs.
- Official Go release history and `go list -m -versions` checks — Go 1.26.3 and
  current module versions.
- Go stdlib documentation for `encoding/json`, `bufio.Scanner`, and `os` — JSON
  tolerance, scanner token limits, and filesystem behavior.

### Secondary (MEDIUM confidence)

- GitHub READMEs/releases for Charm packages — ecosystem fit and current release
  confirmation.
- Local observations of Claude statusline cache and Codex rollout JSONL shapes —
  useful for fixtures but not durable contracts.

### Tertiary (LOW confidence)

- None identified. Unknowns are better treated as validation gaps inside the
  implementation phases than as low-confidence external sources.

---

*Research completed: 2026-05-16*
*Ready for roadmap: yes*
