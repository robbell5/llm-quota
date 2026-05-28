package tui

// glyphSet holds the runes/strings that differ between the safe (default) and
// Nerd Font display modes.
type glyphSet struct {
	stripe     string
	claudeMark string
	codexMark  string
	clock      string
	warning    string
	spinner    []string
}

// spinnerHold is how many animation ticks each spinner glyph dwells before advancing.
const spinnerHold = 2

func glyphsFor(icons bool) glyphSet {
	g := glyphSet{
		stripe:  "▎",
		clock:   "",
		warning: "⚠",
		spinner: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
	if icons {
		g.claudeMark = "" // nf-fa-bolt
		g.codexMark = ""  // nf-fa-code
		g.clock = ""      // nf-fa-clock_o
		g.warning = ""    // nf-fa-warning
	}
	return g
}

// spinnerFrame returns the spinner glyph for the given animation phase.
func spinnerFrame(phase int, g glyphSet) string {
	if len(g.spinner) == 0 {
		return ""
	}
	return g.spinner[(phase/spinnerHold)%len(g.spinner)]
}

func (m Model) glyphs() glyphSet {
	return glyphsFor(m.prefs.Icons)
}
