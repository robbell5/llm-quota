package trend

import (
	"math"
	"testing"
	"time"

	"github.com/robbell5/llm-quota/internal/sources"
)

func TestWindowDuration(t *testing.T) {
	cases := map[sources.WindowKind]time.Duration{
		sources.WindowFiveHour:       5 * time.Hour,
		sources.WindowSevenDay:       7 * 24 * time.Hour,
		sources.WindowSonnetSevenDay: 7 * 24 * time.Hour,
	}
	for kind, want := range cases {
		if got := WindowDuration(kind); got != want {
			t.Fatalf("WindowDuration(%s) = %v, want %v", kind, got, want)
		}
	}
}

func TestElapsedFractionClamps(t *testing.T) {
	now := at(0)
	resetsAt := now.Add(2 * time.Hour) // 5h window, 3h elapsed -> 0.6
	if got := ElapsedFraction(sources.WindowFiveHour, resetsAt, now); math.Abs(got-0.6) > 0.001 {
		t.Fatalf("fraction = %v, want ~0.6", got)
	}
	past := now.Add(-time.Hour) // already reset -> clamp to 1
	if got := ElapsedFraction(sources.WindowFiveHour, past, now); got != 1 {
		t.Fatalf("fraction = %v, want 1 (clamped)", got)
	}
}

func TestRateMeasuredFromLookbackBaseline(t *testing.T) {
	reset := at(1000)
	samples := []Sample{
		{CapturedAt: at(0), UsedPct: 10, ResetsAt: reset},
		{CapturedAt: at(60), UsedPct: 28, ResetsAt: reset}, // 18% over 60 min from baseline
	}
	now := at(60)
	windowStart := reset.Add(-WindowDuration(sources.WindowFiveHour))
	rate, measured := Rate(samples, now, 45*time.Minute, windowStart)
	if !measured {
		t.Fatalf("expected measured rate")
	}
	// baseline = newest sample <= now-45m = at(0) (at(60) is too recent).
	// rate = (28-10) / 1h = 18%/hr.
	if math.Abs(rate-18) > 0.001 {
		t.Fatalf("rate = %v, want 18", rate)
	}
}

func TestRateFallsBackToWindowAverage(t *testing.T) {
	reset := at(180)
	// Only a recent reading; nothing older than now-45m.
	samples := []Sample{{CapturedAt: at(170), UsedPct: 30, ResetsAt: reset}}
	now := at(180)
	windowStart := reset.Add(-WindowDuration(sources.WindowFiveHour)) // 5h before reset
	rate, measured := Rate(samples, now, 45*time.Minute, windowStart)
	if measured {
		t.Fatalf("expected window-average fallback (measured=false)")
	}
	// windowStart = reset-5h; now=at(180)=reset; elapsed=5h; rate=30/5=6%/hr.
	if math.Abs(rate-6) > 0.001 {
		t.Fatalf("rate = %v, want 6", rate)
	}
}

func TestRateEmptySamplesIsZero(t *testing.T) {
	rate, measured := Rate(nil, at(0), 45*time.Minute, at(-300))
	if rate != 0 || measured {
		t.Fatalf("empty samples: got (%v,%v), want (0,false)", rate, measured)
	}
}
