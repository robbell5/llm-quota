# Testing Patterns

**Analysis Date:** 2026-05-21

## Test Framework

**Runner:**
- Go `testing` package from Go 1.26.3.
- Config: `go.mod`; no separate test config file is present.
- Test packages are the same package as implementation (`package sources`, `package tui`, `package install`, `package main`) in `internal/sources/claude_test.go`, `internal/tui/update_test.go`, `internal/install/claude_hook_test.go`, and `cmd/llm-quota/main_test.go`.

**Assertion Library:**
- Standard library only. Assertions use `t.Fatalf`, `t.Fatal`, `errors.As`, `reflect.DeepEqual`, `strings.Contains`, and custom helper functions in test files such as `assertWindows` in `internal/sources/claude_test.go` and `assertRenderedLineWidths` in `internal/tui/view_test.go`.

**Run Commands:**
```bash
go test ./...              # Run all tests; passes on 2026-05-21
go test -race ./...        # Race detector; passes on 2026-05-21
go test ./... -cover       # Coverage view; currently fails in this environment with Go coverage package resolution errors
go vet ./...               # Static checks documented in AGENTS.md
go fmt ./...               # Standard formatting
```

## Test File Organization

**Location:**
- Tests are co-located with the package under test: `cmd/llm-quota/main_test.go`, `internal/install/claude_hook_test.go`, `internal/sources/claude_test.go`, `internal/sources/codex_test.go`, `internal/tui/update_test.go`, `internal/tui/view_test.go`.
- There is no separate `test/`, `fixtures/`, or `__snapshots__/` directory.

**Naming:**
- Test files use `_test.go` suffix and mirror package areas: `claude_test.go` tests `internal/sources/claude.go`, `codex_test.go` tests `internal/sources/codex.go`, `view_test.go` tests `internal/tui/view.go`, `update_test.go` tests `internal/tui/update.go`.
- Test functions use `Test<Subject><Behavior>` names: `TestCodexFetchNewestUsableRollout`, `TestInstallClaudeHookPreservesSymlinkedConfig`, `TestRenderResponsiveQuotaLayouts`, `TestRunNoArgStartupConstructsSourceBackedModelWithoutStartingRealTUI`.

**Structure:**
```text
cmd/llm-quota/
├── main.go
└── main_test.go

internal/sources/
├── claude.go
├── claude_test.go
├── codex.go
└── codex_test.go

internal/tui/
├── update.go
├── update_test.go
├── view.go
└── view_test.go

internal/install/
├── claude_hook.go
└── claude_hook_test.go
```

## Test Structure

**Suite Organization:**
```go
func TestClaudeFetch(t *testing.T) {
	now := time.Unix(1_778_940_000, 0)

	tests := []struct {
		name         string
		writeCache   bool
		cache        string
		wantWindows  []Window
		wantCategory ErrorCategory
	}{
		{name: "missing cache", wantCategory: ErrorMissing},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cachePath := filepath.Join(t.TempDir(), "claude.json")
			windows, err := NewClaudeReader(cachePath).Fetch(now)
			assertSourceError(t, err, tt.wantCategory)
			if len(windows) != 0 {
				t.Fatalf("expected no windows on error, got %#v", windows)
			}
		})
	}
}
```

Pattern source: `internal/sources/claude_test.go`.

