---
phase: 06-i-think-we-are-missing-an-uninstaller
status: passed
validated_at: 2026-05-21T19:40:36Z
---

# Phase 06 Human UAT

## Scope

Validate that the Claude setup uninstaller safely removes only the app-owned integration and that install → uninstall → reinstall remains reversible in a real local Claude setup.

Do not paste private Claude settings contents into this file. Record only command outcomes, marker presence/absence, and whether expected files remain present.

## Checklist

- [x] Build local binary with `go build ./cmd/llm-quota`.
- [x] Run `./llm-quota install-claude-hook`.
- [x] Confirm Claude settings contain an app-owned `llm_quota_marker`.
- [x] Run `./llm-quota uninstall-claude-hook`.
- [x] Confirm the app-owned marker is absent.
- [x] Confirm any previous statusline command is restored when present.
- [x] Confirm `~/.cache/llm-quota/claude.json is not deleted` by uninstall.
- [x] Rerun `./llm-quota install-claude-hook`.
- [x] Run `go test ./... -count=1`.

## Results

- **Status:** passed
- **Validator:** Human checkpoint approved by Rob
- **Validated at:** 2026-05-21T19:40:36Z
- **Notes:** Real local install → uninstall → reinstall flow approved. No private Claude settings contents were recorded.

## Summary

- total: 9
- passed: 9
- issues: 0
- pending: 0
- skipped: 0
- blocked: 0
