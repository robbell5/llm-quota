---
phase: 06-i-think-we-are-missing-an-uninstaller
status: pending
validated_at: null
---

# Phase 06 Human UAT

## Scope

Validate that the Claude setup uninstaller safely removes only the app-owned integration and that install → uninstall → reinstall remains reversible in a real local Claude setup.

Do not paste private Claude settings contents into this file. Record only command outcomes, marker presence/absence, and whether expected files remain present.

## Checklist

- [ ] Build local binary with `go build ./cmd/llm-quota`.
- [ ] Run `./llm-quota install-claude-hook`.
- [ ] Confirm Claude settings contain an app-owned `llm_quota_marker`.
- [ ] Run `./llm-quota uninstall-claude-hook`.
- [ ] Confirm the app-owned marker is absent.
- [ ] Confirm any previous statusline command is restored when present.
- [ ] Confirm `~/.cache/llm-quota/claude.json is not deleted` by uninstall.
- [ ] Rerun `./llm-quota install-claude-hook`.
- [ ] Run `go test ./... -count=1`.

## Results

- **Status:** pending
- **Validator:** pending
- **Validated at:** pending
- **Notes:** pending

## Summary

- total: 9
- passed: 0
- issues: 0
- pending: 9
- skipped: 0
- blocked: 0
