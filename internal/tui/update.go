package tui

import (
	"errors"
	"time"

	tea "charm.land/bubbletea/v2"
	"golang.org/x/sync/errgroup"

	"github.com/robbell5/llm-quota/internal/sources"
)

type refreshRequestedMsg struct{}

type tickMsg time.Time

type refreshMsg struct {
	results   []sourceRefreshResult
	fetchedAt time.Time
}

type sourceRefreshResult struct {
	product sources.Product
	windows []sources.Window
	err     sources.SourceError
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(requestRefreshCmd(), tickCmd(m.refreshEvery))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			if m.refreshing {
				return m, nil
			}
			return m, requestRefreshCmd()
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case refreshRequestedMsg:
		if m.refreshing {
			return m, nil
		}
		m.refreshing = true
		return m, refreshCmd(m.claudeReader, m.codexReader, m.now)
	case tickMsg:
		return m, tea.Batch(requestRefreshCmd(), tickCmd(m.refreshEvery))
	case refreshMsg:
		m.refreshing = false
		m.mergeRefresh(msg)
		return m, nil
	}

	return m, nil
}

func requestRefreshCmd() tea.Cmd {
	return func() tea.Msg {
		return refreshRequestedMsg{}
	}
}

func tickCmd(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func refreshCmd(claude SourceReader, codex SourceReader, now func() time.Time) tea.Cmd {
	return func() tea.Msg {
		if now == nil {
			now = time.Now
		}
		fetchedAt := now()

		results := []sourceRefreshResult{
			{product: sources.ProductClaude},
			{product: sources.ProductCodex},
		}

		var group errgroup.Group
		group.Go(func() error {
			results[0] = fetchSource(sources.ProductClaude, claude, fetchedAt)
			return nil
		})
		group.Go(func() error {
			results[1] = fetchSource(sources.ProductCodex, codex, fetchedAt)
			return nil
		})
		_ = group.Wait()

		return refreshMsg{results: results, fetchedAt: fetchedAt}
	}
}

func fetchSource(product sources.Product, reader SourceReader, now time.Time) sourceRefreshResult {
	if reader == nil {
		return sourceRefreshResult{
			product: product,
			err: sources.SourceError{
				Source:   product,
				Category: sources.ErrorMissing,
			},
		}
	}

	windows, err := reader.Fetch(now)
	if err != nil {
		return sourceRefreshResult{product: product, err: normalizeSourceError(product, err)}
	}

	return sourceRefreshResult{product: product, windows: windows}
}

func normalizeSourceError(product sources.Product, err error) sources.SourceError {
	var sourceErr sources.SourceError
	if errors.As(err, &sourceErr) {
		if sourceErr.Source == "" {
			sourceErr.Source = product
		}
		return sourceErr
	}

	return sources.SourceError{
		Source:   product,
		Category: sources.ErrorRead,
		Err:      err,
	}
}

func (m *Model) mergeRefresh(msg refreshMsg) {
	if m.windows == nil {
		m.windows = make(map[sources.Product][]sources.Window)
	}
	if m.errors == nil {
		m.errors = make(map[sources.Product]sources.SourceError)
	}

	for _, result := range msg.results {
		if result.err.Category != "" {
			m.errors[result.product] = result.err
			continue
		}

		m.windows[result.product] = markStale(result.windows, msg.fetchedAt, m.staleAfter)
		delete(m.errors, result.product)
	}
}

func markStale(windows []sources.Window, now time.Time, staleAfter time.Duration) []sources.Window {
	marked := make([]sources.Window, len(windows))
	for i, window := range windows {
		age := now.Sub(window.CapturedAt)
		if age < 0 {
			age = 0
		}
		window.StaleAge = age
		window.Stale = age > staleAfter
		marked[i] = window
	}

	return marked
}

func (m Model) View() tea.View {
	v := tea.NewView(render(m))
	v.AltScreen = true
	return v
}
