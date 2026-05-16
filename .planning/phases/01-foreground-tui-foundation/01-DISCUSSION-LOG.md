# Phase 1: Foreground TUI Foundation - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or
> execution agents. Decisions are captured in CONTEXT.md; this log preserves
> the alternatives considered.

**Date:** 2026-05-16
**Phase:** 1-Foreground TUI Foundation
**Areas discussed:** Startup Screen, Quit Feedback, Key Hint Scope, Run Target

---

## Startup Screen

### Startup Screen: Pre-Data Display

**Question:** What should the Phase 1 startup screen show before real quota
data exists?

- **Skeleton Rows (Recommended):** Show the app title plus four placeholder
  quota rows so the foundation already resembles the final pane without source
  data. **Selected:** Yes.
- **Title Only:** Keep Phase 1 minimal with a stable title and quit
  instructions only.
- **Setup Message:** Show a short explanation that data source setup arrives in
  later phases.
- **You decide:** Let the planner choose the smallest version that still proves
  foreground TUI behavior.

### Startup Screen: Placeholder Copy

**Question:** How explicit should the placeholder copy be about quota data not
existing yet?

- **Short Phase Note (Recommended):** Say data sources land in later phases,
  without turning the foundation screen into documentation.
- **User-Ready Hint:** Use final-product style missing-data hints now, even
  before source readers exist. **Selected:** Yes.
- **No Caveat:** Render placeholders only and avoid explaining why they are
  empty.
- **You decide:** Let the planner choose the least distracting copy.

### Startup Screen: Visual Styling

**Question:** How much visual styling should Phase 1 establish?

- **Minimal Styled Shell (Recommended):** Use the final stack and a simple
  polished shell, but defer detailed quota colors and responsive rendering.
- **Plain Text First:** Prioritize the smallest compiling TUI and leave styling
  almost entirely for Phase 4.
- **Final Visual Direction:** Set the initial palette, divider, row alignment,
  and footer feel now so later phases fill in behavior. **Selected:** Yes.
- **You decide:** Let the planner choose based on the smallest useful
  foundation.

### Startup Screen: Footer Content

**Question:** What footer should the startup screen include in Phase 1?

- **Keys + Data Hints (Recommended):** Show q/Ctrl-C plus concise
  Claude/Codex placeholder hints that will evolve later. **Selected:** Yes.
- **Keys Only:** Only show working quit keys; avoid data hints until data
  source phases.
- **Phase Status:** Show keys plus a clear foundation-only status line for
  development transparency.
- **You decide:** Let the planner choose a low-clutter footer.

**User's choice:** Skeleton rows, user-ready hints, final visual direction,
and keys plus data hints.
**Notes:** User moved to the next area after these decisions.

---

## Quit Feedback

### Quit Feedback: Shell Return

**Question:** After a normal quit, what should the user see back at the shell?

- **Silent Return (Recommended):** Use alt-screen cleanup so q/Ctrl-C leaves no
  extra app text behind. **Selected:** Yes.
- **One-Line Exit:** Leave a concise message like `llm-quota exited` after
  returning to the shell.
- **Last Screen Remains:** Prefer normal-screen behavior where the final TUI
  frame remains visible in scrollback.
- **You decide:** Let the planner choose the cleanest behavior for a dedicated
  tmux pane.

### Quit Feedback: Quit Key Equivalence

**Question:** Should q and Ctrl-C behave exactly the same from the user's
perspective?

- **Same Behavior (Recommended):** Both keys quit cleanly, return silently, and
  leave no traceback or partial screen. **Selected:** Yes.
- **Ctrl-C More Abrupt:** q is the normal path; Ctrl-C exits as an interrupt
  with minimal special handling.
- **You decide:** Let the planner implement the simplest clean quit semantics.

### Quit Feedback: Startup Failure

**Question:** If the TUI fails before it can start, how should that surface?

- **Plain Error (Recommended):** Print a concise shell error and return
  non-zero; do not attempt a fallback UI. **Selected:** Yes.
- **Friendly Message:** Print a more user-oriented explanation with a suggested
  next step.
- **Silent Nonzero:** Return a failing exit code without extra text.
- **You decide:** Let the planner choose based on standard Go CLI behavior.

### Quit Feedback: Testing

**Question:** Should Phase 1 explicitly test quit behavior, or is manual
validation enough for this foundation?

- **Test It (Recommended):** Add focused model/update tests for q and Ctrl-C
  returning the quit command where practical. **Selected:** Yes.
- **Manual Only:** Keep Phase 1 tests lighter and validate quit behavior by
  running the TUI.
- **Build Smoke Only:** Only require the app to compile and launch; defer key
  behavior tests.
- **You decide:** Let the planner set the smallest responsible verification
  boundary.