**Patterns:**
- Use table-driven tests for parser variants and argument validation: `TestClaudeFetch` in `internal/sources/claude_test.go`, `TestRunClaudeHookCacheWriterCommandRejectsMissingOrExtraArgs` in `cmd/llm-quota/main_test.go`.
- Use named subtests when one behavior has several state transitions: `TestRefresh` in `internal/tui/update_test.go`.
- Use fixed clocks to make reset and stale calculations deterministic: `fixedNow` in `internal/tui/update_test.go`, `WithClock` in `internal/tui/view_test.go`, explicit `time.Unix` values in `internal/sources/claude_test.go`.
- Use `t.TempDir()` for all filesystem fixtures. No tests read or write real `~/.claude`, `~/.codex`, or app cache paths.
- Use helper functions with `t.Helper()` for repeated assertions and fixture writes: `writeRollout` in `internal/sources/codex_test.go`, `writeJSON` and `readJSONMap` in `internal/install/claude_hook_test.go`, `assertWindows` in `internal/sources/claude_test.go`.
- Assert user-visible CLI output and exit codes directly with injected buffers in `cmd/llm-quota/main_test.go`.
- Assert TUI rendering on plain text by stripping ANSI escape sequences with a local regexp in `internal/tui/view_test.go`.

## Mocking

**Framework:** Standard library fakes and dependency injection.

**Patterns:**
```go
type fakeReader struct {
	windows []sources.Window
	err     error
	calls   int
	seenNow []time.Time
}

func (r *fakeReader) Fetch(now time.Time) ([]sources.Window, error) {
	r.calls++
	r.seenNow = append(r.seenNow, now)
	return cloneWindows(r.windows), r.err
}
```

Pattern source: `internal/tui/update_test.go`.

```go
code := run(nil, appStreams{
	Stdin:  strings.NewReader("yes\n"),
	Stdout: &stdout,
	Stderr: &stderr,
}, appDeps{
	InstallClaudeHook: func(paths install.ClaudeHookPaths) (install.InstallResult, error) {
		events = append(events, "install")
		return install.InstallResult{Changed: true, Message: "installed llm-quota Claude hook"}, nil
	},
	StartTUI: func(model tui.Model) error {
		events = append(events, "tui")
		return nil
	},
})
```

Pattern source: `cmd/llm-quota/main_test.go`.

**What to Mock:**
- Mock source readers at the TUI boundary through `SourceReader` and `WithReaders` in `internal/tui/model.go`.
- Mock CLI streams and dependencies through `appStreams` and `appDeps` in `cmd/llm-quota/main.go`.
- Mock time through `WithClock` in `internal/tui/model.go` or explicit `now time.Time` arguments in `internal/sources` and `internal/install`.
- Mock filesystem roots with `t.TempDir()` and path constructors, not global home directories.

**What NOT to Mock:**
- Do not mock JSON parsing, file writes, symlink behavior, or cache compatibility. `internal/install/claude_hook_test.go` and `internal/sources/*.go` tests exercise real files in temporary directories.
- Do not mock Bubble Tea message types. Tests use real `tea.KeyPressMsg`, `tea.WindowSizeMsg`, `tea.BatchMsg`, and `tea.QuitMsg` in `internal/tui/update_test.go`.
- Do not add gomock/testify-style dependencies. Current tests are standard library only.

## Fixtures and Factories

**Test Data:**
```go
func writeRollout(t *testing.T, root, relativePath string, modified time.Time, lines []string) {
	t.Helper()

	path := filepath.Join(root, relativePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("make rollout dir: %v", err)
	}

	contents := ""
	for _, line := range lines {
		contents += line + "\n"
	}

	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write rollout: %v", err)
	}
	if err := os.Chtimes(path, modified, modified); err != nil {
		t.Fatalf("set rollout mtime: %v", err)
	}
}
```

Pattern source: `internal/sources/codex_test.go`.

**Location:**
- Fixtures are inline strings and helper-built temp files in the relevant `_test.go` file.
- JSON payloads are embedded near the tests that use them: Claude cache payloads in `internal/sources/claude_test.go`, Codex JSONL rollout lines in `internal/sources/codex_test.go`, Claude settings maps in `internal/install/claude_hook_test.go`.
- No committed golden files are present. Render tests assert substrings, forbidden strings, and line widths in `internal/tui/view_test.go`.

## Coverage

