package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/robbell5/llm-quota/internal/sources"
)

func TestTitleBandHasClock(t *testing.T) {
	at := time.Date(2026, 5, 27, 10, 37, 0, 0, time.Local)
	band := renderTitleBand(60, at, glyphsFor(false))
	if !strings.Contains(band, "LLM QUOTA") {
		t.Fatalf("title band missing title: %q", band)
	}
	if !strings.Contains(band, "10:37 AM") {
		t.Fatalf("title band missing clock: %q", band)
	}
}

func TestGroupHeadersShownWhenWideEnough(t *testing.T) {
	now := time.Date(2026, 5, 27, 12, 0, 0, 0, time.Local)
	m := NewModel(WithClock(func() time.Time { return now }))
	m.windows[sources.ProductClaude] = []sources.Window{{
		Product: sources.ProductClaude, Kind: sources.WindowFiveHour,
		UsedPercent: 20, ResetsAt: now.Add(2 * time.Hour), CapturedAt: now,
	}}
	m.width = 80
	full := render(m)
	if !strings.Contains(full, "CLAUDE") {
		t.Fatalf("expected CLAUDE group header at width 80:\n%s", full)
	}
}
