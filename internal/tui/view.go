package tui

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/lipgloss/v2"

	"github.com/robbell5/llm-quota/internal/sources"
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
	normalPercentWidth = 4
	normalResetWidth   = 7
	compactResetWidth  = 4
	normalGap          = "  "
	compactGap         = " "
	minProgressWidth   = 6
)

type quotaRowSpec struct {
	full    string
	short   string
	product sources.Product
	kind    sources.WindowKind
}

var quotaRowSpecs = []quotaRowSpec{
	{full: "Claude 5h", short: "Cl 5h", product: sources.ProductClaude, kind: sources.WindowFiveHour},
	{full: "Claude 7d", short: "Cl 7d", product: sources.ProductClaude, kind: sources.WindowSevenDay},
	{full: "Sonnet 7d", short: "Sn 7d", product: sources.ProductClaude, kind: sources.WindowSonnetSevenDay},
	{full: "Codex 5h", short: "Cx 5h", product: sources.ProductCodex, kind: sources.WindowFiveHour},
	{full: "Codex 7d", short: "Cx 7d", product: sources.ProductCodex, kind: sources.WindowSevenDay},
}

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
	rows := make([]string, 0, len(quotaRowSpecs)+2)
	now := time.Now
	if m.now != nil {
		now = m.now
	}

	for _, spec := range quotaRowSpecs {
		if window, ok := findWindow(m, spec.product, spec.kind); ok {
			rows = append(rows, renderDataRow(spec.full, spec.short, window, now(), width))
		} else {
			rows = append(rows, renderMissingRow(spec.full, spec.short, width))
		}

		if spec.kind == sources.WindowSonnetSevenDay {
			if line, ok := renderFreshnessLine(m, sources.ProductClaude, "Claude", "Cl", now(), width); ok {
				rows = append(rows, line)
			}
		}
		if spec.product == sources.ProductCodex && spec.kind == sources.WindowSevenDay {
			if line, ok := renderFreshnessLine(m, sources.ProductCodex, "Codex", "Cx", now(), width); ok {
				rows = append(rows, line)
			}
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
	percentText := fmt.Sprintf("%.0f%%", math.Round(window.UsedPercent))
	percent := lipgloss.NewStyle().Foreground(thresholdColor(window.UsedPercent)).Render(formatCell(percentText, normalPercentWidth, true))
	reset := resetText(window.ResetsAt, now)

	switch {
	case width >= 46:
		barWidth := width - fullRowLabelWidth - normalPercentWidth - normalResetWidth - 3*len(normalGap)

		return fmt.Sprintf(
			"%s  %s  %s  %s",
			labelStyle.Render(formatCell(fullLabel, fullRowLabelWidth, false)),
			renderProgressBar(window.UsedPercent, barWidth),
			percent,
			formatCell(reset, normalResetWidth, true),
		)
	case width >= 26:
		compactReset := compactResetText(window.ResetsAt, now)
		compactPercent := lipgloss.NewStyle().Foreground(thresholdColor(window.UsedPercent)).Render(formatCell(percentText, normalPercentWidth, true))
		barWidth := width - shortRowLabelWidth - normalPercentWidth - compactResetWidth - 3*len(compactGap)
		if barWidth >= minProgressWidth {
			return fmt.Sprintf(
				"%s %s %s %s",
				labelStyle.Render(formatCell(shortLabel, shortRowLabelWidth, false)),
				renderProgressBar(window.UsedPercent, barWidth),
				compactPercent,
				formatCell(compactReset, compactResetWidth, true),
			)
		}

		return fmt.Sprintf(
			"%s  %s  %s",
			labelStyle.Render(formatCell(shortLabel, shortRowLabelWidth, false)),
			compactPercent,
			formatCell(compactReset, compactResetWidth, true),
		)
	default:
		compactReset := compactResetText(window.ResetsAt, now)
		compactPercent := lipgloss.NewStyle().Foreground(thresholdColor(window.UsedPercent)).Render(formatCell(percentText, normalPercentWidth, true))
		withReset := fmt.Sprintf(
			"%s %s %s",
			labelStyle.Render(formatCell(shortLabel, shortRowLabelWidth, false)),
			compactPercent,
			formatCell(compactReset, compactResetWidth, true),
		)
		if lipgloss.Width(withReset) <= width {
			return withReset
		}

		return fmt.Sprintf("%s %s", labelStyle.Render(formatCell(shortLabel, shortRowLabelWidth, false)), compactPercent)
	}
}

func renderMissingRow(fullLabel string, shortLabel string, width int) string {
	percent := missingStyle.Render(formatCell("—", normalPercentWidth, true))
	reset := missingStyle.Render(formatCell("—", normalResetWidth, true))

	switch {
	case width >= 46:
		barWidth := width - fullRowLabelWidth - normalPercentWidth - normalResetWidth - 3*len(normalGap)
		barText := formatCell("missing local data", barWidth, false)
		return fmt.Sprintf(
			"%s  %s  %s  %s",
			labelStyle.Render(formatCell(fullLabel, fullRowLabelWidth, false)),
			hintStyle.Render(barText),
			percent,
			reset,
		)
	case width >= 26:
		compactReset := missingStyle.Render(formatCell("—", compactResetWidth, true))
		barWidth := width - shortRowLabelWidth - normalPercentWidth - compactResetWidth - 3*len(compactGap)
		if barWidth >= minProgressWidth {
			return fmt.Sprintf(
				"%s %s %s %s",
				labelStyle.Render(formatCell(shortLabel, shortRowLabelWidth, false)),
				hintStyle.Render(formatCell("pending", barWidth, false)),
				percent,
				compactReset,
			)
		}

		return fmt.Sprintf(
			"%s  %s  %s",
			labelStyle.Render(formatCell(shortLabel, shortRowLabelWidth, false)),
			percent,
			compactReset,
		)
	default:
		withReset := fmt.Sprintf(
			"%s %s %s",
			labelStyle.Render(formatCell(shortLabel, shortRowLabelWidth, false)),
			percent,
			missingStyle.Render(formatCell("—", compactResetWidth, true)),
		)
		if lipgloss.Width(withReset) <= width {
			return withReset
		}

		return fmt.Sprintf("%s %s", labelStyle.Render(formatCell(shortLabel, shortRowLabelWidth, false)), percent)
	}
}

func formatCell(value string, width int, alignRight bool) string {
	if width <= 0 {
		return ""
	}
	for lipgloss.Width(value) > width && len(value) > 0 {
		value = value[:len(value)-1]
	}
	padding := width - lipgloss.Width(value)
	if padding <= 0 {
		return value
	}
	if alignRight {
		return strings.Repeat(" ", padding) + value
	}
	return value + strings.Repeat(" ", padding)
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
		return "—"
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

func compactResetText(resetsAt time.Time, now time.Time) string {
	if resetsAt.IsZero() {
		return "—"
	}

	remaining := resetsAt.Sub(now)
	if remaining <= 0 {
		return "now"
	}

	totalMinutes := int(remaining / time.Minute)
	if remaining >= 24*time.Hour {
		days := totalMinutes / int((24 * time.Hour / time.Minute))
		return fmt.Sprintf("%dd", days)
	}

	hours := totalMinutes / int(time.Hour/time.Minute)
	return fmt.Sprintf("%dh", hours)
}

func renderFreshnessLine(m Model, product sources.Product, fullLabel string, shortLabel string, now time.Time, width int) (string, bool) {
	capturedAt, age, ok := sourceFreshness(m, product, now)
	if !ok {
		return "", false
	}

	statuses := make([]string, 0, 2)
	if sourceIsStale(m, product, now) {
		statuses = append(statuses, ageText(age)+" old")
	} else if age > 0 {
		statuses = append(statuses, ageText(age)+" ago")
	}
	if err, hasError := m.errors[product]; hasError && err.Category != "" {
		statuses = append(statuses, "refresh failed")
	}

	timeText := capturedAt.Local().Format("3:04 PM")
	if width >= 46 {
		candidates := []string{
			freshnessText(fullLabel, "updated "+timeText, statuses),
			freshnessText(fullLabel, "updated "+timeText, nil),
		}
		return renderFirstFittingFreshness(candidates, width)
	}
	if width >= 26 {
		candidates := []string{
			freshnessText(fullLabel, timeText, statuses),
			freshnessText(fullLabel, timeText, nil),
			freshnessText(shortLabel, timeText, nil),
		}
		return renderFirstFittingFreshness(candidates, width)
	}

	veryShortTime := capturedAt.Local().Format("3:04")
	return renderFirstFittingFreshness([]string{
		freshnessText(shortLabel, veryShortTime, nil),
	}, width)
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

func freshnessText(label string, updated string, statuses []string) string {
	text := label + " " + updated
	if len(statuses) > 0 {
		text += " (" + strings.Join(statuses, ", ") + ")"
	}
	return text
}

func renderFirstFittingFreshness(candidates []string, width int) (string, bool) {
	for _, candidate := range candidates {
		if lipgloss.Width(candidate) <= width {
			return hintStyle.Render(candidate), true
		}
	}

	return "", false
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
			if err.Category == sources.ErrorMissing && !m.claudeHookInstalled {
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
