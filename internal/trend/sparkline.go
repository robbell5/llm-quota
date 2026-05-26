package trend

import (
	"math"
	"strings"
)

var sparkLevels = []rune("▁▂▃▄▅▆▇█")

// Sparkline renders the last `width` samples as block runes, left-padded with
// spaces so the result is exactly `width` cells wide.
func Sparkline(samples []Sample, width int) string {
	if width <= 0 {
		return ""
	}
	used := samples
	if len(used) > width {
		used = used[len(used)-width:]
	}

	var b strings.Builder
	for i := 0; i < width-len(used); i++ {
		b.WriteRune(' ')
	}
	for _, s := range used {
		b.WriteRune(sparkRune(s.UsedPct))
	}
	return b.String()
}

func sparkRune(p float64) rune {
	if p < 0 {
		p = 0
	}
	if p > 100 {
		p = 100
	}
	idx := int(math.Round(p / 100 * float64(len(sparkLevels)-1)))
	return sparkLevels[idx]
}
