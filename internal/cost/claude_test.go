package cost

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/robbell5/llm-quota/internal/sources"
)

func writeFile(t *testing.T, path string, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestClaudeCostReaderWindowCosts(t *testing.T) {
	root := t.TempDir()
	// One in-window opus turn: 1M output = $25. A duplicate requestId must not
	// double count. An older line (pre-window) must be excluded.
	body := `{"type":"assistant","timestamp":"2026-05-28T11:00:00Z","requestId":"r1","message":{"id":"m1","model":"claude-opus-4-8","usage":{"output_tokens":1000000}}}
{"type":"assistant","timestamp":"2026-05-28T11:00:05Z","requestId":"r1","message":{"id":"m1","model":"claude-opus-4-8","usage":{"output_tokens":1000000}}}
{"type":"assistant","timestamp":"2026-05-20T11:00:00Z","requestId":"r0","message":{"id":"m0","model":"claude-opus-4-8","usage":{"output_tokens":1000000}}}
`
	writeFile(t, filepath.Join(root, "proj", "session.jsonl"), body)

	pricing, _ := LoadPricing()
	now := ts("2026-05-28T13:00:00Z")
	windows := []sources.Window{
		{Product: sources.ProductClaude, Kind: sources.WindowFiveHour, ResetsAt: ts("2026-05-28T16:00:00Z")},
		{Product: sources.ProductClaude, Kind: sources.WindowSevenDay, ResetsAt: ts("2026-05-29T00:00:00Z")},
	}
	got := NewClaudeCostReader(root, pricing).WindowCosts(now, windows)

	// 5h window starts 11:00 → includes the deduped r1 turn only ($25).
	approx(t, got[sources.WindowFiveHour].Amount, 25.0)
	// 7d window starts 2026-05-22 → still excludes the 05-20 line, also $25.
	approx(t, got[sources.WindowSevenDay].Amount, 25.0)
}

func TestClaudeCostReaderCacheTokens(t *testing.T) {
	root := t.TempDir()
	body := `{"type":"assistant","timestamp":"2026-05-28T12:00:00Z","requestId":"r1","message":{"id":"m1","model":"claude-opus-4-8","usage":{"cache_read_input_tokens":1000000,"cache_creation_input_tokens":1000000,"cache_creation":{"ephemeral_5m_input_tokens":1000000,"ephemeral_1h_input_tokens":0}}}}
`
	writeFile(t, filepath.Join(root, "p", "s.jsonl"), body)
	pricing, _ := LoadPricing()
	now := ts("2026-05-28T13:00:00Z")
	windows := []sources.Window{{Kind: sources.WindowFiveHour, ResetsAt: ts("2026-05-28T16:00:00Z")}}
	got := NewClaudeCostReader(root, pricing).WindowCosts(now, windows)
	// cache_write_5m 1M ($6.25) + cache_read 1M ($0.5) = $6.75.
	approx(t, got[sources.WindowFiveHour].Amount, 6.75)
}

func TestClaudeCostReaderCacheCreationFallback(t *testing.T) {
	root := t.TempDir()
	// No nested cache_creation object: the flat cache_creation_input_tokens must
	// be priced as 5m cache-write. 1M × $6.25/M = $6.25, no read component.
	body := `{"type":"assistant","timestamp":"2026-05-28T12:00:00Z","requestId":"r2","message":{"id":"m2","model":"claude-opus-4-8","usage":{"cache_creation_input_tokens":1000000}}}
`
	writeFile(t, filepath.Join(root, "p", "s.jsonl"), body)
	pricing, _ := LoadPricing()
	now := ts("2026-05-28T13:00:00Z")
	windows := []sources.Window{{Kind: sources.WindowFiveHour, ResetsAt: ts("2026-05-28T16:00:00Z")}}
	got := NewClaudeCostReader(root, pricing).WindowCosts(now, windows)
	approx(t, got[sources.WindowFiveHour].Amount, 6.25)
}
