package tui

import (
	"testing"

	"github.com/robbell5/llm-quota/internal/sources"
)

func TestVisibilityNextCycles(t *testing.T) {
	got := []Visibility{}
	v := VisibilityBoth
	for i := 0; i < 4; i++ {
		got = append(got, v)
		v = v.next()
	}
	want := []Visibility{VisibilityBoth, VisibilityClaudeOnly, VisibilityCodexOnly, VisibilityBoth}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("step %d: expected %v, got %v", i, want[i], got[i])
		}
	}
}

func TestVisibilityShows(t *testing.T) {
	cases := []struct {
		v      Visibility
		claude bool
		codex  bool
	}{
		{VisibilityBoth, true, true},
		{VisibilityClaudeOnly, true, false},
		{VisibilityCodexOnly, false, true},
	}
	for _, c := range cases {
		if c.v.shows(sources.ProductClaude) != c.claude {
			t.Errorf("%v: claude visibility expected %v", c.v, c.claude)
		}
		if c.v.shows(sources.ProductCodex) != c.codex {
			t.Errorf("%v: codex visibility expected %v", c.v, c.codex)
		}
	}
}

func TestWithDisplayPrefs(t *testing.T) {
	m := NewModel(WithDisplayPrefs(DisplayPrefs{Visibility: VisibilityCodexOnly, HideTrend: true}))
	if m.prefs.Visibility != VisibilityCodexOnly || !m.prefs.HideTrend {
		t.Fatalf("expected prefs applied, got %#v", m.prefs)
	}
}

func TestTrendVisibleDefaultsOn(t *testing.T) {
	var p DisplayPrefs // zero value
	if !p.trendVisible() {
		t.Fatalf("zero-value DisplayPrefs should show the trend line")
	}
	p.HideTrend = true
	if p.trendVisible() {
		t.Fatalf("HideTrend=true should hide the trend line")
	}
}

func TestDisplayPrefsIconsDefaultOff(t *testing.T) {
	var p DisplayPrefs
	if p.Icons {
		t.Fatalf("Icons should default to false (safe glyphs)")
	}
}

func TestCostVisibleDefaultsOn(t *testing.T) {
	var p DisplayPrefs
	if !p.costVisible() {
		t.Fatalf("cost should be visible by default (zero value)")
	}
	p.HideCost = true
	if p.costVisible() {
		t.Fatalf("HideCost=true should hide cost")
	}
}
