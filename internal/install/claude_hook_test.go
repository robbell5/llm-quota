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
	if len(postToolUse) != 1 {
		t.Fatalf("expected unrelated hook to remain, got %#v", postToolUse)
	}
	assertContainsHook(t, postToolUse, unrelatedHook)
	assertManagedStatusLineShape(t, updated, cachePath, "")

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
	if len(postToolUse) != 1 {
		t.Fatalf("expected only unowned entry after removing managed tool hook, got %#v", postToolUse)
	}
	assertContainsHook(t, postToolUse, looksSimilarButUnowned)
}

func TestInstallClaudeHookPreservesMarkerlessLLMQuotaNamedHook(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "settings.json")
	cachePath := filepath.Join(tempDir, "quota cache", "claude cache.json")

	markerlessHook := map[string]any{
		"name":    "llm-quota",
		"matcher": "*",
		"hooks": []any{
			map[string]any{
				"type":    "command",
				"command": "echo user-owned",
			},
		},
	}
	writeJSON(t, configPath, map[string]any{
		"hooks": map[string]any{
			"PostToolUse": []any{markerlessHook},
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
		t.Fatalf("expected changed result when appending managed llm-quota hook")
	}

	postToolUse := hookEntries(t, readJSONMap(t, configPath), "PostToolUse")
	if len(postToolUse) != 1 {
		t.Fatalf("expected markerless hook to remain without managed tool hook, got %#v", postToolUse)
	}
	assertContainsHook(t, postToolUse, markerlessHook)
}

func TestInstallClaudeHookUsesConfiguredExecutablePath(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "settings.json")
	cachePath := filepath.Join(tempDir, "quota cache", "claude cache.json")
	executablePath := filepath.Join(tempDir, "bin dir", "llm-quota")

	if _, err := InstallClaudeHook(ClaudeHookPaths{
		ClaudeConfigPath: configPath,
		StatePath:        filepath.Join(tempDir, "state.json"),
		CachePath:        cachePath,
		ExecutablePath:   executablePath,
	}); err != nil {
		t.Fatalf("InstallClaudeHook returned error: %v", err)
	}

	statusLine := assertManagedStatusLineShape(t, readJSONMap(t, configPath), cachePath, "")
	command, _ := statusLine["command"].(string)
	if !strings.HasPrefix(command, shellQuote(executablePath)+" ") {
		t.Fatalf("statusLine command = %q, want executable path prefix %q", command, shellQuote(executablePath))
	}
}

