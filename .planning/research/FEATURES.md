# Feature Research

**Domain:** Tiny local terminal quota/status dashboard for Claude Code and Codex rolling limits
**Researched:** 2026-05-16
**Confidence:** HIGH for project-specific v1 scope; MEDIUM for broader terminal dashboard ecosystem

## Feature Landscape

Small terminal status dashboards win by being glanceable, resilient, and boring. For `llm-quota`, the v1 feature set should stay tightly centered on one tmux-pane use case: show the current four subscription windows, refresh without fuss, tolerate missing local files, and exit predictably.

The feature boundary should be unusually strict. Anything that introduces accounts, auth, network calls, background processes, persistent history, or user configuration undermines the core value of a tiny local-only monitor.

### Table Stakes for v1

Features users expect. Missing these means the pane does not solve the stated job.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Four quota rows: Claude 5h, Claude 7d, Codex 5h, Codex 7d | The whole product exists to compare all rolling limits in one glance. | MEDIUM | Non-negotiable v1 scope; partial product support would force users back to separate status views. |
| Percent-used readout per window | Users need exact enough status, not only color or bar length. | LOW | Render as whole percentages unless source precision proves useful. |
| Colored progress bar per window | Visual scan is faster than reading four numbers in a tmux pane. | MEDIUM | Use green/yellow/red thresholds from the design spec; drop bars only at very narrow widths. |
| Reset countdown per window | Quota usage is only actionable if the user knows when capacity returns. | MEDIUM | Format as `Xh Ym` under 24h and `Xd YYh` after that; show `now` for negative durations. |
| Automatic refresh loop | A status pane should stay current without user intervention. | LOW | Use 30 seconds for v1; validate whether 60 seconds feels calmer after real use. |
| Manual refresh key | Users expect a way to force reread after opening Claude/Codex or changing pane focus. | LOW | `r` only; do not add extra controls. |
| Clean quit keys | Foreground TUIs must have obvious exits. | LOW | `q` and `Ctrl-C`; no confirmation prompt. |
| Local Codex rollout reader | Codex quota data is available locally during interactive sessions. | MEDIUM | Read newest rollout JSONL and use the last usable token-count event with rate limits. |
| Claude hook installation | The app must work for users without Rob's custom statusline. | MEDIUM | Installing or first launching the TUI should prompt for permission to install a small Claude hook/cache writer. |
| Local Claude cache reader | Claude data comes from an app-owned hook without prompting for credentials. | MEDIUM | Read `~/.cache/llm-quota/claude.json`; the hook is installed by `llm-quota`. |
| Missing-data placeholders | First run, malformed files, and absent sessions are expected conditions, not crashes. | MEDIUM | Render `—` rows plus short footer hints. |
| Last-known-good retention | Temporary source failure should not blank a useful status pane. | MEDIUM | Keep prior values and annotate stale/error state. |
| Stale-data warnings | Old quota data can mislead unless age is visible. | LOW | Footer warning after 1 hour; do not blank stale data. |
| Responsive narrow-pane layout | The target environment is a dedicated tmux pane, not a full terminal. | MEDIUM | Comfortable around 50 columns; below about 30 columns drop bars and preserve labels/percent/reset. |
| Minimal footer hints | The user needs to know refresh/quit keys and why a source is missing. | LOW | Keep footer terse; no full help overlay. |
| Source parsing and render tests | Local-file parsing and failure states are the main correctness risk. | MEDIUM | Cover Codex JSONL variants, Claude cache errors, stale data, and golden render output. |

### Differentiators and Deferred Improvements

Features that could be valuable, but should not expand v1 unless they directly support the tiny pane workflow.

| Feature | Value Proposition | Complexity | Recommendation |
|---------|-------------------|------------|----------------|
| Display source freshness per product | Builds trust when data comes from local files instead of live APIs. | LOW | Include minimal footer freshness in v1 if it fits; otherwise add in v1.x. |
| Codex plan display | Helps interpret which quota family is being shown without branching behavior. | LOW | Include only as passive footer text when available. |
| Configurable refresh interval | Lets user tune distraction vs freshness. | MEDIUM | Defer until the hardcoded 30-second interval proves wrong. |
| Normal-screen mode option | Preserves tmux scrollback better than alt-screen for some workflows. | LOW | Defer behind implementation spike; default to the design spec choice. |
| Hand-rolled progress bars | More visual control than the Bubbles progress component. | MEDIUM | Defer unless Bubbles progress cannot match the desired compact layout. |
| Terminal title update | Makes tmux/window lists more identifiable. | LOW | Optional v1.x polish; not needed for validation. |
| README troubleshooting section | Reduces confusion about missing local data. | LOW | Add with v1 or immediately after; document hook installation and how to refresh Claude/Codex data. |
| Install convenience | Smooths repeated use from a tmux config or shell. | LOW | `go install` plus a first-launch/setup prompt for the Claude hook is enough for v1; package managers are not. |

