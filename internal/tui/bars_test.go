package tui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/lucasb-eyer/go-colorful"
)

func TestPulseColorVariesAndStaysReddish(t *testing.T) {
	a := pulseColorful(pulsePeriod / 4)     // near the peak of the cycle
	b := pulseColorful(3 * pulsePeriod / 4) // near the trough
	if a.AlmostEqualRgb(b) {
		t.Fatalf("pulse should vary across the cycle: %v vs %v", a.Hex(), b.Hex())
	}
	for _, c := range []colorful.Color{a, b} {
		if c.R < c.G || c.R < c.B {
			t.Fatalf("pulse color should stay reddish: %+v", c)
		}
	}
}

func TestRenderGradientBarDisplayWidth(t *testing.T) {
	for _, w := range []int{1, 6, 14, 20, 22} {
		for _, f := range []float64{0, 0.01, 0.2, 0.35, 0.5, 0.93, 1} {
			got := lipgloss.Width(renderGradientBar(f, w))
			if got != w {
				t.Fatalf("width=%d frac=%v: display width %d, want %d", w, f, got, w)
			}
		}
	}
}

func TestRenderGradientBarFillCount(t *testing.T) {
	plain := ansiEscapeRE.ReplaceAllString(renderGradientBar(0.5, 20), "")
	full := strings.Count(plain, "█")
	if full < 9 || full > 11 {
		t.Fatalf("frac 0.5 width 20: got %d full cells, want ~10", full)
	}
	empties := strings.Count(plain, string(barEmptyRune))
	if empties == 0 {
		t.Fatalf("expected a dim empty track, got none in %q", plain)
	}
}

func TestRenderGradientBarFullHasNoEmptyTrack(t *testing.T) {
	plain := ansiEscapeRE.ReplaceAllString(renderGradientBar(1, 12), "")
	if strings.Contains(plain, string(barEmptyRune)) {
		t.Fatalf("a full bar should have no empty cells: %q", plain)
	}
}
