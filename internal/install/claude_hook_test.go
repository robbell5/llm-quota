package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/rob/llm-quota/internal/sources"
)

func TestInstallClaudeHookPreservesUnrelatedHooksAndCreatesBackupOnlyOnChange(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "settings.json")
	statePath := filepath.Join(tempDir, "state.json")
	cachePath := filepath.Join(tempDir, "quota cache", "claude cache.json")

	unrelatedHook := map[string]any{
		"name":    "user-statusline",
		"matcher": "*",
		"command": "echo keep-me",
	}
	existingConfig := map[string]any{
		"theme": "dark",
		"hooks": map[string]any{
			"PostToolUse": []any{unrelatedHook},
		},
	}
	writeJSON(t, configPath, existingConfig)

	result, err := InstallClaudeHook(ClaudeHookPaths{
		ClaudeConfigPath: configPath,
		StatePath:        statePath,
		CachePath:        cachePath,
	})
	if err != nil {
		t.Fatalf("InstallClaudeHook returned error: %v", err)
	}
	if !result.Changed {
		t.Fatalf("expected first install to change config")
	}
	if result.BackupPath == "" {
		t.Fatalf("expected backup path on changed write")
	}
	if _, err := os.Stat(result.BackupPath); err != nil {
		t.Fatalf("expected backup file to exist: %v", err)
	}

	updated := readJSONMap(t, configPath)
	postToolUse := hookEntries(t, updated, "PostToolUse")
	if len(postToolUse) != 2 {
		t.Fatalf("expected unrelated hook plus llm-quota hook, got %#v", postToolUse)
	}
	assertContainsHook(t, postToolUse, unrelatedHook)
	managed := findManagedHook(t, postToolUse)
	if managed["name"] != "llm-quota" {
		t.Fatalf("managed hook name = %#v, want llm-quota", managed["name"])
	}
	if managed["llm_quota_marker"] != "llm-quota" {
		t.Fatalf("managed hook marker = %#v, want llm-quota", managed["llm_quota_marker"])
	}
	assertManagedCommandHookShape(t, managed, cachePath)

	second, err := InstallClaudeHook(ClaudeHookPaths{
		ClaudeConfigPath: configPath,
		StatePath:        statePath,
		CachePath:        cachePath,
	})
	if err != nil {
		t.Fatalf("second InstallClaudeHook returned error: %v", err)
	}
	if second.Changed {
		t.Fatalf("expected idempotent second install to report unchanged")
	}
	if second.BackupPath != "" {
		t.Fatalf("expected no backup when unchanged, got %q", second.BackupPath)
	}
}

func TestInstallClaudeHookUpdatesOnlyExplicitlyManagedLLMQuotaEntry(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "settings.json")
	cachePath := filepath.Join(tempDir, "quota cache", "claude cache.json")

	looksSimilarButUnowned := map[string]any{
		"name":    "quota-helper",
		"matcher": "*",
		"command": "echo writes quota but is not app-owned",
	}
	managedOld := map[string]any{
		"name":             "llm-quota",
		"llm_quota_marker": "llm-quota",
		"matcher":          "*",
		"command":          "echo old-cache-path",
	}
	writeJSON(t, configPath, map[string]any{
		"hooks": map[string]any{
			"PostToolUse": []any{looksSimilarButUnowned, managedOld},
		},
	})

	result, err := InstallClaudeHook(ClaudeHookPaths{
		ClaudeConfigPath: configPath,
		StatePath:        filepath.Join(tempDir, "state.json"),
		CachePath:        cachePath,
	})
	if err != nil {
		t.Fatalf("InstallClaudeHook returned error: %v", err)
	}
	if !result.Changed {
		t.Fatalf("expected changed result when managed llm-quota entry is updated")
	}

	postToolUse := hookEntries(t, readJSONMap(t, configPath), "PostToolUse")
	if len(postToolUse) != 2 {
		t.Fatalf("expected unowned entry and updated managed entry, got %#v", postToolUse)
	}
	assertContainsHook(t, postToolUse, looksSimilarButUnowned)
	managed := findManagedHook(t, postToolUse)
	assertManagedCommandHookShape(t, managed, cachePath)
}

