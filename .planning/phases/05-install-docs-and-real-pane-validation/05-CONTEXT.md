# Phase 5: Install, Docs, and Real-Pane Validation - Context

**Gathered:** 2026-05-20T12:00:27Z
**Status:** Ready for planning

## Phase Boundary

Phase 5 makes the completed `llm-quota` TUI usable as a small local
tmux-pane tool: document how to install it, document standalone Claude hook
setup without relying on Rob's custom statusline, explain missing/stale
local-data recovery using the same language as the TUI, and record real
tmux-pane validation for cadence, quit keys, responsive layout, and terminal
colors.

This phase does not add network, OAuth, Keychain, daemon, statusline
integration, release-binary packaging, demo/fixture runtime modes, history,
forecasting, alerts, or new data sources.

## Implementation Decisions

### Install Path

- **D-01:** The primary documented v1 install path should be `go install`.
- **D-02:** Keep CLI command handling narrow. Do not add `--help`, `help`, or
  broader setup aliases in Phase 5 unless planning finds an unavoidable
  blocker.
- **D-03:** Document `llm-quota install-claude-hook` as the explicit Claude
  setup path after install. The first-launch prompt may still exist, but docs
  should not depend on users discovering it.
- **D-04:** Verify install instructions locally through repo/build output rather
  than requiring mutation of a broader real user PATH or shell setup.

### Docs Shape

- **D-05:** Create the primary user-facing documentation in `README.md`.
- **D-06:** README depth should be quickstart plus troubleshooting: install,
  hook setup, run, quit/refresh keys, tmux-pane expectation, and missing/stale
  data recovery.
- **D-07:** Documentation may mention user-visible local paths such as
  `~/.cache/llm-quota/claude.json` and `~/.codex/sessions`, but should avoid
  detailed private Claude/Codex file schemas unless necessary for
  troubleshooting.
- **D-08:** Phase 5 should update both user-facing README content and
  planning/validation artifacts so the release state is traceable.

### Troubleshooting Coverage

- **D-09:** Troubleshooting docs should mirror the TUI footer hints and
  user-facing states rather than exposing raw internal `SourceError` category
  names.
- **D-10:** Claude troubleshooting should explain the hook/cache basics: run
  `llm-quota install-claude-hook`, then open Claude so the app-owned hook can
  write `~/.cache/llm-quota/claude.json`.
- **D-11:** Codex troubleshooting should tell users to open Codex locally so
  session rollout data is produced under `~/.codex/sessions`.
- **D-12:** Stale-but-valid data should be explained calmly as last-known local
  data that remains useful; users refresh it by opening Claude or Codex, not by
  treating it as a hard failure.

### Real Pane Validation

- **D-13:** Real tmux-pane validation should be a manual checklist, not
  screenshots or automation-only evidence.
- **D-14:** Record real-pane validation results in a Phase 5 planning UAT or
  validation artifact, not in the README as release-test evidence.
- **D-15:** Do not add a public demo mode, fixture mode, or temporary validation
  mode just to force quota colors for manual checks.
- **D-16:** The manual checklist must cover the default refresh cadence, quit
  keys, responsive widths, and perceived terminal colors. It should carry
  forward Phase 4's human-needed checks for widths 50, 49, 30, and 29 plus
  green/yellow/red color perception.

### the agent's Discretion

No decisions were explicitly delegated with "you decide." Downstream agents may
choose exact README wording, validation artifact filenames, checklist
formatting, and local verification commands as long as the decisions above are
respected.

## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Scope And Requirements

- `.planning/PROJECT.md` — Defines the local-only product, core tmux-pane
  value, standalone setup constraint, failure tolerance, and validated
  decisions through Phase 4.
- `.planning/REQUIREMENTS.md` — Defines Phase 5 requirement IDs `DOC-01` and
  `DOC-02`, plus out-of-scope constraints such as no network fallback, no
  statusline integration, and no daemon.
