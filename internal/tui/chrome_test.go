package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/robbell5/llm-quota/internal/cost"
	"github.com/robbell5/llm-quota/internal/sources"
)

func TestTitleBandHasClock(t *testing.T) {
	at := time.Date(2026, 5, 27, 10, 37, 0, 0, time.Local)
	band := renderTitleBand(60, at, glyphsFor(false))
	if !strings.Contains(band, "LLM QUOTA") {
		t.Fatalf("title band missing title: %q", band)
	}
	if !strings.Contains(band, "10:37 AM") {
		t.Fatalf("title band missing clock: %q", band)
	}
}

func TestGroupHeaderShowsValueClusterWhenCostActive(t *testing.T) {
	now := time.Date(2026, 5, 29, 12, 0, 0, 0, time.Local)
	m := NewModel(WithClock(func() time.Time { return now }), WithCosts(
		map[sources.Product]map[sources.WindowKind]cost.WindowCost{
			sources.ProductClaude: {
				sources.WindowFiveHour: {Amount: 3.2},
				sources.WindowSevenDay: {Amount: 47.5},
			},
		}))
	if !m.costActive() {
		t.Fatal("precondition: cost should be active for this test")
	}
	header := renderGroupHeader(m, sources.ProductClaude, "CLAUDE", now, 46)
	plain := ansiEscapeRE.ReplaceAllString(header, "")
	if !strings.Contains(plain, "5h $3.20 · 7d $47.50") {
		t.Fatalf("header missing value cluster: %q", plain)
	}
	if strings.Contains(plain, "updated") {
		t.Fatalf("freshness should be relocated, not on header: %q", plain)
	}
}

func TestGroupHeaderDropsClusterWhenTooNarrow(t *testing.T) {
	now := time.Date(2026, 5, 29, 12, 0, 0, 0, time.Local)
	m := NewModel(WithClock(func() time.Time { return now }), WithCosts(
		map[sources.Product]map[sources.WindowKind]cost.WindowCost{
			sources.ProductClaude: {
				sources.WindowFiveHour: {Amount: 3.2},
				sources.WindowSevenDay: {Amount: 47.5},
			},
		}))
	header := renderGroupHeader(m, sources.ProductClaude, "CLAUDE", now, 8)
	plain := ansiEscapeRE.ReplaceAllString(header, "")
	if strings.Contains(plain, "$") {
		t.Fatalf("expected cluster dropped at narrow width (label only), got %q", plain)
	}
}

func TestGroupHeaderShowsFreshnessWhenCostInactive(t *testing.T) {
	now := time.Date(2026, 5, 29, 12, 0, 0, 0, time.Local)
	m := NewModel(WithClock(func() time.Time { return now })) // no WithCosts → cost inactive
	m.windows[sources.ProductClaude] = []sources.Window{{
		Product: sources.ProductClaude, Kind: sources.WindowFiveHour,
		CapturedAt: now.Add(-2 * time.Minute), ResetsAt: now.Add(time.Hour),
	}}
	header := renderGroupHeader(m, sources.ProductClaude, "CLAUDE", now, 60)
	plain := ansiEscapeRE.ReplaceAllString(header, "")
	if !strings.Contains(plain, "updated") {
		t.Fatalf("cost-inactive header should show freshness, got %q", plain)
	}
}

func TestFooterHasCostHint(t *testing.T) {
	footer := renderFooter(NewModel(), 80)
	if !strings.Contains(footer, "c cost") {
		t.Fatalf("footer missing cost hint: %q", footer)
	}
}

func TestGroupHeadersShownWhenWideEnough(t *testing.T) {
	now := time.Date(2026, 5, 27, 12, 0, 0, 0, time.Local)
	m := NewModel(WithClock(func() time.Time { return now }))
	m.windows[sources.ProductClaude] = []sources.Window{{
		Product: sources.ProductClaude, Kind: sources.WindowFiveHour,
		UsedPercent: 20, ResetsAt: now.Add(2 * time.Hour), CapturedAt: now,
	}}
	m.width = 80
	full := render(m)
	if !strings.Contains(full, "CLAUDE") {
		t.Fatalf("expected CLAUDE group header at width 80:\n%s", full)
	}
}
