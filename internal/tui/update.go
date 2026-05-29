package tui

import (
	"errors"
	"math"
	"time"

	tea "charm.land/bubbletea/v2"
	"golang.org/x/sync/errgroup"

	"github.com/robbell5/llm-quota/internal/cost"
	"github.com/robbell5/llm-quota/internal/sources"
	"github.com/robbell5/llm-quota/internal/trend"
)

type refreshRequestedMsg struct{}

type tickMsg time.Time

type refreshMsg struct {
	results   []sourceRefreshResult
	fetchedAt time.Time
	costs     map[sources.Product]map[sources.WindowKind]cost.WindowCost
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
		case "v":
			m.prefs.Visibility = m.prefs.Visibility.next()
			return m, nil
		case "t":
			m.prefs.HideTrend = !m.prefs.HideTrend
			return m, nil
		case "c":
			m.prefs.HideCost = !m.prefs.HideCost
			return m, nil
		case "i":
			m.prefs.Icons = !m.prefs.Icons
			return m, nil
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
		return m, tea.Batch(refreshCmd(m.claudeReader, m.codexReader, m.claudeCost, m.codexCost, m.now), m.ensureAnim())
	case tickMsg:
		return m, tea.Batch(requestRefreshCmd(), tickCmd(m.refreshEvery))
	case refreshMsg:
		m.refreshing = false
		cmds := m.mergeRefresh(msg)
		return m, tea.Batch(cmds...)
	case animTickMsg:
		for i := range m.bars {
			if m.bars[i].target < 0 {
				continue
			}
			m.bars[i].pos, m.bars[i].vel = m.spring.Update(m.bars[i].pos, m.bars[i].vel, m.bars[i].target)
		}
		m.animPhase++
		if m.animating() {
			return m, animTickCmd()
		}
		m.animRunning = false
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

type animTickMsg time.Time

func animTickCmd() tea.Cmd {
	return tea.Tick(time.Second/animFPS, func(t time.Time) tea.Msg {
		return animTickMsg(t)
	})
}

// ensureAnim starts the animation loop if it is not already running.
func (m *Model) ensureAnim() tea.Cmd {
	if m.animRunning {
		return nil
	}
	m.animRunning = true
	return animTickCmd()
}

// animating reports whether anything still needs frames: an unsettled spring, an
// active value-change highlight, an at-risk row (pulse), or an in-flight refresh.
// Assumes a NewModel-constructed model (non-nil now); only reached via Update.
func (m Model) animating() bool {
	if m.refreshing {
		return true
	}
	now := m.now()
	for _, until := range m.highlightUntil {
		if now.Before(until) {
			return true
		}
	}
	if m.anyAtRisk(now) {
		return true
	}
	for _, b := range m.bars {
		if b.target < 0 {
			continue
		}
		if math.Abs(b.pos-b.target) > springSettleEpsilon || math.Abs(b.vel) > springSettleEpsilon {
			return true
		}
	}
	return false
}

func refreshCmd(claude SourceReader, codex SourceReader, claudeCost CostReader, codexCost CostReader, now func() time.Time) tea.Cmd {
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

		costs := computeCosts(claudeCost, codexCost, results, fetchedAt)
		return refreshMsg{results: results, fetchedAt: fetchedAt, costs: costs}
	}
}

// computeCosts prices each product whose fetch produced windows. A nil reader or
// errored/empty fetch yields no entry (the renderer then hides that cluster).
func computeCosts(claudeCost CostReader, codexCost CostReader, results []sourceRefreshResult, now time.Time) map[sources.Product]map[sources.WindowKind]cost.WindowCost {
	out := map[sources.Product]map[sources.WindowKind]cost.WindowCost{}
	add := func(reader CostReader, res sourceRefreshResult) {
		if reader == nil || res.err.Category != "" || len(res.windows) == 0 {
			return
		}
		if wc := reader.WindowCosts(now, res.windows); len(wc) > 0 {
			out[res.product] = wc
		}
	}
	add(claudeCost, results[0])
	add(codexCost, results[1])
	return out
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

func (m *Model) mergeRefresh(msg refreshMsg) []tea.Cmd {
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

		stamped := markStale(result.windows, msg.fetchedAt, m.staleAfter)
		m.windows[result.product] = stamped
		for _, w := range stamped {
			m.history.Append(trend.Key(w.Product, w.Kind), trend.Sample{
				CapturedAt: w.CapturedAt,
				UsedPct:    w.UsedPercent,
				ResetsAt:   w.ResetsAt,
			})
		}
		delete(m.errors, result.product)
	}

	if m.store != nil {
		_ = m.store.Save(m.history)
	}

	// msg.costs is nil only when a refreshMsg is built without cost data (e.g.
	// test helpers or a refresh with no cost readers); skip then to preserve any
	// previously computed costs rather than clobbering them with an empty map.
	if msg.costs != nil {
		m.costs = msg.costs
	}

	m.syncBarTargets()
	return nil
}

// syncBarTargets assumes a NewModel-constructed model (non-nil now); only
// reached via Update through mergeRefresh.
func (m *Model) syncBarTargets() {
	for i, spec := range quotaRowSpecs {
		window, ok := findWindow(*m, spec.product, spec.kind)
		if !ok {
			continue
		}
		target := progressFraction(window.UsedPercent)
		if target != m.bars[i].target {
			if m.bars[i].target >= 0 { // not the first load
				m.highlightUntil[i] = m.now().Add(highlightDuration)
			}
			m.bars[i].target = target
		}
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
