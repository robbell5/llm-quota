package trend

import (
	"time"

	"github.com/robbell5/llm-quota/internal/sources"
)

// RateLookback is the window over which the recent burn rate is measured.
const RateLookback = 45 * time.Minute

// WindowDuration is the fixed length of a quota window by kind.
func WindowDuration(kind sources.WindowKind) time.Duration {
	switch kind {
	case sources.WindowFiveHour:
		return 5 * time.Hour
	case sources.WindowSevenDay, sources.WindowSonnetSevenDay:
		return 7 * 24 * time.Hour
	default:
		return 0
	}
}

// ElapsedFraction is how far through its window `now` is, clamped to [0,1].
func ElapsedFraction(kind sources.WindowKind, resetsAt, now time.Time) float64 {
	d := WindowDuration(kind)
	if d <= 0 {
		return 0
	}
	start := resetsAt.Add(-d)
	f := now.Sub(start).Seconds() / d.Seconds()
	if f < 0 {
		return 0
	}
	if f > 1 {
		return 1
	}
	return f
}

// Rate returns the burn rate in percent-per-hour and whether it was measured
// from history (true) or derived from the window average (false). samples must
// be the current epoch's points, chronological (oldest first). The result is
// clamped to >= 0.
func Rate(samples []Sample, now time.Time, lookback time.Duration, windowStart time.Time) (float64, bool) {
	if len(samples) == 0 {
		return 0, false
	}
	latest := samples[len(samples)-1]

	cutoff := now.Add(-lookback)
	var baseline *Sample
	for i := len(samples) - 1; i >= 0; i-- {
		if !samples[i].CapturedAt.After(cutoff) {
			baseline = &samples[i]
			break
		}
	}
	if baseline != nil {
		dt := latest.CapturedAt.Sub(baseline.CapturedAt).Hours()
		if dt > 0 {
			return clampRate((latest.UsedPct - baseline.UsedPct) / dt), true
		}
	}

	elapsed := now.Sub(windowStart).Hours()
	if elapsed <= 0 {
		return 0, false
	}
	return clampRate(latest.UsedPct / elapsed), false
}

func clampRate(r float64) float64 {
	if r < 0 {
		return 0
	}
	return r
}
