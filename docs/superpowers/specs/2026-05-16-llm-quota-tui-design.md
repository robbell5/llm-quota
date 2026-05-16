# `llm-quota` — Tiny TUI for Claude + Codex Subscription Limits

**Date:** 2026-05-16
**Status:** Draft for review
**Owner:** Rob

## Goal

A small terminal tool that shows current consumption of **all four** rolling
subscription windows in one screen:

- Claude Code — 5-hour rolling window
- Claude Code — 7-day rolling window
- Codex — 5-hour rolling window
- Codex — 7-day rolling window

Each row shows: percent used, a colored progress bar, and "resets in Xh Ym".

The tool is **always running** — it's meant to live in a dedicated tmux pane.
There is no one-shot mode. Refreshes every 30 seconds (or sooner on a window
resize / explicit refresh keypress). Exits on `q` or `Ctrl-C`.

Out of scope: history, projections, alerting/notifications, multi-account
support, per-model breakdowns. If those become wanted, they're separate specs.

## Data Sources

Both products turn out to have clean local data sources, so the steady-state
implementation never makes a network call.

### Codex — read the most-recent rollout JSONL

- Location: `~/.codex/sessions/YYYY/MM/DD/rollout-<ts>-<uuid>.jsonl`
- Strategy: glob the tree, pick the file with the most recent mtime, read it
  end-to-end (these files are typically under 5 MB), keep the last line where
  `type == "event_msg"` AND `payload.type == "token_count"` AND
  `payload.rate_limits` is non-null. Extract `payload.rate_limits`.
- Confirmed shape (from this machine, 2026-05-16):

  ```json
  {
    "limit_id": "codex",
    "limit_name": null,
    "primary":   {"used_percent": 40.0, "window_minutes": 300,   "resets_at": 1778942485},
    "secondary": {"used_percent": 18.0, "window_minutes": 10080, "resets_at": 1779382265},
    "credits": null,
    "plan_type": "prolite",
    "rate_limit_reached_type": null
  }
  ```

- `primary` is the 5h window (300 min); `secondary` is the 7d window (10080 min).
- `resets_at` is unix seconds.
- `plan_type` is informational only; we render it in the footer but never branch
  on it.
- Known caveat: `codex exec` invocations may emit `rate_limits: null`. We tolerate
  this by skipping such lines; the interactive sessions the user actually runs do
  populate the field.

### Claude Code — cache file written by the existing statusline

- The user's `~/.claude/statusline-command.sh` (symlinked from
  `~/dotfiles/claude/.claude/statusline-command.sh`) already extracts
  `rate_limits.five_hour.used_percentage` and `rate_limits.seven_day.used_percentage`
  from the JSON Claude pipes on stdin (lines 134 and 142).
- We extend that script to ALSO atomically write a small JSON file:
  `~/.cache/llm-quota/claude.json`. Shape:

  ```json
  {
    "five_hour":  {"used_percentage": 42.3, "resets_at": 1778942485},
    "seven_day":  {"used_percentage": 85.7, "resets_at": 1779382265},
    "written_at": 1778940000
  }
  ```

- The cache write happens **after** the statusline has already emitted its
  output, so any I/O latency cannot affect display. Atomic write = tmpfile +
  rename. The cache write is additive — the existing script behavior is
  untouched.
- The TUI reads this cache directly. `stale_seconds = now - written_at`.

### Why no OAuth/network fallback for Claude

We deliberately do **not** add an OAuth fallback to `api.anthropic.com/api/oauth/usage`,
even though it exists. The reason: on this machine Claude Code stores its
credentials in macOS Keychain, not in `~/.claude/.credentials.json` (verified
2026-05-16 — no such file exists). Reading from Keychain requires a
`security find-generic-password` call which can prompt the user the first time,
introduces a platform-specific code path, and adds a runtime dep (`httpx`) for
the network call itself.

The "very small TUI" goal outweighs the value of the fallback: if the cache is
missing or stale, the Claude rows render `—` with a footer hint
(`Claude: open a Claude session to refresh`). Opening Claude triggers a
statusline render within seconds and the cache is fresh again.

### Why no network call for Codex

The rollout JSONL is updated by Codex itself every model turn, so the local
file is as fresh as the data behind Codex's own `/status`. A `wham/usage`
endpoint exists as a possible fallback but isn't needed in the common case.

## Architecture

Four small pieces with one tiny shared data shape between them.

### Components

1. **statusline extension** — ~6 added lines in `statusline-command.sh` that
   pipe the rate-limit object to a tmpfile and atomically rename it to
   `~/.cache/llm-quota/claude.json`. Lives in the dotfiles repo, committed
   separately from the Go project.

2. **`internal/sources/codex.go`** — `Fetch() ([]Window, error)`. Pure function
   over the filesystem; no network. Returns a non-nil error on read failure so
   the model can render `—` for that source without crashing the program.

