---
phase: 05-install-docs-and-real-pane-validation
status: complete
research_level: 0
created: 2026-05-20
---

<!-- markdownlint-disable MD013 -->

# Phase 05 Research: Install, Docs, and Real-Pane Validation

## Research Decision

No new external technical research is required for Phase 5. The phase is documentation, install smoke verification, and human real-pane validation over behavior already implemented in Phases 1-4.

## Inputs Used

- `.planning/phases/05-install-docs-and-real-pane-validation/05-CONTEXT.md` locks the install, troubleshooting, and real-pane validation decisions.
- `.planning/phases/04-quota-display-and-responsive-rendering/04-VERIFICATION.md` and `04-HUMAN-UAT.md` identify the remaining human terminal/tmux checks to close.
- `cmd/llm-quota/main.go` confirms the intentionally narrow command surface: no-arg TUI launch, `install-claude-hook`, and internal `claude-hook-cache-writer`.
- `internal/tui/view.go` defines the user-facing footer hints README troubleshooting must mirror: `q / Ctrl-C quit`, `r refresh`, `Claude: run install-claude-hook`, `Claude: open Claude`, `Codex: open Codex`, and stale age hints.

## Planning Implications

- Plan README work without adding CLI help, demo mode, fixture mode, network fallback, statusline integration, or release-binary packaging.
- Verify install instructions through local build/test commands and documented `go install` instructions, not by mutating the user's shell PATH.
- Treat real tmux-pane validation as a blocking human-verification checkpoint recorded in Phase 5 planning artifacts rather than as README content.
- Preserve source-error trust boundaries by documenting user-facing hints and recovery actions, not raw `SourceError` category names or private Claude/Codex schemas.

## Security Notes

- Documentation is a trust boundary: incorrect setup commands can lead users to mutate Claude configuration unexpectedly. Plans must keep `install-claude-hook` explicit and describe app-owned hook behavior.
- Troubleshooting docs must avoid encouraging users to paste or expose private Claude/Codex local payloads.
