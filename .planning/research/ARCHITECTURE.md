# Architecture Research

**Domain:** Small local-file-backed Go/Bubble Tea quota monitor
**Researched:** 2026-05-16
**Confidence:** HIGH for component boundaries and data flow; MEDIUM for exact Charm API
version until the project pins Bubble Tea/Bubbles/Lip Gloss major versions.

## Recommendation

Build `llm-quota` as a deliberately small foreground Bubble Tea program with three
internal seams:

1. source readers that turn local files into a shared quota-window shape;
2. a TUI model/update layer that owns refresh cadence and last-known-good state;
3. a view renderer that is pure enough to test with fixed input and golden files.

Do not introduce a daemon, service layer, database, watcher framework, network
client, or broad plugin architecture. The highest-risk behavior is not UI
complexity; it is tolerating missing, malformed, null, stale, or temporarily
unavailable local source data without crashing or blanking useful values.

## Standard Architecture

### System Overview

```text
┌──────────────────────────────────────────────────────────────────────┐
│ External local producers                                              │
│                                                                      │
│  Claude statusline script            Codex interactive sessions       │
│  ~/dotfiles/.../statusline.sh        ~/.codex/sessions/**/rollout*.jsonl
│           │                                      │                    │
│           │ atomic write                         │ append JSONL        │
│           ▼                                      ▼                    │
│  ~/.cache/llm-quota/claude.json       newest rollout JSONL            │
└───────────┬──────────────────────────────────────┬───────────────────┘
            │                                      │
            ▼                                      ▼
┌──────────────────────────────────────────────────────────────────────┐
│ internal/sources                                                     │
│                                                                      │
│  ClaudeReader.Fetch(now)              CodexReader.Fetch(now)          │
│  - read one cache JSON file           - find newest rollout file       │
│  - validate shape                     - scan for last usable event     │
│  - compute staleness                  - skip null rate_limits          │
│           │                                      │                    │
│           └──────────────┬───────────────────────┘                    │
│                          ▼                                            │
│                 []sources.Window + error                              │
└──────────────────────────┬───────────────────────────────────────────┘
                           │ refreshMsg
                           ▼
┌──────────────────────────────────────────────────────────────────────┐
│ internal/tui                                                         │
│                                                                      │
│  Model                                                               │
│  - last-known-good Claude windows                                    │
│  - last-known-good Codex windows                                     │
│  - source errors and hints                                           │
│  - terminal width/height                                             │
│  - clock dependency for deterministic rendering                      │
│                                                                      │
│  Update(msg)                                                         │
│  - tickMsg: refresh + schedule next tick                             │
│  - refreshMsg: merge successful data, preserve old data on errors     │
│  - KeyMsg/KeyPressMsg: q/Ctrl-C quit, r refresh                      │
│  - WindowSizeMsg: store dimensions                                   │
│                                                                      │
│  View() / Render(model, now)                                         │
│  - rows, bars, reset countdowns, footer hints                        │
│  - degrade layout for narrow panes                                   │
└──────────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Boundary |
|-----------|----------------|----------|
| `cmd/llm-quota/main.go` | Wire defaults and run the Bubble Tea program. | No parsing, rendering, or policy logic. |
| `internal/sources/window.go` | Define the small shared `Window` value and source identifiers. | No filesystem access. |
| `internal/sources/claude.go` | Read and validate the Claude statusline cache. | Local cache file only; no Keychain or network fallback. |
| `internal/sources/codex.go` | Locate the newest rollout JSONL and extract the last usable rate-limit event. | Local rollout files only; skip null events. |
| `internal/tui/model.go` | Hold durable UI state and injectable dependencies. | No formatting-heavy view code. |
| `internal/tui/update.go` | Own event handling, refresh commands, ticks, and last-known-good merge policy. | Does not parse source formats. |
| `internal/tui/view.go` | Convert model state to terminal output. | No filesystem reads or mutation. |
| `internal/tui/colors.go` | Copy Catppuccin Mocha colors used by this app. | No dependency on dotfiles script at runtime. |
| `testdata/` | Synthetic source files and golden render snapshots. | No secrets and no reads from the user's home directory. |

## Recommended Project Structure

```text
cmd/llm-quota/
└── main.go                 # defaults, tea.NewProgram, program run/error handling

