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
	if _, ok := p.models["claude-opus-4-8"]; !ok {
		t.Fatalf("expected claude-opus-4-8 in table")
	}
	if len(p.models) != 10 {
		t.Fatalf("expected 10 seeded models, got %d", len(p.models))
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