3. **`internal/sources/claude.go`** — `Fetch() ([]Window, error)`. Reads the
   cache file. No network. Returns an error on missing/malformed cache.

4. **`internal/tui/`** — the Bubble Tea program:
   - `model.go` — `Model` struct holding last-known-good windows + errors
   - `update.go` — handles `tickMsg`, `refreshMsg`, `tea.KeyMsg`,
     `tea.WindowSizeMsg`
   - `view.go` — `lipgloss`-based rendering, with `bubbles/progress` for bars
   - `colors.go` — Catppuccin Mocha hex constants

### Shared data shape

Defined in `internal/sources/window.go`:

```go
type Window struct {
    Label        string    // "Claude 5h"
    UsedPct      float64   // 0-100
    ResetsAt     time.Time
    StaleSeconds int       // age of underlying datum
    Source       string    // "cache" or "rollout"
}
```

Each source returns `[]Window` (length 2: the 5h and 7d window for that
product) or an `error`. The model holds `claudeWindows`, `codexWindows`, and
the most recent error for each. A failed refresh leaves the previous good
data in place rather than blanking the display.

### Bubble Tea loop

```text
main.go
  ├─ tea.NewProgram(initialModel, tea.WithAltScreen())
  └─ p.Run()
       │
       ├─ Init() returns tea.Batch(refreshCmd, tickCmd)
       │
       └─ Update(msg) handles:
            ├─ tickMsg       → refreshCmd + new tickCmd
            ├─ refreshMsg    → update model.claudeWindows / codexWindows
            ├─ tea.KeyMsg    → "q"/"ctrl+c" → tea.Quit; "r" → refreshCmd
            └─ tea.WindowSizeMsg → store width/height for view sizing
```

`refreshCmd` is a `tea.Cmd` that runs both source `Fetch()` calls in
goroutines (`errgroup.Group`) and returns a `refreshMsg` with both results.
`tickCmd` is `tea.Tick(30*time.Second, ...)`. Refresh and tick are
independent — the user pressing `r` doesn't reset the next tick.

## UX

One screen, monospace, ~12 rows tall:

```text
LLM Quota                                          2026-05-16 14:07
──────────────────────────────────────────────────────────────────
Claude   5h  ████████░░░░░░░░░░░░░░░░  42%   resets in 2h 14m
Claude   7d  █████████████████████░░░  85%   resets in 4d 06h
Codex    5h  ████████░░░░░░░░░░░░░░░░  40%   resets in 2h 18m
Codex    7d  ███░░░░░░░░░░░░░░░░░░░░░  18%   resets in 5d 02h

   Codex plan: prolite     Claude data cached 12s ago
```

### Color thresholds

- Green: `used_pct < 60`
- Yellow: `60 <= used_pct < 85`
- Red: `used_pct >= 85`

Uses the same Catppuccin Mocha hex values already defined in the user's
statusline script, copied into `internal/tui/colors.go` to avoid coupling.

### Keys

- `q` or `Ctrl-C` — exit
- `r` — force a refresh now (the next 30s tick is unaffected)
- No other keybindings.

### Sizing

The view is laid out for a pane as narrow as ~50 columns. Bars shrink/grow
with available width via `lipgloss` width calculations. On a tmux pane resize
the `tea.WindowSizeMsg` triggers a re-layout.

### "Resets in" format

- Under 24h until reset: `Xh Ym` (e.g. `2h 14m`, `0h 47m`).
- 24h or more: `Xd YYh` (e.g. `4d 06h`).
- Negative remaining (clock skew or stale data): show `now` rather than a
  negative duration.

### Stale-data rendering