internal/
├── sources/
│   ├── window.go           # shared Window and Source types
│   ├── claude.go           # cache reader
│   ├── claude_test.go      # cache fixtures, missing/malformed/stale cases
│   ├── codex.go            # rollout discovery and JSONL parser
│   └── codex_test.go       # newest-file and last-valid-event fixtures
└── tui/
    ├── model.go            # Bubble Tea model, message types, dependencies
    ├── update.go           # Init/Update, tick and refresh commands
    ├── update_test.go      # last-known-good and key/resize behavior
    ├── view.go             # pure-ish renderer
    ├── view_test.go        # fixed-width golden render tests
    └── colors.go           # local palette constants

testdata/
├── claude/
│   ├── valid.json
│   ├── malformed.json
│   └── stale.json
├── codex/
│   ├── rollout-old.jsonl
│   ├── rollout-new.jsonl
│   └── rollout-null-rate-limits.jsonl
└── golden/
    ├── all_available_80.txt
    ├── source_missing_80.txt
    └── narrow_28.txt
```

### Structure Rationale

- **`sources/` stays Bubble Tea-free:** source tests should exercise ordinary Go
  functions without terminal fixtures or event-loop setup.
- **`tui/` depends on source interfaces, not file paths:** tests can inject fake
  readers and scripted results to prove refresh failure behavior.
- **`main.go` owns real defaults:** home-directory expansion, default paths, and
  program options belong at the edge so tests never touch `~/.claude` or
  `~/.codex`.
- **`testdata/` uses synthetic fixtures:** quota files can contain account- or
  session-adjacent data, so tests should never copy real rollout files.

## Architectural Patterns

### Pattern 1: Bubble Tea MVU With I/O Isolated in Commands

**What:** Keep model mutation inside `Update`, rendering inside `View`, and file
I/O inside `tea.Cmd` functions that return typed messages.

**Why here:** Bubble Tea's documented model is `Init`, `Update`, and `View`, with
commands used for asynchronous work. This app has exactly two asynchronous
activities: initial/forced refreshes and periodic ticks.

**Example shape:**

```go
type sourceReader interface {
	Fetch(ctx context.Context, now time.Time) ([]sources.Window, error)
}

type refreshMsg struct {
	At     time.Time
	Claude sourceResult
	Codex  sourceResult
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.refreshCmd(), tickCmd(m.refreshEvery))
}
```

**Trade-off:** This keeps the app easy to test, but requires one small interface
and typed message structs. That is worthwhile because failure-tolerant refresh is
a product requirement, not incidental plumbing.

### Pattern 2: Last-Known-Good Merge Policy

**What:** Treat successful source results as replacements for that source's
window slice. Treat failed source results as an error update only; do not clear
previous windows.

**When to use:** Every refresh, including initial load, timer refresh, manual
refresh, and refresh triggered after resize if implemented.

**Policy:**

```text
if result.err == nil and result.windows valid:
    model.windows[source] = result.windows
    model.errors[source] = nil
else:
    model.errors[source] = result.err
    keep model.windows[source] unchanged
```

If no previous windows exist, the view renders placeholders and hints. If old
windows exist, the view renders them with staleness/error hints.

**Trade-off:** The screen may show old data for a while, but that is explicitly
better for this product than flashing blanks during transient parse/read errors.

### Pattern 3: Path-Injection for Source Tests

**What:** Construct readers with explicit paths in tests and default paths only
in `main.go`.

```go
reader := sources.NewClaudeReader(cachePath)
reader := sources.NewCodexReader(sessionsRoot)
```

**Why here:** It prevents accidental reads from Rob's real home directory during
tests and lets fixtures cover malformed, missing, and null-rate-limit cases.

### Pattern 4: Pure Rendering Helper

**What:** Make the rendered output a deterministic function of model state,
terminal width, and `now`.

```go
func Render(m Model, now time.Time) string
```

Bubble Tea's `View` method can call `Render(m, m.clock.Now())`, but tests should
call `Render` directly with fixed time and fixed width.

**Why here:** Reset countdowns, stale ages, and width-sensitive bars are all
otherwise flaky in golden tests.

### Pattern 5: Major-Version API Decision Before Implementation

**What:** Pin one Charm major version before coding the TUI. Current Context7
docs show Bubble Tea/Bubbles/Lip Gloss v2 APIs using `charm.land/.../v2`,
`tea.KeyPressMsg`, `tea.View`, and view fields for alt-screen. Older examples and
the project spec use v1-style `github.com/charmbracelet/...`, `tea.KeyMsg`,
string `View()`, and `tea.WithAltScreen()`.

**Recommendation:** Start the first implementation phase by deciding and pinning
the Charm stack in `go.mod`. Prefer current v2 APIs if available and stable for
the project, then update examples from the design spec accordingly. If choosing
v1 for simplicity, pin it intentionally and avoid mixing v2 snippets.

## Data Flow

### Refresh Flow

```text
Init / tick / r key
    ↓
