# Effective-Dated Claude Pricing -- Design

**Date:** 2026-07-01
**Status:** Approved

## Why

Anthropic has shipped new Claude models and updated the published pricing table.
`llm-quota` embeds `internal/cost/pricing_table.json` so equivalent API-value
figures work offline, but the current table is now missing current Claude model
IDs and cannot represent a scheduled price change.

The immediate gap is Claude Sonnet 5. Anthropic's pricing docs list introductory
Sonnet 5 rates through August 31, 2026, then standard rates starting September
1, 2026. A static row would either become stale on September 1 or misprice
current usage before then. The tool should switch automatically and should
price mixed 7-day windows correctly when they span the changeover.

Official pricing references checked during brainstorming:

- Anthropic pricing docs: `https://platform.claude.com/docs/en/about-claude/pricing`
- Anthropic models overview: `https://platform.claude.com/docs/en/about-claude/models/overview`

## Scope

Add effective-dated model pricing in `internal/cost` and update the bundled
Claude pricing table for current first-party Claude API model IDs.

In scope:

1. Keep existing flat pricing rows valid.
2. Add a scheduled pricing shape for models with date-based rates.
3. Price each transcript entry using that entry's timestamp.
4. Add missing current Claude rows:
   - `claude-mythos-5`
   - `claude-opus-4-5`
   - `claude-sonnet-5`
5. Preserve exact-model lookup plus the existing bracketed context suffix
   fallback, for example `claude-sonnet-5[1m]` -> `claude-sonnet-5`.

Out of scope:

- Runtime network pricing fetches.
- Fuzzy model-family fallback.
- Batch, fast-mode, partner-cloud regional, or data-residency multipliers. The
  local transcript data path used by `llm-quota` does not expose enough request
  metadata to apply those safely.
- TUI layout changes.

## Pricing Data

Keep `Rates` as the token-rate unit:

```go
type Rates struct {
    Input        float64
    Output       float64
    CacheWrite5m float64
    CacheWrite1h float64
    CacheRead    float64
    Estimated    bool
}
```

`pricing_table.json` should accept two model-entry shapes:

```json
{
  "claude-opus-4-8": {
    "input": 5.0,
    "output": 25.0,
    "cache_write_5m": 6.25,
    "cache_write_1h": 10.0,
    "cache_read": 0.5
  },
  "claude-sonnet-5": {
    "rates": [
      {
        "input": 2.0,
        "output": 10.0,
        "cache_write_5m": 2.5,
        "cache_write_1h": 4.0,
        "cache_read": 0.2
      },
      {
        "effective_from": "2026-09-01T00:00:00Z",
        "input": 3.0,
        "output": 15.0,
        "cache_write_5m": 3.75,
        "cache_write_1h": 6.0,
        "cache_read": 0.3
      }
    ]
  }
}
```

For scheduled entries, a missing `effective_from` means the period applies from
the beginning of time. Period dates use RFC3339 timestamps in UTC. `LoadPricing`
normalizes both shapes into a private `modelPricing` representation with sorted
periods.

Rates to add:

| Model | Input | 5m cache write | 1h cache write | Cache read | Output |
|---|---:|---:|---:|---:|---:|
| `claude-mythos-5` | 10.0 | 12.5 | 20.0 | 1.0 | 50.0 |
| `claude-opus-4-5` | 5.0 | 6.25 | 10.0 | 0.5 | 25.0 |
| `claude-sonnet-5`, through 2026-08-31 | 2.0 | 2.5 | 4.0 | 0.2 | 10.0 |
| `claude-sonnet-5`, starting 2026-09-01 | 3.0 | 3.75 | 6.0 | 0.3 | 15.0 |

Claude rows are official published pricing and should not set `estimated`.
Codex/GPT rows remain flat entries with `estimated: true`.

## Data Flow

Today, `aggregate` filters entries by window and then calls `p.price(model,
usage)`. Change that to price at the entry timestamp:

```go
amount, known, estimated := p.price(e.model, e.usage, e.ts)
```

`Pricing.lookup(model, at)` should:

1. Try the exact model ID.
2. If the model has a bracketed suffix, strip from the first `[` and retry.
3. Select the last pricing period whose `effective_from` is less than or equal
   to `at`.
4. Return unknown if the model is absent or has no applicable period.

This makes a seven-day window crossing September 1 behave correctly: August
Sonnet 5 entries use introductory rates, while September entries use standard
rates in the same displayed total.

## Error Handling

Malformed embedded pricing JSON remains a `LoadPricing` error. The application
already handles that by degrading to an empty pricing table, which hides or marks
equivalent API-value figures rather than crashing.

Scheduled pricing validation should reject:

- invalid `effective_from` timestamps,
- scheduled models with no periods,
- scheduled models whose earliest period starts after the usage timestamp being
  priced.

Unknown non-zero usage remains incomplete and renders with the existing trailing
`*` marker. Do not add fuzzy fallback to avoid silently pricing a new model at
the wrong rate.

## Tests

Add focused unit coverage in `internal/cost`:

1. `LoadPricing` still accepts existing flat rows and includes the expected
   seeded models.
2. `claude-mythos-5`, `claude-opus-4-5`, and `claude-sonnet-5` are seeded.
3. `claude-sonnet-5` prices 1M input + 1M output at `$12.00` before
   `2026-09-01T00:00:00Z`.
4. `claude-sonnet-5` prices the same usage at `$18.00` on or after
   `2026-09-01T00:00:00Z`.
5. A mixed aggregate window with one August Sonnet 5 entry and one September
   Sonnet 5 entry sums both rates.
6. `claude-sonnet-5[1m]` resolves through the context-suffix fallback and uses
   the scheduled rate for the entry timestamp.
7. Unknown models with non-zero usage still set `Incomplete`.

Run the normal Go gate after implementation: `gofmt`, `go test ./...`, and
`go vet ./...`.

## Acceptance

Implementation is complete when:

1. The bundled table contains the current Claude rows above.
2. Sonnet 5 automatically switches at `2026-09-01T00:00:00Z`.
3. Mixed windows use per-entry timestamps, not app startup time.
4. Existing Codex estimated pricing behavior is unchanged.
5. The full Go verification gate passes.

## Decisions

- Use effective-dated rates per model rather than app-start date switching,
  because quota windows can span a pricing boundary.
- Keep the JSON table human-editable and backward-compatible by allowing both
  flat and scheduled model entries.
- Keep exact model IDs as the pricing contract. Unknown models should stay
  visible via the incomplete marker instead of being guessed.
