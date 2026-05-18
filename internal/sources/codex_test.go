package sources

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCodexFetchNewestUsableRollout(t *testing.T) {
	now := time.Unix(1_778_940_000, 0)
	sessionsRoot := t.TempDir()

	writeRollout(t, sessionsRoot, "2026/05/18/rollout-newest.jsonl", time.Unix(300, 0), []string{
		`{"type":"session_configured","payload":{"rate_limits":{"primary":{"used_percent":99}}}}`,
		`not-json`,
		`{"type":"event_msg","payload":{"type":"token_count","rate_limits":null}}`,
		`{"type":"event_msg","payload":{"type":"token_count","rate_limits":{"primary":{"used_percent":41.5,"window_minutes":300,"resets_at":1778942485},"secondary":{"used_percent":18.25,"window_minutes":10080,"resets_at":1779382265},"plan_type":"prolite"}}}`,
	})
	writeRollout(t, sessionsRoot, "2026/05/17/rollout-older.jsonl", time.Unix(200, 0), []string{
		`{"type":"event_msg","payload":{"type":"token_count","rate_limits":{"primary":{"used_percent":9,"window_minutes":300,"resets_at":1778940001},"secondary":{"used_percent":10,"window_minutes":10080,"resets_at":1779380001},"plan_type":"older"}}}`,
	})

	windows, err := NewCodexReader(sessionsRoot).Fetch(now)
	if err != nil {
		t.Fatalf("Fetch returned error: %v", err)
	}

	assertWindows(t, windows, []Window{
		{
			Product:     ProductCodex,
			Kind:        WindowFiveHour,
			Label:       "Codex 5h",
			UsedPercent: 41.5,
			ResetsAt:    time.Unix(1_778_942_485, 0),
			CapturedAt:  now,
			Metadata:    Metadata{"plan_type": "prolite"},
		},
		{
			Product:     ProductCodex,
			Kind:        WindowSevenDay,
			Label:       "Codex 7d",
			UsedPercent: 18.25,
			ResetsAt:    time.Unix(1_779_382_265, 0),
			CapturedAt:  now,
			Metadata:    Metadata{"plan_type": "prolite"},
		},
	})
}

func TestCodexFetchFallsBackToOlderUsableRollout(t *testing.T) {
	now := time.Unix(1_778_940_000, 0)
	sessionsRoot := t.TempDir()

	writeRollout(t, sessionsRoot, "2026/05/18/rollout-newest.jsonl", time.Unix(300, 0), []string{
		`{"type":"event_msg","payload":{"type":"token_count","rate_limits":null}}`,
		`not-json`,
	})
	writeRollout(t, sessionsRoot, "2026/05/17/rollout-older.jsonl", time.Unix(200, 0), []string{
		`{"type":"event_msg","payload":{"type":"token_count","rate_limits":{"primary":{"used_percent":33,"window_minutes":300,"resets_at":1778942485},"secondary":{"used_percent":66,"window_minutes":10080,"resets_at":1779382265},"plan_type":"fallback"}}}`,
	})

	windows, err := NewCodexReader(sessionsRoot).Fetch(now)
	if err != nil {
		t.Fatalf("Fetch returned error: %v", err)
	}

	assertWindows(t, windows, []Window{
		{
			Product:     ProductCodex,
			Kind:        WindowFiveHour,
			Label:       "Codex 5h",
			UsedPercent: 33,
			ResetsAt:    time.Unix(1_778_942_485, 0),
			CapturedAt:  now,
			Metadata:    Metadata{"plan_type": "fallback"},
		},
		{
			Product:     ProductCodex,
			Kind:        WindowSevenDay,
			Label:       "Codex 7d",
			UsedPercent: 66,
			ResetsAt:    time.Unix(1_779_382_265, 0),
			CapturedAt:  now,
			Metadata:    Metadata{"plan_type": "fallback"},
		},
	})
}

func TestCodexFetchNoUsableRollout(t *testing.T) {
	sessionsRoot := t.TempDir()

	writeRollout(t, sessionsRoot, "2026/05/18/rollout-empty.jsonl", time.Unix(300, 0), []string{
		`{"type":"event_msg","payload":{"type":"token_count","rate_limits":null}}`,
		`{"type":"event_msg","payload":{"type":"other","rate_limits":{"primary":{}}}}`,
		`malformed`,
	})

	windows, err := NewCodexReader(sessionsRoot).Fetch(time.Unix(1_778_940_000, 0))
	assertCodexSourceError(t, err, ErrorNoUsableEvent)
	if len(windows) != 0 {
		t.Fatalf("expected no windows on error, got %#v", windows)
	}
}

func writeRollout(t *testing.T, root, relativePath string, modified time.Time, lines []string) {
	t.Helper()

	path := filepath.Join(root, relativePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("make rollout dir: %v", err)
	}

	contents := ""
	for _, line := range lines {
		contents += line + "\n"
	}

	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write rollout: %v", err)
	}
	if err := os.Chtimes(path, modified, modified); err != nil {
		t.Fatalf("set rollout mtime: %v", err)
	}
}

func assertCodexSourceError(t *testing.T, err error, category ErrorCategory) {
	t.Helper()

	var sourceErr SourceError
	if !errors.As(err, &sourceErr) {
		t.Fatalf("expected SourceError, got %T: %v", err, err)
	}

	if sourceErr.Source != ProductCodex {
		t.Fatalf("expected source %q, got %q", ProductCodex, sourceErr.Source)
	}

	if sourceErr.Category != category {
		t.Fatalf("expected category %q, got %q", category, sourceErr.Category)
	}
}
