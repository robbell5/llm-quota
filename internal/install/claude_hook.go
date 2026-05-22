package install

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

const (
	managedHookName              = "llm-quota"
	managedHookMarker            = "llm-quota"
	managedStatusLineOriginalKey = "llm_quota_original_statusLine"
	claudeHookEvent              = "PostToolUse"
)

type ClaudeHookPaths struct {
	ClaudeConfigPath string
	StatePath        string
	CachePath        string
	ExecutablePath   string
}

type InstallResult struct {
	Changed    bool
	BackupPath string
	Message    string
}

func InstallClaudeHook(paths ClaudeHookPaths) (InstallResult, error) {
	if paths.ClaudeConfigPath == "" {
		return InstallResult{}, errors.New("claude config path is required")
	}
	if paths.CachePath == "" {
		return InstallResult{}, errors.New("cache path is required")
	}

	config, existed, err := readClaudeConfig(paths.ClaudeConfigPath)
	if err != nil {
		return InstallResult{}, err
	}
	original := cloneJSONMap(config)

	if err := installManagedStatusLine(config, paths.ExecutablePath, paths.CachePath); err != nil {
		return InstallResult{}, err
	}
	removeManagedToolHook(config)
	if reflect.DeepEqual(original, config) {
		return InstallResult{Message: "llm-quota Claude hook already installed"}, nil
	}

	var backupPath string
	if existed {
		backupPath, err = backupFile(paths.ClaudeConfigPath)
		if err != nil {
			return InstallResult{}, err
		}
	}
	if err := writeJSONAtomic(paths.ClaudeConfigPath, config); err != nil {
		return InstallResult{}, err
	}

	return InstallResult{
		Changed:    true,
		BackupPath: backupPath,
		Message:    "installed llm-quota Claude hook",
	}, nil
}

func UninstallClaudeHook(paths ClaudeHookPaths) (InstallResult, error) {
	if paths.ClaudeConfigPath == "" {
		return InstallResult{}, errors.New("claude config path is required")
	}

	config, existed, err := readClaudeConfig(paths.ClaudeConfigPath)
	if err != nil {
		return InstallResult{}, err
	}
	if !existed || len(config) == 0 {
		return InstallResult{Message: "llm-quota Claude hook is not installed"}, nil
	}
	original := cloneJSONMap(config)

	if statusLine, ok := config["statusLine"].(map[string]any); ok && statusLine["llm_quota_marker"] == managedHookMarker {
		passthrough, _ := statusLine["llm_quota_passthrough"].(string)
		if originalStatusLine, ok := statusLine[managedStatusLineOriginalKey].(map[string]any); ok {
			config["statusLine"] = cloneJSONMap(originalStatusLine)
		} else if passthrough != "" {
			config["statusLine"] = map[string]any{
				"type":    "command",
				"command": passthrough,
			}
		} else {
			delete(config, "statusLine")
		}
	}
	removeManagedToolHook(config)

	if reflect.DeepEqual(original, config) {
		return InstallResult{Message: "llm-quota Claude hook is not installed"}, nil
	}

	backupPath, err := backupFile(paths.ClaudeConfigPath)
	if err != nil {
		return InstallResult{}, err
	}
	if err := writeJSONAtomic(paths.ClaudeConfigPath, config); err != nil {
		return InstallResult{}, err
	}

	return InstallResult{
		Changed:    true,
		BackupPath: backupPath,
		Message:    "uninstalled llm-quota Claude hook",
	}, nil
}

func RecordClaudeHookDeclined(statePath string) error {
	if statePath == "" {
		return errors.New("state path is required")
	}

	state, err := readDeclineState(statePath)
	if err != nil {
		return err
	}
	state.ClaudeHookDeclined = true
	return writeJSONAtomic(statePath, state)
}

func ClaudeHookDeclined(statePath string) (bool, error) {
	if statePath == "" {
		return false, errors.New("state path is required")
	}

	state, err := readDeclineState(statePath)
	if err != nil {
		return false, err
	}
	return state.ClaudeHookDeclined, nil
}

type declineState struct {
	ClaudeHookDeclined bool `json:"claude_hook_declined"`
}

