package tui

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/robbell5/llm-quota/internal/sources"
)

var ansiEscapeRE = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

func TestRenderStartupScreen(t *testing.T) {
	full := render(Model{width: 80, height: 12})

	if !strings.Contains(full, "LLM Quota") {
		t.Fatalf("expected title in startup screen, got:\n%s", full)
	}

	for _, label := range []string{"Claude 5h", "Claude 7d", "Sonnet 7d", "Codex 5h", "Codex 7d"} {
		if count := strings.Count(full, label); count != 1 {
			t.Fatalf("expected row label %q exactly once, got %d in:\n%s", label, count, full)
		}
	}

	if count := strings.Count(full, "—"); count < 8 {
		t.Fatalf("expected missing-data placeholders for percent and reset values, got %d em dashes in:\n%s", count, full)
	}

	if !strings.Contains(full, "missing local data") {
		t.Fatalf("expected missing local data hint in rows, got:\n%s", full)
	}

	const baselineFooter = "q / Ctrl-C quit · r refresh"
	if !strings.Contains(full, baselineFooter) {
		t.Fatalf("expected baseline footer %q in:\n%s", baselineFooter, full)
	}

	if strings.Contains(full, "Claude: run install-claude-hook") || strings.Contains(full, "Codex: open Codex") {
		t.Fatalf("startup screen should wait for source errors before showing recovery hints:\n%s", full)
	}

	compact := render(Model{width: 49, height: 12})
	if !strings.Contains(compact, baselineFooter) {
		t.Fatalf("expected compact baseline footer %q in:\n%s", baselineFooter, compact)
	}

	atFifty := render(Model{width: 50, height: 12})
	if !strings.Contains(atFifty, baselineFooter) {
		t.Fatalf("expected baseline footer at width 50 in:\n%s", atFifty)
	}

	if strings.Contains(compact, "Claude: run install-claude-hook") || strings.Contains(compact, "Codex: open Codex") {
		t.Fatalf("compact footer should avoid source hints that wrap:\n%s", compact)
	}

	for _, width := range []int{80, 50, 49, 30, 29, 20} {
		assertRenderedLineWidths(t, render(Model{width: width, height: 12}), width)
	}
}

func TestRenderSourceBackedRows(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	model := NewModel(WithClock(func() time.Time { return now }))
	model.width = 80
	model.height = 12
	model.windows[sources.ProductClaude] = []sources.Window{
		{
			Product:     sources.ProductClaude,
			Kind:        sources.WindowFiveHour,
			Label:       "Claude 5h",
			UsedPercent: 42,
			ResetsAt:    now.Add(2 * time.Hour),
			CapturedAt:  now.Add(-2 * time.Hour),
			Stale:       true,
			StaleAge:    2 * time.Hour,
		},
		{
			Product:     sources.ProductClaude,
			Kind:        sources.WindowSonnetSevenDay,
			Label:       "Sonnet 7d",
			UsedPercent: 91,
			ResetsAt:    now.Add(21*time.Hour + time.Minute),
			CapturedAt:  now,
		},
	}
	model.windows[sources.ProductCodex] = []sources.Window{
		{
			Product:     sources.ProductCodex,
			Kind:        sources.WindowSevenDay,
			Label:       "Codex 7d",
			UsedPercent: 17.4,
			ResetsAt:    now.Add(7 * 24 * time.Hour),
			CapturedAt:  now,
		},
	}

	rendered := render(model)
	plain := ansiEscapeRE.ReplaceAllString(rendered, "")
	for _, want := range []string{"Claude 5h", "42%", "Sonnet 7d", "91%", "21h 1m", "Codex 7d", "17%"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("expected rendered source-backed rows to contain %q, got:\n%s", want, plain)
		}
	}
	if strings.Contains(plain, "Claude 5h  —  missing local data") {
		t.Fatalf("stale-but-valid Claude data should render as data, not placeholder:\n%s", plain)
	}

	if !strings.Contains(plain, "Claude updated") {
		t.Fatalf("expected Claude freshness line in source-backed output, got:\n%s", plain)
	}
	if strings.Contains(plain, "Claude data 2h old; open Claude") {
		t.Fatalf("source-backed stale Claude output should not duplicate stale status in footer:\n%s", plain)
	}

	for _, forbidden := range []string{"malformed", "read_error", "no_usable_event"} {
		if strings.Contains(strings.ToLower(plain), forbidden) {
			t.Fatalf("render output should not expose raw error category %q:\n%s", forbidden, plain)
		}
	}
}

