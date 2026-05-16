# Pitfalls Research

**Domain:** Tiny local-file-backed Go/Bubble Tea TUI quota monitor
**Researched:** 2026-05-16
**Confidence:** HIGH for Bubble Tea/layout/Go parsing pitfalls; MEDIUM for
Claude/Codex file-shape drift because those formats are local observations, not
stable public contracts.

## Critical Pitfalls

### Pitfall 1: Treating local quota files as stable APIs

**What goes wrong:**
The Codex rollout reader hard-codes one observed JSONL shape or assumes the
newest file always contains usable quota data. The Claude reader assumes the
cache is always present and complete. A harmless upstream shape change, a
`codex exec` run with `rate_limits: null`, or a missing first-run Claude cache
then blanks the TUI or crashes it.

**Why it happens:**
Local files feel safer than network APIs, so small tools often skip defensive
schema handling. In this project, both sources are intentionally unofficial:
Codex rollout JSONL is a session artifact, and Claude data is a cache written by
an app-installed hook.

**How to avoid:**
Implement sources as tolerant parsers with narrow success criteria:

- Codex: scan the chosen rollout for the last line where `type == "event_msg"`,
  `payload.type == "token_count"`, and `payload.rate_limits != null`.
- Validate `primary.window_minutes == 300` and
  `secondary.window_minutes == 10080`; if labels drift, return a source error
  instead of silently swapping 5-hour and 7-day windows.
- Claude: require both windows and `written_at`; missing/malformed cache returns
  an error that the TUI can render as placeholder rows.
- Keep parsing tests fixture-driven, including null rate limits, malformed JSON,
  missing fields, swapped windows, and no usable events.

**Warning signs:**

- Tests only cover the happy-path JSON shown in the design spec.
- Parser structs expose every observed field instead of the few fields needed.
- A source error replaces previously good data with empty rows.
- `plan_type` or `limit_name` starts influencing logic.

**Phase to address:**
Phase 1: Source readers and fixtures. This must be solved before TUI polish,
because the render layer depends on trustworthy `Window` values and source
errors.

---

### Pitfall 2: Losing last-known-good data on refresh failure

**What goes wrong:**
The 30-second refresh path overwrites the model with whatever the latest fetch
returns. If one source has a transient read error, malformed partial write, or
no current session, the display regresses from useful stale data to `—` rows.

**Why it happens:**
The simple model shape is tempting: fetch both sources, assign both results,
render. That ignores the project's central UX rule: old data with a warning is
better than no data.

**How to avoid:**
Make refresh merging explicit in `internal/tui/update.go`:

- Successful source result replaces that source's windows and clears its error.
- Failed source result stores the error but preserves existing windows.
- Placeholder rows render only when there is no last-known-good data for that
  source.
- Staleness is calculated from source metadata or fetch time, not from when the
  row happened to be rendered.

**Warning signs:**

- `refreshMsg` contains a single combined error instead of per-source results.
- The model has one `windows []Window` field rather than separate Claude and
  Codex state.
- Tests assert missing-source placeholders but not recovery after a successful
  fetch followed by a failed fetch.

**Phase to address:**
Phase 2: Bubble Tea model/update loop. Add model-level tests before view tests
so stale-data behavior is locked down independently of ANSI rendering.

---

### Pitfall 3: Blocking or racing the Bubble Tea update loop

**What goes wrong:**
File reads, globbing, or status checks run directly inside `Update`, making the
TUI feel frozen during slow filesystem operations. Conversely, ad hoc goroutines
mutate model state outside Bubble Tea messages, creating races that tests miss.

**Why it happens:**
Tiny TUIs blur the boundary between "read a file" and "handle an event." Bubble
Tea's architecture expects side effects to be represented as `tea.Cmd` values
that return messages; direct mutation from goroutines bypasses that contract.

**How to avoid:**
Keep the MVU boundary strict:

- `Update` handles `tickMsg`, `refreshMsg`, key messages, and
  `tea.WindowSizeMsg`; it does not touch the filesystem.
- `refreshCmd` performs both fetches and returns one message containing both
  source outcomes.
