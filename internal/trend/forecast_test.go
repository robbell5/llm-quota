package trend

import (
	"testing"
	"time"
)

func TestForecastAtRiskWhenExhaustsBeforeReset(t *testing.T) {
	now := at(0)
	resetsAt := now.Add(2 * time.Hour)
	// 40% used, 60% remaining, 40%/hr -> full in 1.5h < 2h to reset.
	f := ComputeForecast(40, 40, now, resetsAt)
	if !f.AtRisk || f.Arrow != '↑' {
		t.Fatalf("expected at-risk with up arrow, got %+v", f)
	}
	if f.Status != "full in 1h 30m" {
		t.Fatalf("Status = %q, want 'full in 1h 30m'", f.Status)
	}
}

func TestForecastSafeShowsProjectedByReset(t *testing.T) {
	now := at(0)
	resetsAt := now.Add(10 * time.Hour)
	// 20% used, 4%/hr over 10h -> +40% -> ~60% by reset, never hits 100.
	f := ComputeForecast(20, 4, now, resetsAt)
	if f.AtRisk || f.Arrow != '→' {
		t.Fatalf("expected safe with right arrow, got %+v", f)
	}
	if f.Status != "~60% by reset" {
		t.Fatalf("Status = %q, want '~60%% by reset'", f.Status)
	}
}

func TestForecastZeroRateProjectsFlat(t *testing.T) {
	now := at(0)
	f := ComputeForecast(33, 0, now, now.Add(time.Hour))
	if f.AtRisk || f.Status != "~33% by reset" {
		t.Fatalf("zero-rate: got %+v", f)
	}
}

func TestForecastMaxedShowsFull(t *testing.T) {
	now := at(0)
	f := ComputeForecast(100, 5, now, now.Add(time.Hour))
	if f.AtRisk || f.Status != "full" {
		t.Fatalf("maxed: got %+v, want Status 'full' and AtRisk false", f)
	}
}

func TestForecastResetPassedHasNoProjection(t *testing.T) {
	now := at(0)
	f := ComputeForecast(50, 10, now, now.Add(-time.Minute))
	if f.Status != "" {
		t.Fatalf("reset-passed: Status = %q, want empty", f.Status)
	}
}

func TestShortDuration(t *testing.T) {
	cases := []struct {
		hours float64
		want  string
	}{
		{0.5, "30m"},
		{2.25, "2h 15m"},
		{50, "2d"},
	}
	for _, c := range cases {
		if got := shortDuration(c.hours); got != c.want {
			t.Fatalf("shortDuration(%v) = %q, want %q", c.hours, got, c.want)
		}
	}
}