- `.planning/ROADMAP.md` — Defines Phase 5 goal and success criteria: install
  binary, complete Claude hook setup, troubleshoot missing data, and validate
  in the intended tmux pane.
- `.planning/STATE.md` — Carries current status, accumulated decisions, and the
  explicit concern that actual tmux ergonomics still need final validation.

### Prior Phase Context

- `.planning/phases/04-quota-display-and-responsive-rendering/04-CONTEXT.md`
  — Locks footer hint behavior, calm color-only urgency, responsive
  breakpoints, and row rendering expectations that docs and validation must
  respect.
- `.planning/phases/03-refresh-and-resilience-loop/03-CONTEXT.md` — Locks
  refresh cadence, manual `r` refresh behavior, resize semantics, stale
  threshold, and last-known-good policy.
- `.planning/phases/02-standalone-local-data-sources/02-CONTEXT.md` — Locks
  app-owned Claude hook setup, explicit install command, hook safety policy,
  Claude cache path, Codex local source behavior, and source-error trust
  boundaries.

### Validation Carry-Forward

- `.planning/phases/04-quota-display-and-responsive-rendering/04-VERIFICATION.md`
  — Identifies the remaining human checks for real tmux-pane layout and
  terminal color perception that Phase 5 must close.
- `.planning/phases/04-quota-display-and-responsive-rendering/04-HUMAN-UAT.md`
  — Tracks the pending Phase 4 manual UAT items that Phase 5 should incorporate
  or supersede in final validation.

### Product Design

- `docs/superpowers/specs/2026-05-16-llm-quota-tui-design.md` — Original
  design spec for local-only data, tmux-pane footprint, four quota windows,
  refresh cadence, setup flow, and failure rendering.

## Existing Code Insights

### Reusable Assets

- `cmd/llm-quota/main.go` — Owns no-arg TUI launch,
  `install-claude-hook`, first-launch consent prompt,
  `claude-hook-cache-writer`, default Claude/Codex local paths, and the
  `tea.NewProgram` startup seam.
- `cmd/llm-quota/main_test.go` — Covers install command behavior,
  first-launch accept/decline flow, unknown-argument behavior, cache writer
  behavior, and source-backed model construction with injected paths.
- `internal/tui/view.go` — Defines the exact footer hints and user-facing
  recovery copy docs should mirror: `q / Ctrl-C quit`, `r refresh`,
  `Claude: run install-claude-hook`, `Claude: open Claude`,
  `Codex: open Codex`, and stale age hints.
- `.planning/phases/04-quota-display-and-responsive-rendering/04-VERIFICATION.md`
  — Provides ready-made human validation expectations for pane widths and
  color perception.

### Established Patterns

- Command dispatch is intentionally small: no-arg TUI launch,
  `install-claude-hook`, and internal `claude-hook-cache-writer` only.
- Real local paths remain at the command edge. Tests use injected paths and
  synthetic fixtures rather than touching real home-directory data.
- UI copy hides raw parser/source categories. Troubleshooting docs should
  preserve that boundary and speak in terms of user actions.
- Render validation uses automated ANSI-stripped tests for content and width,
  with human UAT reserved for real terminal/tmux perception.

### Integration Points

- Add `README.md` for install, setup, run, and troubleshooting documentation.
- Add or update Phase 5 planning artifacts to record manual real-pane
  validation results.
- Keep any command-line changes minimal; if planning proposes CLI help, it must
  reconcile with the locked decision to keep command handling narrow.

## Specific Ideas

- The README should be practical and short: install with `go install`, run
  `llm-quota install-claude-hook`, run `llm-quota` in a tmux pane, use `r` to
  refresh, and quit with `q` or `Ctrl-C`.
- Troubleshooting should explain what the visible hints mean and what the user
  should do next, not how to parse internal JSON/JSONL schemas.
- Final manual validation should close the Phase 4 human-needed checks rather
  than inventing new visual capabilities.

## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 5-Install, Docs, and Real-Pane Validation*
*Context gathered: 2026-05-20T12:00:27Z*