func TestRenderNormalQuotaRowsAlignRightColumns(t *testing.T) {
	fixedNow := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	model := NewModel(WithClock(func() time.Time { return fixedNow }))
	model.width = 80
	model.height = 12
	model.windows[sources.ProductClaude] = []sources.Window{
		{
			Product:     sources.ProductClaude,
			Kind:        sources.WindowFiveHour,
			Label:       "Claude 5h",
			UsedPercent: 0,
			ResetsAt:    fixedNow.Add(54 * time.Minute),
			CapturedAt:  fixedNow,
		},
		{
			Product:     sources.ProductClaude,
			Kind:        sources.WindowSevenDay,
			Label:       "Claude 7d",
			UsedPercent: 100,
			ResetsAt:    fixedNow.Add(21*time.Hour + time.Minute),
			CapturedAt:  fixedNow,
		},
		{
			Product:     sources.ProductClaude,
			Kind:        sources.WindowSonnetSevenDay,
			Label:       "Sonnet 7d",
			UsedPercent: 64,
			ResetsAt:    fixedNow.Add(4*24*time.Hour + 6*time.Hour),
			CapturedAt:  fixedNow,
		},
	}

	plain := ansiEscapeRE.ReplaceAllString(render(model), "")
	lines := quotaLines(plain)
	if len(lines) < 3 {
		t.Fatalf("expected at least three quota lines, got:\n%s", plain)
	}

	percentColumnEnd := strings.Index(lines[0], "0%") + len("0%")
	resetColumn := strings.Index(lines[0], "0h 54m")
	if percentColumnEnd < len("0%") || resetColumn < 0 {
		t.Fatalf("could not locate percent/reset columns in first line: %q", lines[0])
	}
	for _, line := range lines[:3] {
		if got := firstPercentEnd(line); got != percentColumnEnd {
			t.Fatalf("percent column end = %d, want %d in line %q", got, percentColumnEnd, line)
		}
		if got := firstResetIndex(line); got != resetColumn {
			t.Fatalf("reset column = %d, want %d in line %q", got, resetColumn, line)
		}
	}

	for _, want := range []string{"0h 54m", "21h 1m", "100%"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("expected aligned output to contain %q, got:\n%s", want, plain)
		}
	}
}

func TestRenderMissingAndStaleFooterHints(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	claudeMissing := NewModel(WithClock(func() time.Time { return now }))
	claudeMissing.width = 80
	claudeMissing.height = 12
	claudeMissing.errors[sources.ProductClaude] = sources.SourceError{Source: sources.ProductClaude, Category: sources.ErrorMissing}
	claudeMissing.errors[sources.ProductCodex] = sources.SourceError{Source: sources.ProductCodex, Category: sources.ErrorMalformed}
	plain := ansiEscapeRE.ReplaceAllString(render(claudeMissing), "")
	if !strings.Contains(plain, "Claude: run install-claude-hook") {
		t.Fatalf("expected Claude setup hint for missing Claude data, got:\n%s", plain)
	}

	installedButMissing := NewModel(
		WithClock(func() time.Time { return now }),
		WithClaudeHookInstalled(true),
	)
	installedButMissing.width = 80
	installedButMissing.height = 12
	installedButMissing.errors[sources.ProductClaude] = sources.SourceError{Source: sources.ProductClaude, Category: sources.ErrorMissing}
	plain = ansiEscapeRE.ReplaceAllString(render(installedButMissing), "")
	if !strings.Contains(plain, "Claude: open Claude") {
		t.Fatalf("expected open-Claude hint when hook is installed but cache is missing, got:\n%s", plain)
	}
	if strings.Contains(plain, "Claude: run install-claude-hook") {
		t.Fatalf("installed hook should not render setup hint for missing cache, got:\n%s", plain)
	}
	for _, forbidden := range []string{"malformed", "read_error", "no_usable_event"} {
		if strings.Contains(strings.ToLower(plain), forbidden) {
			t.Fatalf("render output should not expose raw error category %q:\n%s", forbidden, plain)
		}
	}

	codexMissing := NewModel(WithClock(func() time.Time { return now }))
	codexMissing.width = 80
	codexMissing.height = 12
	codexMissing.errors[sources.ProductCodex] = sources.SourceError{Source: sources.ProductCodex, Category: sources.ErrorNoUsableEvent}
	plain = ansiEscapeRE.ReplaceAllString(render(codexMissing), "")
	if !strings.Contains(plain, "Codex: open Codex") {
		t.Fatalf("expected Codex recovery hint when no Claude setup hint applies, got:\n%s", plain)
	}
	for _, forbidden := range []string{"malformed", "read_error", "no_usable_event"} {
		if strings.Contains(strings.ToLower(plain), forbidden) {
			t.Fatalf("render output should not expose raw error category %q:\n%s", forbidden, plain)
		}
	}

	staleClaude := NewModel(WithClock(func() time.Time { return now }))
	staleClaude.width = 80
	staleClaude.height = 12
	staleClaude.windows[sources.ProductClaude] = []sources.Window{
		{
			Product:     sources.ProductClaude,
			Kind:        sources.WindowFiveHour,
			Label:       "Claude 5h",
			UsedPercent: 42,
			ResetsAt:    now.Add(2 * time.Hour),
			CapturedAt:  now.Add(-2 * time.Hour),
			Stale:       true,
			StaleAge:    2 * time.Hour,
		},
	}
	plain = ansiEscapeRE.ReplaceAllString(render(staleClaude), "")
	for _, want := range []string{"Claude 5h", "42%", "Claude updated"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("expected stale-but-valid Claude output to contain %q, got:\n%s", want, plain)
		}
	}
	if strings.Contains(plain, "Claude data 2h old; open Claude") {
		t.Fatalf("stale-but-valid Claude output should carry stale status on freshness line, not footer:\n%s", plain)
	}
}