- Use `tea.Batch(refreshCmd, tickCmd)` for concurrent commands, but remember
  Bubble Tea gives no ordering guarantee for batched commands.
- Never mutate `Model` from inside source goroutines; only the returned
  `refreshMsg` changes state.

**Warning signs:**

- `Update` imports `os`, `filepath`, or source package filesystem helpers.
- A goroutine closes over `*Model` or writes to model fields.
- Manual refresh resets the periodic tick unintentionally.
- Race detector failures appear only under resize or rapid `r` presses.

**Phase to address:**
Phase 2: Bubble Tea event loop. Treat this as an architectural phase gate before
adding visual polish.

---

### Pitfall 4: Terminal resize support exists but layout math is wrong

**What goes wrong:**
The app receives `tea.WindowSizeMsg`, but rows still overflow in a narrow tmux
pane, bars wrap onto a second line, or footer hints push important rows off
screen. The TUI technically "handles resize" while failing the glanceable-pane
goal.

**Why it happens:**
Terminal cell width is not string length. ANSI sequences, styled progress bars,
padding, and Unicode bar characters make naive `len()` math unreliable. Lip
Gloss provides cell-aware width helpers, but it is easy to bypass them while
assembling fixed-format rows.

**How to avoid:**
Define width breakpoints in the view layer and test each one:

- `>= 50 cols`: full layout with progress bars.
- `30-49 cols`: shortened labels and smaller bars.
- `< 30 cols`: drop bars; show product/window, percent, and reset text only.
- Use `lipgloss.Width`, `lipgloss.Size`, `MaxWidth`, and inline rendering for
  measuring/truncating rendered strings.
- Keep footer hints short and prioritize source-action hints over decorative
  metadata.

**Warning signs:**

- View tests run at only one fixed width.
- Row-building uses hard-coded bar widths with no lower bound.
- The footer includes plan, cache age, all errors, and key hints at every width.
- Snapshot output differs after stripping ANSI because style codes affected
  alignment assumptions.

**Phase to address:**
Phase 3: Rendering and responsive layout. Do not defer narrow-pane behavior to a
later polish pass; it is core to the tmux-pane product.

---

### Pitfall 5: Rendering tests become brittle or meaningless

**What goes wrong:**
Golden tests either fail constantly due to timestamps, ANSI color changes, and
terminal width differences, or they become too weak and only check that `View()`
returns a non-empty string.

**Why it happens:**
TUI output is stringly and environment-sensitive. Without injected time, fixed
width, deterministic model fixtures, and ANSI normalization, tests measure the
host terminal instead of the renderer.

**How to avoid:**
Make rendering deterministic by design:

- Inject `now` into the model/view path or provide a `Now func() time.Time` hook
  used by tests.
- Build test models directly; do not read real `~/.claude` or `~/.codex` data in
  view tests.
- Strip ANSI before comparing golden files unless color output itself is the
  behavior under test.
- Maintain golden cases for all-green, mixed thresholds, one source missing,
  stale source, and narrow widths.

**Warning signs:**

- Golden files include the current wall-clock date or cache age.
- Tests depend on the actual terminal width or color profile.
- Failures are resolved by weakening assertions to substring checks everywhere.
- `go test` passes locally only when a real Claude or Codex session has run.

**Phase to address:**
Phase 3: Rendering and test fixtures. Rendering tests should be added with the
first real `View()` implementation, not after the UI stabilizes by hand.

---

### Pitfall 6: Implementing network, history, alerts, or multi-account fallback
inside v1

**What goes wrong:**
The project expands from a tiny foreground pane into a quota product: OAuth or
Keychain reads, network fallback endpoints, history storage, alerting,
multi-account config, model breakdowns, or daemon behavior. The app becomes
harder to install, harder to trust, and less useful as a simple learning project
for Bubble Tea.

**Why it happens:**
Quota data invites "just one more" feature. Missing local data makes network
fallback especially tempting, but the spec explicitly rejects it because the
tool is valuable precisely because it avoids prompts, credentials, and services.

**How to avoid:**
Use explicit scope guards in README and roadmap acceptance criteria:

