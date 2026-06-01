package sources

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"time"
)

const defaultCodexAccountTimeout = 15 * time.Second

type codexAppServerRunner func(context.Context, []byte) ([]byte, error)

// CodexAccountReader fetches the same ChatGPT-managed Codex rate limits shown by
// the Codex app. It is intentionally separate from CodexReader so the default
// source can remain local rollout files unless the caller explicitly opts in.
type CodexAccountReader struct {
	run     codexAppServerRunner
	timeout time.Duration
}

func NewCodexAccountReader() CodexAccountReader {
	return CodexAccountReader{run: runCodexAppServer, timeout: defaultCodexAccountTimeout}
}

func (r CodexAccountReader) Fetch(now time.Time) ([]Window, error) {
	run := r.run
	if run == nil {
		run = runCodexAppServer
	}
	timeout := r.timeout
	if timeout <= 0 {
		timeout = defaultCodexAccountTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	out, err := run(ctx, codexAccountRequest())
	if err != nil {
		return nil, SourceError{Source: ProductCodex, Category: ErrorRead, Err: err}
	}

	windows, err := windowsFromCodexAccountOutput(out, now)
	if err != nil {
		return nil, SourceError{Source: ProductCodex, Category: ErrorNoUsableEvent, Err: err}
	}
	return windows, nil
}

func codexAccountRequest() []byte {
	return []byte(stringsJoinLines(
		`{"method":"initialize","id":0,"params":{"clientInfo":{"name":"llm_quota","title":"llm-quota","version":"0.1.0"}}}`,
		`{"method":"initialized","params":{}}`,
		`{"method":"account/read","id":1,"params":{"refreshToken":false}}`,
		`{"method":"account/rateLimits/read","id":2}`,
	))
}

func stringsJoinLines(lines ...string) string {
	var b bytes.Buffer
	for _, line := range lines {
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

func runCodexAppServer(ctx context.Context, _ []byte) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "codex", "app-server")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(stdout)
	encoder := json.NewEncoder(stdin)
	var out bytes.Buffer
	send := func(v any) error {
		return encoder.Encode(v)
	}
	if err := send(map[string]any{
		"method": "initialize",
		"id":     0,
		"params": map[string]any{"clientInfo": map[string]string{"name": "llm_quota", "title": "llm-quota", "version": "0.1.0"}},
	}); err != nil {
		return nil, finishCodexAppServer(cmd, stdin, err)
	}
	if err := readCodexAppServerResponse(decoder, &out, 0); err != nil {
		return nil, finishCodexAppServer(cmd, stdin, err)
	}
	if err := send(map[string]any{"method": "initialized", "params": map[string]any{}}); err != nil {
		return nil, finishCodexAppServer(cmd, stdin, err)
	}
	if err := send(map[string]any{"method": "account/read", "id": 1, "params": map[string]bool{"refreshToken": false}}); err != nil {
		return nil, finishCodexAppServer(cmd, stdin, err)
	}
	if err := readCodexAppServerResponse(decoder, &out, 1); err != nil {
		return nil, finishCodexAppServer(cmd, stdin, err)
	}
	if err := send(map[string]any{"method": "account/rateLimits/read", "id": 2}); err != nil {
		return nil, finishCodexAppServer(cmd, stdin, err)
	}
	if err := readCodexAppServerResponse(decoder, &out, 2); err != nil {
		return nil, finishCodexAppServer(cmd, stdin, err)
	}
	return out.Bytes(), finishCodexAppServer(cmd, stdin, nil)
}

func readCodexAppServerResponse(decoder *json.Decoder, out *bytes.Buffer, wantID int) error {
	for {
		var env codexAccountEnvelope
		if err := decoder.Decode(&env); err != nil {
			return err
		}
		raw, err := json.Marshal(env)
		if err == nil {
			out.Write(raw)
			out.WriteByte('\n')
		}
		if env.ID == nil || *env.ID != wantID {
			continue
		}
		if env.Error != nil {
			return errors.New(env.Error.Message)
		}
		return nil
	}
}

func finishCodexAppServer(cmd *exec.Cmd, stdin io.Closer, prior error) error {
	_ = stdin.Close()
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case err := <-done:
		if prior != nil {
			return prior
		}
		return err
	case <-time.After(2 * time.Second):
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		<-done
		return prior
	}
}

