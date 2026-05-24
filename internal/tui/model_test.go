package tui

import "testing"

func TestNewModelCreatesOneBarPerRowSpec(t *testing.T) {
	m := NewModel()
	if len(m.bars) != len(quotaRowSpecs) {
		t.Fatalf("expected %d bars, got %d", len(quotaRowSpecs), len(m.bars))
	}
	if len(m.barTargets) != len(quotaRowSpecs) {
		t.Fatalf("expected %d bar targets, got %d", len(quotaRowSpecs), len(m.barTargets))
	}
}
