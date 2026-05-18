package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

const (
	defaultWidth           = 50
	shellHorizontalPadding = 4
	shellPaddingX          = shellHorizontalPadding / 2
)

const (
	fullFooter    = "q / Ctrl-C quit · Claude: install hook · Codex: open Codex"
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
		renderRows(innerWidth),
		"",
		footerStyle.Width(innerWidth).Render(renderFooter(width)),
	}, "\n")

	return shellStyle.Render(content) + "\n"
}

func renderRows(width int) string {
	rows := make([]string, 0, 4)
	rowLabels := []struct {
		full  string
		short string
	}{
		{full: "Claude 5h", short: "Cl 5h"},
		{full: "Claude 7d", short: "Cl 7d"},
		{full: "Codex 5h", short: "Cx 5h"},
		{full: "Codex 7d", short: "Cx 7d"},
	}

	for _, label := range rowLabels {
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

func renderFooter(width int) string {
	if lipgloss.Width(fullFooter)+shellHorizontalPadding > width {
		return compactFooter
	}

	return fullFooter
}
