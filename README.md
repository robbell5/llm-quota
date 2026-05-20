# llm-quota

`llm-quota` is a small foreground terminal UI for watching Claude Code and Codex subscription quota windows in a dedicated tmux pane. It reads local usage files only, refreshes while it runs, and shows placeholder rows with recovery hints when local data is missing or stale.

## What it shows

- Claude Code 5-hour usage
- Claude Code 7-day usage
- Codex 5-hour usage
- Codex 7-day usage

Each available row shows percent used, a colored progress bar, and the reset countdown. Missing rows stay visible so the pane remains useful while local data is being produced.

## Install

Install the command with Go:

```sh
go install github.com/rob/llm-quota/cmd/llm-quota@latest
```

For a local repository smoke check without changing a wider shell setup, build the command from the repo root:

```sh
go build ./cmd/llm-quota
```

That writes a local `llm-quota` binary in the current directory.

## Set up Claude quota data

After installing, run the explicit Claude setup command:

```sh
llm-quota install-claude-hook
```

The command installs or updates only the app-owned Claude hook/cache writer. It preserves unrelated Claude configuration and writes a backup path when it changes the Claude settings file.

After Claude runs, the hook writes local quota data to:

```text
~/.cache/llm-quota/claude.json
```

Normal `llm-quota` launches may also offer to install this app-owned hook on first run, but the documented setup path is the explicit command above.

## Run in a tmux pane

Start the always-running TUI in a dedicated pane:

```sh
llm-quota
```

The display is designed for a small pane around 50 columns and degrades for narrower panes. It refreshes quota data every 30 seconds while it remains in the foreground.

Codex quota data comes from local Codex session rollout data under:

```text
~/.codex/sessions
```

Open Codex locally when Codex rows need fresh data.

## Keys

- `r` refreshes immediately.
- `q` quits.
- `Ctrl-C` quits.

## Troubleshooting

If the footer shows `Claude: run install-claude-hook`, run:

```sh
llm-quota install-claude-hook
```

Then open Claude so the app-owned hook can write `~/.cache/llm-quota/claude.json`.

If the footer shows `Claude: open Claude`, open Claude locally so it can produce newer local quota data.

If the footer shows `Codex: open Codex`, open Codex locally so rollout data appears under `~/.codex/sessions`.

If the footer says data is old, such as `Claude data 2h old; open Claude` or `Codex data 1d old; open Codex`, the TUI is keeping last-known local data on screen. Open the named tool and press `r`, or wait for the next 30-second refresh, to pick up newer local data.

The TUI does not print private Claude or Codex payloads. It keeps recovery actions focused on the same user-facing hints shown in the footer.

## Scope

v1 is a local-only foreground tmux-pane monitor. It does not use network or OAuth fallback, macOS Keychain reads, statusline integration, a daemon, alerts, forecasting, demo mode, or fixture mode.
