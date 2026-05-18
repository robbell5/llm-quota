package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/rob/llm-quota/internal/install"
	"github.com/rob/llm-quota/internal/tui"
)

type appStreams struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type appDeps struct {
	Paths                    func() (install.ClaudeHookPaths, error)
	ClaudeHookInstalled      func(install.ClaudeHookPaths) (bool, error)
	ClaudeHookDeclined       func(string) (bool, error)
	RecordClaudeHookDeclined func(string) error
	InstallClaudeHook        func(install.ClaudeHookPaths) (install.InstallResult, error)
	StartTUI                 func() error
}

func main() {
	os.Exit(run(os.Args[1:], appStreams{}, appDeps{}))
}

func run(args []string, streams appStreams, deps appDeps) int {
	streams = streams.withDefaults()
	deps = deps.withDefaults()

	if len(args) > 0 {
		switch args[0] {
		case "claude-hook-cache-writer":
			return runClaudeHookCacheWriter(args[1:], streams)
		case "install-claude-hook":
			if len(args) > 1 {
				fmt.Fprintf(streams.Stderr, "llm-quota: unknown argument: %s\n", args[1])
				return 2
			}
			return runInstallClaudeHook(streams, deps)
		default:
			fmt.Fprintf(streams.Stderr, "llm-quota: unknown argument: %s\n", args[0])
			return 2
		}
	}

	if code, ok := offerFirstLaunchInstall(streams, deps); ok {
		return code
	}

	if err := deps.StartTUI(); err != nil {
		fmt.Fprintf(streams.Stderr, "llm-quota: %v\n", err)
		return 1
	}
	return 0
}

func runInstallClaudeHook(streams appStreams, deps appDeps) int {
	paths, err := deps.Paths()
	if err != nil {
		fmt.Fprintf(streams.Stderr, "llm-quota: %v\n", err)
		return 1
	}
	result, err := deps.InstallClaudeHook(paths)
	if err != nil {
		fmt.Fprintf(streams.Stderr, "llm-quota: %v\n", err)
		return 1
	}
	fmt.Fprintln(streams.Stdout, result.Message)
	if result.BackupPath != "" {
		fmt.Fprintf(streams.Stdout, "backup: %s\n", result.BackupPath)
	}
	return 0
}

func runClaudeHookCacheWriter(args []string, streams appStreams) int {
	if len(args) != 2 || args[0] != "--cache" || args[1] == "" {
		fmt.Fprintln(streams.Stderr, "llm-quota: usage: claude-hook-cache-writer --cache <path>")
		return 2
	}
	if err := install.RunClaudeHookCacheWriter(streams.Stdin, args[1], time.Now()); err != nil {
		fmt.Fprintf(streams.Stderr, "llm-quota: %v\n", err)
		return 1
	}
	return 0
}

func offerFirstLaunchInstall(streams appStreams, deps appDeps) (int, bool) {
	paths, err := deps.Paths()
	if err != nil {
		fmt.Fprintf(streams.Stderr, "llm-quota: %v\n", err)
		return 1, true
	}
	installed, err := deps.ClaudeHookInstalled(paths)
	if err != nil {
		fmt.Fprintf(streams.Stderr, "llm-quota: skipping Claude hook prompt: %v\n", err)
		return 0, false
	}
	if installed {
		return 0, false
	}
	declined, err := deps.ClaudeHookDeclined(paths.StatePath)
	if err != nil {
		fmt.Fprintf(streams.Stderr, "llm-quota: skipping Claude hook prompt: %v\n", err)
		return 0, false
	}
	if declined {
		return 0, false
	}

	fmt.Fprintln(streams.Stdout, "llm-quota can install an app-owned Claude hook to write local quota data.")
	fmt.Fprintln(streams.Stdout, "It preserves unrelated Claude configuration and only updates the llm-quota hook entry.")
	fmt.Fprint(streams.Stdout, "Install llm-quota Claude hook now? [y/N] ")

	answer, err := bufio.NewReader(streams.Stdin).ReadString('\n')
	if err != nil && err != io.EOF {
		fmt.Fprintf(streams.Stderr, "llm-quota: %v\n", err)
		return 1, true
	}
	if isYes(answer) {
		result, err := deps.InstallClaudeHook(paths)
		if err != nil {
			fmt.Fprintf(streams.Stderr, "llm-quota: %v\n", err)
			return 1, true
		}
		fmt.Fprintln(streams.Stdout, result.Message)
		if result.BackupPath != "" {
			fmt.Fprintf(streams.Stdout, "backup: %s\n", result.BackupPath)
		}
		return 0, false
	}

	if err := deps.RecordClaudeHookDeclined(paths.StatePath); err != nil {
		fmt.Fprintf(streams.Stderr, "llm-quota: could not record Claude hook decline: %v\n", err)
	}
	return 0, false
}

