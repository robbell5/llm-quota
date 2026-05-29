package cost

import (
	"bytes"
	"encoding/json"
	"os"
	"time"

	"github.com/robbell5/llm-quota/internal/sources"
	"github.com/robbell5/llm-quota/internal/trend"
)

// CodexCostReader prices Codex rollout usage per quota window (an estimate:
// ChatGPT-plan token pricing is unofficial, so all entries map to API rates).
type CodexCostReader struct {
	sessionsRoot string
	pricing      Pricing
	cache        *parseCache
}

func NewCodexCostReader(sessionsRoot string, pricing Pricing) *CodexCostReader {
	return &CodexCostReader{sessionsRoot: sessionsRoot, pricing: pricing, cache: newParseCache()}
}

func (r *CodexCostReader) WindowCosts(now time.Time, windows []sources.Window) map[sources.WindowKind]WindowCost {
	if len(windows) == 0 {
		return nil
	}
	earliest := earliestStart(windows, now)
	entries := collectEntries(r.sessionsRoot, "rollout-", earliest, r.cache, parseCodexFile)
	// No cross-file dedup: each rollout is a distinct session, each token_count a
	// distinct turn.

	out := make(map[sources.WindowKind]WindowCost, len(windows))
	for _, w := range windows {
		start := w.ResetsAt.Add(-trend.WindowDuration(w.Kind))
		out[w.Kind] = aggregate(entries, start, now, r.pricing)
	}
	return out
}

type codexRecord struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Payload   struct {
		Type  string `json:"type"`
		Model string `json:"model"`
		Info  *struct {
			LastTokenUsage codexTokens `json:"last_token_usage"`
		} `json:"info"`
	} `json:"payload"`
}

type codexTokens struct {
	InputTokens           int64 `json:"input_tokens"`
	OutputTokens          int64 `json:"output_tokens"`
	CachedInputTokens     int64 `json:"cached_input_tokens"`
	ReasoningOutputTokens int64 `json:"reasoning_output_tokens"`
}

func parseCodexFile(path string) ([]entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []entry
	currentModel := ""
	for _, line := range bytes.Split(data, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var rec codexRecord
		if json.Unmarshal(line, &rec) != nil {
			continue
		}
		// turn_context is metadata: capture the model and skip (not a billable
		// event). An empty model leaves currentModel unchanged — a token_count
		// with no preceding model is priced as unknown (Incomplete), the safe path.
		if rec.Type == "turn_context" && rec.Payload.Model != "" {
			currentModel = rec.Payload.Model
			continue
		}
		if rec.Payload.Type != "token_count" || rec.Payload.Info == nil {
			continue
		}
		when, err := time.Parse(time.RFC3339, rec.Timestamp)
		if err != nil {
			continue
		}
		lu := rec.Payload.Info.LastTokenUsage
		uncachedInput := lu.InputTokens - lu.CachedInputTokens
		if uncachedInput < 0 {
			uncachedInput = 0
		}
		out = append(out, entry{
			ts:    when,
			model: currentModel,
			usage: Usage{
				Input:     uncachedInput,
				Output:    lu.OutputTokens + lu.ReasoningOutputTokens,
				CacheRead: lu.CachedInputTokens,
			},
		})
	}
	return out, nil
}
