package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/robbell5/llm-quota/internal/install"
	"github.com/robbell5/llm-quota/internal/sources"
	"github.com/robbell5/llm-quota/internal/tui"
)

func TestRunInstallClaudeHookCommandInstallsWithoutStartingTUI(t *testing.T) {
	var stdout, stderr bytes.Buffer
	var installed bool
	var tuiStarted bool

	code := run([]string{"install-claude-hook"}, appStreams{
		Stdout: &stdout,
		Stderr: &stderr,
	}, appDeps{
		InstallClaudeHook: func(paths install.ClaudeHookPaths) (install.InstallResult, error) {
			installed = true
			return install.InstallResult{Changed: true, Message: "installed llm-quota Claude hook"}, nil
		},
		StartTUI: func(model tui.Model) error {
			tuiStarted = true
			return nil
		},
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if !installed {
		t.Fatal("expected install-claude-hook to call installer")
	}
	if tuiStarted {
		t.Fatal("install-claude-hook should not start the TUI")
	}
	if !strings.Contains(stdout.String(), "installed llm-quota Claude hook") {
		t.Fatalf("expected installer result message on stdout, got %q", stdout.String())
	}
}

func TestRunUninstallClaudeHookCommandUninstallsWithoutStartingTUI(t *testing.T) {
	var stdout, stderr bytes.Buffer
	var uninstalled bool
	var tuiStarted bool
	backupPath := filepath.Join(t.TempDir(), "settings.json.llm-quota-backup")

	code := run([]string{"uninstall-claude-hook"}, appStreams{
		Stdout: &stdout,
		Stderr: &stderr,
	}, appDeps{
		UninstallClaudeHook: func(paths install.ClaudeHookPaths) (install.InstallResult, error) {
			uninstalled = true
			return install.InstallResult{Changed: true, BackupPath: backupPath, Message: "uninstalled llm-quota Claude hook"}, nil
		},
		StartTUI: func(model tui.Model) error {
			tuiStarted = true
			return nil
		},
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if !uninstalled {
		t.Fatal("expected uninstall-claude-hook to call uninstaller")
	}
	if tuiStarted {
		t.Fatal("uninstall-claude-hook should not start the TUI")
	}
	if !strings.Contains(stdout.String(), "uninstalled llm-quota Claude hook") {
		t.Fatalf("expected uninstaller result message on stdout, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "backup: "+backupPath) {
		t.Fatalf("expected backup path on stdout, got %q", stdout.String())
	}
}

func TestRunUninstallClaudeHookRejectsExtraArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	var uninstalled bool
	var tuiStarted bool

	code := run([]string{"uninstall-claude-hook", "extra"}, appStreams{
		Stdout: &stdout,
		Stderr: &stderr,
	}, appDeps{
		UninstallClaudeHook: func(paths install.ClaudeHookPaths) (install.InstallResult, error) {
			uninstalled = true
			return install.InstallResult{Changed: true, Message: "uninstalled llm-quota Claude hook"}, nil
		},
		StartTUI: func(model tui.Model) error {
			tuiStarted = true
			return nil
		},
	})

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if got, want := stderr.String(), "llm-quota: unknown argument: extra\n"; got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if uninstalled {
		t.Fatal("invalid uninstall-claude-hook args should not call uninstaller")
	}
	if tuiStarted {
		t.Fatal("invalid uninstall-claude-hook args should not start the TUI")
	}
}

func TestRunFirstLaunchDeclineRecordsDeclineBeforeStartingTUI(t *testing.T) {
	var stdout, stderr bytes.Buffer
	var events []string

	code := run(nil, appStreams{
		Stdin:  strings.NewReader("n\n"),
		Stdout: &stdout,
		Stderr: &stderr,
	}, appDeps{
		CodexSessionsRoot: func() (string, error) {
			return filepath.Join(t.TempDir(), ".codex", "sessions"), nil
		},
		ClaudeHookInstalled: func(paths install.ClaudeHookPaths) (bool, error) {
			return false, nil
		},
		ClaudeHookDeclined: func(statePath string) (bool, error) {
			return false, nil
		},
		RecordClaudeHookDeclined: func(statePath string) error {
			events = append(events, "decline")
			return nil
		},
		StartTUI: func(model tui.Model) error {
			events = append(events, "tui")
			return nil
		},
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Install llm-quota Claude hook now? [y/N]") {
		t.Fatalf("expected first-launch consent prompt, got %q", stdout.String())
	}
	if strings.Contains(stdout.String(), "installed llm-quota Claude hook") {
		t.Fatalf("decline path should not install hook, got stdout %q", stdout.String())
	}
	assertEvents(t, events, []string{"decline", "tui"})
}

func TestRunFirstLaunchAcceptInstallsBeforeStartingTUI(t *testing.T) {
	var stdout, stderr bytes.Buffer
	var events []string

	code := run(nil, appStreams{
		Stdin:  strings.NewReader("yes\n"),
		Stdout: &stdout,
		Stderr: &stderr,
	}, appDeps{
		CodexSessionsRoot: func() (string, error) {
			return filepath.Join(t.TempDir(), ".codex", "sessions"), nil
		},
		ClaudeHookInstalled: func(paths install.ClaudeHookPaths) (bool, error) {
			return false, nil
		},
		ClaudeHookDeclined: func(statePath string) (bool, error) {
			return false, nil
		},
		InstallClaudeHook: func(paths install.ClaudeHookPaths) (install.InstallResult, error) {
			events = append(events, "install")
			return install.InstallResult{Changed: true, Message: "installed llm-quota Claude hook"}, nil
		},
		StartTUI: func(model tui.Model) error {
			events = append(events, "tui")
			return nil
		},
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Install llm-quota Claude hook now? [y/N]") {
		t.Fatalf("expected first-launch consent prompt, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "installed llm-quota Claude hook") {
		t.Fatalf("expected installer result message, got %q", stdout.String())
	}
	assertEvents(t, events, []string{"install", "tui"})
}

func TestRunFirstLaunchUpgradesOldManagedHookBeforeStartingTUI(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "settings.json")
	statePath := filepath.Join(tempDir, "state.json")
	cachePath := filepath.Join(tempDir, "claude.json")
	if err := os.WriteFile(configPath, []byte(`{"hooks":{"PostToolUse":[{"name":"llm-quota","llm_quota_marker":"llm-quota","matcher":"*","command":"cat > old-cache.json"}]}}`), 0o600); err != nil {
		t.Fatalf("write old config: %v", err)
	}

	var stdout, stderr bytes.Buffer
	var events []string

	code := run(nil, appStreams{
		Stdin:  strings.NewReader("yes\n"),
		Stdout: &stdout,
		Stderr: &stderr,
	}, appDeps{
		CodexSessionsRoot: func() (string, error) {
			return filepath.Join(tempDir, ".codex", "sessions"), nil
		},
		Paths: func() (install.ClaudeHookPaths, error) {
			return install.ClaudeHookPaths{
				ClaudeConfigPath: configPath,
				StatePath:        statePath,
				CachePath:        cachePath,
			}, nil
		},
		ClaudeHookDeclined: func(statePath string) (bool, error) {
			return false, nil
		},
		InstallClaudeHook: func(paths install.ClaudeHookPaths) (install.InstallResult, error) {
			events = append(events, "install")
			return install.InstallResult{Changed: true, Message: "installed llm-quota Claude hook"}, nil
		},
		StartTUI: func(model tui.Model) error {
			events = append(events, "tui")
			return nil
		},
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Install llm-quota Claude hook now? [y/N]") {
		t.Fatalf("expected first-launch consent prompt for old managed hook, got %q", stdout.String())
	}
	assertEvents(t, events, []string{"install", "tui"})
}

func TestClaudeHookInstalledIgnoresMarkerlessLLMQuotaNamedHook(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "settings.json")
	cachePath := filepath.Join(tempDir, "claude.json")
	config := `{"hooks":{"PostToolUse":[{"name":"llm-quota","matcher":"*","hooks":[{"type":"command","command":"llm-quota claude-hook-cache-writer --cache ` + cachePath + `"}]}]}}`
	if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	installed, err := claudeHookInstalled(install.ClaudeHookPaths{
		ClaudeConfigPath: configPath,
		CachePath:        cachePath,
	})
	if err != nil {
		t.Fatalf("claudeHookInstalled returned error: %v", err)
	}
	if installed {
		t.Fatalf("markerless llm-quota hook should not count as installed")
	}
}

func TestClaudeHookInstalledMatchesQuotedCachePath(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "settings.json")
	cachePath := filepath.Join(tempDir, "rob's cache.json")
	command := install.ManagedHookCommand("", cachePath)
	if !strings.Contains(command, "'\\''") {
		t.Fatalf("test setup expected shell-quoted apostrophe in command, got %q", command)
	}
	config, err := json.Marshal(map[string]any{
		"hooks": map[string]any{
			"PostToolUse": []any{
				map[string]any{
					"name":             "llm-quota",
					"llm_quota_marker": "llm-quota",
					"matcher":          "*",
					"hooks": []any{
						map[string]any{"type": "command", "command": command},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(configPath, config, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	installed, err := claudeHookInstalled(install.ClaudeHookPaths{
		ClaudeConfigPath: configPath,
		CachePath:        cachePath,
	})
	if err != nil {
		t.Fatalf("claudeHookInstalled returned error: %v", err)
	}
	if !installed {
		t.Fatalf("shell-quoted cache path should count as installed")
	}
}

func TestClaudeHookInstalledMatchesManagedStatusLine(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "settings.json")
	cachePath := filepath.Join(tempDir, "claude.json")
	executablePath := filepath.Join(tempDir, "llm-quota")
	config, err := json.Marshal(map[string]any{
		"statusLine": map[string]any{
			"type":                  "command",
			"command":               install.ManagedStatusLineCommand(executablePath, cachePath, "statusline.sh"),
			"llm_quota_marker":      "llm-quota",
			"llm_quota_passthrough": "statusline.sh",
		},
	})
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(configPath, config, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	installed, err := claudeHookInstalled(install.ClaudeHookPaths{
		ClaudeConfigPath: configPath,
		CachePath:        cachePath,
		ExecutablePath:   executablePath,
	})
	if err != nil {
		t.Fatalf("claudeHookInstalled returned error: %v", err)
	}
	if !installed {
		t.Fatalf("managed statusline should count as installed")
	}
}

func TestClaudeHookInstalledRejectsWrongStatusLineCachePath(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "settings.json")
	executablePath := filepath.Join(tempDir, "llm-quota")
	config, err := json.Marshal(map[string]any{
		"statusLine": map[string]any{
			"type":                  "command",
			"command":               install.ManagedStatusLineCommand(executablePath, filepath.Join(tempDir, "old.json"), "statusline.sh"),
			"llm_quota_marker":      "llm-quota",
			"llm_quota_passthrough": "statusline.sh",
		},
	})
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if err := os.WriteFile(configPath, config, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	installed, err := claudeHookInstalled(install.ClaudeHookPaths{
		ClaudeConfigPath: configPath,
		CachePath:        filepath.Join(tempDir, "claude.json"),
		ExecutablePath:   executablePath,
	})
	if err != nil {
		t.Fatalf("claudeHookInstalled returned error: %v", err)
	}
	if installed {
		t.Fatalf("managed statusline with wrong cache path should not count as installed")
	}
}

func TestRunUnknownArgumentPreservesErrorAndExitCode(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"bogus"}, appStreams{
		Stdout: &stdout,
		Stderr: &stderr,
	}, appDeps{
		StartTUI: func(model tui.Model) error {
			return errors.New("should not start TUI")
		},
	})

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if got, want := stderr.String(), "llm-quota: unknown argument: bogus\n"; got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunClaudeHookCacheWriterCommandWritesCacheWithoutStartingTUI(t *testing.T) {
	var stdout, stderr bytes.Buffer
	var tuiStarted bool
	cachePath := filepath.Join(t.TempDir(), "quota cache", "claude cache.json")
	stdin := strings.NewReader(`{"rate_limits":{"five_hour":{"used_percentage":42.3,"resets_at":1778942485},"seven_day":{"used_percentage":85.7,"resets_at":1779382265}}}`)

	code := run([]string{"claude-hook-cache-writer", "--cache", cachePath}, appStreams{
		Stdin:  stdin,
		Stdout: &stdout,
		Stderr: &stderr,
	}, appDeps{
		StartTUI: func(model tui.Model) error {
			tuiStarted = true
			return errors.New("should not start TUI")
		},
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if tuiStarted {
		t.Fatal("claude-hook-cache-writer should not start the TUI")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	windows, err := sources.NewClaudeReader(cachePath).Fetch(time.Unix(1778930000, 0))
	if err != nil {
		t.Fatalf("ClaudeReader could not read generated cache: %v", err)
	}
	if len(windows) != 2 {
		t.Fatalf("ClaudeReader returned %d windows, want 2: %#v", len(windows), windows)
	}
}

func TestRunClaudeHookCacheWriterCommandRejectsMissingOrExtraArgs(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{name: "missing cache flag", args: []string{"claude-hook-cache-writer"}},
		{name: "missing cache path", args: []string{"claude-hook-cache-writer", "--cache"}},
		{name: "wrong flag", args: []string{"claude-hook-cache-writer", "--path", "claude.json"}},
		{name: "extra arg", args: []string{"claude-hook-cache-writer", "--cache", "claude.json", "extra"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			var tuiStarted bool
			args := make([]string, len(tc.args))
			copy(args, tc.args)
			for i, arg := range args {
				if arg == "claude.json" {
					args[i] = filepath.Join(t.TempDir(), "quota cache", "claude cache.json")
				}
			}

			code := run(args, appStreams{
				Stdin:  strings.NewReader(`{"rate_limits":{"five_hour":{"used_percentage":42.3,"resets_at":1778942485},"seven_day":{"used_percentage":85.7,"resets_at":1779382265}}}`),
				Stdout: &stdout,
				Stderr: &stderr,
			}, appDeps{
				StartTUI: func(model tui.Model) error {
					tuiStarted = true
					return errors.New("should not start TUI")
				},
			})

			if code != 2 {
				t.Fatalf("exit code = %d, want 2; stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
			if tuiStarted {
				t.Fatal("invalid claude-hook-cache-writer args should not start the TUI")
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
		})
	}
}

func TestRunNoArgStartupConstructsSourceBackedModelWithoutStartingRealTUI(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".claude", "settings.json")
	statePath := filepath.Join(tempDir, ".cache", "llm-quota", "state.json")
	cachePath := filepath.Join(tempDir, ".cache", "llm-quota", "claude.json")
	codexSessions := filepath.Join(tempDir, ".codex", "sessions")

	var stdout, stderr bytes.Buffer
	var captured tui.Model
	var started bool

	code := run(nil, appStreams{
		Stdin:  strings.NewReader(""),
		Stdout: &stdout,
		Stderr: &stderr,
	}, appDeps{
		Paths: func() (install.ClaudeHookPaths, error) {
			return install.ClaudeHookPaths{
				ClaudeConfigPath: configPath,
				StatePath:        statePath,
				CachePath:        cachePath,
				ExecutablePath:   filepath.Join(tempDir, "llm-quota"),
			}, nil
		},
		CodexSessionsRoot: func() (string, error) {
			return codexSessions, nil
		},
		ClaudeHookInstalled: func(paths install.ClaudeHookPaths) (bool, error) {
			return true, nil
		},
		StartTUI: func(model tui.Model) error {
			started = true
			captured = model
			return nil
		},
	})

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if !started {
		t.Fatal("expected no-arg startup to start TUI through injected seam")
	}
	modelDebug := fmt.Sprintf("%#v", captured)
	if !strings.Contains(modelDebug, "sources.ClaudeReader") || !strings.Contains(modelDebug, "sources.CodexReader") {
		t.Fatalf("expected source-backed model, got %s", modelDebug)
	}
	if !strings.Contains(modelDebug, cachePath) {
		t.Fatalf("expected Claude cache path %q in model readers, got %q", cachePath, modelDebug)
	}
	if !strings.Contains(modelDebug, codexSessions) {
		t.Fatalf("expected Codex sessions path %q in model readers, got %q", codexSessions, modelDebug)
	}
	if !strings.Contains(modelDebug, "claudeHookInstalled:true") {
		t.Fatalf("expected source-backed model to remember installed Claude hook, got %q", modelDebug)
	}
}

func TestParseDisplayFlags(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		wantBar  tui.BarStyle
		wantVis  tui.Visibility
		wantHelp bool
		wantErr  bool
	}{
		{"defaults", nil, tui.BarSegmented, tui.VisibilityBoth, false, false},
		{"solid", []string{"--solid-bars"}, tui.BarSolid, tui.VisibilityBoth, false, false},
		{"claude only", []string{"--only=claude"}, tui.BarSegmented, tui.VisibilityClaudeOnly, false, false},
		{"codex only", []string{"--only=codex"}, tui.BarSegmented, tui.VisibilityCodexOnly, false, false},
		{"combined", []string{"--solid-bars", "--only=codex"}, tui.BarSolid, tui.VisibilityCodexOnly, false, false},
		{"help", []string{"--help"}, tui.BarSegmented, tui.VisibilityBoth, true, false},
		{"short help", []string{"-h"}, tui.BarSegmented, tui.VisibilityBoth, true, false},
		{"bad only value", []string{"--only=both"}, tui.BarSegmented, tui.VisibilityBoth, false, true},
		{"unknown flag", []string{"--nope"}, tui.BarSegmented, tui.VisibilityBoth, false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			prefs, help, err := parseDisplayFlags(c.args)
			if (err != nil) != c.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, c.wantErr)
			}
			if c.wantErr {
				return
			}
			if help != c.wantHelp {
				t.Fatalf("help = %v, want %v", help, c.wantHelp)
			}
			if prefs.BarStyle != c.wantBar || prefs.Visibility != c.wantVis {
				t.Fatalf("prefs = %#v, want bar=%v vis=%v", prefs, c.wantBar, c.wantVis)
			}
		})
	}
}

func TestRunHelpAndBadFlagExitCodes(t *testing.T) {
	t.Run("help exits 0 and prints usage", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		code := run([]string{"--help"}, appStreams{Stdin: strings.NewReader(""), Stdout: &stdout, Stderr: &stderr}, appDeps{})
		if code != 0 {
			t.Fatalf("expected exit 0, got %d", code)
		}
		if !strings.Contains(stdout.String(), "Usage:") {
			t.Fatalf("expected usage text, got: %s", stdout.String())
		}
	})

	t.Run("unknown flag exits 2", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		code := run([]string{"--nope"}, appStreams{Stdin: strings.NewReader(""), Stdout: &stdout, Stderr: &stderr}, appDeps{})
		if code != 2 {
			t.Fatalf("expected exit 2, got %d", code)
		}
		if !strings.Contains(stderr.String(), "unknown flag") {
			t.Fatalf("expected 'unknown flag' on stderr, got: %s", stderr.String())
		}
	})

	t.Run("invalid --only value exits 2 with guidance", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		code := run([]string{"--only=both"}, appStreams{Stdin: strings.NewReader(""), Stdout: &stdout, Stderr: &stderr}, appDeps{})
		if code != 2 {
			t.Fatalf("expected exit 2, got %d", code)
		}
		if !strings.Contains(stderr.String(), "invalid --only value") {
			t.Fatalf("expected '--only' guidance on stderr, got: %s", stderr.String())
		}
	})
}

func TestParseDisplayFlagsNoTrend(t *testing.T) {
	prefs, showHelp, err := parseDisplayFlags([]string{"--no-trend"})
	if err != nil || showHelp {
		t.Fatalf("unexpected err=%v showHelp=%v", err, showHelp)
	}
	if !prefs.HideTrend {
		t.Fatalf("--no-trend should set HideTrend")
	}
}

func assertEvents(t *testing.T, got, want []string) {
	t.Helper()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("events = %v, want %v", got, want)
	}
}
