# llm-quota

`llm-quota` is a small foreground terminal UI for watching Claude Code and Codex subscription quota windows in a dedicated tmux pane. It reads local usage files only, refreshes while it runs, and shows placeholder rows with recovery hints when local data is missing or stale.

## What it shows

- Claude Code 5-hour usage
- Claude Code 7-day usage
- Codex 5-hour usage
- Codex 7-day usage

Each available row shows percent used, a color-gradient progress bar (green → amber → red), and the reset countdown. Missing rows stay visible so the pane remains useful while local data is being produced.

Each row also shows a recent burn rate, a forecast (projected fill at reset, or time-to-100% when a window is on pace to exhaust early), and a sparkline of usage within the current window. A window projected to hit 100% before it resets is flagged with a ⚠ and red forecast text.

Each provider's group header also shows the **equivalent API value** of usage in its 5-hour and 7-day windows — what those tokens would have cost at pay-as-you-go API rates. This is a value/ROI figure (your flat-rate subscription has no per-token cost), not money spent. Codex values are an estimate (prefixed `~`), since ChatGPT-plan token pricing is unofficial. Toggle the values with `c` or hide them at launch with `--no-cost`.

## Install

Choose one install path.

To install with Homebrew, run:

```sh
brew install robbell5/tap/llm-quota
llm-quota install-claude-hook
llm-quota
```

Homebrew links the `llm-quota` command into its managed bin directory, so no Go `PATH` setup is required. Pick up new releases with:

```sh
brew update && brew upgrade robbell5/tap/llm-quota
```

Check the installed version any time with `llm-quota --version`.

For Go developers who already have Go's install bin directory on `PATH`, this also works:

```sh
go install github.com/robbell5/llm-quota/cmd/llm-quota@latest
```

For the latest unreleased code from `main`, replace `@latest` with `@main`.

For a local repository smoke check without changing a wider shell setup, build and run the local binary from the repo root:

```sh
go build ./cmd/llm-quota
./llm-quota install-claude-hook
./llm-quota
```

`go build ./cmd/llm-quota` writes a local `llm-quota` binary in the current directory. If you remove that file or run from another directory, run `go build ./cmd/llm-quota` again before using `./llm-quota`.

## Set up Claude quota data

If you installed with Homebrew or with a Go install bin directory already on `PATH`, the explicit Claude setup command is:

```sh
llm-quota install-claude-hook
```

If you built the local binary, the explicit Claude setup command is:

```sh
./llm-quota install-claude-hook
```

The command installs or updates only the app-owned Claude statusline cache writer. It preserves unrelated Claude configuration, wraps any existing statusline command, preserves a symlinked `~/.claude/settings.json` by writing through to its target, and writes a backup path when it changes the Claude settings file. The cache writer is registered as the `statusLine.command` in `~/.claude/settings.json`; v1 does not create a separate hook script file.

After Claude runs, the hook writes local quota data to:

```text
~/.cache/llm-quota/claude.json
```

The pace forecast and sparkline are backed by a small local history file:

```text
~/.cache/llm-quota/history.json
```

It is written automatically while `llm-quota` runs and is safe to delete; the tool rebuilds it from new samples.

Normal `llm-quota` launches may also offer to install this app-owned cache writer on first run, but the documented setup path is the explicit command above.

### Uninstall Claude quota data setup

If `llm-quota` is on your `PATH`, remove the app-owned Claude quota capture setup with:

```sh
llm-quota uninstall-claude-hook
```

For the local build path, run:

```sh
./llm-quota uninstall-claude-hook
```

Uninstall removes the app-owned Claude statusline cache writer from `~/.claude/settings.json`, preserves unrelated Claude configuration, and restores the previously wrapped statusline command when one was present. It does not delete ~/.cache/llm-quota/claude.json or ~/.cache/llm-quota/state.json, so rerunning setup can re-enable quota capture without wiping local cache or state files.

## Run in a tmux pane

Start the always-running TUI in a dedicated pane:

