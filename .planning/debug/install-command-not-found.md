---
status: resolved
trigger: "Step 1 is failing: `go build ./cmd/llm-quota` succeeds, then `llm-quota install-claude-hook` fails with `zsh: command not found: llm-quota`."
created: 2026-05-20
updated: 2026-05-20
---

# Debug Session: install-command-not-found

## Symptoms

- Expected behavior: Following README Step 1 should install or make available the `llm-quota` command so `llm-quota install-claude-hook` works.
- Actual behavior: `go build ./cmd/llm-quota` completes but `llm-quota install-claude-hook` fails in zsh with `command not found`.
- Error messages: `zsh: command not found: llm-quota`.
- Timeline: Found during Phase 5 human real-pane validation after README docs were added.
- Reproduction: From repository root, run `go build ./cmd/llm-quota`, then run `llm-quota install-claude-hook`.

## Current Focus

- hypothesis: documentation mixed the local `go build` smoke path with the installed `llm-quota` command path
- test: inspect README instructions, module command layout, and Go build output behavior
- expecting: `go build ./cmd/llm-quota` writes `./llm-quota`, not a command resolvable as `llm-quota` unless the repo root is on PATH
- next_action: resolved; ask human to retry Phase 5 UAT with the matching installed or local command path
- reasoning_checkpoint:
- tdd_checkpoint:

## Evidence

- timestamp: 2026-05-20T12:37:26Z
  observation: README documented `go build ./cmd/llm-quota` as a local smoke path and correctly said it writes a local `llm-quota` binary, but the following setup and run commands used bare `llm-quota`.
  implication: A user choosing the local build path will get `zsh: command not found: llm-quota` unless the repo root is on PATH; they must run `./llm-quota ...` or use `go install`.
- timestamp: 2026-05-20T12:37:26Z
  observation: `cmd/llm-quota/main.go` dispatches `install-claude-hook` as an implemented command, and `cmd/llm-quota/main_test.go` covers that path.
  implication: The command implementation is present; the failure happens before program startup during shell command lookup.
- timestamp: 2026-05-20T12:37:26Z
  observation: Phase 5 human UAT also paired `go build ./cmd/llm-quota` with bare `llm-quota install-claude-hook`.
  implication: The human UAT instructions needed the same installed-vs-local command distinction as README.
- timestamp: 2026-05-20T12:45:00Z
  observation: After verification cleanup removed the generated local binary, `./llm-quota install-claude-hook` failed with `zsh: no such file or directory: ./llm-quota`.
  implication: The local smoke path must be presented as a sequence: build first, then run `./llm-quota`; if the binary is missing, rebuild before retrying.

## Eliminated

## Resolution

- root_cause: Documentation and Phase 5 UAT instructions conflated the local `go build` smoke path with an installed command; `go build ./cmd/llm-quota` creates `./llm-quota` in the repo root but does not make `llm-quota` available on PATH.
- fix: Updated README and Phase 5 human UAT instructions to present the installed and local smoke paths as explicit sequences; local smoke path now rebuilds before using `./llm-quota` and notes that a missing local binary means the build must be rerun.
- verification: README command documentation check passed; `go test ./... -count=1` passed; `go build ./cmd/llm-quota` passed; `test -x ./llm-quota` passed after build.
- files_changed: README.md; .planning/phases/05-install-docs-and-real-pane-validation/05-HUMAN-UAT.md; .planning/debug/install-command-not-found.md