- No network clients in this repository.
- No credential, Keychain, OAuth, or `DATABASE_URL` handling.
- No history database, daemon, alert subsystem, one-shot mode, or per-model
  drill-down.
- Claude cache writing stays inside the small app-owned hook installer and does
  not depend on Rob's custom statusline.

**Warning signs:**

- New dependencies appear for HTTP clients, config files, SQLite, notifications,
  cron/launchd, or keychain access.
- Placeholder copy says "trying network fallback" instead of "open Claude" or
  "start a Codex session."
- A phase plan contains "while we're here" features not listed in active
  requirements.

**Phase to address:**
Every phase. Add a scope check to each phase review; reject work that contradicts
the non-goals unless a new spec intentionally supersedes v1.

---

## Technical Debt Patterns

Shortcuts that seem reasonable but create long-term problems.

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Parse Codex JSONL into `map[string]any` everywhere | Fast first parser | Runtime panics, silent field swaps, hard-to-test null handling | Only in a small private helper before converting to typed structs |
| Store all rows in one flat slice | Simple render loop | Cannot preserve one source when the other fails | Never for model state; acceptable for final render assembly |
| Use `len()` for row width | Quick alignment | Broken layout with ANSI/styled output | Never in layout code; use Lip Gloss width helpers |
| Read real home-directory files in tests | Easy manual validation | Flaky tests and accidental secret/session coupling | Never in automated tests; use fixtures only |
| Add a config file early | Feels flexible | Creates a product surface before defaults are validated | Defer until a real need appears after v1 |
| Treat stale data as an error state | Simple branching | Hides useful quota data | Never; stale data is displayable with a warning |

## Integration Gotchas

Common mistakes when connecting to local data producers.

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| Codex rollout JSONL | Picking newest file and assuming latest line contains quota data | Pick newest rollout, then scan for the last usable `event_msg`/`token_count` with non-null `rate_limits` |
| Codex rollout JSONL | Treating `rate_limits: null` as fatal | Skip null entries and keep looking backward for the last usable event |
| Claude hook cache | Reading cache while the writer is mid-write | Hook must write tmpfile then rename; reader treats malformed cache as a source error and keeps last-known-good data |
| Claude hook cache | Assuming Rob's custom statusline exists | Install an app-owned hook after prompting for permission; use the statusline script only as implementation inspiration |
| tmux pane | Assuming resize means only width changed | Store both width and height; test narrow, normal, and extra-tall panes |

## Performance Traps

Patterns that work at tiny scale but fail when the pane runs all day.

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Full recursive glob every few seconds | CPU/disk spikes, battery drain, laggy pane | Keep refresh at 30s; only scan Codex sessions on refresh; consider day-bucket pruning if needed | Large `~/.codex/sessions` tree over months |
| `bufio.Scanner` with default token size for JSONL | Rare failure on long JSONL line | Either read the small file and split lines, or set `Scanner.Buffer` above expected line size | A rollout line exceeds 64 KiB |
| Animated progress bars | Constant redraws and distraction in tmux | Use static bars for quota percentages; no animation loop besides refresh/tick | Immediately noticeable in always-on pane |
| Serial source reads in refresh | One slow/missing source delays the other | Fetch Claude and Codex independently and merge per-source outcomes | Slow filesystem or deep Codex tree |

## Security Mistakes

Domain-specific security issues beyond general CLI security.

| Mistake | Risk | Prevention |
|---------|------|------------|
| Reading Claude credentials or Keychain from the TUI | Prompts, secret exposure, platform-specific failures | Do not implement network/OAuth fallback; read only the hook-written cache |
| Logging raw rollout/cache data | Local session metadata leaks into terminal logs or test output | Log only concise source errors; fixtures must be synthetic |
| Modifying arbitrary statusline scripts | Cross-repo ownership confusion and accidental user regressions | Install or update only the `llm-quota`-owned Claude hook entry after user permission |

## UX Pitfalls

Common user experience mistakes for this specific always-running quota pane.

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Blanking stale rows | User loses the last useful quota estimate | Keep stale rows visible and add age/hint in footer |
| Over-explaining every source error | The pane stops being glanceable | Short row placeholders plus one concise footer hint per source |
| Too many keybindings | Turns monitor into an app to learn | Only `r`, `q`, and `Ctrl-C` for v1 |
| Showing per-model details | Dilutes the four-window goal | Product-level rows only |
| Letting plan type drive layout or logic | Incorrect behavior if plan labels change | Render plan type as optional footer metadata only |