func TestRunClaudeHookCacheWriterWritesReaderCompatibleCache(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "quota cache", "claude cache.json")
	writtenAt := int64(1778930000)
	input := strings.NewReader(`{"rate_limits":{"five_hour":{"used_percentage":42.3,"resets_at":1778942485},"seven_day":{"used_percentage":85.7,"resets_at":1779382265}}}`)

	if err := RunClaudeHookCacheWriter(input, cachePath, time.Unix(writtenAt, 0)); err != nil {
		t.Fatalf("RunClaudeHookCacheWriter returned error: %v", err)
	}

	contents, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("read generated cache: %v", err)
	}
	var cache map[string]any
	if err := json.Unmarshal(contents, &cache); err != nil {
		t.Fatalf("unmarshal generated cache: %v", err)
	}
	if got, want := sortedKeys(cache), []string{"five_hour", "seven_day", "written_at"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("generated cache top-level keys = %v, want %v; cache=%s", got, want, contents)
	}
	writtenAtValue, ok := cache["written_at"].(float64)
	if !ok {
		t.Fatalf("written_at has type %T, want number", cache["written_at"])
	}
	if got := int64(writtenAtValue); got != writtenAt {
		t.Fatalf("written_at = %d, want %d", got, writtenAt)
	}

	windows, err := sources.NewClaudeReader(cachePath).Fetch(time.Unix(writtenAt, 0))
	if err != nil {
		t.Fatalf("ClaudeReader could not read generated cache: %v", err)
	}
	if len(windows) != 2 {
		t.Fatalf("ClaudeReader returned %d windows, want 2: %#v", len(windows), windows)
	}
	if windows[0].Product != sources.ProductClaude || windows[1].Product != sources.ProductClaude {
		t.Fatalf("ClaudeReader returned non-Claude windows: %#v", windows)
	}
}

func TestRunClaudeHookCacheWriterAcceptsPayloadRateLimits(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "claude.json")
	writtenAt := int64(1778930000)
	input := strings.NewReader(`{"payload":{"rate_limits":{"five_hour":{"used_percentage":42.3,"resets_at":1778942485},"seven_day":{"used_percentage":85.7,"resets_at":1779382265}}}}`)

	if err := RunClaudeHookCacheWriter(input, cachePath, time.Unix(writtenAt, 0)); err != nil {
		t.Fatalf("RunClaudeHookCacheWriter returned error: %v", err)
	}

	windows, err := sources.NewClaudeReader(cachePath).Fetch(time.Unix(writtenAt, 0))
	if err != nil {
		t.Fatalf("ClaudeReader could not read generated payload cache: %v", err)
	}
	if len(windows) != 2 {
		t.Fatalf("ClaudeReader returned %d windows, want 2: %#v", len(windows), windows)
	}
}

func TestRunClaudeHookCacheWriterRejectsTrailingGarbageWithoutChangingCache(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "quota cache", "claude cache.json")
	original := []byte(`{"five_hour":{"used_percentage":1,"resets_at":1778942485},"seven_day":{"used_percentage":2,"resets_at":1779382265},"written_at":1778930000}` + "\n")
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o700); err != nil {
		t.Fatalf("create cache dir: %v", err)
	}
	if err := os.WriteFile(cachePath, original, 0o600); err != nil {
		t.Fatalf("write original cache: %v", err)
	}

	input := strings.NewReader(`{"rate_limits":{"five_hour":{"used_percentage":42.3,"resets_at":1778942485},"seven_day":{"used_percentage":85.7,"resets_at":1779382265}}} garbage`)
	if err := RunClaudeHookCacheWriter(input, cachePath, time.Unix(1778930000, 0)); err == nil {
		t.Fatalf("RunClaudeHookCacheWriter returned nil error for trailing garbage")
	}

	contents, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("read cache after rejected input: %v", err)
	}
	if string(contents) != string(original) {
		t.Fatalf("cache changed after rejected input; got %s want %s", contents, original)
	}
}

