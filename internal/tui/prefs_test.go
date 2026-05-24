package tui

import (
	"testing"

	"github.com/robbell5/llm-quota/internal/sources"
)

func TestBarStyleToggled(t *testing.T) {
	if BarSegmented.toggled() != BarSolid {
		t.Fatal("segmented should toggle to solid")
	}
	if BarSolid.toggled() != BarSegmented {
		t.Fatal("solid should toggle to segmented")
	}
}

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
	m := NewModel(WithDisplayPrefs(DisplayPrefs{BarStyle: BarSolid, Visibility: VisibilityCodexOnly}))
	if m.prefs.BarStyle != BarSolid || m.prefs.Visibility != VisibilityCodexOnly {
		t.Fatalf("expected prefs applied, got %#v", m.prefs)
	}
}
