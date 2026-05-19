package tui

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/lipgloss/v2"

	"github.com/rob/llm-quota/internal/sources"
)

const (
	defaultWidth           = 50
	shellHorizontalPadding = 4
	shellPaddingX          = shellHorizontalPadding / 2
)

const (
	quitHint          = "q / Ctrl-C quit"
	refreshHint       = "r refresh"
	claudeInstallHint = "Claude: run install-claude-hook"
	claudeOpenHint    = "Claude: open Claude"
	codexOpenHint     = "Codex: open Codex"
)

var (
	shellStyle = lipgloss.NewStyle().
			Background(mochaBase).
			Foreground(mochaText).
			Padding(1, shellPaddingX)
	titleStyle = lipgloss.NewStyle().
			Foreground(mochaBlue).
			Bold(true)
	dividerStyle = lipgloss.NewStyle().Foreground(mochaSurface0)
	labelStyle   = lipgloss.NewStyle().Foreground(mochaText).Bold(true)
	missingStyle = lipgloss.NewStyle().Foreground(mochaYellow)
	hintStyle    = lipgloss.NewStyle().Foreground(mochaSubtext0)
	footerStyle  = lipgloss.NewStyle().Foreground(mochaSubtext0)
)

const (
	fullRowLabelWidth  = 9
	shortRowLabelWidth = 5
	minProgressWidth   = 6
)

func render(m Model) string {
	width := m.width
	if width <= 0 {
		width = defaultWidth
	}
	innerWidth := width - shellHorizontalPadding
	if innerWidth < 1 {
		innerWidth = 1
	}

	content := strings.Join([]string{
		titleStyle.Render("LLM Quota"),
		dividerStyle.Render(strings.Repeat("─", innerWidth)),
		renderRows(m, innerWidth),
		"",
		footerStyle.Width(innerWidth).Render(renderFooter(m, innerWidth)),
	}, "\n")

	return shellStyle.Render(content) + "\n"
}

func renderRows(m Model, width int) string {
	rows := make([]string, 0, 4)
	rowLabels := []struct {
		full    string
		short   string
		product sources.Product
		kind    sources.WindowKind
	}{
		{full: "Claude 5h", short: "Cl 5h", product: sources.ProductClaude, kind: sources.WindowFiveHour},
		{full: "Claude 7d", short: "Cl 7d", product: sources.ProductClaude, kind: sources.WindowSevenDay},
		{full: "Codex 5h", short: "Cx 5h", product: sources.ProductCodex, kind: sources.WindowFiveHour},
		{full: "Codex 7d", short: "Cx 7d", product: sources.ProductCodex, kind: sources.WindowSevenDay},
	}
	now := time.Now
	if m.now != nil {
		now = m.now
	}

	for _, label := range rowLabels {
		if window, ok := findWindow(m, label.product, label.kind); ok {
			rows = append(rows, renderDataRow(label.full, label.short, window, now(), width))
			continue
		}

		switch {
		case width >= 46:
			rows = append(rows, fmt.Sprintf(
				"%s  %s  %s  reset %s",
				labelStyle.Render(fmt.Sprintf("%-9s", label.full)),
				missingStyle.Render("—"),
				hintStyle.Render("missing local data"),
				missingStyle.Render("—"),
			))
		case width >= 26:
			rows = append(rows, fmt.Sprintf(
				"%s  %s  %s",
				labelStyle.Render(fmt.Sprintf("%-5s", label.short)),
				missingStyle.Render("—"),
				hintStyle.Render("pending"),
			))
		default:
			rows = append(rows, fmt.Sprintf(
				"%s %s",
				labelStyle.Render(fmt.Sprintf("%-5s", label.short)),
				missingStyle.Render("—"),
			))
		}
	}

	return strings.Join(rows, "\n")
}

func findWindow(m Model, product sources.Product, kind sources.WindowKind) (sources.Window, bool) {
	for _, window := range m.windows[product] {
		if window.Kind == kind {
			return window, true
		}
	}

	return sources.Window{}, false
}

