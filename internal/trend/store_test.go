package trend

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStoreRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "history.json")
	store := NewStore(path)

	h := NewHistory()
	reset := at(120)
	h.Append("claude:five_hour", Sample{CapturedAt: at(0), UsedPct: 10, ResetsAt: reset})
	h.Append("claude:five_hour", Sample{CapturedAt: at(10), UsedPct: 25, ResetsAt: reset})
	if err := store.Save(h); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded := NewStore(path).Load()
	got := loaded.EpochSamples("claude:five_hour", reset)
	if len(got) != 2 || got[1].UsedPct != 25 {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
	if !got[0].CapturedAt.Equal(at(0)) {
		t.Fatalf("captured_at not preserved: %v", got[0].CapturedAt)
	}
}

func TestStoreLoadMissingReturnsEmpty(t *testing.T) {
	h := NewStore(filepath.Join(t.TempDir(), "nope.json")).Load()
	if h == nil || len(h.EpochSamples("k", at(0))) != 0 {
		t.Fatalf("missing file should load empty, non-nil history")
	}
}

func TestStoreLoadMalformedReturnsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if h := NewStore(path).Load(); h == nil {
		t.Fatalf("malformed file should load empty, non-nil history")
	}
}

func TestStoreLoadVersionMismatchReturnsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	contents := `{"version":999,"windows":{"k":[{"captured_at":1,"used_pct":5,"resets_at":2}]}}`
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	h := NewStore(path).Load()
	if h == nil {
		t.Fatalf("version mismatch should return non-nil history")
	}
	if n := len(h.EpochSamples("k", time.Unix(2, 0))); n != 0 {
		t.Fatalf("version mismatch should discard data, got %d samples", n)
	}
}
