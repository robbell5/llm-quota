# Phase 1: Foreground TUI Foundation - Context

**Gathered:** 2026-05-16
**Status:** Ready for planning

## Phase Boundary

Phase 1 delivers the minimal runnable foreground `llm-quota` Bubble Tea TUI
spine. The user must be able to start a stable screen and quit cleanly with
`q` or `Ctrl-C`. This phase does not implement quota data sources, refresh
cadence, real quota rows, responsive rendering, install/setup flows, or
documentation beyond what is needed for the foundation to compile and run.

## Implementation Decisions

### Startup Screen

- **D-01:** Render a future-shaped startup screen with the app title and four
  placeholder quota rows, not a title-only placeholder.
- **D-02:** Use user-ready missing-data hints even in Phase 1, so the skeleton
  screen resembles the eventual first-run experience.
- **D-03:** Establish the final visual direction now: polished terminal shell,
  row alignment, divider/footer feel, and palette direction may be introduced in
  Phase 1. Detailed quota behavior still belongs to later phases.
- **D-04:** Include compact footer content with working quit keys plus concise
  Claude/Codex placeholder hints.

### Quit Feedback

- **D-05:** Normal quit returns silently to the shell. Do not leave a final
  message or stale TUI frame behind.
- **D-06:** `q` and `Ctrl-C` must behave the same from the user's perspective:
  clean quit, no panic, no partial terminal state, no traceback.
- **D-07:** If the TUI fails before startup, print a concise plain error and
  return a non-zero exit code. Do not attempt a fallback UI.
- **D-08:** Phase 1 should explicitly test quit behavior where practical, in
  addition to a manual launch/quit smoke check.

### Key Hint Scope

- **D-09:** Show only working keys in Phase 1. Do not preview `r` until manual
  refresh exists in a later phase.
- **D-10:** Use compact key copy: `q / Ctrl-C quit`.
- **D-11:** Keep key hints and data-source hints in the same compact footer.
- **D-12:** Avoid footer wrapping in narrow terminals. Use a shorter compact
  footer variant if needed, even before full responsive rendering lands.

### Run Target

- **D-13:** Phase 1 is done only when both `go run ./cmd/llm-quota` and an
  installed `llm-quota` binary from `go install ./cmd/llm-quota` can launch the
  TUI.
- **D-14:** Keep Phase 1 command handling minimal: no arguments launches the
  TUI; unknown arguments should fail plainly. Do not add help text, install
  commands, or stub future commands in this phase.
- **D-15:** Initialize the module as `github.com/rob/llm-quota`.
- **D-16:** Verification before completion should include `go test ./...`,
  `go install ./cmd/llm-quota`, and a quick interactive launch/quit smoke check.

### Agent Discretion

No decisions were explicitly delegated with "you decide." Downstream agents may
choose exact copy, spacing, file split, and test mechanics as long as the locked
decisions above and canonical references are respected.

## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Scope And Requirements

- `.planning/ROADMAP.md` — Defines Phase 1 goal, success criteria, dependency
  status, and phase boundary.
- `.planning/REQUIREMENTS.md` — Maps Phase 1 to `TUI-01` and `TUI-04` and keeps
  later TUI, data-source, setup, and rendering requirements out of this phase.
- `.planning/PROJECT.md` — Defines core value, local-only constraints, key
  decisions, and out-of-scope behavior.

### Stack And Architecture

- `.planning/research/STACK.md` — Locks Go, Bubble Tea v2, Bubbles, Lip Gloss,
  and `golang.org/x/sync/errgroup`; includes current Charm v2 API notes.
- `.planning/research/ARCHITECTURE.md` — Defines recommended package seams,
  Bubble Tea MVU pattern, test seams, and anti-patterns to avoid.

### Product Design

- `docs/superpowers/specs/2026-05-16-llm-quota-tui-design.md` — Original design
  spec covering runtime model, keys, startup/quit expectations, future data
  source behavior, one-screen UX, and proposed project layout.

## Existing Code Insights

### Reusable Assets

- No Go source exists yet. Phase 1 should create the initial module, command
  entrypoint, and TUI package from scratch.
- Existing planning and research docs are the reusable assets for this phase;
  downstream agents should not infer behavior from absent code.

### Established Patterns

- Use Bubble Tea v2 import paths and APIs consistently; do not mix v1
  `github.com/charmbracelet/...` examples with v2 `charm.land/.../v2` code.
- Keep `main.go` thin. Put TUI model/update/view behavior under `internal/tui`.
- Prefer deterministic unit tests for update behavior and render helpers before
  relying on manual TUI checks.
- The project favors small direct Go code over broad CLI/config frameworks.

### Integration Points

- `cmd/llm-quota/main.go` will become the user-facing command entrypoint.
- `internal/tui` will own the Bubble Tea model, key handling, startup view, and
  clean quit behavior.
- Later phases will connect source readers, refresh commands, responsive quota
  rendering, and Claude hook installation to this foundation.

## Specific Ideas

- The Phase 1 screen should look like the future quota pane with placeholder
  rows, not like a temporary debug screen.
- Silent alt-screen cleanup is preferred for the dedicated tmux-pane workflow.
- Do not display inactive future keys; the visible footer should only document
  behavior that works now.

## Deferred Ideas

None — discussion stayed within phase scope.

---

*Phase: 1-Foreground TUI Foundation*
*Context gathered: 2026-05-16*