func TestInstallClaudeHookPreservesSymlinkedConfig(t *testing.T) {
	tempDir := t.TempDir()
	dotfilesDir := filepath.Join(tempDir, "dotfiles")
	if err := os.MkdirAll(dotfilesDir, 0o700); err != nil {
		t.Fatalf("create dotfiles dir: %v", err)
	}
	realConfigPath := filepath.Join(dotfilesDir, "settings.json")
	writeJSON(t, realConfigPath, map[string]any{"theme": "dark"})

	configPath := filepath.Join(tempDir, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	if err := os.Symlink(realConfigPath, configPath); err != nil {
		t.Fatalf("create settings symlink: %v", err)
	}

	result, err := InstallClaudeHook(ClaudeHookPaths{
		ClaudeConfigPath: configPath,
		StatePath:        filepath.Join(tempDir, "state.json"),
		CachePath:        filepath.Join(tempDir, "claude.json"),
	})
	if err != nil {
		t.Fatalf("InstallClaudeHook returned error: %v", err)
	}
	if !result.Changed {
		t.Fatalf("expected symlinked config install to report changed")
	}

	info, err := os.Lstat(configPath)
	if err != nil {
		t.Fatalf("lstat config symlink: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("config path mode = %v, want symlink preserved", info.Mode())
	}
	linkTarget, err := os.Readlink(configPath)
	if err != nil {
		t.Fatalf("read config symlink: %v", err)
	}
	if linkTarget != realConfigPath {
		t.Fatalf("config symlink target = %q, want %q", linkTarget, realConfigPath)
	}

	assertManagedStatusLineShape(t, readJSONMap(t, realConfigPath), filepath.Join(tempDir, "claude.json"), "")
}

func TestInstallClaudeHookWrapsExistingStatusLineAndRemovesManagedToolHook(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "settings.json")
	cachePath := filepath.Join(tempDir, "claude.json")
	executablePath := filepath.Join(tempDir, "llm-quota")
	existingStatusLine := filepath.Join(tempDir, "statusline.sh")

	writeJSON(t, configPath, map[string]any{
		"statusLine": map[string]any{
			"type":    "command",
			"command": existingStatusLine,
		},
		"hooks": map[string]any{
			"PostToolUse": []any{
				map[string]any{
					"name":             "llm-quota",
					"llm_quota_marker": "llm-quota",
					"matcher":          "*",
					"hooks": []any{
						map[string]any{"type": "command", "command": "old broken hook"},
					},
				},
				map[string]any{"name": "user-hook", "matcher": "Read"},
			},
		},
	})

	if _, err := InstallClaudeHook(ClaudeHookPaths{
		ClaudeConfigPath: configPath,
		StatePath:        filepath.Join(tempDir, "state.json"),
		CachePath:        cachePath,
		ExecutablePath:   executablePath,
	}); err != nil {
		t.Fatalf("InstallClaudeHook returned error: %v", err)
	}

	updated := readJSONMap(t, configPath)
	statusLine, ok := updated["statusLine"].(map[string]any)
	if !ok {
		t.Fatalf("statusLine missing or wrong type: %#v", updated["statusLine"])
	}
	command, ok := statusLine["command"].(string)
	if !ok {
		t.Fatalf("statusLine command missing or wrong type: %#v", statusLine["command"])
	}
	for _, want := range []string{shellQuote(executablePath), "claude-statusline-cache-writer --cache", shellQuote(cachePath), "--passthrough", shellQuote(existingStatusLine)} {
		if !strings.Contains(command, want) {
			t.Fatalf("statusLine command = %q, want to contain %q", command, want)
		}
	}

	postToolUse := hookEntries(t, updated, "PostToolUse")
	if len(postToolUse) != 1 {
		t.Fatalf("expected only unrelated PostToolUse hook to remain, got %#v", postToolUse)
	}
	assertContainsHook(t, postToolUse, map[string]any{"name": "user-hook", "matcher": "Read"})
}

func TestUninstallClaudeHookRestoresWrappedStatusLine(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "settings.json")
	statusLineCommand := filepath.Join(tempDir, "statusline.sh")
	writeJSON(t, configPath, map[string]any{
		"theme": "dark",
		"statusLine": map[string]any{
			"type":                  "command",
			"command":               ManagedStatusLineCommand(filepath.Join(tempDir, "llm-quota"), filepath.Join(tempDir, "claude.json"), statusLineCommand),
			"llm_quota_marker":      "llm-quota",
			"llm_quota_passthrough": statusLineCommand,
		},
	})

	result, err := UninstallClaudeHook(ClaudeHookPaths{ClaudeConfigPath: configPath})
	if err != nil {
		t.Fatalf("UninstallClaudeHook returned error: %v", err)
	}
	if !result.Changed {
		t.Fatalf("expected uninstall to change managed statusLine")
	}
	if result.BackupPath == "" {
		t.Fatalf("expected backup path on changed uninstall")
	}
	if _, err := os.Stat(result.BackupPath); err != nil {
		t.Fatalf("expected backup file to exist: %v", err)
	}
	if result.Message != "uninstalled llm-quota Claude hook" {
		t.Fatalf("message = %q, want uninstall confirmation", result.Message)
	}

	updated := readJSONMap(t, configPath)
	statusLine, ok := updated["statusLine"].(map[string]any)
	if !ok {
		t.Fatalf("statusLine missing or wrong type: %#v", updated["statusLine"])
	}
	if got, want := statusLine["type"], "command"; got != want {
		t.Fatalf("statusLine type = %#v, want %q", got, want)
	}
	if got := statusLine["command"]; got != statusLineCommand {
		t.Fatalf("statusLine command = %#v, want restored passthrough %q", got, statusLineCommand)
	}
	if _, ok := statusLine["llm_quota_marker"]; ok {
		t.Fatalf("restored statusLine should not retain llm-quota marker: %#v", statusLine)
	}
}

func TestUninstallClaudeHookRestoresFullOriginalStatusLine(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "settings.json")
	originalStatusLine := map[string]any{
		"type":    "command",
		"command": filepath.Join(tempDir, "statusline.sh"),
		"padding": true,
		"theme":   "compact",
	}
	writeJSON(t, configPath, map[string]any{
		"statusLine": originalStatusLine,
	})

	if _, err := InstallClaudeHook(ClaudeHookPaths{
		ClaudeConfigPath: configPath,
		CachePath:        filepath.Join(tempDir, "claude.json"),
		ExecutablePath:   filepath.Join(tempDir, "llm-quota"),
	}); err != nil {
		t.Fatalf("InstallClaudeHook returned error: %v", err)
	}

	result, err := UninstallClaudeHook(ClaudeHookPaths{ClaudeConfigPath: configPath})
	if err != nil {
		t.Fatalf("UninstallClaudeHook returned error: %v", err)
	}
	if !result.Changed {
		t.Fatalf("expected uninstall to restore original statusLine")
	}

	updated := readJSONMap(t, configPath)
	if got := updated["statusLine"]; !reflect.DeepEqual(got, originalStatusLine) {
		t.Fatalf("statusLine = %#v, want full original %#v", got, originalStatusLine)
	}
}

