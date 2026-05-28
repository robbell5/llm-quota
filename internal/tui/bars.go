package tui

import (
	"image/color"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/lucasb-eyer/go-colorful"
)

const barEmptyRune = '░'

const pulsePeriod = 30 // anim frames per full pulse cycle

var pulseDim = mustHex("#7d4859") // dimmed red endpoint

// pulseColorful oscillates between a dim and full red over pulsePeriod frames.
func pulseColorful(phase int) colorful.Color {
	t := (math.Sin(2*math.Pi*float64(phase)/pulsePeriod) + 1) / 2
	return pulseDim.BlendHcl(rampHigh, t).Clamped()
}

func pulseColor(phase int) color.Color {
	return lipgloss.Color(pulseColorful(phase).Hex())
}

const paceMarkerRune = "╎"

func pacedBar(fraction float64, width int, evenUse float64) string {
	bar := renderGradientBar(fraction, width)
	return overlayEvenUseTick(bar, width, evenUse)
}

// fractionalTipRunes[i] is the left-aligned partial block for i/8 of a cell.
var fractionalTipRunes = []rune{' ', '▏', '▎', '▍', '▌', '▋', '▊', '▉'}

var barTrackStyle = lipgloss.NewStyle().Foreground(mochaSurface0)

// renderGradientBar draws a width-cell bar filled to fraction (0..1) using the
// green->amber->red ramp. The ramp is anchored to the full track: cell i takes
// the color at i/(width-1), so the visible tip color encodes the percentage.
// A fractional tip block gives sub-cell precision; remaining cells are the dim
// empty track.
func renderGradientBar(fraction float64, width int) string {
	if width < 1 {
		width = 1
	}
	if fraction < 0 {
		fraction = 0
	}
	if fraction > 1 {
		fraction = 1
	}

	cellColor := func(i int) color.Color {
		if width == 1 {
			return rampColor(fraction)
		}
		return rampColor(float64(i) / float64(width-1))
	}

	filled := fraction * float64(width)
	fullCells := int(filled)
	if fullCells > width {
		fullCells = width
	}

	var b strings.Builder
	rendered := 0
	for i := 0; i < fullCells; i++ {
		// Per-cell style allocation is acceptable: the animation loop self-suspends
		// when idle, so this path is not always-hot.
		b.WriteString(lipgloss.NewStyle().Foreground(cellColor(i)).Render("█"))
		rendered++
	}
	if rendered < width {
		if tip := int((filled - float64(fullCells)) * 8); tip > 0 {
			b.WriteString(lipgloss.NewStyle().Foreground(cellColor(rendered)).Render(string(fractionalTipRunes[tip])))
			rendered++
		}
	}
	for i := rendered; i < width; i++ {
		b.WriteString(barTrackStyle.Render(string(barEmptyRune)))
	}
	return b.String()
}

// overlayEvenUseTick replaces the bar cell at the elapsed-fraction position
// with a tick marker, preserving the bar's visible width.
func overlayEvenUseTick(bar string, width int, fraction float64) string {
	if width <= 0 {
		return bar
	}
	if fraction < 0 {
		fraction = 0
	}
	if fraction > 1 {
		fraction = 1
	}
	col := int(math.Round(fraction * float64(width-1)))
	if col < 0 {
		col = 0
	}
	if col > width-1 {
		col = width - 1
	}
	tick := lipgloss.NewStyle().Foreground(mochaText).Render(paceMarkerRune)
	return replaceCellAt(bar, col, tick)
}

// replaceCellAt swaps the visible rune at display column `col` for
// `replacement`, copying ANSI escape sequences verbatim (they have zero width)
// and re-emitting the active style after the replacement so trailing cells keep
// their color.
func replaceCellAt(s string, col int, replacement string) string {
	var b strings.Builder
	runes := []rune(s)
	visible := 0
	lastEscape := ""
	for i := 0; i < len(runes); {
		if runes[i] == '\x1b' {
			j := i + 1
			if j < len(runes) && runes[j] == '[' {
				j++
				for j < len(runes) && !(runes[j] >= '@' && runes[j] <= '~') {
					j++
				}
				if j < len(runes) {
					j++
				}
			}
			esc := string(runes[i:j])
			lastEscape = esc
			b.WriteString(esc)
			i = j
			continue
		}
		if visible == col {
			b.WriteString(replacement)
			b.WriteString(lastEscape)
		} else {
			b.WriteRune(runes[i])
		}
		visible++
		i++
	}
	return b.String()
}

func progressFraction(percent float64) float64 {
	if percent < 0 {
		return 0
	}
	if percent > 100 {
		return 1
	}

	return percent / 100
}
