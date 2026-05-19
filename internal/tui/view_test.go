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

	const fullFooter = "q / Ctrl-C quit · Claude: run install-claude-hook · Codex: open Codex"
	if !strings.Contains(full, fullFooter) {
		t.Fatalf("expected full footer %q in:\n%s", fullFooter, full)
	}

	if strings.Contains(full, "r refresh") || strings.Contains(full, " refresh") || strings.Contains(full, " · r") {
		t.Fatalf("startup screen should not advertise an r refresh key:\n%s", full)
	}

	compact := render(Model{width: 49, height: 12})
	const compactFooter = "q / Ctrl-C quit · data pending"
	if !strings.Contains(compact, compactFooter) {
		t.Fatalf("expected compact footer %q in:\n%s", compactFooter, compact)
	}

	atFifty := render(Model{width: 50, height: 12})
	if !strings.Contains(atFifty, compactFooter) {
		t.Fatalf("expected compact footer at width 50 in:\n%s", atFifty)
	}

	if strings.Contains(compact, "Claude: run install-claude-hook") || strings.Contains(compact, "Codex: open Codex") {
		t.Fatalf("compact footer should avoid source hints that wrap:\n%s", compact)
	}

	for _, width := range []int{50, 49, 29} {
		assertRenderedLineWidths(t, render(Model{width: width, height: 12}), width)
	}
}

func TestRenderSourceBackedRowsWithoutPhaseFourCopy(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	model := NewModel()
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

	for _, forbidden := range []string{"refreshing", "last updated", "r refresh", "stale"} {
		if strings.Contains(strings.ToLower(plain), forbidden) {
			t.Fatalf("render output should not contain Phase 4 copy %q:\n%s", forbidden, plain)
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

func assertRenderedLineWidths(t *testing.T, rendered string, maxWidth int) {
	t.Helper()

	for lineNumber, line := range strings.Split(strings.TrimSuffix(rendered, "\n"), "\n") {
		plain := ansiEscapeRE.ReplaceAllString(line, "")
		if width := lipgloss.Width(plain); width > maxWidth {
			t.Fatalf("line %d width = %d, want <= %d: %q", lineNumber+1, width, maxWidth, plain)
		}
	}
}
