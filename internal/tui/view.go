package tui

import (
	"strings"
	"time"

	"charm.land/lipgloss/v2"
)

const (
	defaultWidth           = 50
	shellHorizontalPadding = 4
	shellPaddingX          = shellHorizontalPadding / 2
	wideThreshold          = 68 // inner width at/above which the wide tier renders
)

var (
	shellStyle = lipgloss.NewStyle().
			Background(mochaBase).
			Foreground(mochaText).
			Padding(1, shellPaddingX)
	titleStyle = lipgloss.NewStyle().
			Foreground(mochaBlue).
			Bold(true)
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

	now := time.Now
	if m.now != nil {
		now = m.now
	}
	content := strings.Join([]string{
		renderTitleBand(innerWidth, now(), m.glyphs()),
		renderRows(m, innerWidth),
		"",
		footerStyle.Width(innerWidth).Render(renderFooter(m, innerWidth)),
	}, "\n")

	return shellStyle.Render(content) + "\n"
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
