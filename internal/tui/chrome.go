package tui

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/robbell5/llm-quota/internal/sources"
)

const (
	quitHint          = "q / Ctrl-C quit"
	refreshHint       = "r refresh"
	claudeInstallHint = "Claude: run install-claude-hook"
	claudeOpenHint    = "Claude: open Claude"
	codexOpenHint     = "Codex: open Codex"
)

var (
	bandRuleStyle = lipgloss.NewStyle().Foreground(mochaSurface0)
	clockStyle    = lipgloss.NewStyle().Foreground(mochaSubtext0)
)

func accentFor(product sources.Product) color.Color {
	if product == sources.ProductCodex {
		return mochaTeal
	}
	return mochaPeach
}

// renderTitleBand is the header: bold title left, wall clock right, over a rule.
// width is the inner content width. Heavy rule when wide, light otherwise. The
// clock is dropped when the title and clock cannot share the line.
func renderTitleBand(width int, now time.Time, g glyphSet) string {
	const titleText = "LLM QUOTA"
	clock := now.Format("3:04 PM")
	if g.clock != "" {
		clock = g.clock + " " + clock
	}
	title := titleStyle.Render(titleText)
	rule := "─"
	if width >= wideThreshold {
		rule = "━"
	}
	ruleLine := bandRuleStyle.Render(strings.Repeat(rule, max(width, 1)))

	gap := width - lipgloss.Width(titleText) - lipgloss.Width(clock)
	if gap < 1 {
		// No room for both: show the title alone.
		return title + "\n" + ruleLine
	}
	header := title + strings.Repeat(" ", gap) + clockStyle.Render(clock)
	return header + "\n" + ruleLine
}

// renderGroupHeader is the accent-striped provider label with its freshness on
// the right. Drops the freshness when it would not fit.
func renderGroupHeader(m Model, product sources.Product, label string, now time.Time, width int) string {
	g := m.glyphs()
	accent := lipgloss.NewStyle().Foreground(accentFor(product))
	left := accent.Render(g.stripe) + " "
	if mark := providerMark(g, product); mark != "" {
		left += accent.Render(mark) + " "
	}
	left += accent.Bold(true).Render(label)

	fresh, ok := groupFreshnessText(m, product, now)
	if m.refreshing {
		sp := spinnerFrame(m.animPhase, g)
		if ok {
			fresh = sp + " " + fresh
		} else {
			fresh = sp + " refreshing"
			ok = true
		}
	}
	if !ok {
		return left
	}
	gap := width - lipgloss.Width(left) - lipgloss.Width(fresh)
	if gap < 1 {
		return left
	}
	return left + strings.Repeat(" ", gap) + hintStyle.Render(fresh)
}

func providerMark(g glyphSet, product sources.Product) string {
	if product == sources.ProductCodex {
		return g.codexMark
	}
	return g.claudeMark
}

// groupFreshnessText is the dim freshness string for a group header, e.g.
// "updated 10:37 AM · 0m ago" or "updated 2:14 PM · 2h old · refresh failed".
// Returns ok=false when there is no capture yet. It mirrors the information the
// old per-row freshness line carried (time, age/staleness, and refresh failure).
func groupFreshnessText(m Model, product sources.Product, now time.Time) (string, bool) {
	capturedAt, age, ok := sourceFreshness(m, product, now)
	if !ok {
		return "", false
	}
	parts := []string{"updated " + capturedAt.Local().Format("3:04 PM")}
	if sourceIsStale(m, product, now) {
		parts = append(parts, ageText(age)+" old")
	} else if age > 0 {
		parts = append(parts, ageText(age)+" ago")
	}
	if err, hasErr := m.errors[product]; hasErr && err.Category != "" {
		parts = append(parts, "refresh failed")
	}
	return strings.Join(parts, " · "), true
}

func sourceFreshness(m Model, product sources.Product, now time.Time) (time.Time, time.Duration, bool) {
	windows := m.windows[product]
	var capturedAt time.Time
	var staleAge time.Duration
	for _, window := range windows {
		if window.CapturedAt.IsZero() {
			continue
		}
		if capturedAt.IsZero() || window.CapturedAt.After(capturedAt) {
			capturedAt = window.CapturedAt
		}
		if window.StaleAge > staleAge {
			staleAge = window.StaleAge
		}
	}
	if capturedAt.IsZero() {
		return time.Time{}, 0, false
	}
	age := staleAge
	if age <= 0 {
		age = now.Sub(capturedAt)
	}
	if age < 0 {
		age = 0
	}

	return capturedAt, age, true
}

func sourceIsStale(m Model, product sources.Product, now time.Time) bool {
	for _, window := range m.windows[product] {
		if window.Stale {
			return true
		}
	}
	_, age, ok := sourceFreshness(m, product, now)
	return ok && age > m.staleAfter
}

func renderFooter(m Model, innerWidth int) string {
	hints := footerRecoveryHints(m)
	if len(hints) == 0 {
		full := []string{quitHint, refreshHint, "v view", "t trend", "i icons"}
		return appendHintWithinWidth("", full, innerWidth)
	}

	footer := appendHintWithinWidth("", hints, innerWidth)
	if footer != "" {
		return footer
	}

	return appendHintWithinWidth("", []string{quitHint}, innerWidth)
}

func footerRecoveryHints(m Model) []string {
	hints := make([]string, 0, 2)

	if m.prefs.Visibility.shows(sources.ProductClaude) && !hasWindows(m, sources.ProductClaude) {
		if err, ok := m.errors[sources.ProductClaude]; ok {
			if err.Category == sources.ErrorMissing && !m.claudeHookInstalled {
				hints = append(hints, claudeInstallHint)
			} else {
				hints = append(hints, claudeOpenHint)
			}
		}
	}
	if m.prefs.Visibility.shows(sources.ProductCodex) && !hasWindows(m, sources.ProductCodex) {
		if _, ok := m.errors[sources.ProductCodex]; ok {
			hints = append(hints, codexOpenHint)
		}
	}

	return hints
}

func hasWindows(m Model, product sources.Product) bool {
	return len(m.windows[product]) > 0
}

func ageText(age time.Duration) string {
	if age < 0 {
		age = 0
	}
	if age >= 24*time.Hour {
		return fmt.Sprintf("%dd", int(age/(24*time.Hour)))
	}
	if age >= time.Hour {
		return fmt.Sprintf("%dh", int(age/time.Hour))
	}

	return fmt.Sprintf("%dm", int(age/time.Minute))
}

func appendHintWithinWidth(base string, hints []string, width int) string {
	footer := base
	for _, hint := range hints {
		candidate := hint
		if footer != "" {
			candidate = footer + " · " + hint
		}
		if lipgloss.Width(candidate) > width {
			continue
		}

		footer = candidate
	}

	return footer
}
