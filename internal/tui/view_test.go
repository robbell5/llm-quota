package tui

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/robbell5/llm-quota/internal/sources"
	"github.com/robbell5/llm-quota/internal/trend"
)

var ansiEscapeRE = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

func TestRenderStartupScreen(t *testing.T) {
	full := render(Model{width: 80, height: 12})

	if !strings.Contains(full, "LLM QUOTA") {
		t.Fatalf("expected title in startup screen, got:\n%s", full)
	}

	// At width 80 (grouped) rows live under CLAUDE/CODEX headers with short
	// per-window labels rather than the full "Claude 5h"/"Codex 7d" labels.
	for _, header := range []string{"CLAUDE", "CODEX"} {
		if count := strings.Count(full, header); count != 1 {
			t.Fatalf("expected group header %q exactly once, got %d in:\n%s", header, count, full)
		}
	}
	plain := ansiEscapeRE.ReplaceAllString(full, "")
	for _, label := range []string{"5h", "7d", "Sonnet 7d"} {
		if !strings.Contains(plain, label) {
			t.Fatalf("expected short window label %q under group header, got:\n%s", label, plain)
		}
	}
	if strings.Contains(plain, "Claude 5h") || strings.Contains(plain, "Codex 7d") {
		t.Fatalf("grouped layout should drop full provider-prefixed labels, got:\n%s", plain)
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

	for _, width := range []int{120, 100, 80, 72, 50, 49, 30, 29, 20} {
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
	// Grouped at width 80: CLAUDE/CODEX headers carry freshness; rows use short
	// per-window labels (5h, Sonnet 7d, 7d) under those headers.
	for _, want := range []string{"CLAUDE", "5h", "42%", "Sonnet 7d", "91%", "21h 1m", "CODEX", "7d", "17%"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("expected rendered source-backed rows to contain %q, got:\n%s", want, plain)
		}
	}
	if strings.Contains(plain, "5h         —  missing local data") {
		t.Fatalf("stale-but-valid Claude data should render as data, not placeholder:\n%s", plain)
	}

	// Freshness now lives in the Claude group header (the 5h window was captured
	// 2h ago and is stale).
	if !strings.Contains(plain, "updated") || !strings.Contains(plain, "2h old") {
		t.Fatalf("expected Claude freshness in group header, got:\n%s", plain)
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
	// Use a width below wideThreshold (inner < 68) so the normal 4-column layout
	// applies; the wide tier's UTF-8 trend characters shift byte offsets and break
	// the column-alignment check which uses strings.Index on plain text.
	model.width = 71
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
	// Grouped at width 80: 5h row carries the data, the CLAUDE header carries the
	// stale freshness ("updated ... · 2h old").
	for _, want := range []string{"CLAUDE", "5h", "42%", "updated", "2h old"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("expected stale-but-valid Claude output to contain %q, got:\n%s", want, plain)
		}
	}
	if strings.Contains(plain, "Claude data 2h old; open Claude") {
		t.Fatalf("stale-but-valid Claude output should carry stale status in group header, not footer:\n%s", plain)
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

	// Freshness now lives in each provider's group header (top of the group),
	// with " · " separators instead of the old "(Nm ago)" parens.
	model.width = 90
	plain := ansiEscapeRE.ReplaceAllString(render(model), "")
	if count := strings.Count(plain, "updated 2:14 PM · 3m ago"); count != 2 {
		t.Fatalf("expected freshness %q in both group headers (got %d):\n%s", "updated 2:14 PM · 3m ago", count, plain)
	}
	for _, header := range []string{"CLAUDE", "CODEX"} {
		if !strings.Contains(plain, header) {
			t.Fatalf("expected %q group header, got:\n%s", header, plain)
		}
	}
	// The group header (with its freshness) leads each provider group, so the
	// header precedes that provider's data rows.
	assertLineOrder(t, plain, "CLAUDE", "CODEX")

	// Below the grouping threshold (terminal < 50) there are no group headers and
	// therefore no freshness — by design freshness lives only in the header.
	for _, width := range []int{49, 30, 20} {
		model.width = width
		plain = ansiEscapeRE.ReplaceAllString(render(model), "")
		assertRenderedLineWidths(t, render(model), width)
		if strings.Contains(plain, "updated") || strings.Contains(plain, "2:14") {
			t.Fatalf("width %d: ungrouped layout should not show freshness, got:\n%s", width, plain)
		}
	}
}

func TestRenderRefreshFailureOnFreshnessLine(t *testing.T) {
	now := time.Date(2026, 5, 19, 16, 14, 0, 0, time.Local)
	captured := time.Date(2026, 5, 19, 14, 14, 0, 0, time.Local)
	model := NewModel(WithClock(func() time.Time { return now }))
	model.width = 90
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
	// Combined stale + refresh-failure status now lives in the CLAUDE group
	// header, joined by " · ".
	if !strings.Contains(plain, "updated 2:14 PM · 2h old · refresh failed") {
		t.Fatalf("expected combined stale/current-error freshness status in group header, got:\n%s", plain)
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

		// Grouped at both 80 and 50: providers carry CLAUDE/CODEX headers; the
		// per-window data (percent + reset) still appears for all four windows.
		for _, want := range []string{
			"CLAUDE", "CODEX",
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
		// Grouped layout uses short per-window labels (5h, 7d, Sonnet 7d) under
		// CLAUDE/CODEX headers. "Sonnet 7d" must precede "7d" so the longer label
		// wins the prefix match.
		for _, label := range []string{"Sonnet 7d", "5h", "7d"} {
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
		// Grouped at width 80: visibility is observable via the CLAUDE/CODEX group
		// headers (per-window labels 5h/7d are shared across providers). Sonnet 7d
		// is Claude-only, so it doubles as a Claude-presence marker.
		{"both", VisibilityBoth, []string{"CLAUDE", "CODEX", "Sonnet 7d"}, nil},
		{"claude only", VisibilityClaudeOnly, []string{"CLAUDE", "Sonnet 7d"}, []string{"CODEX"}},
		{"codex only", VisibilityCodexOnly, []string{"CODEX"}, []string{"CLAUDE", "Sonnet 7d"}},
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
	// Freshness lives in the group header, so a hidden provider has no header and
	// therefore no freshness at all.
	t.Run("claude only", func(t *testing.T) {
		m := sampleBothProviders()
		m.prefs.Visibility = VisibilityClaudeOnly
		m.width = 80
		plain := ansiEscapeRE.ReplaceAllString(render(m), "")
		if !strings.Contains(plain, "CLAUDE") || !strings.Contains(plain, "updated") {
			t.Fatalf("expected Claude group header with freshness, got:\n%s", plain)
		}
		if strings.Contains(plain, "CODEX") {
			t.Fatalf("expected no Codex group header (and thus no Codex freshness), got:\n%s", plain)
		}
	})

	t.Run("codex only", func(t *testing.T) {
		m := sampleBothProviders()
		m.prefs.Visibility = VisibilityCodexOnly
		m.width = 80
		plain := ansiEscapeRE.ReplaceAllString(render(m), "")
		if !strings.Contains(plain, "CODEX") || !strings.Contains(plain, "updated") {
			t.Fatalf("expected Codex group header with freshness, got:\n%s", plain)
		}
		if strings.Contains(plain, "CLAUDE") {
			t.Fatalf("expected no Claude group header (and thus no Claude freshness), got:\n%s", plain)
		}
	})
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

func TestReplaceCellAtPlainString(t *testing.T) {
	if got := replaceCellAt("abcde", 2, "X"); got != "abXde" {
		t.Fatalf("replaceCellAt = %q, want abXde", got)
	}
}

func TestReplaceCellAtPreservesWidthWithANSI(t *testing.T) {
	styled := lipgloss.NewStyle().Foreground(mochaGreen).Render("█████")
	out := replaceCellAt(styled, 2, lipgloss.NewStyle().Foreground(mochaText).Render("▕"))
	plain := ansiEscapeRE.ReplaceAllString(out, "")
	if plain != "██▕██" {
		t.Fatalf("plain = %q, want ██▕██", plain)
	}
	if got := lipgloss.Width(out); got != 5 {
		t.Fatalf("width = %d, want 5", got)
	}
}

func TestEvenUseTickAppearsInDataRow(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	m := NewModel(WithClock(func() time.Time { return now }))
	m.width = 80
	m.height = 12
	m.windows[sources.ProductClaude] = []sources.Window{
		{Product: sources.ProductClaude, Kind: sources.WindowFiveHour, Label: "Claude 5h",
			UsedPercent: 50, ResetsAt: now.Add(2 * time.Hour), CapturedAt: now},
	}
	plain := ansiEscapeRE.ReplaceAllString(render(m), "")
	if !strings.Contains(plain, paceMarkerRune) {
		t.Fatalf("expected even-use pace marker in rendered row, got:\n%s", plain)
	}
	for _, width := range []int{80, 50, 49, 30} {
		m.width = width
		assertRenderedLineWidths(t, render(m), width)
	}
}

func TestPacedBarIsWidthSafeAcrossFractions(t *testing.T) {
	// The spring drives pos to intermediate (and briefly out-of-range) fractions
	// each frame. pacedBar must always emit exactly `width` cells regardless.
	const width = 20
	for _, fraction := range []float64{-0.2, 0, 0.001, 0.37, 0.6, 0.999, 1, 1.3} {
		plain := ansiEscapeRE.ReplaceAllString(pacedBar(fraction, width, 0.5), "")
		if got := len([]rune(plain)); got != width {
			t.Fatalf("pacedBar(%v) width = %d runes, want %d: %q", fraction, got, width, plain)
		}
	}
}

func modelWithHistory(now time.Time, kind sources.WindowKind, percents []float64, capStepMin int, usedPct float64, resetsAt time.Time) Model {
	m := NewModel(WithClock(func() time.Time { return now }))
	m.width = 80
	m.height = 12
	key := trend.Key(sources.ProductClaude, kind)
	for i, p := range percents {
		m.history.Append(key, trend.Sample{
			CapturedAt: now.Add(time.Duration(-(len(percents)-1-i)*capStepMin) * time.Minute),
			UsedPct:    p,
			ResetsAt:   resetsAt,
		})
	}
	m.windows[sources.ProductClaude] = []sources.Window{
		{Product: sources.ProductClaude, Kind: kind, Label: "Claude 5h",
			UsedPercent: usedPct, ResetsAt: resetsAt, CapturedAt: now},
	}
	return m
}

func TestSecondLineSafeShowsProjection(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	reset := now.Add(4 * time.Hour)
	// Slow climb 10 -> 12 over the last hour: lots of headroom, never hits 100.
	m := modelWithHistory(now, sources.WindowFiveHour, []float64{10, 12}, 60, 12, reset)
	plain := ansiEscapeRE.ReplaceAllString(render(m), "")
	if !strings.Contains(plain, "by reset") {
		t.Fatalf("expected safe projection 'by reset', got:\n%s", plain)
	}
	if strings.Contains(plain, "⚠") {
		t.Fatalf("safe window should not show the warning marker:\n%s", plain)
	}
}

func TestSecondLineAtRiskShowsWarning(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	reset := now.Add(2 * time.Hour)
	// Fast climb 20 -> 60 over the last hour (40%/hr) -> full in 1h < 2h to reset.
	m := modelWithHistory(now, sources.WindowFiveHour, []float64{20, 60}, 60, 60, reset)
	plain := ansiEscapeRE.ReplaceAllString(render(m), "")
	if !strings.Contains(plain, "⚠") || !strings.Contains(plain, "full in") {
		t.Fatalf("expected at-risk warning with 'full in', got:\n%s", plain)
	}
}

func TestSecondLineHiddenWhenTrendOff(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	reset := now.Add(4 * time.Hour)
	m := modelWithHistory(now, sources.WindowFiveHour, []float64{10, 12}, 60, 12, reset)
	m.prefs.HideTrend = true
	plain := ansiEscapeRE.ReplaceAllString(render(m), "")
	if strings.Contains(plain, "by reset") || strings.Contains(plain, "%/hr") {
		t.Fatalf("trend-off should omit the second line, got:\n%s", plain)
	}
}

func TestWideTierFoldsTrendInline(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	m := NewModel(WithClock(func() time.Time { return now }))
	m.windows[sources.ProductClaude] = []sources.Window{{
		Product: sources.ProductClaude, Kind: sources.WindowFiveHour,
		UsedPercent: 35, ResetsAt: now.Add(2 * time.Hour), CapturedAt: now,
	}}
	m.bars[0].target = 0.35
	m.bars[0].pos = 0.35

	wide := render(withWidth(m, 100))
	skinny := render(withWidth(m, 50))

	if !lineWithRateIsDataRow(wide) {
		t.Fatalf("wide tier should fold the rate onto the data row:\n%s", wide)
	}
	if lineWithRateIsDataRow(skinny) {
		t.Fatalf("skinny tier should keep the rate on its own indented line:\n%s", skinny)
	}
}

func TestHighlightedRowBoldsPercent(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	reset := now.Add(2 * time.Hour)
	m := modelWithHistory(now, sources.WindowFiveHour, []float64{10, 20}, 60, 20, reset)
	// boldPercent matches a bold SGR (the "1;" parameter) wrapping the percent
	// token; a non-highlighted percent is wrapped in a plain (non-bold) SGR.
	boldPercent := regexp.MustCompile(`\x1b\[1;[0-9;]*m\s*20%`)

	if boldPercent.MatchString(render(m)) {
		t.Fatalf("percent should not be bold without an active highlight:\n%s", render(m))
	}

	m.highlightUntil[0] = now.Add(highlightDuration)
	if !boldPercent.MatchString(render(m)) {
		t.Fatalf("a just-changed value should bold its percent:\n%s", render(m))
	}
}

func TestRefreshingShowsSpinnerInGroupHeader(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	m := NewModel(WithClock(func() time.Time { return now }))
	m.windows[sources.ProductClaude] = []sources.Window{{
		Product: sources.ProductClaude, Kind: sources.WindowFiveHour,
		UsedPercent: 20, ResetsAt: now.Add(2 * time.Hour), CapturedAt: now,
	}}
	m.refreshing = true
	full := render(withWidth(m, 100))
	g := glyphsFor(false)
	if !strings.Contains(full, spinnerFrame(m.animPhase, g)) {
		t.Fatalf("refreshing should show a spinner glyph:\n%s", full)
	}
}

func withWidth(m Model, w int) Model { m.width = w; return m }

// lineWithRateIsDataRow reports whether the line containing "/hr" also contains a
// usage percent sign that appears before the rate token (i.e., the rate sits on
// the data row, not a standalone trend line). On a data row the usage percent
// ("35%") precedes the rate ("5%/hr") by many characters; on a standalone trend
// line the only percent is the one directly before "/hr" (as in "0%/hr").
func lineWithRateIsDataRow(view string) bool {
	for _, line := range strings.Split(view, "\n") {
		hrIdx := strings.Index(line, "/hr")
		if hrIdx < 0 {
			continue
		}
		pctIdx := strings.Index(line, "%")
		// True only when a percent sign appears well before "/hr", meaning it is a
		// separate usage percent rather than being part of the rate token itself.
		return pctIdx >= 0 && pctIdx < hrIdx-1
	}
	return false
}

func TestMissingRowDimUnderGroup(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	m := NewModel(WithClock(func() time.Time { return now }))
	m.windows[sources.ProductClaude] = []sources.Window{{
		Product: sources.ProductClaude, Kind: sources.WindowFiveHour,
		UsedPercent: 20, ResetsAt: now.Add(2 * time.Hour), CapturedAt: now,
	}}
	full := render(withWidth(m, 100))
	// Sonnet 7d has no window -> a missing row with its short grouped label.
	if !strings.Contains(full, "Sonnet") {
		t.Fatalf("expected Sonnet missing row under CLAUDE:\n%s", full)
	}
	if !strings.Contains(full, "missing local data") {
		t.Fatalf("expected missing-data hint:\n%s", full)
	}
	assertRenderedLineWidths(t, full, 100)
}

// TestWideTierColumnsAlign verifies that data rows and missing rows in the wide
// tier (inner width ≥ 68) share the same display column for the percent/dash
// placeholder. A regression where renderMissingRow used a 4-column layout (no
// trend cell) would shift the "—" leftward and this test would fail.
//
// Column math at width 100 (inner 96):
//
//	label(9) + gap(2) + bar(42) + gap(2) + percent(4) + gap(2) + trend(26) + gap(2) + reset(7)
//	                                        ↑ percent starts at display col 55
func TestWideTierColumnsAlign(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	m := NewModel(WithClock(func() time.Time { return now }))

	// Two real Claude windows with distinct percents so we get two data rows.
	// quotaRowSpecs[0] = Claude 5h (20%), quotaRowSpecs[1] = Claude 7d (60%).
	// Sonnet 7d and all Codex windows are absent → missing rows.
	m.windows[sources.ProductClaude] = []sources.Window{
		{
			Product:     sources.ProductClaude,
			Kind:        sources.WindowFiveHour,
			Label:       "Claude 5h",
			UsedPercent: 20,
			ResetsAt:    now.Add(2 * time.Hour),
			CapturedAt:  now,
		},
		{
			Product:     sources.ProductClaude,
			Kind:        sources.WindowSevenDay,
			Label:       "Claude 7d",
			UsedPercent: 60,
			ResetsAt:    now.Add(24 * time.Hour),
			CapturedAt:  now,
		},
	}
	// Set bar positions to the matching fractions so renderDataRow draws a bar.
	m.bars[0].target = 0.20
	m.bars[0].pos = 0.20
	m.bars[1].target = 0.60
	m.bars[1].pos = 0.60

	full := render(withWidth(m, 100))
	plain := ansiEscapeRE.ReplaceAllString(full, "")
	lines := strings.Split(plain, "\n")

	// percentRightEdge returns the display column immediately to the right of the
	// percent value sub in line. Using lipgloss.Width on both the prefix and sub
	// means multibyte block runes '█'/'░' in the progress bar and the trend arrow
	// '→' are measured by visual width, not byte length. Returns -1 when sub is
	// absent.
	//
	// We measure the RIGHT edge rather than the left edge because the percent cell
	// is right-aligned inside a 4-column field: "20%" pads to " 20%" (left edge
	// col 58) while "—" pads to "   —" (left edge col 60), yet both end at the
	// same right-edge column (61). Comparing right edges is the correct invariant.
	percentRightEdge := func(line, sub string) int {
		idx := strings.Index(line, sub)
		if idx < 0 {
			return -1
		}
		return lipgloss.Width(line[:idx]) + lipgloss.Width(sub)
	}

	// Locate the two data-row lines and one missing-row line.
	var line20, line60, lineMissing string
	for _, l := range lines {
		switch {
		case strings.Contains(l, "20%") && line20 == "":
			line20 = l
		case strings.Contains(l, "60%") && line60 == "":
			line60 = l
		case strings.Contains(l, "missing local data") && lineMissing == "":
			lineMissing = l
		}
	}
	if line20 == "" {
		t.Fatalf("could not find a line containing '20%%' in:\n%s", plain)
	}
	if line60 == "" {
		t.Fatalf("could not find a line containing '60%%' in:\n%s", plain)
	}
	if lineMissing == "" {
		t.Fatalf("could not find a line containing 'missing local data' in:\n%s", plain)
	}

	// Right edge of the percent cell in data rows ("20%" ends the 4-col field).
	edge20 := percentRightEdge(line20, "20%")
	edge60 := percentRightEdge(line60, "60%")
	// The "—" in a missing row is the percent placeholder (right-aligned in the
	// same 4-col field). There is a second "—" for the reset cell; strings.Index
	// finds the first, which is the percent placeholder.
	edgeMissing := percentRightEdge(lineMissing, "—")

	if edge20 < 0 {
		t.Fatalf("could not locate '20%%' in line: %q", line20)
	}
	if edge60 < 0 {
		t.Fatalf("could not locate '60%%' in line: %q", line60)
	}
	if edgeMissing < 0 {
		t.Fatalf("could not locate '—' in missing line: %q", lineMissing)
	}

	// All three right edges must be equal. If the missing row used a 4-column
	// layout (no trend cell) its percent "—" would end ~28 columns earlier.
	if edge20 != edge60 {
		t.Fatalf("20%% right edge col=%d, 60%% right edge col=%d — data rows are misaligned", edge20, edge60)
	}
	if edge20 != edgeMissing {
		t.Fatalf("data row percent right edge col=%d, missing row '—' right edge col=%d — percent cell misaligned between data and missing rows", edge20, edgeMissing)
	}

	// Sanity-check: the right edge should be ~61
	// (2 shell-pad + 9 label + 2 gap + 42 bar + 2 gap + 4 percent = col 61).
	// Not a hard assertion — wideTrendWidth tuning may shift it — but a large
	// deviation signals unintended breakage.
	const expectedPercentRightEdge = 61
	if edge20 < expectedPercentRightEdge-4 || edge20 > expectedPercentRightEdge+4 {
		t.Logf("WARN: percent right-edge col is %d, expected ~%d — check wideTrendWidth/barWidth math", edge20, expectedPercentRightEdge)
	}

	// Every wide-tier data/missing row should have display width exactly equal to
	// the full terminal width (100). The shell style adds 2-cell left/right
	// padding around the 96-column inner content, so each content line fills the
	// full 100-column width. Skip blank lines and the footer; identify data/missing
	// rows by the presence of "%" or "missing local data".
	const wantWidth = 100
	for _, l := range lines {
		stripped := strings.TrimSpace(l)
		isDataOrMissing := strings.Contains(l, "%") || strings.Contains(l, "missing local data")
		if !isDataOrMissing || stripped == "" {
			continue
		}
		w := lipgloss.Width(l)
		if w != wantWidth {
			t.Errorf("wide-tier row display width = %d, want %d: %q", w, wantWidth, l)
		}
	}
}

func TestSecondLineWidthSafeAcrossTiers(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	reset := now.Add(2 * time.Hour)
	m := modelWithHistory(now, sources.WindowFiveHour, []float64{20, 60}, 60, 60, reset)
	for _, width := range []int{80, 50, 49, 30, 29, 20} {
		m.width = width
		assertRenderedLineWidths(t, render(m), width)
	}
	// Very narrow: the second line is dropped entirely.
	m.width = 29
	plain := ansiEscapeRE.ReplaceAllString(render(m), "")
	if strings.Contains(plain, "%/hr") {
		t.Fatalf("width 29 should omit the second line, got:\n%s", plain)
	}
}

func TestIconModeKeepsLineWidths(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	for _, w := range []int{120, 100, 80, 72, 50, 49, 30, 29, 20} {
		m := NewModel(WithClock(func() time.Time { return now }), WithDisplayPrefs(DisplayPrefs{Icons: true}))
		m.windows[sources.ProductClaude] = []sources.Window{{Product: sources.ProductClaude, Kind: sources.WindowFiveHour, UsedPercent: 35, ResetsAt: now.Add(2 * time.Hour), CapturedAt: now}}
		m.width = w
		assertRenderedLineWidths(t, render(m), w)
	}
}

func TestAtRiskPercentPulsesWithoutTrendLine(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	mk := func(phase int) string {
		m := NewModel(WithClock(func() time.Time { return now }), WithDisplayPrefs(DisplayPrefs{HideTrend: true}))
		m.windows[sources.ProductClaude] = []sources.Window{{
			Product: sources.ProductClaude, Kind: sources.WindowFiveHour,
			UsedPercent: 100, ResetsAt: now.Add(30 * time.Minute), CapturedAt: now,
		}}
		m.width = 80
		m.animPhase = phase
		return render(m)
	}
	a := mk(pulsePeriod / 4)     // pulse peak
	b := mk(3 * pulsePeriod / 4) // pulse trough
	if a == b {
		t.Fatalf("at-risk percent should pulse (differ) across animation phases even with the trend line hidden")
	}
}
