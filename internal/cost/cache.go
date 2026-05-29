package cost

import (
	"os"
	"sync"
	"time"
)

// parseCache memoizes parsed entries per file, keyed by (size, mtime). It is
// safe for the single-flight refresh goroutine; the mutex guards against any
// future concurrent caller. Tail parsing is intentionally not implemented:
// unchanged files are never reparsed and files older than the window are
// skipped before reaching here, so only the active file is fully reparsed each
// 30s tick — acceptable at that cadence (revisit if profiling says otherwise).
// There is no eviction; callers pass only currently-relevant files.
type parseCache struct {
	mu    sync.Mutex
	files map[string]cachedFile
}

type cachedFile struct {
	size    int64
	mtime   time.Time
	entries []entry
}

func newParseCache() *parseCache {
	return &parseCache{files: map[string]cachedFile{}}
}

// load returns cached entries for path when size+mtime match the last parse,
// otherwise calls parse and stores the result.
func (c *parseCache) load(path string, info os.FileInfo, parse func() ([]entry, error)) ([]entry, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if cached, ok := c.files[path]; ok &&
		cached.size == info.Size() && cached.mtime.Equal(info.ModTime()) {
		return cached.entries, nil
	}

	entries, err := parse()
	if err != nil {
		return nil, err
	}
	c.files[path] = cachedFile{size: info.Size(), mtime: info.ModTime(), entries: entries}
	return entries, nil
}
