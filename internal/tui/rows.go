package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/robbell5/llm-quota/internal/sources"
	"github.com/robbell5/llm-quota/internal/trend"
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

const (
	sparkWidth         = 6
	fullTrendIndent    = fullRowLabelWidth + len(normalGap)
	compactTrendIndent = shortRowLabelWidth + len(compactGap)
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

func renderRows(m Model, width int) string {
	rows := make([]string, 0, len(quotaRowSpecs)+4)
	now := time.Now
	if m.now != nil {
		now = m.now
	}
	grouped := width >= 46
	var lastProduct sources.Product = ""

	// m.bars is a parallel slice aligned 1:1 with quotaRowSpecs (both built in NewModel).
	for i, spec := range quotaRowSpecs {
		if !m.prefs.Visibility.shows(spec.product) {
			continue
		}
		if grouped && spec.product != lastProduct {
			label := "CLAUDE"
			if spec.product == sources.ProductCodex {
				label = "CODEX"
			}
			rows = append(rows, renderGroupHeader(m, spec.product, label, now(), width))
			lastProduct = spec.product
		}

		fullLabel, shortLabel := spec.full, spec.short
		if grouped {
			fullLabel, shortLabel = groupedLabels(spec)
		}

		if window, ok := findWindow(m, spec.product, spec.kind); ok {
			evenUse := trend.ElapsedFraction(spec.kind, window.ResetsAt, now())
			highlighted := now().Before(m.highlightUntil[i])
			rows = append(rows, renderDataRow(m, spec, fullLabel, shortLabel, window, m.bars[i].pos, evenUse, now(), width, highlighted))
			if m.prefs.trendVisible() && width < wideThreshold {
				if line, ok := renderTrendLine(m, spec, window, now(), width); ok {
					rows = append(rows, line)
				}
			}
		} else {
			rows = append(rows, renderMissingRow(fullLabel, shortLabel, width))
		}
	}

	return strings.Join(rows, "\n")
}

// groupedLabels returns the short per-row labels used under a group header.
func groupedLabels(spec quotaRowSpec) (full string, short string) {
	switch spec.kind {
	case sources.WindowFiveHour:
		return "5h", "5h"
	case sources.WindowSonnetSevenDay:
		return "Sonnet 7d", "Sonnet"
	default:
		return "7d", "7d"
	}
}

func findWindow(m Model, product sources.Product, kind sources.WindowKind) (sources.Window, bool) {
	for _, window := range m.windows[product] {
		if window.Kind == kind {
			return window, true
		}
	}

	return sources.Window{}, false
}

// windowForecast computes the trend forecast for a window using recent history
// samples. ok is false when there is no history to compute from.
func windowForecast(m Model, spec quotaRowSpec, window sources.Window, now time.Time) (trend.Forecast, bool) {
	if m.history == nil {
		return trend.Forecast{}, false
	}
	samples := m.history.EpochSamples(trend.Key(spec.product, spec.kind), window.ResetsAt)
	windowStart := window.ResetsAt.Add(-trend.WindowDuration(spec.kind))
	rate, _ := trend.Rate(samples, now, trend.RateLookback, windowStart)
	return trend.ComputeForecast(window.UsedPercent, rate, now, window.ResetsAt), true
}

// windowAtRisk reports whether a window is full or projected to exhaust before
// reset. It is the single source of truth for the at-risk pulse and the ⚠ flag.
func windowAtRisk(m Model, spec quotaRowSpec, window sources.Window, now time.Time) bool {
	if window.UsedPercent >= 100 {
		return true
	}
	f, ok := windowForecast(m, spec, window, now)
	return ok && f.AtRisk
}

func (m Model) anyAtRisk(now time.Time) bool {
	for _, spec := range quotaRowSpecs {
		if !m.prefs.Visibility.shows(spec.product) {
			continue
		}
		if window, ok := findWindow(m, spec.product, spec.kind); ok {
			if windowAtRisk(m, spec, window, now) {
				return true
			}
		}
	}
	return false
}

const wideTrendWidth = 26 // trend cell columns; "↑ 100%/hr · ~100% by reset" = 26

// wideTrendCell renders the inline trend cluster for the wide tier, padded to
// width. It prefers the rich form (rate + projection) and falls back to the
// plain rate when the rich form would not fit, mirroring renderTrendLine.
func wideTrendCell(m Model, spec quotaRowSpec, window sources.Window, now time.Time, width int) string {
	atRisk := false
	if window.UsedPercent < 100 {
		if f, ok := windowForecast(m, spec, window, now); ok {
			atRisk = f.AtRisk
		}
	}

	// pulseStyle drives the at-risk/maxed text; it pulses across animation frames.
	pulseStyle := lipgloss.NewStyle().Foreground(pulseColor(m.animPhase))
	// coreStyle is the pulse-red when the window is full (≥100%) or at risk; hint otherwise.
	coreStyle := hintStyle
	if window.UsedPercent >= 100 || atRisk {
		coreStyle = pulseStyle
	}
	// warnPrefix is the glyph-set warning plus a space (e.g. "⚠ " in safe mode).
	warnPrefix := m.glyphs().warning + " "

	for _, rich := range []bool{true, false} {
		core := inlineTrendCore(m, spec, window, now, rich)
		if atRisk {
			// The warning prefix takes its display columns; the core fills the rest.
			if lipgloss.Width(warnPrefix+core) <= width {
				padded := formatCell(core, width-lipgloss.Width(warnPrefix), false)
				return pulseStyle.Render(warnPrefix) + pulseStyle.Render(padded)
			}
		} else {
			if lipgloss.Width(core) <= width {
				return coreStyle.Render(formatCell(core, width, false))
			}
		}
	}
	return coreStyle.Render(formatCell("", width, false))
}

// inlineTrendCore is the wide-tier trend text (arrow + rate, plus the projection
// clause when rich), reusing the shared windowForecast pipeline.
func inlineTrendCore(m Model, spec quotaRowSpec, window sources.Window, now time.Time, rich bool) string {
	if window.UsedPercent >= 100 {
		return "full"
	}
	forecast, ok := windowForecast(m, spec, window, now)
	if !ok {
		return ""
	}
	return forecastCore(forecast, window.UsedPercent, rich)
}

func renderDataRow(m Model, spec quotaRowSpec, fullLabel string, shortLabel string, window sources.Window, fraction float64, evenUse float64, now time.Time, width int, highlighted bool) string {
	percentText := fmt.Sprintf("%.0f%%", math.Round(window.UsedPercent))
	atRisk := windowAtRisk(m, spec, window, now)
	percentStyle := lipgloss.NewStyle().Foreground(thresholdColor(window.UsedPercent))
	if atRisk {
		percentStyle = lipgloss.NewStyle().Foreground(pulseColor(m.animPhase))
	}
	if highlighted {
		// A just-changed value briefly bolds (zero-width) so the eye catches it.
		percentStyle = percentStyle.Bold(true)
	}
	percent := percentStyle.Render(formatCell(percentText, normalPercentWidth, true))
	reset := resetText(window.ResetsAt, now)

	switch {
	case width >= wideThreshold:
		barWidth := width - fullRowLabelWidth - normalPercentWidth - normalResetWidth - wideTrendWidth - 4*len(normalGap)
		if barWidth < minProgressWidth {
			barWidth = minProgressWidth
		}
		barCell := pacedBar(fraction, barWidth, evenUse)
		var trendCell string
		if m.prefs.trendVisible() {
			trendCell = wideTrendCell(m, spec, window, now, wideTrendWidth)
		} else {
			trendCell = formatCell("", wideTrendWidth, false)
		}
		return fmt.Sprintf(
			"%s  %s  %s  %s  %s",
			labelStyle.Render(formatCell(fullLabel, fullRowLabelWidth, false)),
			barCell,
			percent,
			trendCell,
			formatCell(reset, normalResetWidth, true),
		)
	case width >= 46:
		barWidth := width - fullRowLabelWidth - normalPercentWidth - normalResetWidth - 3*len(normalGap)
		barCell := pacedBar(fraction, barWidth, evenUse)

		return fmt.Sprintf(
			"%s  %s  %s  %s",
			labelStyle.Render(formatCell(fullLabel, fullRowLabelWidth, false)),
			barCell,
			percent,
			formatCell(reset, normalResetWidth, true),
		)
	case width >= 26:
		compactReset := compactResetText(window.ResetsAt, now)
		compactPercent := percentStyle.Render(formatCell(percentText, normalPercentWidth, true))
		barWidth := width - shortRowLabelWidth - normalPercentWidth - compactResetWidth - 3*len(compactGap)
		if barWidth >= minProgressWidth {
			return fmt.Sprintf(
				"%s %s %s %s",
				labelStyle.Render(formatCell(shortLabel, shortRowLabelWidth, false)),
				pacedBar(fraction, barWidth, evenUse),
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
		compactPercent := percentStyle.Render(formatCell(percentText, normalPercentWidth, true))
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
	case width >= wideThreshold:
		barWidth := width - fullRowLabelWidth - normalPercentWidth - normalResetWidth - wideTrendWidth - 4*len(normalGap)
		if barWidth < minProgressWidth {
			barWidth = minProgressWidth
		}
		return fmt.Sprintf(
			"%s  %s  %s  %s  %s",
			labelStyle.Render(formatCell(fullLabel, fullRowLabelWidth, false)),
			hintStyle.Render(formatCell("missing local data", barWidth, false)),
			percent,
			formatCell("", wideTrendWidth, false),
			reset,
		)
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

// renderTrendLine builds the second line (sparkline + rate + forecast) for a
// data row, choosing the richest variant that fits `width`. Returns ("", false)
// when there is no room (very narrow tier).
func renderTrendLine(m Model, spec quotaRowSpec, window sources.Window, now time.Time, width int) (string, bool) {
	if width < 26 {
		return "", false
	}
	// Defensive: production builds history via NewModel (never nil), but guard
	// against Model literals so a trend row can never panic the TUI. windowForecast
	// returns ok == false exactly when m.history == nil.
	forecast, ok := windowForecast(m, spec, window, now)
	if !ok {
		return "", false
	}
	indent := fullTrendIndent
	if width < 46 {
		indent = compactTrendIndent
	}

	samples := m.history.EpochSamples(trend.Key(spec.product, spec.kind), window.ResetsAt)
	spark := trend.Sparkline(samples, sparkWidth)

	alert := window.UsedPercent >= 100 || forecast.AtRisk
	prefix := ""
	if forecast.AtRisk {
		prefix = m.glyphs().warning + " "
	}
	coreStyle := hintStyle
	if alert {
		coreStyle = lipgloss.NewStyle().Foreground(pulseColor(m.animPhase))
	}
	prefixStyle := lipgloss.NewStyle().Foreground(pulseColor(m.animPhase))

	for _, rich := range []bool{true, false} {
		core := forecastCore(forecast, window.UsedPercent, rich)
		plain := strings.Repeat(" ", indent) + prefix + spark + "  " + core
		if lipgloss.Width(plain) > width {
			continue
		}
		line := strings.Repeat(" ", indent) +
			prefixStyle.Render(prefix) +
			hintStyle.Render(spark) +
			"  " +
			coreStyle.Render(core)
		return line, true
	}
	return "", false
}

// forecastCore is the text after the sparkline: arrow + rate, optionally with
// the projection clause. A maxed window shows just "full".
func forecastCore(f trend.Forecast, usedPct float64, rich bool) string {
	if usedPct >= 100 {
		return "full"
	}
	core := fmt.Sprintf("%c %s/hr", f.Arrow, rateText(f.Rate))
	if rich && f.Status != "" {
		core += " · " + f.Status
	}
	return core
}

func rateText(rate float64) string {
	if rate > 0 && rate < 1 {
		return "<1%"
	}
	return fmt.Sprintf("%.0f%%", rate)
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