func TestClaudeHookDeclineStateIsRemembered(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "llm-quota-state.json")

	declined, err := ClaudeHookDeclined(statePath)
	if err != nil {
		t.Fatalf("ClaudeHookDeclined before state returned error: %v", err)
	}
	if declined {
		t.Fatalf("expected missing state to mean not declined")
	}

	if err := RecordClaudeHookDeclined(statePath); err != nil {
		t.Fatalf("RecordClaudeHookDeclined returned error: %v", err)
	}

	declined, err = ClaudeHookDeclined(statePath)
	if err != nil {
		t.Fatalf("ClaudeHookDeclined after record returned error: %v", err)
	}
	if !declined {
		t.Fatalf("expected recorded decline to be remembered")
	}
}

func writeJSON(t *testing.T, path string, value any) {
	t.Helper()

	contents, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}

func readJSONMap(t *testing.T, path string) map[string]any {
	t.Helper()

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var value map[string]any
	if err := json.Unmarshal(contents, &value); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}
	return value
}

func hookEntries(t *testing.T, config map[string]any, event string) []any {
	t.Helper()

	hooks, ok := config["hooks"].(map[string]any)
	if !ok {
		t.Fatalf("config hooks missing or wrong type: %#v", config["hooks"])
	}
	entries, ok := hooks[event].([]any)
	if !ok {
		t.Fatalf("hook event %q missing or wrong type: %#v", event, hooks[event])
	}
	return entries
}

func findManagedHook(t *testing.T, hooks []any) map[string]any {
	t.Helper()

	for _, entry := range hooks {
		hook, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		if hook["llm_quota_marker"] == "llm-quota" {
			return hook
		}
	}
	t.Fatalf("managed llm-quota hook not found in %#v", hooks)
	return nil
}

func assertContainsHook(t *testing.T, hooks []any, want map[string]any) {
	t.Helper()

	wantJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal want hook: %v", err)
	}
	for _, entry := range hooks {
		gotJSON, err := json.Marshal(entry)
		if err != nil {
			t.Fatalf("marshal got hook: %v", err)
		}
		if string(gotJSON) == string(wantJSON) {
			return
		}
	}
	t.Fatalf("expected hook %#v in %#v", want, hooks)
}

func assertManagedCommandHookShape(t *testing.T, managed map[string]any, cachePath string) {
	t.Helper()

	if managed["matcher"] != "*" {
		t.Fatalf("managed hook matcher = %#v, want *", managed["matcher"])
	}
	if managed["name"] != "llm-quota" {
		t.Fatalf("managed hook name = %#v, want llm-quota", managed["name"])
	}
	if managed["llm_quota_marker"] != "llm-quota" {
		t.Fatalf("managed hook marker = %#v, want llm-quota", managed["llm_quota_marker"])
	}
	if _, ok := managed["command"]; ok {
		t.Fatalf("managed hook has top-level command field: %#v", managed)
	}

	hooks, ok := managed["hooks"].([]any)
	if !ok {
		t.Fatalf("managed hook nested hooks missing or wrong type: %#v", managed["hooks"])
	}
	if len(hooks) != 1 {
		t.Fatalf("managed hook nested hooks length = %d, want 1: %#v", len(hooks), hooks)
	}
	commandHook, ok := hooks[0].(map[string]any)
	if !ok {
		t.Fatalf("managed hook nested command hook wrong type: %#v", hooks[0])
	}
	if commandHook["type"] != "command" {
		t.Fatalf("nested hook type = %#v, want command", commandHook["type"])
	}
	command, ok := commandHook["command"].(string)
	if !ok || command == "" {
		t.Fatalf("nested hook command = %#v, want command pointing at cache path %q", commandHook["command"], cachePath)
	}
	if !strings.Contains(command, "claude-hook-cache-writer --cache") {
		t.Fatalf("nested hook command = %q, want claude-hook-cache-writer --cache", command)
	}
	if !strings.Contains(command, shellQuote(cachePath)) {
		t.Fatalf("nested hook command = %q, want shell-quoted cache path %q", command, shellQuote(cachePath))
	}
}

func sortedKeys(value map[string]any) []string {
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}

func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
