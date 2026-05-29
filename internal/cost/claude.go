package cost

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robbell5/llm-quota/internal/sources"
	"github.com/robbell5/llm-quota/internal/trend"
)

// ClaudeCostReader prices Claude Code transcript usage per quota window.
type ClaudeCostReader struct {
	projectsRoot string
	pricing      Pricing
	cache        *parseCache
}

func NewClaudeCostReader(projectsRoot string, pricing Pricing) *ClaudeCostReader {
	return &ClaudeCostReader{projectsRoot: projectsRoot, pricing: pricing, cache: newParseCache()}
}

// WindowCosts returns the equivalent API value for each given window. Missing
// or unreadable data yields a zero-value WindowCost (the caller hides it).
func (r *ClaudeCostReader) WindowCosts(now time.Time, windows []sources.Window) map[sources.WindowKind]WindowCost {
	if len(windows) == 0 {
		return nil
	}
	earliest := earliestStart(windows, now)
	entries := collectEntries(r.projectsRoot, "", earliest, r.cache, parseClaudeFile)
	entries = dedup(entries)

	out := make(map[sources.WindowKind]WindowCost, len(windows))
	for _, w := range windows {
		start := w.ResetsAt.Add(-trend.WindowDuration(w.Kind))
		out[w.Kind] = aggregate(entries, start, now, r.pricing)
	}
	return out
}

// earliestStart returns the widest window's start (resets_at − duration), which
// bounds how far back the file scan and entry collection must reach.
func earliestStart(windows []sources.Window, now time.Time) time.Time {
	earliest := now
	for _, w := range windows {
		start := w.ResetsAt.Add(-trend.WindowDuration(w.Kind))
		if start.Before(earliest) {
			earliest = start
		}
	}
	return earliest
}

// collectEntries walks root for *.jsonl files, skips those last modified before
// `since` (no in-window entries possible), and parses the rest through the
// cache. namePrefix, when non-empty, additionally filters by filename prefix
// (used by the Codex reader for "rollout-").
func collectEntries(root, namePrefix string, since time.Time, cache *parseCache, parse func(string) ([]entry, error)) []entry {
	var all []entry
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		name := d.Name()
		if filepath.Ext(name) != ".jsonl" {
			return nil
		}
		if namePrefix != "" && !strings.HasPrefix(name, namePrefix) {
			return nil
		}
		info, statErr := d.Info()
		if statErr != nil {
			return nil
		}
		if info.ModTime().Before(since) {
			return nil // last write predates the widest window
		}
		entries, parseErr := cache.load(path, info, func() ([]entry, error) { return parse(path) })
		if parseErr != nil {
			return nil
		}
		all = append(all, entries...)
		return nil
	})
	return all
}

type claudeRecord struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	RequestID string `json:"requestId"`
	Message   *struct {
		ID    string      `json:"id"`
		Model string      `json:"model"`
		Usage claudeUsage `json:"usage"`
	} `json:"message"`
}

type claudeUsage struct {
	InputTokens              int64 `json:"input_tokens"`
	OutputTokens             int64 `json:"output_tokens"`
	CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
	CacheCreation            *struct {
		Ephemeral5m int64 `json:"ephemeral_5m_input_tokens"`
		Ephemeral1h int64 `json:"ephemeral_1h_input_tokens"`
	} `json:"cache_creation"`
}

func parseClaudeFile(path string) ([]entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []entry
	for _, line := range bytes.Split(data, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var rec claudeRecord
		if json.Unmarshal(line, &rec) != nil || rec.Type != "assistant" || rec.Message == nil {
			continue
		}
		when, err := time.Parse(time.RFC3339, rec.Timestamp)
		if err != nil {
			continue
		}
		id := rec.RequestID
		if id == "" {
			id = rec.Message.ID
		}
		u := rec.Message.Usage
		write5m := u.CacheCreationInputTokens // fallback: treat all cache-write as 5m
		write1h := int64(0)
		if u.CacheCreation != nil {
			write5m = u.CacheCreation.Ephemeral5m
			write1h = u.CacheCreation.Ephemeral1h
		}
		out = append(out, entry{
			ts:    when,
			model: rec.Message.Model,
			id:    id,
			usage: Usage{
				Input:        u.InputTokens,
				Output:       u.OutputTokens,
				CacheRead:    u.CacheReadInputTokens,
				CacheWrite5m: write5m,
				CacheWrite1h: write1h,
			},
		})
	}
	return out, nil
}