func renderDataRow(fullLabel string, shortLabel string, window sources.Window, now time.Time, width int) string {
	percent := lipgloss.NewStyle().Foreground(thresholdColor(window.UsedPercent)).Render(fmt.Sprintf("%.0f%%", math.Round(window.UsedPercent)))
	reset := resetText(window.ResetsAt, now)

	switch {
	case width >= 46:
		barWidth := width - fullRowLabelWidth - lipgloss.Width(fmt.Sprintf("%.0f%%", math.Round(window.UsedPercent))) - lipgloss.Width(reset) - 6
		if barWidth < minProgressWidth {
			barWidth = minProgressWidth
		}

		return fmt.Sprintf(
			"%s  %s  %s  %s",
			labelStyle.Render(fmt.Sprintf("%-9s", fullLabel)),
			renderProgressBar(window.UsedPercent, barWidth),
			percent,
			reset,
		)
	case width >= 26:
		barWidth := width - shortRowLabelWidth - lipgloss.Width(fmt.Sprintf("%.0f%%", math.Round(window.UsedPercent))) - lipgloss.Width(reset) - 3
		if barWidth >= minProgressWidth {
			return fmt.Sprintf(
				"%s %s %s %s",
				labelStyle.Render(fmt.Sprintf("%-5s", shortLabel)),
				renderProgressBar(window.UsedPercent, barWidth),
				percent,
				reset,
			)
		}

		return fmt.Sprintf(
			"%s  %s  %s",
			labelStyle.Render(fmt.Sprintf("%-5s", shortLabel)),
			percent,
			reset,
		)
	default:
		withReset := fmt.Sprintf(
			"%s %s %s",
			labelStyle.Render(fmt.Sprintf("%-5s", shortLabel)),
			percent,
			reset,
		)
		if lipgloss.Width(withReset) <= width {
			return withReset
		}

		return fmt.Sprintf("%s %s", labelStyle.Render(fmt.Sprintf("%-5s", shortLabel)), percent)
	}
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

func progressFraction(percent float64) float64 {
	if percent < 0 {
		return 0
	}
	if percent > 100 {
		return 1
	}

	return percent / 100
}

func renderProgressBar(percent float64, width int) string {
	if width < 1 {
		width = 1
	}
	p := progress.New(progress.WithWidth(width), progress.WithColors(thresholdColor(percent)), progress.WithoutPercentage())
	p.EmptyColor = mochaSurface0

	return p.ViewAs(progressFraction(percent))
}

func resetText(resetsAt time.Time, now time.Time) string {
	if resetsAt.IsZero() {
		return missingStyle.Render("—")
	}

	remaining := resetsAt.Sub(now)
	if remaining <= 0 {
		return "now"
	}

	totalMinutes := int(remaining / time.Minute)
	if remaining >= 24*time.Hour {
		days := totalMinutes / int((24 * time.Hour / time.Minute))
		hours := (totalMinutes % int((24 * time.Hour / time.Minute))) / int(time.Hour/time.Minute)
		return fmt.Sprintf("%dd %02dh", days, hours)
	}

	hours := totalMinutes / int(time.Hour/time.Minute)
	minutes := totalMinutes % int(time.Hour/time.Minute)
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func renderFooter(m Model, innerWidth int) string {
	hints := footerRecoveryHints(m)
	if len(hints) == 0 {
		return appendHintWithinWidth("", []string{quitHint, refreshHint}, innerWidth)
	}

	footer := appendHintWithinWidth("", hints, innerWidth)
	if footer != "" {
		return footer
	}

	return appendHintWithinWidth("", []string{quitHint}, innerWidth)
}

func footerRecoveryHints(m Model) []string {
	hints := make([]string, 0, 2)

	if !hasWindows(m, sources.ProductClaude) {
		if err, ok := m.errors[sources.ProductClaude]; ok {
			if err.Category == sources.ErrorMissing {
				hints = append(hints, claudeInstallHint)
			} else {
				hints = append(hints, claudeOpenHint)
			}
		}
	}
	if !hasWindows(m, sources.ProductCodex) {
		if _, ok := m.errors[sources.ProductCodex]; ok {
			hints = append(hints, codexOpenHint)
		}
	}

	if len(hints) > 0 {
		return hints
	}

	if hint, ok := staleHint(m, sources.ProductClaude, "Claude", "open Claude"); ok {
		hints = append(hints, hint)
	}
	if hint, ok := staleHint(m, sources.ProductCodex, "Codex", "open Codex"); ok {
		hints = append(hints, hint)
	}

	return hints
}

func hasWindows(m Model, product sources.Product) bool {
	return len(m.windows[product]) > 0
}

func staleHint(m Model, product sources.Product, label string, action string) (string, bool) {
	now := time.Now
	if m.now != nil {
		now = m.now
	}

	for _, window := range m.windows[product] {
		if !window.Stale {
			continue
		}

		age := window.StaleAge
		if age <= 0 && !window.CapturedAt.IsZero() {
			age = now().Sub(window.CapturedAt)
		}

		return fmt.Sprintf("%s data %s old; %s", label, ageText(age), action), true
	}

	return "", false
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