func TestRenderSourceFreshnessLines(t *testing.T) {
	now := time.Date(2026, 5, 19, 14, 17, 0, 0, time.Local)
	captured := time.Date(2026, 5, 19, 14, 14, 0, 0, time.Local)
	model := NewModel(WithClock(func() time.Time { return now }))
	model.width = 80
	model.height = 12
	model.windows[sources.ProductClaude] = []sources.Window{
		{
			Product:     sources.ProductClaude,
			Kind:        sources.WindowFiveHour,
			Label:       "Claude 5h",
			UsedPercent: 42,
			ResetsAt:    now.Add(2 * time.Hour),
			CapturedAt:  captured,
		},
	}
	model.windows[sources.ProductCodex] = []sources.Window{
		{
			Product:     sources.ProductCodex,
			Kind:        sources.WindowSevenDay,
			Label:       "Codex 7d",
			UsedPercent: 18,
			ResetsAt:    now.Add(2 * time.Hour),
			CapturedAt:  captured,
		},
	}

	plain := ansiEscapeRE.ReplaceAllString(render(model), "")
	for _, want := range []string{
		"Claude updated 2:14 PM (3m ago)",
		"Codex updated 2:14 PM (3m ago)",
	} {
		if !strings.Contains(plain, want) {
			t.Fatalf("expected freshness line %q, got:\n%s", want, plain)
		}
	}
	assertLineOrder(t, plain, "Sonnet 7d", "Claude updated 2:14 PM")
	assertLineOrder(t, plain, "Codex 7d", "Codex updated 2:14 PM")

	for _, width := range []int{49, 30, 20} {
		model.width = width
		plain = ansiEscapeRE.ReplaceAllString(render(model), "")
		assertRenderedLineWidths(t, render(model), width)
		switch width {
		case 49, 30:
			if !strings.Contains(plain, "Claude 2:14 PM") || !strings.Contains(plain, "Codex 2:14 PM") {
				t.Fatalf("width %d: expected compact freshness labels, got:\n%s", width, plain)
			}
		case 20:
			if !strings.Contains(plain, "Cl 2:14") || !strings.Contains(plain, "Cx 2:14") {
				t.Fatalf("width %d: expected very narrow freshness labels, got:\n%s", width, plain)
			}
		}
	}
}