Render whatever data we have, regardless of age. Append a footer note if any
row's `stale_seconds > 3600` (1h) — e.g. `Claude data is 2h old, open Claude
to refresh`. We do not blank out stale data; the user is better served by an
old number with a warning than by no number at all.

### Failure rendering

If a source's `Fetch()` returns an error AND we have no last-known-good data
for it, render its two rows as `—  —  —` and append a short hint to the
footer. If we have last-known-good data, keep showing it (annotated as stale)
rather than blanking. The TUI never crashes on missing data.

## Tech Choices

### Go + Bubble Tea

Chosen primarily because Rob is using this project to learn Bubble Tea. The
"always running in a tmux pane" model is also a natural fit for Bubble Tea's
event-loop architecture — the alternatives (Python + rich, bash + ANSI) would
either need a hand-rolled loop or sit in `rich.live.Live` with manual
ticking.

Runtime deps (all from the Charm ecosystem):

- `github.com/charmbracelet/bubbletea` — the framework
- `github.com/charmbracelet/bubbles` — for the `progress` component
- `github.com/charmbracelet/lipgloss` — for styling and layout

Plus `golang.org/x/sync/errgroup` for parallel source fetches.

Go stdlib covers everything else (`encoding/json` for sources, `os`/`filepath`
for cache + rollouts, `time` for window math).

### Install

```bash
cd ~/Personal/llm-quota
go install ./cmd/llm-quota
# → installs to $GOBIN (or $GOPATH/bin), already on PATH per ~/go layout
```

## Project Layout

```text
~/Personal/llm-quota/
├── go.mod
├── go.sum
├── README.md
├── docs/superpowers/specs/
│   └── 2026-05-16-llm-quota-tui-design.md   # this file
├── cmd/llm-quota/
│   └── main.go                              # entrypoint, tea.NewProgram(...).Run()
├── internal/
│   ├── sources/
│   │   ├── window.go                        # Window struct
│   │   ├── claude.go                        # cache reader
│   │   ├── claude_test.go
│   │   ├── codex.go                         # rollout JSONL reader
│   │   └── codex_test.go
│   └── tui/
│       ├── model.go                         # tea.Model + msg types
│       ├── update.go                        # Update()
│       ├── view.go                          # View() via lipgloss
│       ├── view_test.go                     # golden-file render test
│       └── colors.go                        # Catppuccin Mocha
└── testdata/                                # Go-idiomatic, toolchain ignores it
    ├── codex_rollout.jsonl                  # synthetic, no secrets
    ├── claude_cache.json
    └── golden/                              # expected View() output snapshots
        ├── all_green.txt
        ├── mixed_thresholds.txt
        └── claude_missing.txt
```

Statusline extension lives separately in `~/dotfiles/claude/.claude/statusline-command.sh`.

## Testing

All tests are stdlib `testing`, colocated with the code they cover.

- **`internal/sources/codex_test.go`:** fixture JSONL with a mix of
  `session_meta`, `response_item`, and several `event_msg`/`token_count`
  events; assert we get the last one's `rate_limits`. Fixture with
  `rate_limits: null` (exec-mode shape) — assert we skip it and fall back to
  the previous valid event.
- **`internal/sources/claude_test.go`:** fixture cache file → assert parsed
  correctly. Missing file → returns an error. Malformed JSON → returns an
  error. Cache with `written_at` 2h ago → parsed with `StaleSeconds ≈ 7200`.
- **`internal/tui/view_test.go`:** call `View()` with a fixed `now`, fixed
  width, and a fixed `Model`; assert the rendered string (after stripping
  ANSI via a small helper) matches a golden file under `testdata/golden/`.
  Cover: all-green, mixed thresholds, one source missing.

Sources accept their file/directory path as a constructor arg (e.g.
`NewCodex(rolloutsDir string)`), defaulting to the real location in `main.go`.
Tests pass `testdata/` paths so no test touches `~/.claude` or `~/.codex`.

## Failure Modes & Mitigations

| Failure                                       | Behavior                                                                |
|-----------------------------------------------|-------------------------------------------------------------------------|
| No Codex rollout files at all                 | Codex rows render `—`; footer: "Codex: no recent session"               |
| Codex rollout exists but no usable event      | Same as above; footer notes "no token_count event found"                |
| Cache file missing                            | Claude rows render `—`; footer: "Claude: open a Claude session to refresh" |
| Cache file present but malformed              | Same as missing                                                         |
| Source recovers mid-loop                      | Next 30s tick succeeds; rows return to normal automatically             |
| Pane resized to very narrow                   | View shrinks bars; if `width < 30`, drops bars and shows percent only   |
| Pane resized to very tall                     | Extra vertical space at bottom; no layout breakage                      |

The cache file is the single biggest fragility. It's also low-stakes: as long
as Claude is run at least once per ~hour, the cache stays current. If the user
goes a day without running Claude, the rows just show `—` with a friendly
hint. Document this in README.

## Non-Goals (explicitly out)

- Multi-account support
- Historical graphing
- Push notifications when thresholds are crossed
- Per-model breakdowns
- Integration into the existing statusline (the statusline already shows the
  numbers; this is a separate stand-alone view)
- A daemon. The tool is a foreground Bubble Tea program meant to run in a tmux
  pane; it is not backgrounded or auto-started.
- A one-shot/non-watch mode. The tool always runs in its event loop.

## Open Spikes

These are small and can be resolved during implementation rather than gating
the design:

1. Decide whether to use Bubble Tea's `WithAltScreen` or normal mode.
   Alt-screen is cleaner for a dedicated pane, but normal mode survives tmux
   scrollback better. Default plan: alt-screen, revisit if it's annoying.
2. Pick between `bubbles/progress` (Charm's bar component) and a hand-rolled
   bar drawn with `lipgloss`. The Charm component handles width/animation
   automatically; the hand-rolled version is more transparent. Default plan:
   `bubbles/progress`, swap if it doesn't match the visual we want.
3. Confirm the tick interval feels right at 30s once running in a pane —
   may bump to 60s if it's visibly distracting.
