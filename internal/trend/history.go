package trend

import (
	"time"

	"github.com/robbell5/llm-quota/internal/sources"
)

const maxSamplesPerWindow = 64

// Sample is one observed reading of a quota window.
type Sample struct {
	CapturedAt time.Time
	UsedPct    float64
	ResetsAt   time.Time
}

// History holds bounded per-window change-points, chronological (oldest first).
type History struct {
	windows map[string][]Sample
}

func NewHistory() *History {
	return &History{windows: make(map[string][]Sample)}
}

// Key identifies a window's history slot, e.g. "claude:five_hour".
func Key(product sources.Product, kind sources.WindowKind) string {
	return string(product) + ":" + string(kind)
}

// Append records a change-point. It is a no-op when used_pct and resets_at are
// unchanged from the last point. Points from prior epochs (a different
// resets_at) are dropped, and the slice is bounded to maxSamplesPerWindow.
func (h *History) Append(key string, s Sample) {
	cur := h.windows[key]
	if n := len(cur); n > 0 {
		last := cur[n-1]
		if last.UsedPct == s.UsedPct && last.ResetsAt.Equal(s.ResetsAt) {
			return
		}
	}
	cur = append(cur, s)

	filtered := cur[:0]
	for _, p := range cur {
		if p.ResetsAt.Equal(s.ResetsAt) {
			filtered = append(filtered, p)
		}
	}
	cur = filtered

	if len(cur) > maxSamplesPerWindow {
		cur = cur[len(cur)-maxSamplesPerWindow:]
	}
	h.windows[key] = cur
}

// EpochSamples returns the points for key whose resets_at matches resetsAt.
func (h *History) EpochSamples(key string, resetsAt time.Time) []Sample {
	var out []Sample
	for _, p := range h.windows[key] {
		if p.ResetsAt.Equal(resetsAt) {
			out = append(out, p)
		}
	}
	return out
}