func TestRenderRefreshFailureOnFreshnessLine(t *testing.T) {
	now := time.Date(2026, 5, 19, 16, 14, 0, 0, time.Local)
	captured := time.Date(2026, 5, 19, 14, 14, 0, 0, time.Local)
	model := NewModel(WithClock(func() time.Time { return now }))
	model.width = 80
	model.height = 12
	model.windows[sources.ProductClaude] = []sources.Window{
		{
			Product:     sources.ProductClaude,
			Kind:        sources.WindowFiveHour,
			Label:       "Claude 5h",
			UsedPercent: 42,
			ResetsAt:    now.Add(2 * time.Hour),
			CapturedAt:  captured,
			Stale:       true,
			StaleAge:    2 * time.Hour,
		},
	}
	model.errors[sources.ProductClaude] = sources.SourceError{Source: sources.ProductClaude, Category: sources.ErrorMalformed}

	plain := ansiEscapeRE.ReplaceAllString(render(model), "")
	if !strings.Contains(plain, "Claude updated 2:14 PM (2h old, refresh failed)") {
		t.Fatalf("expected combined stale/current-error freshness status, got:\n%s", plain)
	}
	if strings.Contains(plain, "Claude data 2h old; open Claude") {
		t.Fatalf("footer should not duplicate source status when Claude rows exist:\n%s", plain)
	}
	for _, forbidden := range []string{"malformed", "read_error", "no_usable_event"} {
		if strings.Contains(strings.ToLower(plain), forbidden) {
			t.Fatalf("render output should not expose raw error category %q:\n%s", forbidden, plain)
		}
	}
}

func TestRenderQuotaRowsWithThresholdProgressBars(t *testing.T) {
	fixedNow := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	model := NewModel(WithClock(func() time.Time { return fixedNow }))
	model.height = 12
	model.windows[sources.ProductClaude] = []sources.Window{
		{
			Product:     sources.ProductClaude,
			Kind:        sources.WindowFiveHour,
			Label:       "Claude 5h",
			UsedPercent: 59,
			ResetsAt:    fixedNow.Add(2*time.Hour + 14*time.Minute),
			CapturedAt:  fixedNow,
		},
		{
			Product:     sources.ProductClaude,
			Kind:        sources.WindowSevenDay,
			Label:       "Claude 7d",
			UsedPercent: 60,
			ResetsAt:    fixedNow.Add(4*24*time.Hour + 6*time.Hour),
			CapturedAt:  fixedNow,
		},
	}
	model.windows[sources.ProductCodex] = []sources.Window{
		{
			Product:     sources.ProductCodex,
			Kind:        sources.WindowFiveHour,
			Label:       "Codex 5h",
			UsedPercent: 85,
			ResetsAt:    fixedNow.Add(-time.Minute),
			CapturedAt:  fixedNow,
		},
		{
			Product:     sources.ProductCodex,
			Kind:        sources.WindowSevenDay,
			Label:       "Codex 7d",
			UsedPercent: 17,
			ResetsAt:    fixedNow.Add(5*24*time.Hour + 2*time.Hour),
			CapturedAt:  fixedNow,
		},
	}

	for _, width := range []int{80, 50} {
		model.width = width
		rendered := render(model)
		plain := ansiEscapeRE.ReplaceAllString(rendered, "")

		for _, want := range []string{
			"Claude 5h", "Claude 7d", "Codex 5h", "Codex 7d",
			"59%", "60%", "85%", "17%",
			"2h 14m", "4d 06h", "now", "5d 02h",
		} {
			if !strings.Contains(plain, want) {
				t.Fatalf("width %d: expected rendered quota rows to contain %q, got:\n%s", width, want, plain)
			}
		}

		assertRenderedLineWidths(t, rendered, width)
	}
}

