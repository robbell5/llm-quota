package tui

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/lucasb-eyer/go-colorful"
)

var (
	mochaBase     = lipgloss.Color("#1e1e2e")
	mochaSurface0 = lipgloss.Color("#313244")
	mochaText     = lipgloss.Color("#cdd6f4")
	mochaSubtext0 = lipgloss.Color("#a6adc8")
	mochaBlue     = lipgloss.Color("#89b4fa")
	mochaGreen    = lipgloss.Color("#a6e3a1")
	mochaYellow   = lipgloss.Color("#f9e2af")
	mochaRed      = lipgloss.Color("#f38ba8")
)

var (
	mochaPeach = lipgloss.Color("#fab387") // Claude accent
	mochaTeal  = lipgloss.Color("#94e2d5") // Codex accent
)

var (
	rampLow  = mustHex("#a6e3a1") // green  (low usage)
	rampMid  = mustHex("#f9e2af") // yellow (mid usage)
	rampHigh = mustHex("#f38ba8") // red    (high usage)
)

func mustHex(s string) colorful.Color {
	c, err := colorful.Hex(s)
	if err != nil {
		panic("invalid ramp hex " + s + ": " + err.Error())
	}
	return c
}

// rampColorful returns the bar fill color at position fraction f in [0,1],
// blending green -> yellow -> red in HCL so the midrange reads as vivid amber
// rather than the muddy gray a direct green->red blend produces.
func rampColorful(f float64) colorful.Color {
	switch {
	case f <= 0:
		return rampLow
	case f >= 1:
		return rampHigh
	case f < 0.5:
		return rampLow.BlendHcl(rampMid, f/0.5).Clamped()
	default:
		return rampMid.BlendHcl(rampHigh, (f-0.5)/0.5).Clamped()
	}
}

// rampColor adapts rampColorful to the lipgloss/color.Color the renderer needs.
func rampColor(f float64) color.Color {
	return lipgloss.Color(rampColorful(f).Hex())
}

func thresholdColor(percent float64) color.Color {
	if percent >= 85 {
		return mochaRed
	}
	if percent >= 60 {
		return mochaYellow
	}

	return mochaGreen
}