func readDeclineState(path string) (declineState, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return declineState{}, nil
		}
		return declineState{}, err
	}

	var state declineState
	if err := json.Unmarshal(contents, &state); err != nil {
		return declineState{}, err
	}
	return state, nil
}

func readClaudeConfig(path string) (map[string]any, bool, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]any{}, false, nil
		}
		return nil, false, err
	}
	if len(bytes.TrimSpace(contents)) == 0 {
		return map[string]any{}, true, nil
	}

	var config map[string]any
	if err := json.Unmarshal(contents, &config); err != nil {
		return nil, false, err
	}
	if config == nil {
		config = map[string]any{}
	}
	return config, true, nil
}

func installManagedStatusLine(config map[string]any, executablePath string, cachePath string) error {
	statusLine, _ := config["statusLine"].(map[string]any)
	passthrough := ""
	var originalStatusLine map[string]any
	if statusLine != nil {
		if statusLine["llm_quota_marker"] == managedHookMarker {
			passthrough, _ = statusLine["llm_quota_passthrough"].(string)
			originalStatusLine, _ = statusLine[managedStatusLineOriginalKey].(map[string]any)
		} else {
			passthrough, _ = statusLine["command"].(string)
			originalStatusLine = cloneJSONMap(statusLine)
		}
	}

	managedStatusLine := map[string]any{
		"type":                  "command",
		"command":               ManagedStatusLineCommand(executablePath, cachePath, passthrough),
		"llm_quota_marker":      managedHookMarker,
		"llm_quota_passthrough": passthrough,
	}
	if originalStatusLine != nil {
		managedStatusLine[managedStatusLineOriginalKey] = originalStatusLine
	}
	config["statusLine"] = managedStatusLine
	return nil
}

func removeManagedToolHook(config map[string]any) {
	hooks, ok := config["hooks"].(map[string]any)
	if !ok {
		return
	}
	entries, ok := hooks[claudeHookEvent].([]any)
	if !ok {
		return
	}
	kept := entries[:0]
	for _, entry := range entries {
		hook, ok := entry.(map[string]any)
		if ok && isManagedHook(hook) {
			continue
		}
		kept = append(kept, entry)
	}
	hooks[claudeHookEvent] = kept
}

func isManagedHook(hook map[string]any) bool {
	return hook["llm_quota_marker"] == managedHookMarker
}

func ManagedHookCommand(executablePath string, cachePath string) string {
	if executablePath == "" {
		executablePath = managedHookName
	}
	return shellQuote(executablePath) + " claude-hook-cache-writer --cache " + shellQuote(cachePath)
}

func ManagedStatusLineCommand(executablePath string, cachePath string, passthrough string) string {
	if executablePath == "" {
		executablePath = managedHookName
	}
	command := shellQuote(executablePath) + " claude-statusline-cache-writer --cache " + shellQuote(cachePath)
	if passthrough != "" {
		command += " --passthrough " + shellQuote(passthrough)
	}
	return command
}

func RunClaudeHookCacheWriter(input io.Reader, cachePath string, now time.Time) error {
	contents, err := io.ReadAll(input)
	if err != nil {
		return err
	}
	return writeClaudeCache(contents, cachePath, now, true)
}

func RunClaudeStatusLineCacheWriter(input io.Reader, stdout io.Writer, stderr io.Writer, cachePath string, passthrough string, now time.Time) error {
	contents, err := io.ReadAll(input)
	if err != nil {
		return err
	}
	cacheErr := writeClaudeCache(contents, cachePath, now, false)
	if passthrough == "" {
		return cacheErr
	}

	command := exec.Command("sh", "-c", passthrough)
	command.Stdin = bytes.NewReader(contents)
	command.Stdout = stdout
	command.Stderr = stderr
	return command.Run()
}

