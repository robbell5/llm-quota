package sources

import "time"

type SourceReader interface {
	Fetch(now time.Time) ([]Window, error)
}

// FirstAvailableReader tries Primary first and falls back to Fallback when the
// primary source cannot produce windows. It lets opt-in live sources improve
// freshness without making existing local rollout data unavailable.
type FirstAvailableReader struct {
	Primary  SourceReader
	Fallback SourceReader
}

func (r FirstAvailableReader) Fetch(now time.Time) ([]Window, error) {
	if r.Primary != nil {
		windows, err := r.Primary.Fetch(now)
		if err == nil && len(windows) > 0 {
			return windows, nil
		}
	}
	if r.Fallback == nil {
		return nil, SourceError{Source: ProductCodex, Category: ErrorMissing}
	}
	return r.Fallback.Fetch(now)
}
