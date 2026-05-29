package cost

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseCacheReusesUnchangedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.jsonl")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(path)

	calls := 0
	parse := func() ([]entry, error) {
		calls++
		return []entry{{model: "m", ts: time.Unix(1, 0)}}, nil
	}
	c := newParseCache()

	if _, err := c.load(path, info, parse); err != nil {
		t.Fatal(err)
	}
	if _, err := c.load(path, info, parse); err != nil { // same size+mtime
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("expected parse called once, got %d", calls)
	}
}

func TestParseCacheReparsesOnChange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.jsonl")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	info1, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	calls := 0
	parse := func() ([]entry, error) { calls++; return nil, nil }
	c := newParseCache()
	if _, err := c.load(path, info1, parse); err != nil {
		t.Fatal(err)
	}

	// Grow the file and bump mtime so the cache key changes.
	if err := os.WriteFile(path, []byte("xy"), 0o600); err != nil {
		t.Fatal(err)
	}
	later := time.Now().Add(time.Minute)
	if err := os.Chtimes(path, later, later); err != nil {
		t.Fatal(err)
	}
	info2, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.load(path, info2, parse); err != nil {
		t.Fatal(err)
	}

	if calls != 2 {
		t.Fatalf("expected reparse on change, got %d calls", calls)
	}
}
