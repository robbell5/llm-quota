package cost

import "time"

// entry is one priced-able turn: when it happened, which model, and its tokens.
// id is the dedup key (Claude requestId/message.id); empty for Codex.
type entry struct {
	ts    time.Time
	model string
	id    string
	usage Usage
}

// WindowCost is the rendered value for one window.
type WindowCost struct {
	Amount     float64 // USD, equivalent API value
	Estimated  bool    // any priced entry was an estimate (Codex) → render "~"
	Incomplete bool    // some in-window tokens were from an unpriced model → "*"
}

// aggregate sums the cost of entries whose timestamp falls in [windowStart, now].
func aggregate(entries []entry, windowStart, now time.Time, p Pricing) WindowCost {
	var wc WindowCost
	for _, e := range entries {
		if e.ts.Before(windowStart) || e.ts.After(now) {
			continue
		}
		if e.usage.isZero() {
			continue // zero-usage entries (e.g. <synthetic> error records) can't affect cost
		}
		amount, known, estimated := p.price(e.model, e.usage)
		if !known {
			wc.Incomplete = true
			continue
		}
		wc.Amount += amount
		if estimated {
			wc.Estimated = true
		}
	}
	return wc
}

// dedup drops entries with a repeated non-empty id (resumed sessions repeat
// assistant lines), keeping the first. Entries with an empty id are all kept.
func dedup(entries []entry) []entry {
	seen := make(map[string]struct{}, len(entries))
	out := make([]entry, 0, len(entries))
	for _, e := range entries {
		if e.id != "" {
			if _, ok := seen[e.id]; ok {
				continue
			}
			seen[e.id] = struct{}{}
		}
		out = append(out, e)
	}
	return out
}