func isYes(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}

func (streams appStreams) withDefaults() appStreams {
	if streams.Stdin == nil {
		streams.Stdin = os.Stdin
	}
	if streams.Stdout == nil {
		streams.Stdout = os.Stdout
	}
	if streams.Stderr == nil {
		streams.Stderr = os.Stderr
	}
	return streams
}

func (deps appDeps) withDefaults() appDeps {
	if deps.Paths == nil {
		deps.Paths = defaultClaudeHookPaths
	}
	if deps.ClaudeHookInstalled == nil {
		deps.ClaudeHookInstalled = claudeHookInstalled
	}
	if deps.ClaudeHookDeclined == nil {
		deps.ClaudeHookDeclined = install.ClaudeHookDeclined
	}
	if deps.RecordClaudeHookDeclined == nil {
		deps.RecordClaudeHookDeclined = install.RecordClaudeHookDeclined
	}
	if deps.InstallClaudeHook == nil {
		deps.InstallClaudeHook = install.InstallClaudeHook
	}
	if deps.StartTUI == nil {
		deps.StartTUI = startTUI
	}
	return deps
}

func defaultClaudeHookPaths() (install.ClaudeHookPaths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return install.ClaudeHookPaths{}, err
	}
	executablePath, err := os.Executable()
	if err != nil {
		return install.ClaudeHookPaths{}, err
	}
	return install.ClaudeHookPaths{
		ClaudeConfigPath: filepath.Join(home, ".claude", "settings.json"),
		StatePath:        filepath.Join(home, ".cache", "llm-quota", "state.json"),
		CachePath:        filepath.Join(home, ".cache", "llm-quota", "claude.json"),
		ExecutablePath:   executablePath,
	}, nil
}

func claudeHookInstalled(paths install.ClaudeHookPaths) (bool, error) {
	contents, err := os.ReadFile(paths.ClaudeConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if strings.TrimSpace(string(contents)) == "" {
		return false, nil
	}

	var config map[string]any
	if err := json.Unmarshal(contents, &config); err != nil {
		return false, err
	}
	hooks, ok := config["hooks"].(map[string]any)
	if !ok {
		return false, nil
	}
	entries, ok := hooks["PostToolUse"].([]any)
	if !ok {
		return false, nil
	}
	for _, entry := range entries {
		hook, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		if isCurrentManagedClaudeHook(hook, paths.ExecutablePath, paths.CachePath) {
			return true, nil
		}
	}
	return false, nil
}

func isCurrentManagedClaudeHook(hook map[string]any, executablePath string, cachePath string) bool {
	if hook["llm_quota_marker"] != "llm-quota" {
		return false
	}
	if hook["matcher"] != "*" {
		return false
	}
	nested, ok := hook["hooks"].([]any)
	if !ok || len(nested) != 1 {
		return false
	}
	commandHook, ok := nested[0].(map[string]any)
	if !ok || commandHook["type"] != "command" {
		return false
	}
	command, ok := commandHook["command"].(string)
	if !ok {
		return false
	}
	return command == install.ManagedHookCommand(executablePath, cachePath)
}

func startTUI() error {
	program := tea.NewProgram(tui.NewModel())
	_, err := program.Run()
	return err
}
