package trend

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const storeVersion = 1

type Store struct {
	path string
}

func NewStore(path string) *Store {
	return &Store{path: path}
}

type storedSample struct {
	CapturedAt int64   `json:"captured_at"`
	UsedPct    float64 `json:"used_pct"`
	ResetsAt   int64   `json:"resets_at"`
}

type storedHistory struct {
	Version int                       `json:"version"`
	Windows map[string][]storedSample `json:"windows"`
}

// Load reads the history file. Any problem (missing, unreadable, malformed,
// unknown version) yields an empty, non-nil History rather than an error.
func (s *Store) Load() *History {
	h := NewHistory()
	contents, err := os.ReadFile(s.path)
	if err != nil {
		return h
	}
	var stored storedHistory
	if err := json.Unmarshal(contents, &stored); err != nil {
		return h
	}
	if stored.Version != storeVersion {
		return h
	}
	for key, samples := range stored.Windows {
		converted := make([]Sample, 0, len(samples))
		for _, ss := range samples {
			converted = append(converted, Sample{
				CapturedAt: time.Unix(ss.CapturedAt, 0),
				UsedPct:    ss.UsedPct,
				ResetsAt:   time.Unix(ss.ResetsAt, 0),
			})
		}
		h.windows[key] = converted
	}
	return h
}

// Save atomically writes the history file (tmpfile + rename). Callers may
// ignore the error; the UI degrades gracefully without persisted history.
func (s *Store) Save(h *History) error {
	stored := storedHistory{Version: storeVersion, Windows: map[string][]storedSample{}}
	for key, samples := range h.windows {
		out := make([]storedSample, 0, len(samples))
		for _, sample := range samples {
			out = append(out, storedSample{
				CapturedAt: sample.CapturedAt.Unix(),
				UsedPct:    sample.UsedPct,
				ResetsAt:   sample.ResetsAt.Unix(),
			})
		}
		stored.Windows[key] = out
	}

	data, err := json.Marshal(stored)
	if err != nil {
		return err
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "history-*.json")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, s.path); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}
