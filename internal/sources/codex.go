package sources

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type CodexReader struct {
	sessionsRoot string
}

func NewCodexReader(sessionsRoot string) CodexReader {
	return CodexReader{sessionsRoot: sessionsRoot}
}

func (r CodexReader) Fetch(now time.Time) ([]Window, error) {
	if _, err := os.Stat(r.sessionsRoot); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, SourceError{Source: ProductCodex, Category: ErrorMissing, Err: err}
		}

		return nil, SourceError{Source: ProductCodex, Category: ErrorRead, Err: err}
	}

	candidates, err := r.rolloutCandidates()
	if err != nil {
		return nil, err
	}

	for _, candidate := range candidates {
		windows, ok, err := windowsFromCodexRollout(candidate.path, now)
		if err != nil {
			return nil, err
		}
		if ok {
			return windows, nil
		}
	}

	return nil, SourceError{Source: ProductCodex, Category: ErrorNoUsableEvent, Err: errors.New("no usable Codex rate-limit event")}
}

type codexRolloutCandidate struct {
	path    string
	modTime time.Time
}

func (r CodexReader) rolloutCandidates() ([]codexRolloutCandidate, error) {
	var candidates []codexRolloutCandidate
	err := filepath.WalkDir(r.sessionsRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}

		name := entry.Name()
		if !strings.HasPrefix(name, "rollout-") || !strings.HasSuffix(name, ".jsonl") {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		candidates = append(candidates, codexRolloutCandidate{
			path:    path,
			modTime: info.ModTime(),
		})
		return nil
	})
	if err != nil {
		return nil, SourceError{Source: ProductCodex, Category: ErrorRead, Err: err}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].modTime.After(candidates[j].modTime)
	})

	return candidates, nil
}

func windowsFromCodexRollout(path string, now time.Time) ([]Window, bool, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, false, SourceError{Source: ProductCodex, Category: ErrorRead, Err: err}
	}

	var selected *codexRateLimits
	for _, line := range strings.Split(string(contents), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		limits, ok := parseCodexRateLimitLine(line)
		if ok {
			selected = &limits
		}
	}

	if selected == nil {
		return nil, false, nil
	}

	windows, err := selected.windows(now)
	if err != nil {
		return nil, false, nil
	}

	return windows, true, nil
}

func parseCodexRateLimitLine(line string) (codexRateLimits, bool) {
	var event codexEvent
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		return codexRateLimits{}, false
	}

	if event.Type != "event_msg" || event.Payload.Type != "token_count" || event.Payload.RateLimits == nil {
		return codexRateLimits{}, false
	}
	if err := event.Payload.RateLimits.validate(); err != nil {
		return codexRateLimits{}, false
	}

	return *event.Payload.RateLimits, true
}

type codexEvent struct {
	Type    string       `json:"type"`
	Payload codexPayload `json:"payload"`
}

type codexPayload struct {
	Type       string           `json:"type"`
	RateLimits *codexRateLimits `json:"rate_limits"`
}

type codexRateLimits struct {
	Primary   codexRateLimitWindow `json:"primary"`
	Secondary codexRateLimitWindow `json:"secondary"`
	PlanType  string               `json:"plan_type"`
}

func (l codexRateLimits) windows(now time.Time) ([]Window, error) {
	if err := l.validate(); err != nil {
		return nil, err
	}

	metadata := Metadata(nil)
	if l.PlanType != "" {
		metadata = Metadata{"plan_type": l.PlanType}
	}

	return []Window{
		l.Primary.window(WindowFiveHour, "Codex 5h", now, metadata),
		l.Secondary.window(WindowSevenDay, "Codex 7d", now, metadata),
	}, nil
}

func (l codexRateLimits) validate() error {
	if err := l.Primary.validate("primary", 300); err != nil {
		return err
	}
	if err := l.Secondary.validate("secondary", 10080); err != nil {
		return err
	}

	return nil
}

type codexRateLimitWindow struct {
	UsedPercent   *float64 `json:"used_percent"`
	WindowMinutes *int     `json:"window_minutes"`
	ResetsAt      *int64   `json:"resets_at"`
}

func (w codexRateLimitWindow) validate(name string, wantMinutes int) error {
	if w.UsedPercent == nil {
		return fmt.Errorf("missing %s used_percent", name)
	}
	if w.WindowMinutes == nil {
		return fmt.Errorf("missing %s window_minutes", name)
	}
	if *w.WindowMinutes != wantMinutes {
		return fmt.Errorf("unexpected %s window_minutes", name)
	}
	if w.ResetsAt == nil {
		return fmt.Errorf("missing %s resets_at", name)
	}

	return nil
}

func (w codexRateLimitWindow) window(kind WindowKind, label string, capturedAt time.Time, metadata Metadata) Window {
	return Window{
		Product:     ProductCodex,
		Kind:        kind,
		Label:       label,
		UsedPercent: *w.UsedPercent,
		ResetsAt:    time.Unix(*w.ResetsAt, 0),
		CapturedAt:  capturedAt,
		Metadata:    metadata,
	}
}