func TestUninstallClaudeHookRemovesManagedStatusLineWithoutPassthrough(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "settings.json")
	writeJSON(t, configPath, map[string]any{
		"statusLine": map[string]any{
			"type":                  "command",
			"command":               ManagedStatusLineCommand("", filepath.Join(tempDir, "claude.json"), ""),
			"llm_quota_marker":      "llm-quota",
			"llm_quota_passthrough": "",
		},
	})

	result, err := UninstallClaudeHook(ClaudeHookPaths{ClaudeConfigPath: configPath})
	if err != nil {
		t.Fatalf("UninstallClaudeHook returned error: %v", err)
	}
	if !result.Changed {
		t.Fatalf("expected uninstall to remove managed statusLine")
	}
	updated := readJSONMap(t, configPath)
	if _, ok := updated["statusLine"]; ok {
		t.Fatalf("managed statusLine without passthrough should be removed, got %#v", updated["statusLine"])
	}
}

func TestUninstallClaudeHookRemovesOldManagedToolHookAndPreservesUnrelatedHooks(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "settings.json")
	unrelatedHook := map[string]any{"name": "user-hook", "matcher": "Read"}
	writeJSON(t, configPath, map[string]any{
		"hooks": map[string]any{
			"PostToolUse": []any{
				map[string]any{
					"name":             "llm-quota",
					"llm_quota_marker": "llm-quota",
					"matcher":          "*",
					"hooks": []any{
						map[string]any{"type": "command", "command": "old managed hook"},
					},
				},
				unrelatedHook,
			},
		},
	})

	result, err := UninstallClaudeHook(ClaudeHookPaths{ClaudeConfigPath: configPath})
	if err != nil {
		t.Fatalf("UninstallClaudeHook returned error: %v", err)
	}
	if !result.Changed {
		t.Fatalf("expected uninstall to remove old managed tool hook")
	}
	postToolUse := hookEntries(t, readJSONMap(t, configPath), "PostToolUse")
	if len(postToolUse) != 1 {
		t.Fatalf("expected only unrelated hook to remain, got %#v", postToolUse)
	}
	assertContainsHook(t, postToolUse, unrelatedHook)
}

func TestUninstallClaudeHookLeavesUnmanagedConfigUnchanged(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "settings.json")
	original := map[string]any{
		"statusLine": map[string]any{"type": "command", "command": "statusline.sh"},
		"hooks": map[string]any{
			"PostToolUse": []any{map[string]any{"name": "llm-quota", "matcher": "*"}},
		},
	}
	writeJSON(t, configPath, original)

	result, err := UninstallClaudeHook(ClaudeHookPaths{ClaudeConfigPath: configPath})
	if err != nil {
		t.Fatalf("UninstallClaudeHook returned error: %v", err)
	}
	if result.Changed {
		t.Fatalf("expected unmanaged config to remain unchanged")
	}
	if result.BackupPath != "" {
		t.Fatalf("expected no backup for unchanged uninstall, got %q", result.BackupPath)
	}
	if result.Message != "llm-quota Claude hook is not installed" {
		t.Fatalf("message = %q, want not-installed message", result.Message)
	}
	if got := readJSONMap(t, configPath); !reflect.DeepEqual(got, original) {
		t.Fatalf("config changed unexpectedly: got %#v want %#v", got, original)
	}
}

