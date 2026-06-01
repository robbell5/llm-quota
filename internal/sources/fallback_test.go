package sources

import (
	"errors"
	"testing"
	"time"
)

type stubSourceReader struct {
	windows []Window
	err     error
}

func (r stubSourceReader) Fetch(time.Time) ([]Window, error) {
	return r.windows, r.err
}

func TestFirstAvailableReaderPrefersPrimary(t *testing.T) {
	now := time.Unix(1_780_350_900, 0)
	primary := []Window{{Product: ProductCodex, Kind: WindowFiveHour, UsedPercent: 12, CapturedAt: now}}
	fallback := []Window{{Product: ProductCodex, Kind: WindowFiveHour, UsedPercent: 9, CapturedAt: now}}

	windows, err := FirstAvailableReader{
		Primary:  stubSourceReader{windows: primary},
		Fallback: stubSourceReader{windows: fallback},
	}.Fetch(now)
	if err != nil {
		t.Fatalf("Fetch returned error: %v", err)
	}
	if windows[0].UsedPercent != 12 {
		t.Fatalf("UsedPercent = %v, want primary value 12", windows[0].UsedPercent)
	}
}

func TestFirstAvailableReaderFallsBackAfterPrimaryError(t *testing.T) {
	now := time.Unix(1_780_350_900, 0)
	fallback := []Window{{Product: ProductCodex, Kind: WindowFiveHour, UsedPercent: 9, CapturedAt: now}}

	windows, err := FirstAvailableReader{
		Primary:  stubSourceReader{err: errors.New("live source unavailable")},
		Fallback: stubSourceReader{windows: fallback},
	}.Fetch(now)
	if err != nil {
		t.Fatalf("Fetch returned error: %v", err)
	}
	if windows[0].UsedPercent != 9 {
		t.Fatalf("UsedPercent = %v, want fallback value 9", windows[0].UsedPercent)
	}
}