refreshCmd
    ↓
parallel source reads
    ├── Claude cache JSON → []Window or error
    └── newest Codex rollout JSONL → []Window or error
    ↓
refreshMsg
    ↓
Update merges each source independently
    ├── success: replace that source's windows, clear its error
    └── failure: keep old windows, store error/hint
    ↓
View renders rows from model state
```

### Resize Flow

```text
Terminal/tmux pane resize
    ↓
tea.WindowSizeMsg
    ↓
Update stores width/height
    ↓
View recalculates row width and bar visibility
```

Resizing should not re-read files unless implementation shows a concrete need.
The requirement says refresh on resize, but the lower-risk interpretation is:
resize immediately re-renders with current state, and a refresh command can also
be batched if the spec owner wants data revalidation on resize. Avoid coupling
layout calculation to source reads.

### Source Data Normalization

```text
Claude five_hour/seven_day cache fields
Codex primary/secondary rollout fields
        ↓
shared Window values
        ↓
four display rows ordered as:
  Claude 5h, Claude 7d, Codex 5h, Codex 7d
```

Normalize source-specific vocabulary at the source boundary. The TUI should not
know that Codex calls its windows `primary` and `secondary`, or that Claude uses
`used_percentage` rather than `used_percent`.

## Build Order to Reduce Risk

1. **Pin Go module and Charm major versions.**
   - Risk reduced: API drift between Bubble Tea v1 examples and v2 docs.
   - Output: compiling empty Bubble Tea app with quit behavior.

2. **Define `sources.Window` and pure formatting helpers.**
   - Risk reduced: inconsistent labels, percentages, reset countdowns, and stale
     age handling.
   - Tests: reset countdown formats, negative reset shows `now`, thresholds.

3. **Build source readers with fixtures before the TUI.**
   - Risk reduced: unknown local file shapes and malformed/null source data.
   - Tests: Claude valid/missing/malformed/stale; Codex newest-file,
     last-valid-event, null `rate_limits`, no usable event.

4. **Build the pure renderer with fake model state.**
   - Risk reduced: narrow tmux-pane layout and placeholder/hint behavior.
   - Tests: golden render at normal width, missing first-run data, stale data,
     very narrow width with bars dropped.

5. **Implement Bubble Tea model/update loop with fake readers.**
   - Risk reduced: last-known-good merge policy, manual refresh, timer refresh,
     resize, and quit keys.
   - Tests: failed refresh preserves existing windows; initial failure produces
     placeholders; `r` returns a refresh command; tick schedules next tick.

6. **Wire `main.go` to real paths.**
   - Risk reduced: home-directory defaults and program startup behavior.
   - Keep this thin; most behavior should already be tested elsewhere.

7. **Separately change the Claude statusline in dotfiles.**
   - Risk reduced: repository-boundary confusion. This repo can ship and test
     against fixtures before the external producer is updated.

8. **Manual tmux-pane validation.**
   - Risk reduced: alt-screen preference, 30-second cadence feel, and actual
     narrow-pane readability.

## Test Seams

| Seam | Test Type | What It Proves |
|------|-----------|----------------|
| `NewClaudeReader(path)` | Unit with fixture files | Cache parsing, missing/malformed errors, staleness calculation. |
| `NewCodexReader(root)` | Unit with fixture tree | Newest rollout selection and last valid token-count extraction. |
| `sourceReader` interface in `tui.Model` | Update unit tests | Success/error merge behavior without filesystem reads. |
| `clock` or `now func() time.Time` | Unit/golden tests | Stable countdown and stale-age rendering. |
| `Render(model, now)` | Golden tests | Width behavior, colors stripped if needed, placeholder rows, footer hints. |
| `tickCmd(interval)` wrapper | Update tests | Tick handling can be observed without sleeping for 30 seconds. |

### Fixture Rules

- Fixtures must be synthetic and small.
- Do not copy real Claude cache or Codex rollout files into the repo.
- Include at least one malformed JSON fixture and one valid JSONL fixture with
  multiple token-count events so the parser proves it selects the last usable
  event.
- Strip ANSI before comparing golden render output unless the golden files are
  explicitly intended to lock color escape sequences.

## External Dotfiles Statusline Relationship

The Claude statusline change is an upstream local data producer, not a component
owned by this repository.

This repo owns:

- the expected cache-file contract;
- fixtures representing that contract;
- tolerant behavior when the cache is absent, malformed, stale, or old;
- README documentation explaining how the cache gets refreshed.

The dotfiles repo owns:

- adding the cache write to `~/dotfiles/claude/.claude/statusline-command.sh`;
- preserving existing statusline output and latency characteristics;
- atomic tmpfile-plus-rename behavior;
- any shell-specific tests or manual validation for that script.

Roadmap implication: do not block Go source-reader and TUI work on the dotfiles
change. Build this repo against fixtures first, then add a short integration
phase or checklist item to verify that the real statusline writes the documented
JSON shape. Commit/review the dotfiles change separately.

## Anti-Patterns

### Anti-Pattern 1: Daemonizing the Monitor

**What people do:** Add a background service, file watcher, IPC channel, or
persistent cache manager.

**Why it is wrong here:** The product is explicitly a foreground tmux-pane tool.
Bubble Tea's event loop plus a 30-second tick is enough.

**Do this instead:** Read local files during refresh commands and keep state in
the Bubble Tea model.

### Anti-Pattern 2: Network or Keychain Fallbacks

**What people do:** Add OAuth, Keychain reads, or HTTP fallbacks when local data
is missing.

**Why it is wrong here:** It expands the permission, failure, and platform matrix
for a glanceable local tool.

**Do this instead:** Render placeholders or last-known-good data with actionable
footer hints.

### Anti-Pattern 3: Letting Source Errors Clear Good Data

**What people do:** Replace rows with blanks whenever a refresh fails.

**Why it is wrong here:** Temporary file states are expected. Old numbers with a
warning are more useful than blank rows.

**Do this instead:** Merge per source and only replace windows on success.

### Anti-Pattern 4: Testing Against Real Home Directories

**What people do:** Unit tests read `~/.claude` or `~/.codex`.

**Why it is wrong here:** Tests become machine-specific and risk exposing local
session data.

**Do this instead:** Inject paths and use `testdata/` fixtures.

### Anti-Pattern 5: Mixing Charm v1 and v2 APIs

**What people do:** Combine `github.com/charmbracelet/...` imports and
`tea.WithAltScreen()` examples with v2-only `tea.View`, `tea.KeyPressMsg`, or
`charm.land/.../v2` imports.

**Why it is wrong here:** It creates compile churn during the earliest phase.

**Do this instead:** Pin the major version first and make all TUI code follow
that version consistently.

## Integration Points

### External Local Inputs

| Input | Integration Pattern | Notes |
|-------|---------------------|-------|
| Claude cache | Read one JSON file from `~/.cache/llm-quota/claude.json`. | Produced by dotfiles statusline; tolerate absence and malformed content. |
| Codex rollout JSONL | Glob sessions tree, pick newest rollout by mtime, scan for last usable event. | Skip `rate_limits: null`; no network fallback. |
| Terminal/tmux pane | Bubble Tea messages for keypresses and window size. | Store dimensions and render narrower layouts; no broad layout framework. |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| `main` ↔ `sources` | Constructor args with real default paths. | Keep defaults out of tests. |
| `sources` ↔ `tui` | `[]Window, error` through source-reader interface. | TUI handles source-independent last-known-good policy. |
| `update` ↔ `view` | Model fields only. | View does not trigger refreshes or parse files. |
| `view` ↔ tests | Render helper and golden strings. | Fixed time and width are required. |

## Sources

- Project context: `.planning/PROJECT.md` (HIGH confidence).
- Design spec:
  `docs/superpowers/specs/2026-05-16-llm-quota-tui-design.md`
  (HIGH confidence for desired behavior and source shapes).
- Context7 `/charmbracelet/bubbletea` docs (HIGH confidence): Bubble Tea uses
  `Init`, `Update`, and `View`; commands return messages; `tea.Batch` runs
  multiple commands; `tea.Tick` supports periodic messages; `WindowSizeMsg`
  carries terminal dimensions. Current docs also show v2 API changes.
- Context7 `/charmbracelet/bubbles` docs (HIGH confidence): progress component
  supports fixed-width/static progress rendering, and v2 uses setter methods and
  `charm.land/bubbles/v2` imports.
- Context7 `/charmbracelet/lipgloss` docs (HIGH confidence): Lip Gloss supports
  width measurement, joining/layout helpers, max-width constraints, and current
  v2 import/color API changes.

---

*Architecture research for: llm-quota*
*Researched: 2026-05-16*
