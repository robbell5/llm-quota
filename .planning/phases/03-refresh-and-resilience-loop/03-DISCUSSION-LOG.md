# Phase 3: Refresh and Resilience Loop - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md -- this log preserves the alternatives considered.

**Date:** 2026-05-19
**Phase:** 3-Refresh and Resilience Loop
**Areas discussed:** Stale warning policy, Last-good failure state, Resize refresh semantics, Refresh interaction feel

---

## Stale Warning Policy

| Question | Selected | Alternatives Considered |
|----------|----------|-------------------------|
| What freshness threshold should Phase 3 use before marking displayed quota data stale? | One hour | Thirty minutes; Source-specific; You decide |
| Should Claude and Codex use the same stale threshold in the TUI? | Same threshold | Claude only; Different thresholds; You decide |
| When data is stale but still otherwise valid, should the TUI keep showing the quota values? | Show values with warning | Replace with placeholder; Show only if recent success; You decide |
| Where should stale information be surfaced in Phase 3, before final Phase 4 rendering polish? | State plus tests | Minimal footer hint; Per-row marker now; You decide |

**Notes:** Stale policy should be locked in model behavior and tests first. Final stale warning copy and styling remain Phase 4 work.

---

## Last-Good Failure State

| Question | Selected | Alternatives Considered |
|----------|----------|-------------------------|
| When a source refresh fails after previous good data exists, what should the model preserve? | Per-source preserve | All-or-nothing screen; Per-window preserve; You decide |
| If both sources fail on the very first refresh and there is no last-good data, what should Phase 3 expect? | Existing placeholders | New error screen; Exit with error; You decide |
| How should a temporary failure be represented internally for later rendering? | Keep typed source errors | Store plain strings; Store only stale flag; You decide |
| How prominent should the failure warning be during Phase 3? | State/test only | Footer hint now; Row-level warning now; You decide |

**Notes:** The user prioritized independent source resilience and preserving useful data over introducing new visible error UI in Phase 3.

---

## Resize Refresh Semantics

| Question | Selected | Alternatives Considered |
|----------|----------|-------------------------|
| Should a terminal resize trigger source file reads, or only rerender the current model? | Rerender only | Rerender plus refresh; Refresh only if stale; You decide |
| How should Phase 3 interpret the existing requirement that data refreshes on resize? | Layout refresh | Strict data refresh; Defer to Phase 4; You decide |
| Should resize affect the 30-second refresh timer schedule? | No timer change | Reset timer after resize; Trigger then reset; You decide |
| What should tests prove about resize in Phase 3? | Stores size only | Stores size and command; No resize tests yet; You decide |

**Notes:** Resize is a layout event for Phase 3, not a source-read trigger.

---

## Refresh Interaction Feel

| Question | Selected | Alternatives Considered |
|----------|----------|-------------------------|
| Should the TUI run an initial source refresh immediately when it starts? | Immediate refresh | Wait for first tick; Manual only initially; You decide |
| When the user presses r, should that reset the next scheduled 30-second refresh? | Do not reset | Reset schedule; Skip next tick if recent; You decide |
| If r is pressed while a refresh is already running, what behavior should Phase 3 prefer? | Coalesce/ignore duplicate | Start another refresh; Queue one follow-up; You decide |
| Should Phase 3 add visible refresh status like 'refreshing...' or a last-updated timestamp? | No visible status yet | Minimal footer status; Last-updated timestamp; You decide |

**Notes:** Phase 3 should make refresh behavior reliable and testable without pulling final display polish forward from Phase 4.

---

## Agent Discretion

No discussion answers delegated decisions with "you decide." Exact implementation names and test fixture naming remain planner/executor discretion.

## Deferred Ideas

None.