func TestRenderResponsiveQuotaLayouts(t *testing.T) {
	fixedNow := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	model := NewModel(WithClock(func() time.Time { return fixedNow }))
	model.height = 12
	model.windows[sources.ProductClaude] = []sources.Window{
		{
			Product:     sources.ProductClaude,
			Kind:        sources.WindowFiveHour,
			Label:       "Claude 5h",
			UsedPercent: 23,
			ResetsAt:    fixedNow.Add(2*time.Hour + 14*time.Minute),
			CapturedAt:  fixedNow,
		},
		{
			Product:     sources.ProductClaude,
			Kind:        sources.WindowSevenDay,
			Label:       "Claude 7d",
			UsedPercent: 61,
			ResetsAt:    fixedNow.Add(4*24*time.Hour + 6*time.Hour),
			CapturedAt:  fixedNow,
		},
	}
	model.windows[sources.ProductCodex] = []sources.Window{
		{
			Product:     sources.ProductCodex,
			Kind:        sources.WindowFiveHour,
			Label:       "Codex 5h",
			UsedPercent: 86,
			ResetsAt:    fixedNow.Add(45 * time.Minute),
			CapturedAt:  fixedNow,
		},
		{
			Product:     sources.ProductCodex,
			Kind:        sources.WindowSevenDay,
			Label:       "Codex 7d",
			UsedPercent: 17,
			ResetsAt:    fixedNow.Add(5*24*time.Hour + 2*time.Hour),
			CapturedAt:  fixedNow,
		},
	}

	for _, width := range []int{50, 49, 30, 29, 20} {
		model.width = width
		rendered := render(model)
		plain := ansiEscapeRE.ReplaceAllString(rendered, "")
		assertRenderedLineWidths(t, rendered, width)

		if !strings.Contains(plain, "23%") || !strings.Contains(plain, "17%") {
			t.Fatalf("width %d: expected narrow layout to preserve percent text, got:\n%s", width, plain)
		}
		if width == 49 {
			for _, want := range []string{"Cl 5h", "Cx 7d"} {
				if !strings.Contains(plain, want) {
					t.Fatalf("width 49: expected abbreviated label %q, got:\n%s", want, plain)
				}
			}
		}
		if width == 29 {
			if strings.Contains(plain, "Claude 5h") {
				t.Fatalf("width 29: expected short labels instead of full labels, got:\n%s", plain)
			}
			if strings.Contains(plain, "━") {
				t.Fatalf("width 29: expected progress bars to be omitted, got:\n%s", plain)
			}
		}
	}
}

func assertRenderedLineWidths(t *testing.T, rendered string, maxWidth int) {
	t.Helper()

	for lineNumber, line := range strings.Split(strings.TrimSuffix(rendered, "\n"), "\n") {
		plain := ansiEscapeRE.ReplaceAllString(line, "")
		if width := lipgloss.Width(plain); width > maxWidth {
			t.Fatalf("line %d width = %d, want <= %d: %q", lineNumber+1, width, maxWidth, plain)
		}
	}
}

func assertLineOrder(t *testing.T, plain string, before string, after string) {
	t.Helper()

	beforeIndex := strings.Index(plain, before)
	afterIndex := strings.Index(plain, after)
	if beforeIndex < 0 || afterIndex < 0 {
		t.Fatalf("could not find %q before %q in:\n%s", before, after, plain)
	}
	if beforeIndex >= afterIndex {
		t.Fatalf("expected %q before %q in:\n%s", before, after, plain)
	}
}

func quotaLines(plain string) []string {
	lines := make([]string, 0, 5)
	for _, line := range strings.Split(plain, "\n") {
		trimmed := strings.TrimSpace(line)
		for _, label := range []string{"Claude 5h", "Claude 7d", "Sonnet 7d", "Codex 5h", "Codex 7d"} {
			if strings.HasPrefix(trimmed, label) {
				lines = append(lines, strings.TrimRight(line, " "))
				break
			}
		}
	}
	return lines
}

func firstPercentEnd(line string) int {
	best := -1
	for _, marker := range []string{"0%", "100%", "64%"} {
		if idx := strings.Index(line, marker); idx >= 0 && (best == -1 || idx < best) {
			best = idx + len(marker)
		}
	}
	return best
}

func firstResetIndex(line string) int {
	best := -1
	for _, marker := range []string{"0h 54m", "21h 1m", "4d 06h"} {
		if idx := strings.Index(line, marker); idx >= 0 && (best == -1 || idx < best) {
			best = idx
		}
	}
	return best
}

func sampleBothProviders() Model {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	m := NewModel(WithClock(func() time.Time { return now }))
	m.height = 12
	m.windows[sources.ProductClaude] = []sources.Window{
		{Product: sources.ProductClaude, Kind: sources.WindowFiveHour, Label: "Claude 5h", UsedPercent: 40, ResetsAt: now.Add(time.Hour), CapturedAt: now},
		{Product: sources.ProductClaude, Kind: sources.WindowSevenDay, Label: "Claude 7d", UsedPercent: 55, ResetsAt: now.Add(48 * time.Hour), CapturedAt: now},
		{Product: sources.ProductClaude, Kind: sources.WindowSonnetSevenDay, Label: "Sonnet 7d", UsedPercent: 25, ResetsAt: now.Add(72 * time.Hour), CapturedAt: now},
	}
	m.windows[sources.ProductCodex] = []sources.Window{
		{Product: sources.ProductCodex, Kind: sources.WindowFiveHour, Label: "Codex 5h", UsedPercent: 70, ResetsAt: now.Add(2 * time.Hour), CapturedAt: now},
		{Product: sources.ProductCodex, Kind: sources.WindowSevenDay, Label: "Codex 7d", UsedPercent: 30, ResetsAt: now.Add(72 * time.Hour), CapturedAt: now},
	}
	return m
}

