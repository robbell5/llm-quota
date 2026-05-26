package trend

import (
	"strings"
	"testing"
)

func samplesFromPercents(percents ...float64) []Sample {
	out := make([]Sample, 0, len(percents))
	for i, p := range percents {
		out = append(out, Sample{CapturedAt: at(i), UsedPct: p, ResetsAt: at(10000)})
	}
	return out
}

func TestSparklineMapsLevels(t *testing.T) {
	got := Sparkline(samplesFromPercents(0, 100), 2)
	if got != "▁█" {
		t.Fatalf("Sparkline = %q, want '▁█'", got)
	}
}

func TestSparklineLeftPadsWhenFewerSamples(t *testing.T) {
	got := Sparkline(samplesFromPercents(100), 4)
	if got != "   █" {
		t.Fatalf("Sparkline = %q, want '   █' (left-padded)", got)
	}
	if len([]rune(got)) != 4 {
		t.Fatalf("width = %d runes, want 4", len([]rune(got)))
	}
}

func TestSparklineKeepsLastNSamples(t *testing.T) {
	got := Sparkline(samplesFromPercents(0, 0, 0, 100), 2)
	if got != "▁█" {
		t.Fatalf("Sparkline = %q, want '▁█' (last 2)", got)
	}
}

func TestSparklineEmpty(t *testing.T) {
	got := Sparkline(nil, 3)
	if got != strings.Repeat(" ", 3) {
		t.Fatalf("Sparkline(nil,3) = %q, want three spaces", got)
	}
}
