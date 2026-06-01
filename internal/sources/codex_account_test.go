package sources

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestCodexAccountReaderFetchMapsAppServerRateLimits(t *testing.T) {
	now := time.Unix(1_780_350_900, 0)
	var request string
	reader := CodexAccountReader{
		run: func(_ context.Context, input []byte) ([]byte, error) {
			request = string(input)
			return []byte(strings.Join([]string{
				`{"id":0,"result":{"codexHome":"/Users/rob/.codex"}}`,
				`{"id":1,"result":{"account":{"type":"chatgpt","planType":"prolite"},"requiresOpenaiAuth":true}}`,
				`{"method":"account/rateLimits/updated","params":{"rateLimits":{"primary":{"usedPercent":11,"windowDurationMins":300,"resetsAt":1780368000}}}}`,
				`{"id":2,"result":{"rateLimits":{"primary":{"usedPercent":12,"windowDurationMins":300,"resetsAt":1780369000},"secondary":{"usedPercent":2,"windowDurationMins":10080,"resetsAt":1780954000},"planType":"pro","rateLimitReachedType":null}}}`,
			}, "\n")), nil
		},
		timeout: time.Second,
	}

	windows, err := reader.Fetch(now)
	if err != nil {
		t.Fatalf("Fetch returned error: %v", err)
	}

	assertWindows(t, windows, []Window{
		{
			Product:     ProductCodex,
			Kind:        WindowFiveHour,
			Label:       "Codex 5h",
			UsedPercent: 12,
			ResetsAt:    time.Unix(1_780_369_000, 0),
			CapturedAt:  now,
			Metadata:    Metadata{"plan_type": "pro"},
		},
		{
			Product:     ProductCodex,
			Kind:        WindowSevenDay,
			Label:       "Codex 7d",
			UsedPercent: 2,
			ResetsAt:    time.Unix(1_780_954_000, 0),
			CapturedAt:  now,
			Metadata:    Metadata{"plan_type": "pro"},
		},
	})
	if !strings.Contains(request, `"method":"account/read"`) {
		t.Fatalf("request should read account plan type, got %s", request)
	}
	if !strings.Contains(request, `"method":"account/rateLimits/read"`) {
		t.Fatalf("request should read account rate limits, got %s", request)
	}
}
