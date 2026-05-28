package tui

import "testing"

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
