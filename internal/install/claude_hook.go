package install

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

const (
	managedHookName   = "llm-quota"
	managedHookMarker = "llm-quota"
	claudeHookEvent   = "PostToolUse"
)

type ClaudeHookPaths struct {
	ClaudeConfigPath string
	StatePath        string
	CachePath        string
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

	if err := installManagedHook(config, paths.CachePath); err != nil {
		return InstallResult{}, err
	}
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

func installManagedHook(config map[string]any, cachePath string) error {
	hooks, err := hooksObject(config)
	if err != nil {
		return err
	}

	entries, err := getHookEntries(hooks, claudeHookEvent)
	if err != nil {
		return err
	}
	managed := managedHook(cachePath)

	for index, entry := range entries {
		hook, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		if isManagedHook(hook) {
			entries[index] = managed
			hooks[claudeHookEvent] = entries
			return nil
		}
	}

	hooks[claudeHookEvent] = append(entries, managed)
	return nil
}

func hooksObject(config map[string]any) (map[string]any, error) {
	raw, ok := config["hooks"]
	if !ok || raw == nil {
		hooks := map[string]any{}
		config["hooks"] = hooks
		return hooks, nil
	}

	hooks, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unsupported Claude hooks shape %T", raw)
	}
	return hooks, nil
}

func getHookEntries(hooks map[string]any, event string) ([]any, error) {
	raw, ok := hooks[event]
	if !ok || raw == nil {
		return []any{}, nil
	}

	entries, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("unsupported Claude hook event %q shape %T", event, raw)
	}
	return entries, nil
}

func isManagedHook(hook map[string]any) bool {
	return hook["name"] == managedHookName || hook["llm_quota_marker"] == managedHookMarker
}

func managedHook(cachePath string) map[string]any {
	return map[string]any{
		"name":             managedHookName,
		"llm_quota_marker": managedHookMarker,
		"matcher":          "*",
		"command":          managedHookCommand(cachePath),
	}
}

func managedHookCommand(cachePath string) string {
	return "mkdir -p " + shellQuote(filepath.Dir(cachePath)) + " && cat > " + shellQuote(cachePath)
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
	defer output.Close()

	if _, err := io.Copy(output, input); err != nil {
		return "", err
	}
	return backupPath, nil
}

func writeJSONAtomic(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	contents, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	contents = append(contents, '\n')

	tempFile, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*.tmp")
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

	return os.Rename(tempPath, path)
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