func writeClaudeCache(contents []byte, cachePath string, now time.Time, requireRateLimits bool) error {
	if cachePath == "" {
		return errors.New("cache path is required")
	}

	var payload claudeHookPayload
	decoder := json.NewDecoder(bytes.NewReader(contents))
	if err := decoder.Decode(&payload); err != nil {
		return fmt.Errorf("decode Claude hook input: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("decode Claude hook input: trailing JSON value")
		}
		return fmt.Errorf("decode Claude hook input: %w", err)
	}

	rateLimits := payload.RateLimits
	if rateLimits == nil && payload.Payload != nil {
		rateLimits = payload.Payload.RateLimits
	}
	if rateLimits == nil {
		if !requireRateLimits {
			return nil
		}
		return errors.New("missing rate_limits")
	}
	if err := rateLimits.validate(); err != nil {
		return err
	}

	writtenAt := now.Unix()
	cache := claudeHookCache{
		FiveHour:       rateLimits.FiveHour,
		SevenDay:       rateLimits.SevenDay,
		SonnetSevenDay: rateLimits.validSonnetSevenDay(),
		WrittenAt:      &writtenAt,
	}
	return writeJSONAtomic(cachePath, cache)
}

type claudeHookPayload struct {
	RateLimits *claudeHookRateLimits `json:"rate_limits"`
	Payload    *struct {
		RateLimits *claudeHookRateLimits `json:"rate_limits"`
	} `json:"payload"`
}

type claudeHookCache struct {
	FiveHour       *claudeHookWindow `json:"five_hour"`
	SevenDay       *claudeHookWindow `json:"seven_day"`
	SonnetSevenDay *claudeHookWindow `json:"sonnet_seven_day,omitempty"`
	WrittenAt      *int64            `json:"written_at"`
}

type claudeHookRateLimits struct {
	FiveHour       *claudeHookWindow `json:"five_hour"`
	SevenDay       *claudeHookWindow `json:"seven_day"`
	SonnetSevenDay *claudeHookWindow `json:"sonnet_seven_day"`
	SonnetWeekly   *claudeHookWindow `json:"sonnet_weekly"`
}

func (r claudeHookRateLimits) validate() error {
	if r.FiveHour == nil {
		return errors.New("missing five_hour rate limit")
	}
	if err := r.FiveHour.validate("five_hour"); err != nil {
		return err
	}
	if r.SevenDay == nil {
		return errors.New("missing seven_day rate limit")
	}
	if err := r.SevenDay.validate("seven_day"); err != nil {
		return err
	}
	return nil
}

func (r claudeHookRateLimits) validSonnetSevenDay() *claudeHookWindow {
	for _, window := range []*claudeHookWindow{r.SonnetSevenDay, r.SonnetWeekly} {
		if window == nil {
			continue
		}
		if err := window.validate("sonnet_seven_day"); err != nil {
			continue
		}
		return window
	}

	return nil
}

type claudeHookWindow struct {
	UsedPercentage *float64 `json:"used_percentage"`
	ResetsAt       *int64   `json:"resets_at"`
}

func (w claudeHookWindow) validate(name string) error {
	if w.UsedPercentage == nil {
		return fmt.Errorf("missing %s used_percentage", name)
	}
	if w.ResetsAt == nil {
		return fmt.Errorf("missing %s resets_at", name)
	}
	return nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func backupFile(path string) (string, error) {
	backupPath := fmt.Sprintf("%s.llm-quota-backup-%s", path, time.Now().UTC().Format("20060102T150405.000000000Z"))

	input, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer input.Close()

	output, err := os.OpenFile(backupPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(output, input); err != nil {
		_ = output.Close()
		return "", err
	}
	if err := output.Close(); err != nil {
		return "", err
	}
	return backupPath, nil
}

func writeJSONAtomic(path string, value any) error {
	writePath, err := resolveWritePath(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(writePath), 0o700); err != nil {
		return err
	}

	contents, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	contents = append(contents, '\n')

	tempFile, err := os.CreateTemp(filepath.Dir(writePath), "."+filepath.Base(writePath)+".*.tmp")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	if _, err := tempFile.Write(contents); err != nil {
		tempFile.Close()
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}

	return os.Rename(tempPath, writePath)
}

func resolveWritePath(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return path, nil
		}
		return "", err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return path, nil
	}

	target, err := os.Readlink(path)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(path), target)
	}
	return target, nil
}

func cloneJSONMap(value map[string]any) map[string]any {
	contents, err := json.Marshal(value)
	if err != nil {
		return map[string]any{}
	}
	var clone map[string]any
	if err := json.Unmarshal(contents, &clone); err != nil {
		return map[string]any{}
	}
	if clone == nil {
		return map[string]any{}
	}
	return clone
}
