package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/rob/llm-quota/internal/install"
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

func assertEvents(t *testing.T, got, want []string) {
	t.Helper()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("events = %v, want %v", got, want)
	}
}
