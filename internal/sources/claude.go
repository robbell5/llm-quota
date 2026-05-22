package sources

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

type ClaudeReader struct {
	cachePath string
}

func NewClaudeReader(cachePath string) ClaudeReader {
	return ClaudeReader{cachePath: cachePath}
}

func (r ClaudeReader) Fetch(now time.Time) ([]Window, error) {
	contents, err := os.ReadFile(r.cachePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, SourceError{Source: ProductClaude, Category: ErrorMissing, Err: err}
		}

		return nil, SourceError{Source: ProductClaude, Category: ErrorRead, Err: err}
	}

	var cache claudeCache
	if err := json.Unmarshal(contents, &cache); err != nil {
		return nil, SourceError{Source: ProductClaude, Category: ErrorMalformed, Err: err}
	}

	if err := cache.validate(); err != nil {
		return nil, SourceError{Source: ProductClaude, Category: ErrorMalformed, Err: err}
	}

	writtenAt := time.Unix(*cache.WrittenAt, 0)
	staleAge := now.Sub(writtenAt)
	if staleAge < 0 {
		staleAge = 0
	}
	stale := staleAge > time.Hour

	windows := []Window{
		cache.FiveHour.window(WindowFiveHour, "Claude 5h", writtenAt, stale, staleAge),
		cache.SevenDay.window(WindowSevenDay, "Claude 7d", writtenAt, stale, staleAge),
	}
	if sonnet, ok := cache.validSonnetSevenDay(); ok {
		windows = append(windows, sonnet.window(WindowSonnetSevenDay, "Sonnet 7d", writtenAt, stale, staleAge))
	}

	return windows, nil
}

type claudeCache struct {
	FiveHour       *claudeCacheWindow `json:"five_hour"`
	SevenDay       *claudeCacheWindow `json:"seven_day"`
	SonnetSevenDay *claudeCacheWindow `json:"sonnet_seven_day"`
	SonnetWeekly   *claudeCacheWindow `json:"sonnet_weekly"`
	WrittenAt      *int64             `json:"written_at"`
}

func (c claudeCache) validate() error {
	if c.WrittenAt == nil {
		return errors.New("missing written_at")
	}

	if c.FiveHour == nil {
		return errors.New("missing five_hour window")
	}
	if err := c.FiveHour.validate("five_hour"); err != nil {
		return err
	}

	if c.SevenDay == nil {
		return errors.New("missing seven_day window")
	}
	if err := c.SevenDay.validate("seven_day"); err != nil {
		return err
	}

	return nil
}

func (c claudeCache) validSonnetSevenDay() (claudeCacheWindow, bool) {
	for _, window := range []*claudeCacheWindow{c.SonnetSevenDay, c.SonnetWeekly} {
		if window == nil {
			continue
		}
		if err := window.validate("sonnet_seven_day"); err != nil {
			continue
		}
		return *window, true
	}

	return claudeCacheWindow{}, false
}

type claudeCacheWindow struct {
	UsedPercentage *float64 `json:"used_percentage"`
	ResetsAt       *int64   `json:"resets_at"`
}

func (w claudeCacheWindow) validate(name string) error {
	if w.UsedPercentage == nil {
		return fmt.Errorf("missing %s used_percentage", name)
	}
	if w.ResetsAt == nil {
		return fmt.Errorf("missing %s resets_at", name)
	}

	return nil
}

func (w claudeCacheWindow) window(kind WindowKind, label string, capturedAt time.Time, stale bool, staleAge time.Duration) Window {
	return Window{
		Product:     ProductClaude,
		Kind:        kind,
		Label:       label,
		UsedPercent: *w.UsedPercentage,
		ResetsAt:    time.Unix(*w.ResetsAt, 0),
		CapturedAt:  capturedAt,
		Stale:       stale,
		StaleAge:    staleAge,
	}
}
