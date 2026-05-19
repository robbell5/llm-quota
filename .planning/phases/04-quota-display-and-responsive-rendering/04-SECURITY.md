---
phase: 04
slug: quota-display-and-responsive-rendering
status: verified
threats_open: 0
asvs_level: 1
created: 2026-05-19
---

<!-- markdownlint-disable MD013 -->

# Phase 04 - Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| local source data -> TUI renderer | Parsed quota percentages and reset timestamps are local but still untrusted for display width and range safety. | Quota percent, reset timestamp, stale metadata |
| terminal renderer -> user | ANSI-styled output must stay bounded to the pane width and must not obscure the true quota value. | Styled terminal text |
| SourceError -> footer copy | Internal local error categories cross into user-facing rendering and must be mapped to safe, actionable copy. | Error category and source identity |
| terminal width -> row layout | External terminal resize messages affect row composition and can cause wrapping if not bounded. | Terminal width |

---

## Threat Register

| Threat ID | Category | Component | Disposition | Mitigation | Status |
|-----------|----------|-----------|-------------|------------|--------|
| T-04-01-01 | Tampering | `internal/tui/view.go` percent/bar rendering | mitigate | `progressFraction` clamps bar input to `0..1`, percent text still renders from `Window.UsedPercent`, and threshold tests cover `59%`, `60%`, `85%`, and `17%`. Evidence: `internal/tui/view.go:196-205`, `internal/tui/view.go:135`, `internal/tui/view_test.go:170-227`. | closed |
| T-04-01-02 | Denial of Service | `internal/tui/view.go` width calculations | mitigate | Bar width is computed from row width, and ANSI-stripped line width tests cover full-row rendering at widths `80` and `50`. Evidence: `internal/tui/view.go:140-154`, `internal/tui/view_test.go:211-227`, `internal/tui/view_test.go:298-307`. | closed |
| T-04-01-03 | Information Disclosure | render tests | mitigate | Render tests use synthetic `sources.Window` fixtures only; auditor found no `.claude`, `.codex`, environment, OS file reads, or real path access in `internal/tui/view_test.go`. Evidence: `internal/tui/view_test.go:170-209`. | closed |
| T-04-02-01 | Information Disclosure | `renderFooter` | mitigate | Source errors are mapped to fixed user-facing hints and tests assert raw categories `malformed`, `read_error`, and `no_usable_event` are absent. Evidence: `internal/tui/view.go:22-28`, `internal/tui/view.go:253-282`, `internal/tui/view_test.go:115-145`. | closed |
| T-04-02-02 | Denial of Service | responsive row assembly | mitigate | Responsive row branches cover full, compact, and narrow layouts; tests verify widths `50`, `49`, `30`, `29`, and `20` remain within ANSI-stripped line bounds. Evidence: `internal/tui/view.go:138-182`, `internal/tui/view_test.go:230-307`. | closed |
| T-04-02-03 | Repudiation | stale data display | mitigate | Footer copy names the stale source and age while preserving the visible quota row; tests verify `Claude data 2h old; open Claude`. Evidence: `internal/tui/view.go:289-323`, `internal/tui/view_test.go:147-168`. | closed |

*Status: open - closed*
*Disposition: mitigate (implementation required) - accept (documented risk) - transfer (third-party)*

---

## Accepted Risks Log

No accepted risks.

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-05-19 | 6 | 6 | 0 | gsd-security-auditor |

---

## Security Audit 2026-05-19

| Metric | Count |
|--------|-------|
| Threats found | 6 |
| Closed | 6 |
| Open | 0 |

## Summary Threat Flags

No unregistered threat flags were found. `04-01-SUMMARY.md` has no Threat Flags section. `04-02-SUMMARY.md` reports no threat flags.

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-05-19
