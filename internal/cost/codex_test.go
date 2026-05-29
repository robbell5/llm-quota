package cost

import (
	"path/filepath"
	"testing"

	"github.com/robbell5/llm-quota/internal/sources"
)

func TestCodexCostReaderUsesLastTokenUsageAndModel(t *testing.T) {
	root := t.TempDir()
	// turn_context sets the model; each token_count contributes last_token_usage.
	// reasoning_output_tokens fold into output; cached_input subtracts from input.
	body := `{"type":"turn_context","timestamp":"2026-05-28T11:00:00Z","payload":{"model":"gpt-5-codex"}}
{"type":"event_msg","timestamp":"2026-05-28T11:30:00Z","payload":{"type":"token_count","info":{"last_token_usage":{"input_tokens":1000000,"cached_input_tokens":200000,"output_tokens":500000,"reasoning_output_tokens":500000}}}}
`
	writeFile(t, filepath.Join(root, "2026", "05", "28", "rollout-x.jsonl"), body)

	pricing, _ := LoadPricing()
	now := ts("2026-05-28T13:00:00Z")
	windows := []sources.Window{{Product: sources.ProductCodex, Kind: sources.WindowFiveHour, ResetsAt: ts("2026-05-28T16:00:00Z")}}
	got := NewCodexCostReader(root, pricing).WindowCosts(now, windows)

	wc := got[sources.WindowFiveHour]
	// input billed = (1,000,000 - 200,000)=800k @ $1.25 = $1.00
	// cache_read   = 200k @ $0.125               = $0.025
	// output       = (500k + 500k reasoning)=1M @ $10 = $10.00
	approx(t, wc.Amount, 11.025)
	if !wc.Estimated {
		t.Fatalf("expected Codex value flagged Estimated")
	}
}

func TestCodexCostReaderTokenCountBeforeModelIsIncomplete(t *testing.T) {
	root := t.TempDir()
	body := `{"type":"event_msg","timestamp":"2026-05-28T11:30:00Z","payload":{"type":"token_count","info":{"last_token_usage":{"output_tokens":1000000}}}}
`
	writeFile(t, filepath.Join(root, "2026", "05", "28", "rollout-y.jsonl"), body)
	pricing, _ := LoadPricing()
	now := ts("2026-05-28T13:00:00Z")
	windows := []sources.Window{{Product: sources.ProductCodex, Kind: sources.WindowFiveHour, ResetsAt: ts("2026-05-28T16:00:00Z")}}
	got := NewCodexCostReader(root, pricing).WindowCosts(now, windows)
	wc := got[sources.WindowFiveHour]
	if wc.Amount != 0 || !wc.Incomplete {
		t.Fatalf("expected Amount=0 and Incomplete=true for unknown-model turn, got %+v", wc)
	}
}

func TestCodexCostReaderSumsMultipleTokenCounts(t *testing.T) {
	root := t.TempDir()
	body := `{"type":"turn_context","timestamp":"2026-05-28T11:00:00Z","payload":{"model":"gpt-5-codex"}}
{"type":"event_msg","timestamp":"2026-05-28T11:15:00Z","payload":{"type":"token_count","info":{"last_token_usage":{"output_tokens":1000000}}}}
{"type":"event_msg","timestamp":"2026-05-28T11:30:00Z","payload":{"type":"token_count","info":{"last_token_usage":{"output_tokens":1000000}}}}
`
	writeFile(t, filepath.Join(root, "2026", "05", "28", "rollout-z.jsonl"), body)
	pricing, _ := LoadPricing()
	now := ts("2026-05-28T13:00:00Z")
	windows := []sources.Window{{Product: sources.ProductCodex, Kind: sources.WindowFiveHour, ResetsAt: ts("2026-05-28T16:00:00Z")}}
	got := NewCodexCostReader(root, pricing).WindowCosts(now, windows)
	approx(t, got[sources.WindowFiveHour].Amount, 20.0)
}
