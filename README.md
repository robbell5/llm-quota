# llm-quota

`llm-quota` is a small foreground terminal UI for watching Claude Code and Codex subscription quota windows in a dedicated tmux pane. It reads local usage files only, refreshes while it runs, and shows placeholder rows with recovery hints when local data is missing or stale.

## What it shows

- Claude Code 5-hour usage
- Claude Code 7-day usage
- Codex 5-hour usage
- Codex 7-day usage

Each available row shows percent used, a colored progress bar, and the reset countdown. Missing rows stay visible so the pane remains useful while local data is being produced.

## Install

Choose one install path.

To install `llm-quota` on your `PATH`, run:

```sh
go install github.com/rob/llm-quota/cmd/llm-quota@latest
llm-quota install-claude-hook
llm-quota
```

For a local repository smoke check without changing a wider shell setup, build and run the local binary from the repo root:

```sh
go build ./cmd/llm-quota
./llm-quota install-claude-hook
./llm-quota
```

`go build ./cmd/llm-quota` writes a local `llm-quota` binary in the current directory. If you remove that file or run from another directory, run `go build ./cmd/llm-quota` again before using `./llm-quota`.

## Set up Claude quota data

If you installed with `go install`, the explicit Claude setup command is:

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

Normal `llm-quota` launches may also offer to install this app-owned cache writer on first run, but the documented setup path is the explicit command above.

## Run in a tmux pane

Start the always-running TUI in a dedicated pane:

```sh
llm-quota
```

For the local build path, run `./llm-quota` from the repository root instead.

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

For the local build path, run `./llm-quota install-claude-hook` from the repository root.

Then open Claude so the app-owned hook can write `~/.cache/llm-quota/claude.json`.

If the footer shows `Claude: open Claude`, open Claude locally so it can produce newer local quota data.

If the footer shows `Codex: open Codex`, open Codex locally so rollout data appears under `~/.codex/sessions`.

If the footer says data is old, such as `Claude data 2h old; open Claude` or `Codex data 1d old; open Codex`, the TUI is keeping last-known local data on screen. Open the named tool and press `r`, or wait for the next 30-second refresh, to pick up newer local data.

The TUI does not print private Claude or Codex payloads. It keeps recovery actions focused on the same user-facing hints shown in the footer.

## Scope

v1 is a local-only foreground tmux-pane monitor. It does not use network or OAuth fallback, macOS Keychain reads, statusline integration, a daemon, alerts, forecasting, demo mode, or fixture mode.