```sh
llm-quota
```

For the local build path, run `./llm-quota` from the repository root instead.

The display is designed for a small pane around 50 columns and degrades for narrower panes; it expands to a richer grouped layout with a banded title bar and wall clock when the pane is wider. It refreshes quota data every 30 seconds while it remains in the foreground.

Codex quota data comes from local Codex session rollout data under:

```text
~/.codex/sessions
```

Open Codex locally when Codex rows need fresh rollout data. If you want the Codex rows to match the Codex app's live account view, launch with `--codex-live`; this opt-in path asks `codex app-server` for `account/rateLimits/read` on refresh and falls back to local rollouts if the live call is unavailable. `--codex-live` updates quota percentages and reset times only; Codex equivalent API-value estimates remain based on local rollout token usage.

## Keys

- `r` refreshes immediately.
- `v` cycles the provider view (both → Claude-only → Codex-only → both).
- `t` toggles the per-row sparkline and pace forecast line.
- `c` toggles the per-provider equivalent API-value clusters.
- `i` toggles Nerd Font icon mode.
- `q` quits.
- `Ctrl-C` quits.

## Display options

Set at launch with flags, or toggle live with keys:

| Flag | Key | Effect |
| ------ | --- | ------ |
| `--only=claude` / `--only=codex` | `v` | Show only one provider (`v` cycles both → Claude-only → Codex-only → both) |
| `--no-trend` | `t` | Hide the per-row sparkline + pace forecast line (one-line rows) |
| `--no-cost` | `c` | Hide the per-window equivalent API-value clusters |
| `--icons` | `i` | Use Nerd Font icons (requires a Nerd Font terminal; default is safe Unicode) |
| `--codex-live` | — | Poll live Codex quota via app-server; cost estimates remain rollout-based |
| `--help` / `-h` | — | Show usage and exit |

`--icons` can also be enabled at startup with the environment variable `LLM_QUOTA_ICONS=1`.

Other keys: `r` refresh, `q` quit. No setting can hide every provider.

## Troubleshooting

If the footer shows `Claude: run install-claude-hook`, run:

```sh
llm-quota install-claude-hook
```

For the local build path, run `./llm-quota install-claude-hook` from the repository root.

Then open Claude so the app-owned hook can write `~/.cache/llm-quota/claude.json`.

If you previously ran `llm-quota uninstall-claude-hook`, rerunning `llm-quota install-claude-hook` re-enables Claude quota capture.

If the footer shows `Claude: open Claude`, open Claude locally so it can produce newer local quota data.

If the footer shows `Codex: open Codex`, open Codex locally so rollout data appears under `~/.codex/sessions`.

If the Codex app shows fresher percentages than `llm-quota`, run `llm-quota --codex-live` to opt in to the same app-server rate-limit source. Without that flag, `llm-quota` stays local-only and shows the newest rollout snapshot. The live source does not include model/token details, so Codex equivalent API-value estimates may still lag until local rollout files include the matching token usage.

If the footer says data is old, such as `Claude data 2h old; open Claude` or `Codex data 1d old; open Codex`, the TUI is keeping last-known local data on screen. Open the named tool and press `r`, or wait for the next 30-second refresh, to pick up newer local data.

The TUI does not print private Claude or Codex payloads. It keeps recovery actions focused on the same user-facing hints shown in the footer.

## Scope

v1 defaults to a local-only foreground tmux-pane monitor. It persists a small local usage history (`~/.cache/llm-quota/history.json`) to power the in-pane pace forecast, burn rate, and trend sparkline. By default it does not use network or OAuth fallback, macOS Keychain reads, OS-level notifications/alerts, a daemon, multi-account support, demo mode, or fixture mode. The optional `--codex-live` flag uses the local `codex app-server` command to ask the Codex account endpoint for live rate-limit percentages and falls back to local rollout files when unavailable; use it only when matching the Codex app's live account view matters more than the default local-only steady state. The equivalent API-value figures are computed locally from the same transcript and rollout files the tool already reads.
