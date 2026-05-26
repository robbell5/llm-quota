package trend

import (
	"testing"
	"time"

	"github.com/robbell5/llm-quota/internal/sources"
)

func at(min int) time.Time {
	return time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC).Add(time.Duration(min) * time.Minute)
}

func TestKeyFormatsProductAndKind(t *testing.T) {
	if got := Key(sources.ProductClaude, sources.WindowFiveHour); got != "claude:five_hour" {
		t.Fatalf("Key = %q, want claude:five_hour", got)
	}
}

func TestAppendDedupsUnchangedReadings(t *testing.T) {
	h := NewHistory()
	reset := at(120)
	h.Append("k", Sample{CapturedAt: at(0), UsedPct: 10, ResetsAt: reset})
	h.Append("k", Sample{CapturedAt: at(5), UsedPct: 10, ResetsAt: reset}) // unchanged -> dropped
	h.Append("k", Sample{CapturedAt: at(10), UsedPct: 12, ResetsAt: reset})

	got := h.EpochSamples("k", reset)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2 (dedup), samples: %+v", len(got), got)
	}
	if got[1].UsedPct != 12 {
		t.Fatalf("last UsedPct = %v, want 12", got[1].UsedPct)
	}
}

func TestAppendClearsPriorEpochOnReset(t *testing.T) {
	h := NewHistory()
	old := at(0)
	newEpoch := at(600)
	h.Append("k", Sample{CapturedAt: at(0), UsedPct: 90, ResetsAt: old})
	h.Append("k", Sample{CapturedAt: at(601), UsedPct: 3, ResetsAt: newEpoch}) // window rolled over

	if n := len(h.EpochSamples("k", old)); n != 0 {
		t.Fatalf("old epoch should be pruned, got %d", n)
	}
	if n := len(h.EpochSamples("k", newEpoch)); n != 1 {
		t.Fatalf("new epoch len = %d, want 1", n)
	}
}

func TestAppendBoundsRingSize(t *testing.T) {
	h := NewHistory()
	reset := at(100000)
	for i := 0; i < maxSamplesPerWindow+20; i++ {
		h.Append("k", Sample{CapturedAt: at(i), UsedPct: float64(i % 100), ResetsAt: reset})
	}
	got := h.EpochSamples("k", reset)
	if len(got) > maxSamplesPerWindow {
		t.Fatalf("ring not bounded: %d > %d", len(got), maxSamplesPerWindow)
	}
	wantNewest := float64((maxSamplesPerWindow + 19) % 100)
	if got[len(got)-1].UsedPct != wantNewest {
		t.Fatalf("newest sample not retained: UsedPct=%v, want %v", got[len(got)-1].UsedPct, wantNewest)
	}
}
