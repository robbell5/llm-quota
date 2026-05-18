package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rob/llm-quota/internal/install"
	"github.com/rob/llm-quota/internal/sources"
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
		StartTUI: func() error {
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

func TestRunFirstLaunchDeclineRecordsDeclineBeforeStartingTUI(t *testing.T) {
	var stdout, stderr bytes.Buffer
	var events []string

	code := run(nil, appStreams{
		Stdin:  strings.NewReader("n\n"),
		Stdout: &stdout,
		Stderr: &stderr,
	}, appDeps{
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
		StartTUI: func() error {
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
		StartTUI: func() error {
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
		StartTUI: func() error {
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

func TestRunUnknownArgumentPreservesErrorAndExitCode(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"--help"}, appStreams{
		Stdout: &stdout,
		Stderr: &stderr,
	}, appDeps{
		StartTUI: func() error {
			return errors.New("should not start TUI")
		},
	})

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if got, want := stderr.String(), "llm-quota: unknown argument: --help\n"; got != want {
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
		StartTUI: func() error {
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
				StartTUI: func() error {
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

func assertEvents(t *testing.T, got, want []string) {
	t.Helper()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("events = %v, want %v", got, want)
	}
}
