package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/rob/llm-quota/internal/sources"
)

const (
	defaultWidth           = 50
	shellHorizontalPadding = 4
	shellPaddingX          = shellHorizontalPadding / 2
)

const (
	fullFooter    = "q / Ctrl-C quit · Claude: run install-claude-hook · Codex: open Codex"
	compactFooter = "q / Ctrl-C quit · data pending"
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
		footerStyle.Width(innerWidth).Render(renderFooter(width)),
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
		case width >= 41:
			rows = append(rows, fmt.Sprintf(
				"%s  %s  %s  reset %s",
				labelStyle.Render(fmt.Sprintf("%-9s", label.full)),
				missingStyle.Render("—"),
				hintStyle.Render("missing local data"),
				missingStyle.Render("—"),
			))
		case width >= 21:
			rows = append(rows, fmt.Sprintf(
				"%s  %s  %s",
				labelStyle.Render(fmt.Sprintf("%-9s", label.full)),
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
	percent := fmt.Sprintf("%.0f%%", math.Round(window.UsedPercent))
	reset := resetText(window.ResetsAt, now)

	switch {
	case width >= 41:
		return fmt.Sprintf(
			"%s  %s  reset %s",
			labelStyle.Render(fmt.Sprintf("%-9s", fullLabel)),
			percent,
			reset,
		)
	case width >= 21:
		return fmt.Sprintf(
			"%s  %s  %s",
			labelStyle.Render(fmt.Sprintf("%-9s", fullLabel)),
			percent,
			reset,
		)
	default:
		return fmt.Sprintf(
			"%s %s",
			labelStyle.Render(fmt.Sprintf("%-5s", shortLabel)),
			percent,
		)
	}
}

func resetText(resetsAt time.Time, now time.Time) string {
	if resetsAt.IsZero() {
		return missingStyle.Render("—")
	}

	remaining := resetsAt.Sub(now)
	if remaining < 0 {
		remaining = 0
	}
	if remaining >= 24*time.Hour {
		days := int(math.Ceil(remaining.Hours() / 24))
		return fmt.Sprintf("%dd", days)
	}
	if remaining >= time.Hour {
		hours := int(math.Ceil(remaining.Hours()))
		return fmt.Sprintf("%dh", hours)
	}
	minutes := int(math.Ceil(remaining.Minutes()))
	if minutes < 1 {
		minutes = 1
	}
	return fmt.Sprintf("%dm", minutes)
}

func renderFooter(width int) string {
	if lipgloss.Width(fullFooter)+shellHorizontalPadding > width {
		return compactFooter
	}

	return fullFooter
}