**Requirements:** None enforced.
- `go test ./...` is the primary correctness gate and passes on 2026-05-21.
- `go test -race ./...` is the concurrency safety gate and passes on 2026-05-21.
- `go test ./... -cover` is the natural coverage command, but it currently fails in this environment with package resolution errors involving `internal/coverage/cfile`, `testmain`, `runtime/coverage`, and `github.com/robbell5/llm-quota/cmd/llm-quota`.

**View Coverage:**
```bash
go test ./... -cover
```

## Test Types

**Unit Tests:**
- Source parser tests cover valid data, malformed JSON, missing files, stale Claude cache data, newest/fallback Codex rollout selection, no usable Codex event, and typed `SourceError` categories in `internal/sources/claude_test.go` and `internal/sources/codex_test.go`.
- TUI update tests cover quit keys, resize state, init command batching, manual refresh coalescing, tick scheduling, refresh result merging, stale marking, and last-known-good preservation in `internal/tui/update_test.go`.
- TUI render tests cover startup placeholders, source-backed rows, responsive widths, threshold progress bars, recovery hints, stale hints, and absence of raw error categories in `internal/tui/view_test.go`.
- Install tests cover statusline install/uninstall, idempotency, backups, symlink preservation, marker ownership, passthrough execution, cache writer compatibility, trailing JSON rejection, and decline state in `internal/install/claude_hook_test.go`.
- CLI tests cover command dispatch, invalid arguments, first-launch install prompt, dependency injection, stderr/stdout behavior, and model construction in `cmd/llm-quota/main_test.go`.

**Integration Tests:**
- Package-level tests exercise real temporary files and real JSON encode/decode flows rather than mocked parsers: `internal/install/claude_hook_test.go`, `internal/sources/claude_test.go`, `internal/sources/codex_test.go`.
- CLI tests in `cmd/llm-quota/main_test.go` integrate `run`, `appDeps`, `internal/install`, `internal/sources`, and `internal/tui` enough to verify command behavior without launching a real terminal UI.

**E2E Tests:**
- Not used. There are no terminal automation, tmux, or snapshot-golden E2E tests.

## Common Patterns

**Async Testing:**
```go
updated, cmd := model.Update(refreshRequestedMsg{})
if cmd == nil {
	t.Fatal("expected refresh command")
}

msg, ok := cmd().(refreshMsg)
if !ok {
	t.Fatalf("expected refreshMsg, got %T", cmd())
}
if claude.calls != 1 || codex.calls != 1 {
	t.Fatalf("expected one fetch per source, got claude=%d codex=%d", claude.calls, codex.calls)
}
```

Pattern source: `internal/tui/update_test.go`.

**Error Testing:**
```go
var sourceErr SourceError
if !errors.As(err, &sourceErr) {
	t.Fatalf("expected SourceError, got %T: %v", err, err)
}

if sourceErr.Category != category {
	t.Fatalf("expected category %q, got %q", category, sourceErr.Category)
}
```

Pattern source: `internal/sources/claude_test.go`.

**Render Testing:**
```go
plain := ansiEscapeRE.ReplaceAllString(rendered, "")
if !strings.Contains(plain, "Claude data 2h old; open Claude") {
	t.Fatalf("expected stale Claude footer hint in source-backed output, got:\n%s", plain)
}
assertRenderedLineWidths(t, rendered, width)
```

Pattern source: `internal/tui/view_test.go`.

**Filesystem Testing:**
```go
tempDir := t.TempDir()
configPath := filepath.Join(tempDir, "settings.json")
writeJSON(t, configPath, existingConfig)

result, err := InstallClaudeHook(ClaudeHookPaths{
	ClaudeConfigPath: configPath,
	StatePath:        filepath.Join(tempDir, "state.json"),
	CachePath:        filepath.Join(tempDir, "quota cache", "claude cache.json"),
})
```

Pattern source: `internal/install/claude_hook_test.go`.

---

*Testing analysis: 2026-05-21*
