package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/robbell5/llm-quota/internal/cost"
	"github.com/robbell5/llm-quota/internal/sources"
)

func TestFormatValue(t *testing.T) {
	cases := []struct {
		wc   cost.WindowCost
		want string
	}{
		{cost.WindowCost{Amount: 3.2}, "$3.20"},
		{cost.WindowCost{Amount: 47.5}, "$47.50"},
		{cost.WindowCost{Amount: 0}, "$0.00"},
		{cost.WindowCost{Amount: 0.9, Estimated: true}, "~$0.90"},
		{cost.WindowCost{Amount: 3.2, Incomplete: true}, "$3.20*"},
		{cost.WindowCost{Amount: 1234, Estimated: true, Incomplete: true}, "~$1.2k*"},
	}
	for _, c := range cases {
		if got := formatValue(c.wc); got != c.want {
			t.Fatalf("formatValue(%+v) = %q, want %q", c.wc, got, c.want)
		}
	}
}

func TestValueClusterText(t *testing.T) {
	m := NewModel(WithCosts(map[sources.Product]map[sources.WindowKind]cost.WindowCost{
		sources.ProductClaude: {
			sources.WindowFiveHour: {Amount: 3.2},
			sources.WindowSevenDay: {Amount: 47.5},
		},
	}))
	got, ok := valueCluster(m, sources.ProductClaude)
	if !ok || got != "5h $3.20 · 7d $47.50" {
		t.Fatalf("valueCluster = %q ok=%v", got, ok)
	}
	if _, ok := valueCluster(m, sources.ProductCodex); ok {
		t.Fatalf("codex has no costs → ok should be false")
	}

	// only the 7d window present — the 5h term must be dropped, ok still true.
	m2 := NewModel(WithCosts(map[sources.Product]map[sources.WindowKind]cost.WindowCost{
		sources.ProductClaude: {sources.WindowSevenDay: {Amount: 47.5}},
	}))
	got2, ok2 := valueCluster(m2, sources.ProductClaude)
	if !ok2 || got2 != "7d $47.50" {
		t.Fatalf("single-window valueCluster = %q ok=%v", got2, ok2)
	}
}

func TestFreshnessLineMarksStale(t *testing.T) {
	now := time.Date(2026, 5, 29, 12, 0, 0, 0, time.Local)
	m := NewModel(WithClock(func() time.Time { return now }), WithCosts(
		map[sources.Product]map[sources.WindowKind]cost.WindowCost{
			sources.ProductClaude: {sources.WindowFiveHour: {Amount: 3.2}},
		}))
	m.windows[sources.ProductClaude] = []sources.Window{{
		Product: sources.ProductClaude, Kind: sources.WindowFiveHour,
		CapturedAt: now.Add(-2 * time.Hour), ResetsAt: now.Add(time.Hour),
	}}
	line, ok := renderFreshnessLine(m, now, 60)
	if !ok {
		t.Fatalf("expected a freshness line")
	}
	plain := ansiEscapeRE.ReplaceAllString(line, "")
	if !strings.Contains(plain, "old") {
		t.Fatalf("expected stale 'old' marker: %q", plain)
	}
}

func TestFreshnessLineMentionsProductsAndLegend(t *testing.T) {
	now := time.Date(2026, 5, 29, 12, 0, 0, 0, time.Local)
	m := NewModel(WithClock(func() time.Time { return now }), WithCosts(
		map[sources.Product]map[sources.WindowKind]cost.WindowCost{
			sources.ProductCodex: {sources.WindowFiveHour: {Amount: 0.9, Estimated: true}},
		}))
	m.windows[sources.ProductCodex] = []sources.Window{{
		Product: sources.ProductCodex, Kind: sources.WindowFiveHour,
		CapturedAt: now.Add(-2 * time.Minute), ResetsAt: now.Add(time.Hour),
	}}
	line, ok := renderFreshnessLine(m, now, 46)
	if !ok {
		t.Fatalf("expected a freshness line")
	}
	plain := ansiEscapeRE.ReplaceAllString(line, "")
	if !strings.Contains(plain, "Codex") || !strings.Contains(plain, "~ est") {
		t.Fatalf("freshness line missing product or legend: %q", plain)
	}
}
