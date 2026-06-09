package cost

import (
	"testing"
	"time"
)

func ts(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestAggregateFiltersToWindowAndSums(t *testing.T) {
	p, _ := LoadPricing()
	start := ts("2026-05-28T10:00:00Z")
	now := ts("2026-05-28T15:00:00Z")
	entries := []entry{
		{ts: ts("2026-05-28T09:59:00Z"), model: "claude-opus-4-8", usage: Usage{Output: 1_000_000}}, // before window
		{ts: ts("2026-05-28T11:00:00Z"), model: "claude-opus-4-8", usage: Usage{Output: 1_000_000}}, // in window: $25
		{ts: ts("2026-05-28T16:00:00Z"), model: "claude-opus-4-8", usage: Usage{Output: 1_000_000}}, // after now
	}
	wc := aggregate(entries, start, now, p)
	approx(t, wc.Amount, 25.0)
	if wc.Estimated || wc.Incomplete {
		t.Fatalf("unexpected flags: %+v", wc)
	}
}

func TestAggregateMarksIncompleteOnUnknownModel(t *testing.T) {
	p, _ := LoadPricing()
	start := ts("2026-05-28T10:00:00Z")
	now := ts("2026-05-28T15:00:00Z")
	entries := []entry{
		{ts: ts("2026-05-28T11:00:00Z"), model: "claude-opus-4-8", usage: Usage{Output: 1_000_000}},
		{ts: ts("2026-05-28T12:00:00Z"), model: "claude-future-9", usage: Usage{Output: 1_000_000}},
	}
	wc := aggregate(entries, start, now, p)
	approx(t, wc.Amount, 25.0) // priced portion only
	if !wc.Incomplete {
		t.Fatalf("expected Incomplete=true")
	}
}

func TestAggregateMarksEstimatedForCodex(t *testing.T) {
	p, _ := LoadPricing()
	start := ts("2026-05-28T10:00:00Z")
	now := ts("2026-05-28T15:00:00Z")
	entries := []entry{{ts: ts("2026-05-28T11:00:00Z"), model: "gpt-5-codex", usage: Usage{Output: 1_000_000}}}
	wc := aggregate(entries, start, now, p)
	if !wc.Estimated {
		t.Fatalf("expected Estimated=true")
	}
	if wc.Incomplete {
		t.Fatalf("codex model should be known/priced, got Incomplete: %+v", wc)
	}
	if wc.Amount <= 0 {
		t.Fatalf("expected non-zero estimated amount, got %v", wc.Amount)
	}
}

func TestAggregateUnknownOnlyWindow(t *testing.T) {
	p, _ := LoadPricing()
	start := ts("2026-05-28T10:00:00Z")
	now := ts("2026-05-28T15:00:00Z")
	entries := []entry{{ts: ts("2026-05-28T11:00:00Z"), model: "totally-unknown", usage: Usage{Output: 1_000_000}}}
	wc := aggregate(entries, start, now, p)
	if wc.Amount != 0 || !wc.Incomplete {
		t.Fatalf("expected Amount=0 and Incomplete=true, got %+v", wc)
	}
}

func TestAggregateSkipsZeroUsageEntries(t *testing.T) {
	p, _ := LoadPricing()
	start := ts("2026-05-28T10:00:00Z")
	now := ts("2026-05-28T15:00:00Z")
	entries := []entry{
		{ts: ts("2026-05-28T11:00:00Z"), model: "claude-opus-4-8", usage: Usage{Output: 1_000_000}},
		{ts: ts("2026-05-28T12:00:00Z"), model: "<synthetic>", usage: Usage{}}, // zero usage: must be ignored
	}
	wc := aggregate(entries, start, now, p)
	approx(t, wc.Amount, 25.0)
	if wc.Incomplete {
		t.Fatalf("zero-usage unknown-model entry must not flag incomplete, got %+v", wc)
	}
}

func TestDedupKeepsFirstByID(t *testing.T) {
	in := []entry{
		{id: "req-1", model: "a", usage: Usage{Output: 10}},
		{id: "req-1", model: "a", usage: Usage{Output: 10}}, // duplicate
		{id: "", model: "b", usage: Usage{Output: 5}},       // no id: always kept
		{id: "", model: "c", usage: Usage{Output: 5}},
	}
	out := dedup(in)
	if len(out) != 3 {
		t.Fatalf("expected 3 entries after dedup, got %d", len(out))
	}
	if out[0].id != "req-1" || out[0].model != "a" {
		t.Fatalf("expected first kept entry to be the first req-1, got %+v", out[0])
	}
	if out[1].id != "" || out[1].model != "b" {
		t.Fatalf("expected second entry b (empty id), got %+v", out[1])
	}
	if out[2].id != "" || out[2].model != "c" {
		t.Fatalf("expected third entry c (empty id), got %+v", out[2])
	}
}