### Anti-Features to Deliberately Avoid

Features that may sound useful but fight the tiny, local-only v1.

| Anti-Feature | Why Requested | Why Problematic | Alternative |
|--------------|---------------|-----------------|-------------|
| Network fallback for Claude | Would make missing cache recover automatically. | Adds auth, macOS Keychain prompts, platform-specific behavior, and network failure modes. | Show Claude placeholders and tell the user to open a Claude session. |
| Network fallback for Codex | Would avoid relying on rollout JSONL files. | Adds API coupling and duplicates data Codex already writes locally during real use. | Read newest rollout JSONL and tolerate absent/null rate-limit events. |
| Multi-account support | Useful for consultants or users with personal/work plans. | Requires account identity, config, source selection, and UI complexity. | v1 monitors Rob's current local accounts only. |
| Historical graphing | People often ask status tools for trends. | Requires persistence, charting, time-series policy, and more screen space. | Show current status only; make history a separate future spec if truly needed. |
| Forecasting or burn-rate projections | Sounds actionable for planning quota use. | Needs assumptions about future usage and can become misleading quickly. | Show percent used and reset countdown only. |
| Alerts or desktop notifications | Threshold warnings seem convenient. | Adds background behavior, notification permissions, threshold config, and distraction. | Use visual red/yellow thresholds in the always-visible pane. |
| Per-model breakdowns | Advanced users may want attribution. | Source data and UI focus are product-window level, not model analytics. | Track Claude/Codex subscription windows only. |
| Daemon/background service | Could keep data warm without an open pane. | Contradicts the foreground tmux-pane model and adds lifecycle management. | Keep a single foreground Bubble Tea program. |
| One-shot mode | Convenient for scripting or shell prompts. | Splits runtime behavior and encourages statusline overlap. | Always run the TUI loop; hook setup remains a separate install action. |
| Full settings UI | Users may expect theming, intervals, sources, or thresholds. | Consumes code and UI budget before the fixed personal workflow is validated. | Hardcode sensible defaults; document future knobs only when pain appears. |
| Mouse support | Common in rich TUIs. | No value for a passive dashboard in a tmux pane. | Keyboard-only: `r`, `q`, `Ctrl-C`. |
| Sorting/filtering/navigation | Common table-dashboard affordances. | There are exactly four fixed rows; navigation is noise. | Fixed row order matching product/window importance. |
| Depending on a custom statusline | Reuses Rob's existing local script. | Other users may not have that script, so the app would not be standalone. | Install an app-owned Claude hook/cache writer after prompting for permission. |
| Integrating into the existing statusline | Avoids a new terminal pane. | The project goal is a standalone pane; statusline already has different constraints. | Keep the TUI standalone and use a dedicated hook only to write the Claude cache. |

## Feature Dependencies

```text
Local source readers
    ├──requires──> Four quota rows
    ├──requires──> Percent-used readout
    ├──requires──> Reset countdowns
    └──requires──> Source parsing tests

Claude hook installer
    ├──requires──> Local Claude cache reader
    └──enhances──> First-run missing-data recovery hints

Automatic refresh loop
    ├──enhances──> Source readers
    └──requires──> Last-known-good retention

Manual refresh key
    └──enhances──> Missing-data recovery hints

Responsive layout
    ├──requires──> Fixed row model
    └──enhances──> Progress bars and footer hints

Network fallbacks
    └──conflict──> Local-only tiny v1

History/forecasting/alerts
    └──conflict──> Glanceable current-status pane
```

### Dependency Notes

- **Hook setup precedes Claude source confidence:** the dashboard cannot be standalone until it can install or guide installation of its own Claude cache producer.
- **Source readers precede UI polish:** the dashboard cannot validate value until both Claude and Codex produce normalized local window data.
- **Last-known-good retention depends on refresh semantics:** once refresh can fail repeatedly, model state must distinguish previous data from current errors.
- **Responsive layout depends on fixed content:** v1 can stay simple because four rows fit without scrolling, tables, pagination, or navigation.
- **Anti-features conflict with v1 constraints:** network fallbacks, history, notifications, settings, and multi-account support all introduce state or external dependencies beyond a local pane.

## MVP Definition

### Launch With v1

Minimum viable product needed to validate the concept.

