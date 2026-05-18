# Phase 2: Standalone Local Data Sources - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-18
**Phase:** 2-Standalone Local Data Sources
**Areas discussed:** Setup prompt flow, hook safety policy, reader tolerance, Codex rollout choice

---

## Setup Prompt Flow

### When should `llm-quota` offer to install the Claude hook/cache writer?

| Option | Description | Selected |
|--------|-------------|----------|
| First launch + command | Recommended: normal launch offers setup when Claude cache is missing, and an explicit command can install/update later. | yes |
| Explicit command only | Normal TUI never prompts; users run setup manually. Less intrusive, but weaker first-run setup. | |
| First launch only | Simpler visible path, but no direct way to update/reinstall just the hook later. | |
| You decide | Let downstream agents choose the smallest implementation that satisfies the requirements. | |

**User's choice:** First launch + command
**Notes:** The explicit command is narrowed to `install-claude-hook`.

### What happens after the user declines the first-launch prompt?

| Option | Description | Selected |
|--------|-------------|----------|
| Ask each launch | Recommended for v1: no extra config state, and the user can keep declining while the TUI still runs with placeholders. | |
| Remember decline | Avoids repeated prompts, but adds a persistent preference/sentinel that downstream agents must design and test. | yes |
| Stop prompting | Only show footer hints after the first decline in that process; explicit command handles future installs. | |
| You decide | Let agents choose based on minimal implementation and user experience constraints. | |

**User's choice:** Remember decline
**Notes:** Normal launches should not repeatedly interrupt after a decline.

### Where should the first-launch permission prompt appear?

| Option | Description | Selected |
|--------|-------------|----------|
| Before TUI starts | Recommended: a plain terminal prompt before entering alt-screen; easiest to keep permission explicit and avoid half-rendered setup UI. | yes |
| Inside TUI screen | Keeps everything in Bubble Tea, but requires a setup state and key handling before source rows exist. | |
| Install command only prompt | Normal launch only shows a hint; the prompt appears when the user runs setup explicitly. | |
| You decide | Let agents choose the smallest robust permission flow. | |

**User's choice:** Before TUI starts
**Notes:** The prompt completes before Bubble Tea enters alt-screen.

### Which explicit setup command surface should Phase 2 add?

| Option | Description | Selected |
|--------|-------------|----------|
| install-claude-hook | Recommended: precise scope, matches CLD-03, and avoids implying broader install behavior before docs/install phase. | yes |
| install alias too | Support `install` plus `install-claude-hook`; friendlier but adds command surface earlier. | |
| install only | Shorter command, but less clear that only Claude hook configuration is affected. | |
| You decide | Let agents pick the least surprising command name. | |

**User's choice:** install-claude-hook
**Notes:** No broader `install` alias is selected for Phase 2.

---

## Hook Safety Policy

### What should count as `llm-quota`-owned?

| Option | Description | Selected |
|--------|-------------|----------|
| Managed marker only | Recommended: only create/update entries carrying an explicit `llm-quota` managed marker/name; preserve everything else untouched. | yes |
| Path match allowed | Treat a hook command/path under the `llm-quota` install/cache location as owned even without a marker. | |
| Any quota hook | Broader: replace any hook that appears to write quota data. Riskier for user-owned config. | |
| You decide | Let agents choose the safest ownership test. | |

**User's choice:** Managed marker only
**Notes:** Unknown or user-owned hook entries must not be modified as owned.

### What if unrelated Claude hook entries already exist?

| Option | Description | Selected |
|--------|-------------|----------|
| Append safely | Recommended: preserve unrelated hooks and add/update only the managed `llm-quota` entry when the config shape is understood. | yes |
| Ask before merge | Show a warning and require confirmation before adding alongside existing hooks. | |
| Abort on existing | Do nothing if any existing hook config is present; safest, but setup may often fail. | |
| You decide | Let agents pick the safest behavior after inspecting Claude config shape. | |

**User's choice:** Append safely
**Notes:** Preserve unrelated hooks when the config can be understood.

### Should the installer create a backup before changing Claude configuration?

| Option | Description | Selected |
|--------|-------------|----------|
| Backup on change | Recommended: create a timestamped backup only when a config write is actually needed. | yes |
| Always backup | Back up even when no effective change is made. More files, simpler story. | |
| No backup | Rely on conservative edits and tests only. Smallest implementation, weaker recovery. | |
| You decide | Let agents choose based on the config file mechanics. | |

**User's choice:** Backup on change
**Notes:** No backup is needed for no-op installs.

### How should `install-claude-hook` behave when a managed hook already exists?

| Option | Description | Selected |
|--------|-------------|----------|
| Update in place | Recommended: idempotently update the managed entry and report whether it changed. | yes |
| Ask before update | Require confirmation before changing an existing managed hook. | |
| Leave unchanged | If an owned hook exists, do not update it unless a future command is added. | |
| You decide | Let agents choose the least surprising idempotent behavior. | |

**User's choice:** Update in place
**Notes:** The install command should be safe to run repeatedly.