func TestVisibilityFiltersRows(t *testing.T) {
	cases := []struct {
		name    string
		vis     Visibility
		present []string
		absent  []string
	}{
		{"both", VisibilityBoth, []string{"Claude 5h", "Sonnet 7d", "Codex 5h", "Codex 7d"}, nil},
		{"claude only", VisibilityClaudeOnly, []string{"Claude 5h", "Claude 7d", "Sonnet 7d"}, []string{"Codex 5h", "Codex 7d"}},
		{"codex only", VisibilityCodexOnly, []string{"Codex 5h", "Codex 7d"}, []string{"Claude 5h", "Sonnet 7d"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := sampleBothProviders()
			m.prefs.Visibility = c.vis
			m.width = 80
			plain := ansiEscapeRE.ReplaceAllString(render(m), "")
			for _, want := range c.present {
				if !strings.Contains(plain, want) {
					t.Fatalf("expected %q present, got:\n%s", want, plain)
				}
			}
			for _, gone := range c.absent {
				if strings.Contains(plain, gone) {
					t.Fatalf("expected %q absent, got:\n%s", gone, plain)
				}
			}
		})
	}
}

func TestVisibilityHidesFreshnessLine(t *testing.T) {
	m := sampleBothProviders()
	m.prefs.Visibility = VisibilityClaudeOnly
	m.width = 80
	plain := ansiEscapeRE.ReplaceAllString(render(m), "")
	if !strings.Contains(plain, "Claude updated") {
		t.Fatalf("expected Claude freshness line, got:\n%s", plain)
	}
	if strings.Contains(plain, "Codex updated") {
		t.Fatalf("expected no Codex freshness line, got:\n%s", plain)
	}
}

func TestBarStyleFillRune(t *testing.T) {
	cases := []struct {
		name     string
		style    BarStyle
		wantRune string
		denyRune string
	}{
		{"segmented", BarSegmented, "▌", "█"},
		{"solid", BarSolid, "█", "▌"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := sampleBothProviders()
			m.prefs.BarStyle = c.style
			m.width = 80
			plain := ansiEscapeRE.ReplaceAllString(render(m), "")
			if !strings.Contains(plain, c.wantRune) {
				t.Fatalf("expected fill rune %q, got:\n%s", c.wantRune, plain)
			}
			if c.denyRune != "" && strings.Contains(plain, c.denyRune) {
				t.Fatalf("did not expect rune %q for %s, got:\n%s", c.denyRune, c.name, plain)
			}
		})
	}
}

func TestVisibilityKeepsLineWidths(t *testing.T) {
	for _, vis := range []Visibility{VisibilityBoth, VisibilityClaudeOnly, VisibilityCodexOnly} {
		for _, width := range []int{80, 50, 49, 30, 29, 20} {
			t.Run(fmt.Sprintf("vis=%s/width=%d", vis, width), func(t *testing.T) {
				m := sampleBothProviders()
				m.prefs.Visibility = vis
				m.width = width
				assertRenderedLineWidths(t, render(m), width)
			})
		}
	}
}

func TestRenderBarWhileAnimatingIsWidthSafe(t *testing.T) {
	m := NewModel()
	i := barIndex(t, sources.ProductClaude, sources.WindowFiveHour)
	// Kick an animation toward 60% and advance one real frame so the bar is mid-flight.
	cmd := m.bars[i].SetPercent(0.6)
	if cmd != nil {
		m.bars[i], _ = m.bars[i].Update(cmd())
	}
	if !m.bars[i].IsAnimating() {
		t.Skip("spring settled immediately; cannot exercise the animating branch")
	}
	// While animating, renderBar takes the View() branch. It must still produce
	// exactly `width` cells (no overflow/underflow). Assert rune count, not content.
	const width = 20
	plain := ansiEscapeRE.ReplaceAllString(renderBar(m.bars[i], 60, width, BarSegmented), "")
	if got := len([]rune(plain)); got != width {
		t.Fatalf("animating bar width = %d runes, want %d: %q", got, width, plain)
	}
}
