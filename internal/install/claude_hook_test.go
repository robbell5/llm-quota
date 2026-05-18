package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallClaudeHookPreservesUnrelatedHooksAndCreatesBackupOnlyOnChange(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "settings.json")
	statePath := filepath.Join(tempDir, "state.json")
	cachePath := filepath.Join(tempDir, "claude.json")

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
	if command, ok := managed["command"].(string); !ok || command == "" || !contains(command, cachePath) {
		t.Fatalf("managed hook command = %#v, want command pointing at cache path %q", managed["command"], cachePath)
	}

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
	cachePath := filepath.Join(tempDir, "claude.json")

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
	if command, ok := managed["command"].(string); !ok || !contains(command, cachePath) {
		t.Fatalf("managed hook command = %#v, want command pointing at cache path %q", managed["command"], cachePath)
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
