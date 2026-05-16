# Requirements: llm-quota

**Defined:** 2026-05-16
**Core Value:** Rob can glance at one tmux pane and immediately know how close Claude Code and Codex are to their 5-hour and 7-day limits.

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Quota Display

- [ ] **DISP-01**: User can see Claude Code 5-hour quota usage in the TUI.
- [ ] **DISP-02**: User can see Claude Code 7-day quota usage in the TUI.
- [ ] **DISP-03**: User can see Codex 5-hour quota usage in the TUI.
- [ ] **DISP-04**: User can see Codex 7-day quota usage in the TUI.
- [ ] **DISP-05**: User can see percent used, a colored progress bar, and reset countdown for each available quota window.
- [ ] **DISP-06**: User can interpret quota urgency from green, yellow, and red usage thresholds.

### Claude Setup

- [ ] **CLD-01**: User is prompted for permission to install the `llm-quota` Claude hook/cache writer during setup or first launch.
- [ ] **CLD-02**: User can decline Claude hook installation and still run the TUI with clear Claude placeholder rows.
- [ ] **CLD-03**: User can install or update only the `llm-quota`-owned Claude hook without overwriting unrelated Claude configuration.
- [ ] **CLD-04**: User can get Claude quota data from `~/.cache/llm-quota/claude.json` after the hook has run.

### Local Sources

- [ ] **SRC-01**: User can get Codex quota data from the most recent local Codex rollout JSONL file.
- [ ] **SRC-02**: User sees Codex placeholder rows and a concise hint when no usable Codex quota event exists.
- [ ] **SRC-03**: User sees Claude placeholder rows and a concise hook/setup hint when the Claude cache is missing, malformed, or unavailable.
- [ ] **SRC-04**: User continues seeing last-known-good rows when a later refresh fails for one source.
- [ ] **SRC-05**: User sees stale-data warnings when displayed quota data is older than the accepted freshness threshold.

### TUI Runtime

- [ ] **TUI-01**: User can run `llm-quota` as an always-running foreground TUI.
- [ ] **TUI-02**: User sees quota data refresh automatically every 30 seconds.
- [ ] **TUI-03**: User can press `r` to refresh quota data immediately without disrupting the next scheduled refresh.
- [ ] **TUI-04**: User can press `q` or `Ctrl-C` to exit cleanly.
- [ ] **TUI-05**: User can resize the terminal pane and see the layout adapt without wrapping or breaking rows.
- [ ] **TUI-06**: User can still read useful quota status in very narrow panes where progress bars are omitted.

### Documentation and Verification

- [ ] **DOC-01**: User can install the binary and complete Claude hook setup from documented instructions.
- [ ] **DOC-02**: User can troubleshoot missing Claude or Codex data from documented placeholder hints.
- [ ] **TEST-01**: Maintainer can verify Claude cache parsing for valid, missing, malformed, and stale cache files without touching real home-directory data.
- [ ] **TEST-02**: Maintainer can verify Codex rollout parsing for newest-file selection, null rate limits, malformed events, and missing usable events.
- [ ] **TEST-03**: Maintainer can verify refresh merge behavior preserves last-known-good data after source failures.
- [ ] **TEST-04**: Maintainer can verify rendered output for normal, mixed-threshold, missing-source, stale-source, and narrow-width states.

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Configuration

- **CFG-01**: User can configure refresh interval after the default cadence has been validated.
- **CFG-02**: User can choose alt-screen or normal-screen mode after tmux ergonomics have been validated.

### Polish

- **POL-01**: User can see richer source freshness details if the v1 footer proves insufficient.
- **POL-02**: User can set a terminal title for easier pane/window identification.

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Network fallback for Claude or Codex | Adds auth, credential, prompt, platform, and network failure modes to a local-only tool. |
| Reading Claude credentials or macOS Keychain data | Risks prompts and secret handling; the hook/cache approach avoids credential access. |
| Multi-account support | Requires account identity, source selection, and UI complexity beyond the tiny v1. |
| Historical graphing | Requires persistence and charting; v1 is current status only. |
| Forecasting or burn-rate projection | Adds assumptions that can mislead; reset countdown is enough for v1. |
| Alerts or notifications | Adds background behavior, thresholds, permissions, and distraction. |
| Per-model breakdowns | The target view is product-level subscription windows only. |
| Daemon/background service | Contradicts the foreground tmux-pane model. |
| One-shot quota output mode | Splits runtime behavior and overlaps with statusline use cases. |
| Statusline integration | The TUI is standalone; the Claude hook only writes the cache consumed by the TUI. |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| DISP-01 | Unmapped | Pending |
| DISP-02 | Unmapped | Pending |
| DISP-03 | Unmapped | Pending |
| DISP-04 | Unmapped | Pending |
| DISP-05 | Unmapped | Pending |
| DISP-06 | Unmapped | Pending |
| CLD-01 | Unmapped | Pending |
| CLD-02 | Unmapped | Pending |
| CLD-03 | Unmapped | Pending |
| CLD-04 | Unmapped | Pending |
| SRC-01 | Unmapped | Pending |
| SRC-02 | Unmapped | Pending |
| SRC-03 | Unmapped | Pending |
| SRC-04 | Unmapped | Pending |
| SRC-05 | Unmapped | Pending |
| TUI-01 | Unmapped | Pending |
| TUI-02 | Unmapped | Pending |
| TUI-03 | Unmapped | Pending |
| TUI-04 | Unmapped | Pending |
| TUI-05 | Unmapped | Pending |
| TUI-06 | Unmapped | Pending |
| DOC-01 | Unmapped | Pending |
| DOC-02 | Unmapped | Pending |
| TEST-01 | Unmapped | Pending |
| TEST-02 | Unmapped | Pending |
| TEST-03 | Unmapped | Pending |
| TEST-04 | Unmapped | Pending |

**Coverage:**

- v1 requirements: 27 total
- Mapped to phases: 0
- Unmapped: 27

---

*Requirements defined: 2026-05-16*
*Last updated: 2026-05-16 after initial definition*