## "Looks Done But Isn't" Checklist

Things that appear complete but are missing critical pieces.

- [ ] **Codex parsing:** Handles null `rate_limits`, malformed lines, no usable
  event, and swapped/missing window metadata.
- [ ] **Claude parsing:** Handles missing cache, malformed cache, stale cache,
  and valid cache without touching the real home directory in tests.
- [ ] **Refresh merging:** Failed refresh preserves last-known-good rows for that
  source and records the current error.
- [ ] **Resize rendering:** Golden tests cover full, narrow, and barless widths.
- [ ] **Reset countdowns:** Negative reset durations render as `now`, not a
  negative value.
- [ ] **Scope:** No network, OAuth, Keychain, daemon, alert, history, one-shot, or
  multi-account code appears in v1.

## Recovery Strategies

When pitfalls occur despite prevention, how to recover.

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Parser assumes wrong local shape | MEDIUM | Add failing fixture from observed shape, narrow parser success criteria, preserve placeholder behavior |
| Last-known-good data is overwritten | MEDIUM | Split model state by source, add update tests for success-fail-success sequences |
| Layout wraps in tmux | LOW | Add width breakpoint fixture, measure rendered cells with Lip Gloss, drop bars below threshold |
| Golden tests are flaky | LOW | Inject fixed time/width, strip ANSI, replace real-source setup with fixture models |
| Scope creep lands in code | HIGH | Revert feature, update roadmap/non-goal notes, require a separate spec before reintroducing |

## Pitfall-to-Phase Mapping

How roadmap phases should address these pitfalls.

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| Local files treated as stable APIs | Phase 1: Source readers | Fixture tests cover shape drift, nulls, malformed input, missing data |
| Last-known-good data lost | Phase 2: Model/update loop | Unit tests verify failed refresh preserves prior successful source state |
| Blocking/racy update loop | Phase 2: Model/update loop | `Update` remains pure over messages; race detector passes under refresh/resize tests |
| Resize math wrong | Phase 3: Rendering/layout | Golden outputs at representative widths, including `< 30` columns |
| Rendering tests brittle | Phase 3: Rendering/layout | Fixed-time, fixed-width, ANSI-stripped golden tests use synthetic models |
| Scope creep | Every phase | Phase reviews check active requirements and explicit non-goals before completion |

## Sources

- Project context: `.planning/PROJECT.md` (read 2026-05-16) -- HIGH confidence
- Design spec: `docs/superpowers/specs/2026-05-16-llm-quota-tui-design.md`
  (read 2026-05-16) -- HIGH confidence for intended behavior, MEDIUM for
  observed local Claude/Codex shape stability
- Bubble Tea package docs via Context7/pkg.go.dev: `WindowSizeMsg` arrives on
  startup/resize, `Tick` must be returned again for recurring ticks,
  `Batch` executes commands concurrently with no ordering guarantee,
  `WithAltScreen` is the correct startup option -- HIGH confidence
- Lip Gloss docs via Context7: use `Width`, `Height`, `Size`, `MaxWidth`, and
  inline/max-width rendering for terminal cell-aware layout -- HIGH confidence
- Bubbles progress docs via Context7: progress width is configurable and v2 uses
  setter/options; static progress bars are supported -- HIGH confidence
- Go `encoding/json` docs: unknown struct keys are ignored by default;
  `Decoder.DisallowUnknownFields` is available; number/duplicate-key behavior has
  caveats -- HIGH confidence
- Go `bufio.Scanner` docs: default maximum token size is 64 KiB unless
  `Scanner.Buffer` is configured; Scanner stops unrecoverably on oversized tokens
  -- HIGH confidence
- Go `os` docs: filesystem errors include path context; file operations are safe
  for concurrent use, but OS limits may still apply -- HIGH confidence

---

*Pitfalls research for: llm-quota local quota monitor TUI*
*Researched: 2026-05-16*