func TestUninstallClaudeHookPreservesSymlinkedConfig(t *testing.T) {
	tempDir := t.TempDir()
	dotfilesDir := filepath.Join(tempDir, "dotfiles")
	if err := os.MkdirAll(dotfilesDir, 0o700); err != nil {
		t.Fatalf("create dotfiles dir: %v", err)
	}
	realConfigPath := filepath.Join(dotfilesDir, "settings.json")
	writeJSON(t, realConfigPath, map[string]any{
		"statusLine": map[string]any{
			"type":                  "command",
			"command":               ManagedStatusLineCommand("", filepath.Join(tempDir, "claude.json"), ""),
			"llm_quota_marker":      "llm-quota",
			"llm_quota_passthrough": "",
		},
	})

	configPath := filepath.Join(tempDir, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	if err := os.Symlink(realConfigPath, configPath); err != nil {
		t.Fatalf("create settings symlink: %v", err)
	}

	result, err := UninstallClaudeHook(ClaudeHookPaths{ClaudeConfigPath: configPath})
	if err != nil {
		t.Fatalf("UninstallClaudeHook returned error: %v", err)
	}
	if !result.Changed {
		t.Fatalf("expected symlinked config uninstall to report changed")
	}
	info, err := os.Lstat(configPath)
	if err != nil {
		t.Fatalf("lstat config symlink: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("config path mode = %v, want symlink preserved", info.Mode())
	}
	if _, ok := readJSONMap(t, realConfigPath)["statusLine"]; ok {
		t.Fatalf("managed statusLine should be removed from symlink target")
	}
}

func TestRunClaudeStatusLineCacheWriterWritesCacheAndPassesThrough(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "claude.json")
	passthroughPath := filepath.Join(tempDir, "statusline.sh")
	if err := os.WriteFile(passthroughPath, []byte("#!/bin/sh\ncat >/dev/null\nprintf 'original statusline'\n"), 0o700); err != nil {
		t.Fatalf("write passthrough script: %v", err)
	}
	input := strings.NewReader(`{"rate_limits":{"five_hour":{"used_percentage":42.3,"resets_at":1778942485},"seven_day":{"used_percentage":85.7,"resets_at":1779382265}}}`)
	var stdout, stderr strings.Builder

	if err := RunClaudeStatusLineCacheWriter(input, &stdout, &stderr, cachePath, passthroughPath, time.Unix(1778930000, 0)); err != nil {
		t.Fatalf("RunClaudeStatusLineCacheWriter returned error: %v", err)
	}
	if stdout.String() != "original statusline" {
		t.Fatalf("stdout = %q, want passthrough output", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	windows, err := sources.NewClaudeReader(cachePath).Fetch(time.Unix(1778930000, 0))
	if err != nil {
		t.Fatalf("ClaudeReader could not read generated statusline cache: %v", err)
	}
	if len(windows) != 2 {
		t.Fatalf("ClaudeReader returned %d windows, want 2: %#v", len(windows), windows)
	}
}

func TestRunClaudeStatusLineCacheWriterRunsPassthroughWhenCacheInputIsInvalid(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "claude.json")
	passthroughPath := filepath.Join(tempDir, "statusline.sh")
	if err := os.WriteFile(passthroughPath, []byte("#!/bin/sh\ncat >/dev/null\nprintf 'original statusline'\n"), 0o700); err != nil {
		t.Fatalf("write passthrough script: %v", err)
	}
	var stdout, stderr strings.Builder

	if err := RunClaudeStatusLineCacheWriter(strings.NewReader(`not json`), &stdout, &stderr, cachePath, passthroughPath, time.Unix(1778930000, 0)); err != nil {
		t.Fatalf("RunClaudeStatusLineCacheWriter returned error: %v", err)
	}
	if stdout.String() != "original statusline" {
		t.Fatalf("stdout = %q, want passthrough output", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		t.Fatalf("cache path stat error = %v, want not exist", err)
	}
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

func assertManagedStatusLineShape(t *testing.T, config map[string]any, cachePath string, passthrough string) map[string]any {
	t.Helper()

	statusLine, ok := config["statusLine"].(map[string]any)
	if !ok {
		t.Fatalf("statusLine missing or wrong type: %#v", config["statusLine"])
	}
	if statusLine["type"] != "command" {
		t.Fatalf("statusLine type = %#v, want command", statusLine["type"])
	}
	if statusLine["llm_quota_marker"] != "llm-quota" {
		t.Fatalf("statusLine marker = %#v, want llm-quota", statusLine["llm_quota_marker"])
	}
	if statusLine["llm_quota_passthrough"] != passthrough {
		t.Fatalf("statusLine passthrough = %#v, want %q", statusLine["llm_quota_passthrough"], passthrough)
	}
	command, ok := statusLine["command"].(string)
	if !ok || command == "" {
		t.Fatalf("statusLine command = %#v, want non-empty command", statusLine["command"])
	}
	if !strings.Contains(command, "claude-statusline-cache-writer --cache") {
		t.Fatalf("statusLine command = %q, want claude-statusline-cache-writer --cache", command)
	}
	if !strings.Contains(command, shellQuote(cachePath)) {
		t.Fatalf("statusLine command = %q, want shell-quoted cache path %q", command, shellQuote(cachePath))
	}
	if passthrough != "" && !strings.Contains(command, shellQuote(passthrough)) {
		t.Fatalf("statusLine command = %q, want shell-quoted passthrough %q", command, shellQuote(passthrough))
	}
	return statusLine
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
