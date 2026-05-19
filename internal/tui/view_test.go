package tui

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/rob/llm-quota/internal/sources"
)

var ansiEscapeRE = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

func TestRenderStartupScreen(t *testing.T) {
	full := render(Model{width: 80, height: 12})

	if !strings.Contains(full, "LLM Quota") {
		t.Fatalf("expected title in startup screen, got:\n%s", full)
	}

	for _, label := range []string{"Claude 5h", "Claude 7d", "Codex 5h", "Codex 7d"} {
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

	for _, width := range []int{50, 49, 29} {
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
	for _, want := range []string{"Claude 5h", "42%", "Codex 7d", "17%"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("expected rendered source-backed rows to contain %q, got:\n%s", want, plain)
		}
	}
	if strings.Contains(plain, "Claude 5h  —  missing local data") {
		t.Fatalf("stale-but-valid Claude data should render as data, not placeholder:\n%s", plain)
	}

	if !strings.Contains(plain, "Claude data 2h old; open Claude") {
		t.Fatalf("expected stale Claude footer hint in source-backed output, got:\n%s", plain)
	}

	for _, forbidden := range []string{"malformed", "read_error", "no_usable_event"} {
		if strings.Contains(strings.ToLower(plain), forbidden) {
			t.Fatalf("render output should not expose raw error category %q:\n%s", forbidden, plain)
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
	for _, want := range []string{"Claude 5h", "42%", "Claude data 2h old; open Claude"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("expected stale-but-valid Claude output to contain %q, got:\n%s", want, plain)
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
