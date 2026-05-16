# Phase 1: Foreground TUI Foundation - Research

**Researched:** 2026-05-16
**Status:** Complete
**Discovery level:** Level 1 quick verification

## Summary

Phase 1 can use the existing project stack and architecture research without a new external decision. The key implementation risk is mixing Charm v1 and v2 APIs while creating the first Bubble Tea spine. Current docs confirm the project should use Bubble Tea v2 import paths and declarative `tea.View` fields.

## Verified Stack Details

- Use `charm.land/bubbletea/v2` for the TUI event loop.
- `View()` returns `tea.View`; create content with `tea.NewView(rendered)`.
- Set alternate-screen behavior by assigning `v.AltScreen = true` inside `View()`.
- Handle quit keys with `tea.KeyPressMsg` and `msg.String()` values `q` and `ctrl+c`.
- Use `charm.land/lipgloss/v2` for pure style values and width-aware rendering.
- Keep Bubbles progress pinned now for stack coherence, but do not render progress bars until the quota display phase needs them.

## Planning Implications

- Start with module/dependency pinning, a thin command entrypoint, and a focused `internal/tui` package.
- Test quit behavior through `Update` using `tea.KeyPressMsg` before relying on manual terminal smoke checks.
- Put the Phase 1 startup screen renderer in its own file so later display work can extend it without rewriting command startup.
- Unknown CLI arguments should fail before starting Bubble Tea, print one concise plain error, and exit non-zero.

## Out of Scope for Phase 1

- No source readers, hook installer, refresh cadence, manual refresh key, real quota percentages, progress bars, setup docs, network calls, Keychain reads, daemon behavior, or one-shot quota output.

## Sources

- `.planning/research/STACK.md`
- `.planning/research/ARCHITECTURE.md`
- `docs/superpowers/specs/2026-05-16-llm-quota-tui-design.md`
- Context7 `/charmbracelet/bubbletea` v2 docs and upgrade guide
- Context7 `/charmbracelet/lipgloss` v2 docs and upgrade guide
- Context7 `/charmbracelet/bubbles` v2 docs and upgrade guide

## RESEARCH COMPLETE
