package cost

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed pricing_table.json
var pricingTableJSON []byte

// Rates are USD per million tokens for one model.
type Rates struct {
	Input        float64 `json:"input"`
	Output       float64 `json:"output"`
	CacheWrite5m float64 `json:"cache_write_5m"`
	CacheWrite1h float64 `json:"cache_write_1h"`
	CacheRead    float64 `json:"cache_read"`
	Estimated    bool    `json:"estimated"`
}

// Usage is a token tally already split into disjoint pricing classes. Input is
// the non-cached input tokens (Codex's cached subset is moved to CacheRead at
// parse time so Input and CacheRead never overlap, matching Claude's layout).
type Usage struct {
	Input        int64
	Output       int64
	CacheWrite5m int64
	CacheWrite1h int64
	CacheRead    int64
}

// isZero reports whether no tokens were recorded in any pricing class.
func (u Usage) isZero() bool {
	return u.Input == 0 && u.Output == 0 && u.CacheWrite5m == 0 && u.CacheWrite1h == 0 && u.CacheRead == 0
}

// Pricing is an exact-match model → Rates lookup loaded from the embedded table.
type Pricing struct {
	models map[string]Rates
}

type pricingFile struct {
	Models map[string]Rates `json:"models"`
}

// LoadPricing parses the embedded pricing table.
func LoadPricing() (Pricing, error) {
	var f pricingFile
	if err := json.Unmarshal(pricingTableJSON, &f); err != nil {
		return Pricing{}, fmt.Errorf("parse embedded pricing table: %w", err)
	}
	if f.Models == nil {
		f.Models = map[string]Rates{}
	}
	return Pricing{models: f.Models}, nil
}

// lookup finds rates for model, falling back to the id with a trailing context
// suffix stripped (e.g. "claude-opus-4-8[1m]" → "claude-opus-4-8") so 1M-context
// model ids price the same as their base model.
func (p Pricing) lookup(model string) (Rates, bool) {
	if r, ok := p.models[model]; ok {
		return r, true
	}
	if base, _, found := strings.Cut(model, "["); found && base != "" {
		if r, ok := p.models[base]; ok {
			return r, true
		}
	}
	return Rates{}, false
}

// price returns the USD cost of u under model's rates. known is false for an
// unpriced model; estimated reflects the model's Estimated flag.
func (p Pricing) price(model string, u Usage) (amount float64, known bool, estimated bool) {
	r, ok := p.lookup(model)
	if !ok {
		return 0, false, false
	}
	const perMillion = 1_000_000.0
	amount = float64(u.Input)/perMillion*r.Input +
		float64(u.Output)/perMillion*r.Output +
		float64(u.CacheWrite5m)/perMillion*r.CacheWrite5m +
		float64(u.CacheWrite1h)/perMillion*r.CacheWrite1h +
		float64(u.CacheRead)/perMillion*r.CacheRead
	return amount, true, r.Estimated
}
