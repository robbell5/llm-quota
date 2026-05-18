# Phase 2: Standalone Local Data Sources - Context

**Gathered:** 2026-05-18
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 2 delivers the standalone local data-source layer for `llm-quota`: the
user can choose whether to install the app-owned Claude hook/cache writer, the
installer safely preserves unrelated Claude configuration, and the app can read
Claude and Codex quota data from local files using defensive source readers.
This phase does not implement the periodic refresh loop, last-known-good merge
behavior, final quota bar rendering, responsive layout polish, or release docs
beyond what is needed to prove setup and source parsing.

</domain>

<decisions>
## Implementation Decisions

### Setup Prompt Flow

- **D-01:** Support both a first-launch setup offer and an explicit
  `llm-quota install-claude-hook` command. Do not add broader setup aliases in
  this phase.
- **D-02:** If the user declines the first-launch Claude hook prompt, remember
  that decline so normal launches do not keep interrupting the user while
  Claude data is unavailable.
- **D-03:** Show the first-launch permission prompt before starting the
  alt-screen TUI. The prompt should be plain, explicit, and complete before
  `tea.NewProgram(...).Run()` enters the TUI.
- **D-04:** A declined install must still let the TUI run with clear Claude
  placeholder rows and setup hints.

### Hook Safety Policy

- **D-05:** Only entries with an explicit `llm-quota` managed marker/name count
  as app-owned. Do not infer ownership from generic quota behavior or unknown
  user hook commands.
- **D-06:** When unrelated Claude hook entries already exist and the config shape
  is understood, preserve them and append or update only the managed
  `llm-quota` entry.
- **D-07:** Create a timestamped backup before writing Claude configuration, but
  only when a config change is actually needed.
- **D-08:** `install-claude-hook` must be idempotent: update an existing managed
  `llm-quota` hook in place and report whether anything changed.

### Reader Tolerance

- **D-09:** Treat the Claude cache contract as all-or-nothing for its two quota
  windows. If one Claude window is missing or invalid, reject the Claude source
  rather than returning partial Claude data or invented zero values.
- **D-10:** Codex JSONL parsing should skip unrelated events, null rate limits,
  malformed trailing or individual lines, and otherwise bad events while
  scanning for the last valid rate-limit event.
- **D-11:** A valid but old Claude cache is data, not a hard error. Return the
  windows with age/stale metadata so the TUI can warn without blanking values.
- **D-12:** Source errors should use typed categories such as missing, malformed,
  no usable event, and read/permission error so footer hints can stay concise
  and actionable.

### Codex Rollout Choice

- **D-13:** If the newest Codex rollout JSONL has no usable rate-limit event,
  fall back to the newest older rollout file that does contain usable rate-limit
  data.
- **D-14:** Search all rollout JSONL files under `~/.codex/sessions` rather than
  limiting the reader to the latest date directory or a fixed recent window.
- **D-15:** Use file modification time to order Codex rollout files. Do not parse
  private filename timestamp conventions for ordering.
- **D-16:** Preserve Codex `plan_type` as optional source metadata for later
  footer rendering, without making the TUI parse Codex-specific source details.

### Agent Discretion

No decisions were explicitly delegated with "you decide." Downstream agents may
choose exact struct names, error type names, prompt copy, backup filename format,
and test fixture filenames as long as the locked decisions above are respected.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Scope And Requirements

- `.planning/ROADMAP.md` — Defines Phase 2 goal, success criteria, dependency
  status, and phase boundary.
- `.planning/REQUIREMENTS.md` — Maps Phase 2 to `CLD-01` through `CLD-04`,
  `SRC-01` through `SRC-03`, `TEST-01`, and `TEST-02`.
- `.planning/PROJECT.md` — Defines core value, local-only constraints, key
  decisions, and out-of-scope network/OAuth/Keychain behavior.
- `.planning/phases/01-foreground-tui-foundation/01-CONTEXT.md` — Carries
  forward the Phase 1 command, startup screen, placeholder, footer, and TUI
  foundation decisions.

### Stack And Architecture

- `.planning/research/STACK.md` — Locks Go 1.26.3, Bubble Tea v2, Bubbles v2,
  Lip Gloss v2, and `golang.org/x/sync/errgroup`.
- `.planning/research/ARCHITECTURE.md` — Defines `internal/sources`,
  `internal/install`, Bubble Tea command boundaries, path-injected tests, and
  local-source anti-patterns.

### Product Design

- `docs/superpowers/specs/2026-05-16-llm-quota-tui-design.md` — Original design
  spec covering Claude cache shape, Codex rollout shape, no-network rationale,
  source-reader behavior, and fixture expectations. Phase 2 overrides one Codex
  detail by allowing fallback to an older usable rollout file.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- `cmd/llm-quota/main.go` is the thin command edge. It currently rejects all
  arguments, so Phase 2 should add `install-claude-hook` handling here while
  keeping parsing, install policy, and source logic out of `main`.
- `internal/tui/view.go` already renders the future-shaped four-row placeholder
  screen and compact footer hints. Phase 2 can use this for declined/missing
  Claude data without needing final quota bars yet.
- `internal/tui/model.go` and `internal/tui/update.go` hold the current Bubble
  Tea model, quit behavior, resize state, and `tea.View`/alt-screen setup. They
  should remain source-format agnostic.
- `internal/tui/*_test.go` provides the existing table-driven update tests,
  ANSI-stripped render assertions, and width guards that new tests should match
  stylistically.

### Established Patterns

- Use Charm v2 import paths and APIs consistently: `charm.land/bubbletea/v2`,
  `tea.KeyPressMsg`, and `View() tea.View` with `v.AltScreen = true`.
- Keep `main.go` at the edge for real defaults, CLI command dispatch, and program
  startup errors. Put hook install policy under `internal/install` and source
  parsing under `internal/sources`.
- Source readers must accept explicit paths/root directories so tests use
  synthetic fixtures and never touch real `~/.claude`, `~/.codex`, or cache data.
- Rendering and source parsing should stay separate. The TUI should consume
  normalized source results and typed source errors, not Claude/Codex JSON shapes.

### Integration Points

- Add `internal/sources/window.go` for the shared normalized window/result shape
  and optional source metadata such as Codex `plan_type`.
- Add `internal/sources/claude.go` and `internal/sources/codex.go` with fixture
  tests covering valid, missing, malformed, stale, null-limit, no-usable-event,
  and fallback-to-older-rollout cases.
- Add `internal/install/claude_hook.go` for prompted installation/update of the
  app-owned Claude hook/cache writer, preserving unrelated Claude config.
- Update `cmd/llm-quota/main.go` to run the pre-TUI first-launch prompt when
  appropriate and to route `install-claude-hook` without entering the TUI.

</code_context>

<specifics>
## Specific Ideas

- The first-launch prompt should happen before the alt-screen TUI starts so the
  user sees a normal terminal permission question before any file mutation.
- Remembering a declined install is a product decision for Phase 2 even if the
  exact storage mechanism is left to planning.
- The Claude source should prefer a clear source-unavailable state over partial
  Claude quota rows if the cache contract is incomplete.
- The Codex source should be pragmatic about local file noise: scan all rollout
  files by mtime and use the newest usable quota event, even if a newer rollout
  file has only null or malformed data.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 2-Standalone Local Data Sources*
*Context gathered: 2026-05-18*
