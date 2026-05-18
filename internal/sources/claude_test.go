package sources

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestClaudeFetch(t *testing.T) {
	now := time.Unix(1_778_940_000, 0)

	tests := []struct {
		name         string
		writeCache   bool
		cache        string
		wantWindows  []Window
		wantCategory ErrorCategory
	}{
		{
			name:       "valid cache",
			writeCache: true,
			cache: `{
				"five_hour": {"used_percentage": 42.3, "resets_at": 1778942485},
				"seven_day": {"used_percentage": 85.7, "resets_at": 1779382265},
				"written_at": 1778940000
			}`,
			wantWindows: []Window{
				{
					Product:     ProductClaude,
					Kind:        WindowFiveHour,
					Label:       "Claude 5h",
					UsedPercent: 42.3,
					ResetsAt:    time.Unix(1_778_942_485, 0),
					CapturedAt:  now,
				},
				{
					Product:     ProductClaude,
					Kind:        WindowSevenDay,
					Label:       "Claude 7d",
					UsedPercent: 85.7,
					ResetsAt:    time.Unix(1_779_382_265, 0),
					CapturedAt:  now,
				},
			},
		},
		{
			name:         "missing cache",
			wantCategory: ErrorMissing,
		},
		{
			name:         "malformed cache",
			writeCache:   true,
			cache:        `{`,
			wantCategory: ErrorMalformed,
		},
		{
			name:       "missing seven day rejects all",
			writeCache: true,
			cache: `{
				"five_hour": {"used_percentage": 42.3, "resets_at": 1778942485},
				"written_at": 1778940000
			}`,
			wantCategory: ErrorMalformed,
		},
		{
			name:       "stale cache returns windows",
			writeCache: true,
			cache: `{
				"five_hour": {"used_percentage": 42.3, "resets_at": 1778942485},
				"seven_day": {"used_percentage": 85.7, "resets_at": 1779382265},
				"written_at": 1778932800
			}`,
			wantWindows: []Window{
				{
					Product:     ProductClaude,
					Kind:        WindowFiveHour,
					Label:       "Claude 5h",
					UsedPercent: 42.3,
					ResetsAt:    time.Unix(1_778_942_485, 0),
					CapturedAt:  time.Unix(1_778_932_800, 0),
					Stale:       true,
					StaleAge:    2 * time.Hour,
				},
				{
					Product:     ProductClaude,
					Kind:        WindowSevenDay,
					Label:       "Claude 7d",
					UsedPercent: 85.7,
					ResetsAt:    time.Unix(1_779_382_265, 0),
					CapturedAt:  time.Unix(1_778_932_800, 0),
					Stale:       true,
					StaleAge:    2 * time.Hour,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cachePath := filepath.Join(t.TempDir(), "claude.json")
			if tt.writeCache {
				if err := os.WriteFile(cachePath, []byte(tt.cache), 0o600); err != nil {
					t.Fatalf("write cache: %v", err)
				}
			}

			windows, err := NewClaudeReader(cachePath).Fetch(now)
			if tt.wantCategory != "" {
				assertSourceError(t, err, tt.wantCategory)
				if len(windows) != 0 {
					t.Fatalf("expected no windows on error, got %#v", windows)
				}
				return
			}

			if err != nil {
				t.Fatalf("Fetch returned error: %v", err)
			}

			assertWindows(t, windows, tt.wantWindows)
		})
	}
}

func TestClaudeFetchUnreadableCacheReturnsReadError(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "claude.json")
	if err := os.WriteFile(cachePath, []byte(`{}`), 0o600); err != nil {
		t.Fatalf("write cache: %v", err)
	}
	if err := os.Chmod(cachePath, 0o000); err != nil {
		t.Fatalf("chmod cache: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(cachePath, 0o600) })
	if _, err := os.ReadFile(cachePath); err == nil {
		t.Skip("test filesystem did not enforce unreadable cache permissions")
	}

	windows, err := NewClaudeReader(cachePath).Fetch(time.Unix(1_778_940_000, 0))
	assertSourceError(t, err, ErrorRead)
	if len(windows) != 0 {
		t.Fatalf("expected no windows on read error, got %#v", windows)
	}
}

func assertSourceError(t *testing.T, err error, category ErrorCategory) {
	t.Helper()

	var sourceErr SourceError
	if !errors.As(err, &sourceErr) {
		t.Fatalf("expected SourceError, got %T: %v", err, err)
	}

	if sourceErr.Source != ProductClaude {
		t.Fatalf("expected source %q, got %q", ProductClaude, sourceErr.Source)
	}

	if sourceErr.Category != category {
		t.Fatalf("expected category %q, got %q", category, sourceErr.Category)
	}
}

func assertWindows(t *testing.T, got, want []Window) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("expected %d windows, got %d: %#v", len(want), len(got), got)
	}

	for i := range want {
		if got[i].Product != want[i].Product {
			t.Fatalf("window %d product = %q, want %q", i, got[i].Product, want[i].Product)
		}
		if got[i].Kind != want[i].Kind {
			t.Fatalf("window %d kind = %q, want %q", i, got[i].Kind, want[i].Kind)
		}
		if got[i].Label != want[i].Label {
			t.Fatalf("window %d label = %q, want %q", i, got[i].Label, want[i].Label)
		}
		if got[i].UsedPercent != want[i].UsedPercent {
			t.Fatalf("window %d used percent = %v, want %v", i, got[i].UsedPercent, want[i].UsedPercent)
		}
		if !got[i].ResetsAt.Equal(want[i].ResetsAt) {
			t.Fatalf("window %d resets at = %s, want %s", i, got[i].ResetsAt, want[i].ResetsAt)
		}
		if !got[i].CapturedAt.Equal(want[i].CapturedAt) {
			t.Fatalf("window %d captured at = %s, want %s", i, got[i].CapturedAt, want[i].CapturedAt)
		}
		if got[i].Stale != want[i].Stale {
			t.Fatalf("window %d stale = %v, want %v", i, got[i].Stale, want[i].Stale)
		}
		if got[i].StaleAge != want[i].StaleAge {
			t.Fatalf("window %d stale age = %s, want %s", i, got[i].StaleAge, want[i].StaleAge)
		}
		if !reflect.DeepEqual(got[i].Metadata, want[i].Metadata) {
			t.Fatalf("window %d metadata = %#v, want %#v", i, got[i].Metadata, want[i].Metadata)
		}
	}
}
