package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/robbell5/llm-quota/internal/cost"
	"github.com/robbell5/llm-quota/internal/sources"
)

// formatValue renders one WindowCost: "$3.20", "~$0.90" (estimate), "$3.20*"
// (some tokens unpriced), "$1.2k" (>= $1000), combinable.
func formatValue(wc cost.WindowCost) string {
	var b strings.Builder
	if wc.Estimated {
		b.WriteByte('~')
	}
	if wc.Amount >= 1000 {
		fmt.Fprintf(&b, "$%.1fk", wc.Amount/1000)
	} else {
		fmt.Fprintf(&b, "$%.2f", wc.Amount)
	}
	if wc.Incomplete {
		b.WriteByte('*')
	}
	return b.String()
}

// valueCluster builds "5h $X · 7d $Y" for a product. ok is false when the
// product has no cost data at all. A missing single window drops just its term.
func valueCluster(m Model, product sources.Product) (string, bool) {
	byKind, ok := m.costs[product]
	if !ok || len(byKind) == 0 {
		return "", false
	}
	var parts []string
	if wc, ok := byKind[sources.WindowFiveHour]; ok {
		parts = append(parts, "5h "+formatValue(wc))
	}
	if wc, ok := byKind[sources.WindowSevenDay]; ok {
		parts = append(parts, "7d "+formatValue(wc))
	}
	if len(parts) == 0 {
		return "", false
	}
	return strings.Join(parts, " · "), true
}

// renderFreshnessLine is the consolidated dim freshness line shown above the
// footer while cost is active: "fresh: Claude 2m · Codex 1h" plus " · ~ est"
// when any Codex value is on screen. ok is false when no product has freshness.
func renderFreshnessLine(m Model, now time.Time, width int) (string, bool) {
	var parts []string
	codexShown := false
	for _, spec := range []struct {
		product sources.Product
		label   string
	}{
		{sources.ProductClaude, "Claude"},
		{sources.ProductCodex, "Codex"},
	} {
		if !m.prefs.Visibility.shows(spec.product) {
			continue
		}
		if byKind, ok := m.costs[spec.product]; ok && len(byKind) > 0 && spec.product == sources.ProductCodex {
			codexShown = true
		}
		_, age, ok := sourceFreshness(m, spec.product, now)
		if !ok {
			continue
		}
		token := spec.label + " " + ageText(age)
		if sourceIsStale(m, spec.product, now) {
			token += " old"
		}
		if err, has := m.errors[spec.product]; has && err.Category != "" {
			token += " " + m.glyphs().warning
		}
		parts = append(parts, token)
	}
	if len(parts) == 0 {
		return "", false
	}
	text := "fresh: " + strings.Join(parts, " · ")
	if codexShown {
		text += " · ~ est"
	}
	return hintStyle.Render(formatCell(text, width, false)), true
}
