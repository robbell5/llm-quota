package cost

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
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

// Pricing is an exact-match model lookup loaded from the embedded table.
type Pricing struct {
	models map[string]modelPricing
}

type modelPricing struct {
	periods []pricingPeriod
}

type pricingPeriod struct {
	effectiveFrom *time.Time
	rates         Rates
}

type pricingFile struct {
	Models map[string]json.RawMessage `json:"models"`
}

type pricingEntryFile struct {
	Rates
	Periods *[]pricingPeriodFile `json:"rates"`
}

type pricingPeriodFile struct {
	EffectiveFrom string `json:"effective_from"`
	Rates
}

// LoadPricing parses the embedded pricing table.
func LoadPricing() (Pricing, error) {
	var f pricingFile
	if err := json.Unmarshal(pricingTableJSON, &f); err != nil {
		return Pricing{}, fmt.Errorf("parse embedded pricing table: %w", err)
	}
	if f.Models == nil {
		f.Models = map[string]json.RawMessage{}
	}
	models := make(map[string]modelPricing, len(f.Models))
	for model, raw := range f.Models {
		pricing, err := parseModelPricing(model, raw)
		if err != nil {
			return Pricing{}, fmt.Errorf("parse embedded pricing table model %q: %w", model, err)
		}
		models[model] = pricing
	}
	return Pricing{models: models}, nil
}

func parseModelPricing(model string, raw json.RawMessage) (modelPricing, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return modelPricing{}, err
	}
	ratesRaw, hasRates := fields["rates"]

	var f pricingEntryFile
	if err := json.Unmarshal(raw, &f); err != nil {
		return modelPricing{}, err
	}
	if !hasRates {
		return modelPricing{periods: []pricingPeriod{{rates: f.Rates}}}, nil
	}
	if strings.TrimSpace(string(ratesRaw)) == "null" {
		return modelPricing{}, fmt.Errorf("%s rates: expected rates to be an array", model)
	}
	if f.Periods == nil {
		return modelPricing{}, fmt.Errorf("%s rates: expected rates to be an array", model)
	}
	if len(*f.Periods) == 0 {
		return modelPricing{}, fmt.Errorf("%s rates: expected at least one period", model)
	}
	periods := make([]pricingPeriod, 0, len(*f.Periods))
	for i, periodFile := range *f.Periods {
		period := pricingPeriod{rates: periodFile.Rates}
		if periodFile.EffectiveFrom != "" {
			effectiveFrom, err := time.Parse(time.RFC3339, periodFile.EffectiveFrom)
			if err != nil {
				return modelPricing{}, fmt.Errorf("%s rates[%d].effective_from: %w", model, i, err)
			}
			period.effectiveFrom = &effectiveFrom
		}
		periods = append(periods, period)
	}
	sort.SliceStable(periods, func(i, j int) bool {
		return pricingPeriodStart(periods[i]).Before(pricingPeriodStart(periods[j]))
	})
	return modelPricing{periods: periods}, nil
}

func pricingPeriodStart(period pricingPeriod) time.Time {
	if period.effectiveFrom == nil {
		return time.Time{}
	}
	return *period.effectiveFrom
}

// lookup finds rates for model at a specific time, falling back to the id with a
// trailing context suffix stripped (e.g. "claude-opus-4-8[1m]" →
// "claude-opus-4-8") so 1M-context model ids price the same as their base model.
func (p Pricing) lookup(model string, at time.Time) (Rates, bool) {
	if r, ok := p.lookupExact(model, at); ok {
		return r, true
	}
	if base, _, found := strings.Cut(model, "["); found && base != "" {
		return p.lookupExact(base, at)
	}
	return Rates{}, false
}

func (p Pricing) lookupExact(model string, at time.Time) (Rates, bool) {
	pricing, ok := p.models[model]
	if !ok {
		return Rates{}, false
	}
	return pricing.lookup(at)
}

func (p modelPricing) lookup(at time.Time) (Rates, bool) {
	var selected *pricingPeriod
	for i := range p.periods {
		if pricingPeriodStart(p.periods[i]).After(at) {
			break
		}
		selected = &p.periods[i]
	}
	if selected == nil {
		return Rates{}, false
	}
	return selected.rates, true
}

// price returns the USD cost of u under model's rates at at. known is false for
// an unpriced model; estimated reflects the model's Estimated flag.
func (p Pricing) price(model string, u Usage, at time.Time) (amount float64, known bool, estimated bool) {
	r, ok := p.lookup(model, at)
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