type codexAccountEnvelope struct {
	ID     *int               `json:"id"`
	Result codexAccountResult `json:"result"`
	Params codexAccountNotify `json:"params"`
	Method string             `json:"method"`
	Error  *codexAccountError `json:"error"`
}

type codexAccountError struct {
	Message string `json:"message"`
}

type codexAccountResult struct {
	Account    *codexAccountInfo       `json:"account"`
	RateLimits *codexAccountRateLimits `json:"rateLimits"`
}

type codexAccountNotify struct {
	RateLimits *codexAccountRateLimits `json:"rateLimits"`
}

type codexAccountInfo struct {
	PlanType string `json:"planType"`
}

type codexAccountRateLimits struct {
	Primary   *codexAccountWindow `json:"primary"`
	Secondary *codexAccountWindow `json:"secondary"`
	PlanType  string              `json:"planType"`
}

type codexAccountWindow struct {
	UsedPercent        *float64 `json:"usedPercent"`
	WindowDurationMins *int     `json:"windowDurationMins"`
	ResetsAt           *int64   `json:"resetsAt"`
}

func windowsFromCodexAccountOutput(out []byte, now time.Time) ([]Window, error) {
	var planType string
	var limits *codexAccountRateLimits
	for _, line := range bytes.Split(out, []byte{'\n'}) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var env codexAccountEnvelope
		if err := json.Unmarshal(line, &env); err != nil {
			continue
		}
		if env.ID != nil && *env.ID == 1 && env.Result.Account != nil {
			planType = env.Result.Account.PlanType
		}
		if env.ID != nil && *env.ID == 2 && env.Result.RateLimits != nil {
			limits = env.Result.RateLimits
			if limits.PlanType != "" {
				planType = limits.PlanType
			}
		}
	}
	if limits == nil {
		return nil, errors.New("no Codex account rate-limit response")
	}
	return limits.windows(now, planType)
}

func (l codexAccountRateLimits) windows(now time.Time, planType string) ([]Window, error) {
	if l.Primary == nil {
		return nil, errors.New("missing primary rate limit")
	}
	if l.Secondary == nil {
		return nil, errors.New("missing secondary rate limit")
	}
	metadata := Metadata(nil)
	if planType != "" {
		metadata = Metadata{"plan_type": planType}
	}
	primary, err := l.Primary.window(WindowFiveHour, "Codex 5h", now, metadata, 300)
	if err != nil {
		return nil, err
	}
	secondary, err := l.Secondary.window(WindowSevenDay, "Codex 7d", now, metadata, 10080)
	if err != nil {
		return nil, err
	}
	return []Window{primary, secondary}, nil
}

func (w codexAccountWindow) window(kind WindowKind, label string, capturedAt time.Time, metadata Metadata, wantMinutes int) (Window, error) {
	if w.UsedPercent == nil {
		return Window{}, fmt.Errorf("missing %s usedPercent", kind)
	}
	if w.WindowDurationMins == nil {
		return Window{}, fmt.Errorf("missing %s windowDurationMins", kind)
	}
	if *w.WindowDurationMins != wantMinutes {
		return Window{}, fmt.Errorf("unexpected %s windowDurationMins", kind)
	}
	if w.ResetsAt == nil {
		return Window{}, fmt.Errorf("missing %s resetsAt", kind)
	}
	return Window{
		Product:     ProductCodex,
		Kind:        kind,
		Label:       label,
		UsedPercent: *w.UsedPercent,
		ResetsAt:    time.Unix(*w.ResetsAt, 0),
		CapturedAt:  capturedAt,
		Metadata:    metadata,
	}, nil
}