**User's choice:** Silent return, identical `q`/`Ctrl-C` behavior, plain
startup errors, and explicit quit tests.
**Notes:** User moved to the next area after these decisions.

---

## Key Hint Scope

### Key Hint Scope: Visible Keys

**Question:** Which keys should Phase 1 show in the footer?

- **Working Keys Only (Recommended):** Show q and Ctrl-C only; avoid
  advertising r until refresh exists. **Selected:** Yes.
- **Preview r:** Show q, Ctrl-C, and mark r as coming later or inactive.
- **No Key Footer:** Rely on known terminal conventions and avoid footer key
  hints for now.
- **You decide:** Let the planner choose the clearest low-clutter hint set.

### Key Hint Scope: Quit Key Copy

**Question:** How should the quit keys be written?

- **q / Ctrl-C (Recommended):** Compact and explicit: `q / Ctrl-C quit`.
  **Selected:** Yes.
- **Separate Labels:** Spell them separately: `q quit` and `Ctrl-C quit`.
- **Natural Text:** Use sentence copy: `Press q or Ctrl-C to quit.`
- **You decide:** Let the planner pick whichever fits the final layout best.

### Key Hint Scope: Footer Placement

**Question:** Should Phase 1 show data-source hints in the same footer as key
hints?

- **Same Footer (Recommended):** Keep a single compact footer with quit keys
  plus Claude/Codex placeholder hints. **Selected:** Yes.
- **Separate Lines:** Put key hints and data hints on separate lines for
  readability.
- **Data Hints Only Later:** Show quit keys now and add data hints when source
  readers exist.
- **You decide:** Let the planner balance clutter against usefulness.

### Key Hint Scope: Narrow Footer

**Question:** If the terminal is too narrow in Phase 1, what should happen to
the footer hints?

- **Keep Compact (Recommended):** Use a short version instead of wrapping, even
  before full responsive layout lands. **Selected:** Yes.
- **Allow Wrap:** Accept wrapped footer text in Phase 1 and refine it during
  responsive rendering.
- **Hide Data Hints:** Keep quit keys visible and drop data hints when width is
  tight.
- **You decide:** Let the planner choose based on minimal rendering complexity.

**User's choice:** Working keys only, compact `q / Ctrl-C quit` copy, shared
footer, compact no-wrap behavior.
**Notes:** User moved to the next area after these decisions.

---

## Run Target

### Run Target: User-Facing Path

**Question:** What should count as the Phase 1 user-facing run path?

- **Go Run + Binary (Recommended):** Verify both `go run ./cmd/llm-quota` and
  `go install ./cmd/llm-quota` produce a runnable `llm-quota`.
  **Selected:** Yes.
- **Go Run Only:** Treat Phase 1 as development-only; installed binary
  validation waits for install/docs.
- **Binary Only:** Focus on the final `llm-quota` command and avoid documenting
  go run as a user path.
- **You decide:** Let the planner choose the smallest path satisfying the
  roadmap.

### Run Target: Command-Line Arguments

**Question:** How should Phase 1 handle command-line arguments?

- **No-Arg TUI Only (Recommended):** No arguments launches the TUI; unknown
  args return a plain error. Setup commands wait for later phases.
  **Selected:** Yes.
- **Add Help Only:** Support `-h`/`--help` now, but no setup/install subcommands
  yet.
- **Stub Future Commands:** Add visible placeholders for install/setup commands
  that are implemented later.
- **You decide:** Let the planner choose the least distracting CLI surface.

### Run Target: Module Path

**Question:** Which module path should downstream planning assume for the Go
module?

- **github.com/rob/llm-quota (Recommended):** Matches the stack research and
  gives imports a stable repository-shaped path. **Selected:** Yes.
- **llm-quota:** Use a short local module path since this is primarily a
  personal local tool.
- **Decide During Plan:** Leave this to planning if repository publishing
  details are not important yet.
- **You decide:** Let the planner pick the most conventional path.

### Run Target: Verification

**Question:** What Phase 1 verification should downstream planning require
before calling the run target done?

- **Test + Build + Run (Recommended):** Require `go test ./...`,
  `go install ./cmd/llm-quota`, and a quick launch/quit smoke check.
  **Selected:** Yes.
- **Test + Build:** Require tests and build/install, but leave interactive
  smoke testing for later.
- **Build Only:** Only require compilation because the TUI is minimal.
- **You decide:** Let the planner set the verification boundary.

**User's choice:** Support both `go run` and installed binary paths, keep CLI
arguments minimal, use `github.com/rob/llm-quota`, and verify with tests,
install, and smoke launch/quit.
**Notes:** User said the selected areas were ready for context.

---

## Agent Discretion

No areas were explicitly delegated with "you decide."

## Deferred Ideas

None.