- [ ] Four fixed quota rows for Claude/Codex 5h and 7d windows.
- [ ] Percent, colored bar, and reset countdown for each available window.
- [ ] Local-only Codex rollout reader and Claude cache reader.
- [ ] TUI install/setup or first launch prompts for permission to install the Claude hook/cache writer.
- [ ] 30-second auto-refresh plus `r` manual refresh.
- [ ] Clean exit on `q` and `Ctrl-C`.
- [ ] Last-known-good data preservation after refresh failures.
- [ ] Placeholder rows and footer hints for missing/malformed data.
- [ ] Narrow-pane responsive rendering.
- [ ] Tests for source parsing, stale/failure behavior, and rendered output.

### Add After Validation v1.x

Features to add only after the pane is useful in daily use.

- [ ] Tunable refresh interval if 30 seconds is distracting or too slow.
- [ ] Normal-screen mode option if alt-screen hurts tmux scrollback workflow.
- [ ] More explicit freshness display if the footer warning is not enough.
- [ ] Terminal title command if identifying the pane is annoying.
- [ ] README troubleshooting examples based on real first-run confusion.

### Future Consideration v2+

Only revisit these if `llm-quota` becomes useful beyond Rob's single local setup.

- [ ] Multi-account support, because it requires account identity and source selection.
- [ ] Historical graphing, because it requires persistence and new display modes.
- [ ] Forecasting, because it requires behavior assumptions and likely causes false precision.
- [ ] Alerts/notifications, because they require thresholds, background semantics, and notification integrations.
- [ ] Network data providers, because they introduce auth and credential handling.

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Four quota rows | HIGH | MEDIUM | P1 |
| Percent readout | HIGH | LOW | P1 |
| Progress bars | HIGH | MEDIUM | P1 |
| Reset countdowns | HIGH | MEDIUM | P1 |
| Local Codex reader | HIGH | MEDIUM | P1 |
| Claude hook installer | HIGH | MEDIUM | P1 |
| Local Claude cache reader | HIGH | MEDIUM | P1 |
| Auto-refresh | HIGH | LOW | P1 |
| Manual refresh | MEDIUM | LOW | P1 |
| Clean quit | HIGH | LOW | P1 |
| Last-known-good retention | HIGH | MEDIUM | P1 |
| Placeholder rows and hints | HIGH | MEDIUM | P1 |
| Narrow-pane layout | HIGH | MEDIUM | P1 |
| Parsing/render tests | HIGH | MEDIUM | P1 |
| Passive plan/freshness footer | MEDIUM | LOW | P2 |
| Configurable refresh interval | LOW | MEDIUM | P3 |
| Normal-screen mode option | LOW | LOW | P3 |
| Network fallbacks | LOW | HIGH | Do not build |
| History/forecasting/alerts | LOW for v1 | HIGH | Do not build |
| Multi-account support | LOW for v1 | HIGH | Do not build |

**Priority key:**

- P1: Must have for launch.
- P2: Useful if it does not expand scope.
- P3: Nice to have after real-use validation.
- Do not build: Explicit non-goal for tiny v1.

## Competitor and Pattern Notes

This research did not identify a direct competitor for Claude Code plus Codex rolling subscription quota monitoring. The useful comparison set is small terminal status dashboards generally: they usually emphasize fixed-row status, refresh cadence, concise key hints, color thresholds, and failure visibility rather than deep configuration.

Bubble Tea supports the exact interaction model needed here: a Model/View/Update loop, key messages, recurring ticks, and window-size messages on startup and resize. Bubbles provides a progress component, but also includes heavier components such as table, list, viewport, paginator, file picker, text input, and help. For this product, most of those heavier components should be treated as evidence of what not to include: a four-row passive dashboard does not need browsing, selection, filtering, mouse support, or scrolling.

## Sources

- `.planning/PROJECT.md` -- project goals, active requirements, constraints, and out-of-scope list. Confidence: HIGH.
- `docs/superpowers/specs/2026-05-16-llm-quota-tui-design.md` -- design spec, data sources, UX, failure modes, and non-goals. Confidence: HIGH.
- Context7: Bubble Tea package docs for Model/View/Update, key messages, tick commands, quit behavior, and window-size messages. Confidence: HIGH.
- Context7: Bubbles docs for progress, help, table, viewport, list, key, timer, and other components. Confidence: HIGH for available components; MEDIUM for feature-selection implications.
- GitHub README for `charmbracelet/bubbletea`, fetched 2026-05-16 -- confirms Bubble Tea is suitable for simple/full-window terminal apps and has current v2 releases. Confidence: MEDIUM.
- GitHub README for `charmbracelet/bubbles`, fetched 2026-05-16 -- confirms available UI components including progress, help, tables, lists, and viewports. Confidence: MEDIUM.

---

*Feature research for: tiny terminal quota/status dashboard*
*Researched: 2026-05-16*
