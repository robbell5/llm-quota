# Phase 5: Install, Docs, and Real-Pane Validation - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-20T12:00:27Z
**Phase:** 5-Install, Docs, and Real-Pane Validation
**Areas discussed:** Install path, Docs shape, Troubleshooting coverage, Real pane validation

---

## Install Path

| Option | Description | Selected |
|--------|-------------|----------|
| go install | Smallest path for a Go tool; documents `go install` and avoids release packaging work. | yes |
| Local build | Document cloning the repo and running `go build`; useful for development. | |
| Release binary | Plan a packaged binary artifact; more polished but adds release automation or manual packaging scope. | |
| Both go install + build | Primary user path plus fallback developer path. | |

**User's choice:** `go install` as the primary documented v1 install path.
**Notes:** Keep command handling narrow, document explicit `llm-quota install-claude-hook`, and verify install instructions locally.

---

## Docs Shape

| Option | Description | Selected |
|--------|-------------|----------|
| README | Best discoverability for install, setup, run, and troubleshooting in a tiny CLI/TUI repo. | yes |
| docs/ guide | Keeps README short, but creates an extra place users must find. | |
| README + docs guide | More complete, but heavier than v1 may need. | |
| You decide | Let the planner choose the smallest useful docs surface. | |

**User's choice:** Put primary user-facing docs in `README.md`.
**Notes:** README should be quickstart plus troubleshooting, mention only user-visible local paths, and Phase 5 should update both README and planning artifacts.

---

## Troubleshooting Coverage

| Option | Description | Selected |
|--------|-------------|----------|
| Mirror hints | Use the same user-facing states as the TUI footer hints. | yes |
| More diagnostic detail | Include source-error categories and file-shape details. | |
| Very terse | Only say rerun setup/open tools. | |
| You decide | Let the planner choose based on existing renderer copy. | |

**User's choice:** Mirror TUI footer hints.
**Notes:** Claude docs should cover hook plus cache basics; Codex docs should tell users to open Codex locally; stale-but-valid data should be explained calmly as last-known local data.

---

## Real Pane Validation

| Option | Description | Selected |
|--------|-------------|----------|
| Manual checklist | Run the real app in the intended pane and record pass/fail notes. | yes |
| Checklist + screenshots | Stronger evidence, but screenshots may be cumbersome in terminal/tmux workflow. | |
| Automated only | Relies on existing tests and misses terminal/color checks. | |
| You decide | Let the planner set the evidence bar. | |

**User's choice:** Manual checklist recorded in planning UAT.
**Notes:** Do not add a demo or fixture mode. Checklist should cover cadence, quit keys, responsive widths, and color perception.

---

## the agent's Discretion

No decisions were delegated with "you decide."

## Deferred Ideas

None.
