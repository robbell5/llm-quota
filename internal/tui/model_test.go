package tui

import (
	"testing"

	"github.com/robbell5/llm-quota/internal/cost"
	"github.com/robbell5/llm-quota/internal/sources"
)

func TestNewModelCreatesOneBarPerRowSpec(t *testing.T) {
	m := NewModel()
	if len(m.bars) != len(quotaRowSpecs) {
		t.Fatalf("expected %d bars, got %d", len(quotaRowSpecs), len(m.bars))
	}
	if len(m.highlightUntil) != len(quotaRowSpecs) {
		t.Fatalf("expected %d highlight timers, got %d", len(quotaRowSpecs), len(m.highlightUntil))
	}
	for i, b := range m.bars {
		if b.target != -1 {
			t.Fatalf("bar %d should start with sentinel target -1, got %v", i, b.target)
		}
	}
}

func TestCostActiveRequiresVisibleAndData(t *testing.T) {
	m := NewModel()
	if m.costActive() {
		t.Fatalf("no cost data → not active")
	}
	m = NewModel(WithCosts(map[sources.Product]map[sources.WindowKind]cost.WindowCost{
		sources.ProductClaude: {sources.WindowFiveHour: {Amount: 3.2}},
	}))
	if !m.costActive() {
		t.Fatalf("visible + data → active")
	}
	m.prefs.HideCost = true
	if m.costActive() {
		t.Fatalf("hidden → not active even with data")
	}
}
