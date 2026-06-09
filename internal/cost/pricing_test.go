package cost

import (
	"math"
	"testing"
)

func approx(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestLoadPricingHasSeededModels(t *testing.T) {
	p, err := LoadPricing()
	if err != nil {
		t.Fatalf("LoadPricing: %v", err)
	}
	for _, model := range []string{"claude-opus-4-8", "claude-fable-5", "claude-haiku-4-5-20251001"} {
		if _, ok := p.models[model]; !ok {
			t.Fatalf("expected %s in table", model)
		}
	}
	if len(p.models) != 13 {
		t.Fatalf("expected 13 seeded models, got %d", len(p.models))
	}
}

func TestPriceClaudeFirm(t *testing.T) {
	p, _ := LoadPricing()
	// 1M input + 1M output for opus = $5 + $25 = $30.
	amount, known, estimated := p.price("claude-opus-4-8", Usage{Input: 1_000_000, Output: 1_000_000})
	if !known {
		t.Fatalf("expected known model")
	}
	if estimated {
		t.Fatalf("claude should not be estimated")
	}
	approx(t, amount, 30.0)
}

func TestPriceFable(t *testing.T) {
	p, _ := LoadPricing()
	// 1M input + 1M output for fable = $10 + $50 = $60.
	amount, known, estimated := p.price("claude-fable-5", Usage{Input: 1_000_000, Output: 1_000_000})
	if !known {
		t.Fatalf("expected known model")
	}
	if estimated {
		t.Fatalf("claude should not be estimated")
	}
	approx(t, amount, 60.0)
	// A 1M-context-suffixed fable id must price like the base id.
	suffixed, known, _ := p.price("claude-fable-5[1m]", Usage{Input: 1_000_000, Output: 1_000_000})
	if !known {
		t.Fatalf("expected suffixed fable id to resolve to base rates")
	}
	approx(t, suffixed, amount)
}

func TestPriceDatedHaikuMatchesAlias(t *testing.T) {
	p, _ := LoadPricing()
	// The dated full id is the same model as the claude-haiku-4-5 alias.
	dated, known, _ := p.price("claude-haiku-4-5-20251001", Usage{Input: 1_000_000, Output: 1_000_000})
	if !known {
		t.Fatalf("expected dated haiku id to be priced")
	}
	alias, _, _ := p.price("claude-haiku-4-5", Usage{Input: 1_000_000, Output: 1_000_000})
	approx(t, dated, alias)
	approx(t, dated, 6.0) // $1 + $5
}

func TestPriceCacheTokens(t *testing.T) {
	p, _ := LoadPricing()
	// 1M cache-write-5m (6.25) + 1M cache-read (0.5) for opus = $6.75.
	amount, _, _ := p.price("claude-opus-4-8", Usage{CacheWrite5m: 1_000_000, CacheRead: 1_000_000})
	approx(t, amount, 6.75)
}

func TestPriceCodexEstimated(t *testing.T) {
	p, _ := LoadPricing()
	amount, known, estimated := p.price("gpt-5-codex", Usage{Output: 1_000_000})
	if !known || !estimated {
		t.Fatalf("expected known+estimated for gpt-5-codex (known=%v est=%v)", known, estimated)
	}
	approx(t, amount, 10.0)
}

func TestPriceCacheWrite1h(t *testing.T) {
	p, _ := LoadPricing()
	// 1M cache-write-1h (10.0) for opus = $10.
	amount, _, _ := p.price("claude-opus-4-8", Usage{CacheWrite1h: 1_000_000})
	approx(t, amount, 10.0)
}

func TestPriceUnknownModel(t *testing.T) {
	p, _ := LoadPricing()
	_, known, _ := p.price("claude-opus-9-9", Usage{Input: 1_000_000})
	if known {
		t.Fatalf("expected unknown model to report known=false")
	}
}

func TestPriceModelWithContextSuffix(t *testing.T) {
	p, _ := LoadPricing()
	// A 1M-context id must price identically to its base model.
	suffixed, known, _ := p.price("claude-opus-4-8[1m]", Usage{Input: 1_000_000, Output: 1_000_000})
	if !known {
		t.Fatalf("expected suffixed model to resolve to base rates")
	}
	base, _, _ := p.price("claude-opus-4-8", Usage{Input: 1_000_000, Output: 1_000_000})
	if suffixed != base {
		t.Fatalf("suffixed price %v != base price %v", suffixed, base)
	}
	// A genuinely unknown base is still unknown even with a suffix.
	if _, known2, _ := p.price("totally-unknown[1m]", Usage{Input: 1}); known2 {
		t.Fatalf("unknown base with suffix should remain unknown")
	}
}
