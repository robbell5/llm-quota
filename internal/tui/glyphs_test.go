package tui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestGlyphsSafeHasNoPrivateUseCodepoints(t *testing.T) {
	g := glyphsFor(false)
	for _, s := range append([]string{g.stripe, g.clock, g.warning, g.claudeMark, g.codexMark}, g.spinner...) {
		for _, r := range s {
			if r >= 0xE000 && r <= 0xF8FF {
				t.Fatalf("safe glyph set leaked a Nerd Font codepoint U+%04X in %q", r, s)
			}
		}
	}
}

func TestGlyphsNerdFontUsesIcons(t *testing.T) {
	g := glyphsFor(true)
	if g.clock == "" {
		t.Fatalf("nerd font set should provide a clock glyph")
	}
	if g.warning == "" {
		t.Fatalf("nerd font set should provide a warning glyph")
	}
}

func TestSpinnerFrameCyclesAndIsSingleWidth(t *testing.T) {
	g := glyphsFor(false)
	a := spinnerFrame(0, g)
	b := spinnerFrame(len(g.spinner)*spinnerHold, g) // one full cycle later
	if a != b {
		t.Fatalf("spinner should cycle back: %q vs %q", a, b)
	}
	if w := lipgloss.Width(spinnerFrame(0, g)); w != 1 {
		t.Fatalf("spinner frame should be single display column, got %d", w)
	}
}

func TestGlyphsSafeMarksAreEmpty(t *testing.T) {
	g := glyphsFor(false)
	if g.claudeMark != "" || g.codexMark != "" {
		t.Fatalf("safe glyph set should have empty provider marks, got claudeMark=%q codexMark=%q", g.claudeMark, g.codexMark)
	}
}

func TestGlyphsNerdFontHasProviderMarks(t *testing.T) {
	g := glyphsFor(true)
	if g.claudeMark == "" {
		t.Fatalf("nerd font set should provide a claudeMark glyph")
	}
	if g.codexMark == "" {
		t.Fatalf("nerd font set should provide a codexMark glyph")
	}
}

func TestGlyphsNerdFontHasNoEmptySpinnerStrings(t *testing.T) {
	g := glyphsFor(true)
	for i, s := range g.spinner {
		if strings.TrimSpace(s) == "" {
			t.Fatalf("nerd font spinner[%d] is empty", i)
		}
	}
}