---

## Reader Tolerance

### What if only one Claude cache window is valid?

| Option | Description | Selected |
|--------|-------------|----------|
| Reject source | Recommended: treat the cache contract as incomplete and show Claude placeholders until both windows are valid. | yes |
| Return partial | Show the valid Claude window and placeholder only the broken one. More useful, but adds partial-state complexity. | |
| Use zero values | Fill missing window fields with zero/now. Simple, but can make bad data look real. | |
| You decide | Let agents choose based on tests and renderer complexity. | |

**User's choice:** Reject source
**Notes:** Do not invent or partially render Claude cache data in Phase 2.

### How tolerant should the Codex JSONL reader be?

| Option | Description | Selected |
|--------|-------------|----------|
| Skip bad events | Recommended: ignore unrelated events, null limits, and malformed trailing/individual lines; use the last valid rate-limit event if present. | yes |
| Fail on malformed | Ignore unrelated events and null limits, but any malformed JSON line makes the source unavailable. | |
| Strict event shape | Require every candidate event to match exactly; safest contract, but brittle against Codex changes. | |
| You decide | Let agents choose the most robust parser behavior. | |

**User's choice:** Skip bad events
**Notes:** The reader should handle append/log noise pragmatically.

### Should stale Claude cache be data or an error?

| Option | Description | Selected |
|--------|-------------|----------|
| Return stale data | Recommended: return windows with age/stale metadata so the TUI can warn without blanking data. | yes |
| Return error | Treat stale cache as unavailable; simpler source contract, but loses useful old quota values. | |
| Both data + warning | Return windows plus a typed warning separate from hard errors. More expressive, more model state. | |
| You decide | Let agents pick the cleanest source contract. | |

**User's choice:** Return stale data
**Notes:** Staleness should be rendered as a warning later, not as missing data.

### How much should source errors be classified?

| Option | Description | Selected |
|--------|-------------|----------|
| Typed categories | Recommended: distinguish missing, malformed, stale/no usable event, and permission/read errors so hints stay useful. | yes |
| Plain errors | Readers return ordinary errors; the view shows generic source-missing hints. | |
| Detailed raw errors | Expose path/parser details in UI hints. Helpful for debugging, noisier and riskier in a tiny pane. | |
| You decide | Let agents choose the minimal classification needed for tests and hints. | |

**User's choice:** Typed categories
**Notes:** Footer hints need enough structure to be concise and useful.

---

## Codex Rollout Choice

### If the newest rollout has no usable rate-limit event, should older files be checked?

| Option | Description | Selected |
|--------|-------------|----------|
| Strict newest only | Matches the current design: if the newest file has no usable event, Codex is unavailable until a better rollout exists. | |
| Fallback to older usable | Use the newest rollout file that contains a valid rate-limit event. More useful after `codex exec`, but less literal. | yes |
| Fallback with warning | Use older usable data but classify it so the footer can say Codex data came from an older rollout. | |
| You decide | Let agents choose after balancing usefulness against the requirement wording. | |

**User's choice:** Fallback to older usable
**Notes:** This intentionally improves on the draft design's strict newest-file behavior.

### How broad should the rollout search be?

| Option | Description | Selected |
|--------|-------------|----------|
| All sessions | Recommended: recursively consider rollout JSONL files across the sessions tree; simple and faithful to the spec. | yes |
| Recent only | Limit search to recent date directories for speed. Requires choosing a cutoff and can miss useful data. | |
| Latest date dir | Only scan the newest dated session directory. Faster, but brittle around date boundaries. | |
| You decide | Let agents pick based on expected file count and testability. | |

**User's choice:** All sessions
**Notes:** Search the whole sessions tree.

### What should define newest rollout file?

| Option | Description | Selected |
|--------|-------------|----------|
| File mtime | Recommended: matches the design and avoids parsing timestamp conventions from private filenames. | yes |
| Filename timestamp | Use the rollout filename timestamp. Stable if naming stays consistent, brittle if Codex changes it. | |
| Newest valid event | Compare parsed event data across files. Most semantically accurate, more parser work. | |
| You decide | Let agents choose the simplest robust ordering. | |

**User's choice:** File mtime
**Notes:** Private filename timestamp parsing is not needed.

### Should `plan_type` be preserved?

| Option | Description | Selected |
|--------|-------------|----------|
| Preserve plan_type | Recommended: carry optional plan metadata alongside windows so Phase 4 can render it without reparsing source-specific details. | yes |
| Ignore metadata | Keep Phase 2 output to quota windows only. Simpler, but the footer plan detail from the design may be dropped or delayed. | |
| Raw metadata map | Expose flexible source metadata. Future-proof, but loose and less testable. | |
| You decide | Let agents decide the cleanest shared result shape. | |

**User's choice:** Preserve plan_type
**Notes:** Preserve only the useful typed metadata, not a broad raw map.

---

## Agent Discretion

No areas were delegated with "you decide."

## Deferred Ideas

None.
