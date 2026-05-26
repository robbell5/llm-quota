package tui

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"

	"github.com/robbell5/llm-quota/internal/sources"
	"github.com/robbell5/llm-quota/internal/trend"
)

var fixedNow = time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

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

func TestUpdateQuits(t *testing.T) {
	tests := []struct {
		name string
		msg  tea.KeyPressMsg
	}{
		{
			name: "q",
			msg:  tea.KeyPressMsg{Text: "q", Code: 'q'},
		},
		{
			name: "ctrl+c",
			msg:  tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cmd := NewModel().Update(tt.msg)
			if cmd == nil {
				t.Fatal("expected quit command, got nil")
			}

			msg := cmd()
			if _, ok := msg.(tea.QuitMsg); !ok {
				t.Fatalf("expected tea.QuitMsg, got %T", msg)
			}
		})
	}
}

func TestUpdateStoresWindowSize(t *testing.T) {
	updated, cmd := NewModel().Update(tea.WindowSizeMsg{Width: 50, Height: 12})
	if cmd != nil {
		t.Fatalf("expected nil command, got %T", cmd())
	}

	model, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", updated)
	}

	if model.width != 50 {
		t.Fatalf("expected width 50, got %d", model.width)
	}

	if model.height != 12 {
		t.Fatalf("expected height 12, got %d", model.height)
	}
}

func TestInitRequestsRefreshAndSchedulesTick(t *testing.T) {
	model := NewModel()
	if model.refreshEvery != 30*time.Second {
		t.Fatalf("expected default refresh interval 30s, got %s", model.refreshEvery)
	}

	cmd := model.Init()
	if cmd == nil {
		t.Fatal("expected init command")
	}

	batch, ok := cmd().(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected tea.BatchMsg, got %T", cmd())
	}
	if len(batch) != 2 {
		t.Fatalf("expected refresh and tick commands, got %d", len(batch))
	}
	if msg := batch[0](); msg != (refreshRequestedMsg{}) {
		t.Fatalf("expected immediate refresh request, got %T", msg)
	}
}

func TestRefresh(t *testing.T) {
	t.Run("manual r requests refresh when idle without scheduling tick", func(t *testing.T) {
		model := NewModel(WithClock(func() time.Time { return fixedNow }))

		updated, cmd := model.Update(tea.KeyPressMsg{Text: "r", Code: 'r'})
		if cmd == nil {
			t.Fatal("expected refresh request command")
		}
		if msg := cmd(); msg != (refreshRequestedMsg{}) {
			t.Fatalf("expected refreshRequestedMsg, got %T", msg)
		}

		got := updated.(Model)
		if got.refreshing {
			t.Fatal("manual key should request refresh without setting refreshing until request message is handled")
		}
	})

	t.Run("duplicate manual r while refreshing returns nil command", func(t *testing.T) {
		model := NewModel()
		model.refreshing = true

		_, cmd := model.Update(tea.KeyPressMsg{Text: "r", Code: 'r'})
		if cmd != nil {
			t.Fatalf("expected duplicate refresh to return nil command, got %T", cmd())
		}
	})

	t.Run("refresh request starts one refresh and coalesces duplicate requests", func(t *testing.T) {
		claude := &fakeReader{windows: []sources.Window{window(sources.ProductClaude, sources.WindowFiveHour, fixedNow)}}
		codex := &fakeReader{windows: []sources.Window{window(sources.ProductCodex, sources.WindowFiveHour, fixedNow)}}
		model := NewModel(WithReaders(claude, codex), WithClock(func() time.Time { return fixedNow }))

		updated, cmd := model.Update(refreshRequestedMsg{})
		if cmd == nil {
			t.Fatal("expected refresh command")
		}
		refreshing := updated.(Model)
		if !refreshing.refreshing {
			t.Fatal("expected model to mark refresh in flight")
		}

		_, duplicate := refreshing.Update(refreshRequestedMsg{})
		if duplicate != nil {
			t.Fatalf("expected duplicate refresh request to coalesce, got %T", duplicate())
		}

		msg, ok := cmd().(refreshMsg)
		if !ok {
			t.Fatalf("expected refreshMsg, got %T", cmd())
		}
		if msg.fetchedAt != fixedNow {
			t.Fatalf("expected fixed fetch time %s, got %s", fixedNow, msg.fetchedAt)
		}
		if claude.calls != 1 || codex.calls != 1 {
			t.Fatalf("expected one fetch per source, got claude=%d codex=%d", claude.calls, codex.calls)
		}
	})

	t.Run("last-known-good is preserved per product after later source failure", func(t *testing.T) {
		claudeWindow := window(sources.ProductClaude, sources.WindowFiveHour, fixedNow)
		codexWindow := window(sources.ProductCodex, sources.WindowFiveHour, fixedNow)
		model := NewModel()

		updated, _ := model.Update(refreshMsg{
			results: []sourceRefreshResult{
				{product: sources.ProductClaude, windows: []sources.Window{claudeWindow}},
				{product: sources.ProductCodex, windows: []sources.Window{codexWindow}},
			},
			fetchedAt: fixedNow,
		})
		model = updated.(Model)

		newCodex := window(sources.ProductCodex, sources.WindowSevenDay, fixedNow.Add(time.Minute))
		claudeErr := sources.SourceError{Source: sources.ProductClaude, Category: sources.ErrorRead, Err: errors.New("cache locked")}
		updated, _ = model.Update(refreshMsg{
			results: []sourceRefreshResult{
				{product: sources.ProductClaude, err: claudeErr},
				{product: sources.ProductCodex, windows: []sources.Window{newCodex}},
			},
			fetchedAt: fixedNow.Add(time.Minute),
		})
		model = updated.(Model)

		if !reflect.DeepEqual(model.windows[sources.ProductClaude], []sources.Window{claudeWindow}) {
			t.Fatalf("expected Claude last-known-good window preserved, got %#v", model.windows[sources.ProductClaude])
		}
		if !reflect.DeepEqual(model.windows[sources.ProductCodex], []sources.Window{newCodex}) {
			t.Fatalf("expected Codex window updated, got %#v", model.windows[sources.ProductCodex])
		}
		if model.errors[sources.ProductClaude].Category != sources.ErrorRead {
			t.Fatalf("expected Claude read error, got %#v", model.errors[sources.ProductClaude])
		}
		if _, ok := model.errors[sources.ProductCodex]; ok {
			t.Fatalf("expected Codex error cleared, got %#v", model.errors[sources.ProductCodex])
		}
	})

	t.Run("initial failures store typed source errors and leave windows empty", func(t *testing.T) {
		missing := sources.SourceError{Source: sources.ProductClaude, Category: sources.ErrorMissing}
		malformed := sources.SourceError{Source: sources.ProductCodex, Category: sources.ErrorMalformed, Err: errors.New("bad json")}

		updated, _ := NewModel().Update(refreshMsg{
			results: []sourceRefreshResult{
				{product: sources.ProductClaude, err: missing},
				{product: sources.ProductCodex, err: malformed},
			},
			fetchedAt: fixedNow,
		})
		model := updated.(Model)

		if len(model.windows[sources.ProductClaude]) != 0 || len(model.windows[sources.ProductCodex]) != 0 {
			t.Fatalf("expected empty windows, got %#v", model.windows)
		}
		if model.errors[sources.ProductClaude].Category != sources.ErrorMissing {
			t.Fatalf("expected missing Claude error, got %#v", model.errors[sources.ProductClaude])
		}
		if model.errors[sources.ProductCodex].Category != sources.ErrorMalformed {
			t.Fatalf("expected malformed Codex error, got %#v", model.errors[sources.ProductCodex])
		}
	})

	t.Run("all typed source error categories can be preserved", func(t *testing.T) {
		categories := []sources.ErrorCategory{
			sources.ErrorMissing,
			sources.ErrorMalformed,
			sources.ErrorNoUsableEvent,
			sources.ErrorRead,
		}

		for _, category := range categories {
			err := sources.SourceError{Source: sources.ProductClaude, Category: category, Err: errors.New(string(category))}
			updated, _ := NewModel().Update(refreshMsg{
				results:   []sourceRefreshResult{{product: sources.ProductClaude, err: err}},
				fetchedAt: fixedNow,
			})
			model := updated.(Model)

			if model.errors[sources.ProductClaude].Category != category {
				t.Fatalf("expected %s error, got %#v", category, model.errors[sources.ProductClaude])
			}
		}
	})

	t.Run("windows older than one hour are marked stale and remain visible data", func(t *testing.T) {
		capturedAt := fixedNow.Add(-time.Hour - time.Minute)
		oldClaude := window(sources.ProductClaude, sources.WindowSevenDay, capturedAt)
		oldCodex := window(sources.ProductCodex, sources.WindowSevenDay, capturedAt)

		updated, _ := NewModel().Update(refreshMsg{
			results: []sourceRefreshResult{
				{product: sources.ProductClaude, windows: []sources.Window{oldClaude}},
				{product: sources.ProductCodex, windows: []sources.Window{oldCodex}},
			},
			fetchedAt: fixedNow,
		})
		model := updated.(Model)

		for _, product := range []sources.Product{sources.ProductClaude, sources.ProductCodex} {
			windows := model.windows[product]
			if len(windows) != 1 {
				t.Fatalf("expected one %s window, got %#v", product, windows)
			}
			if !windows[0].Stale {
				t.Fatalf("expected %s window to be stale", product)
			}
			if windows[0].StaleAge != time.Hour+time.Minute {
				t.Fatalf("expected stale age 1h1m, got %s", windows[0].StaleAge)
			}
		}
	})

	t.Run("stale state stays in model without Phase 4 visible warning copy", func(t *testing.T) {
		model := NewModel()
		model.windows[sources.ProductClaude] = []sources.Window{{
			Product:     sources.ProductClaude,
			Kind:        sources.WindowFiveHour,
			Label:       "Claude 5h",
			UsedPercent: 42,
			CapturedAt:  fixedNow.Add(-2 * time.Hour),
			Stale:       true,
			StaleAge:    2 * time.Hour,
		}}

		view := render(model)
		for _, forbidden := range []string{"refreshing", "last updated", "stale"} {
			if contains(view, forbidden) {
				t.Fatalf("did not expect visible %q copy in Phase 3 view: %q", forbidden, view)
			}
		}
	})
}

func window(product sources.Product, kind sources.WindowKind, capturedAt time.Time) sources.Window {
	return sources.Window{
		Product:     product,
		Kind:        kind,
		Label:       string(product) + " " + string(kind),
		UsedPercent: 25,
		ResetsAt:    capturedAt.Add(time.Hour),
		CapturedAt:  capturedAt,
	}
}

func cloneWindows(windows []sources.Window) []sources.Window {
	cloned := make([]sources.Window, len(windows))
	copy(cloned, windows)
	return cloned
}

func contains(s string, needle string) bool {
	return strings.Contains(s, needle)
}

func TestToggleKeys(t *testing.T) {
	t.Run("b toggles bar style", func(t *testing.T) {
		updated, cmd := NewModel().Update(tea.KeyPressMsg{Text: "b", Code: 'b'})
		if cmd != nil {
			t.Fatalf("expected nil command, got %T", cmd())
		}
		if got := updated.(Model).prefs.BarStyle; got != BarSolid {
			t.Fatalf("expected BarSolid after one toggle, got %v", got)
		}
	})

	t.Run("v cycles visibility", func(t *testing.T) {
		m := NewModel()
		updated, _ := m.Update(tea.KeyPressMsg{Text: "v", Code: 'v'})
		if got := updated.(Model).prefs.Visibility; got != VisibilityClaudeOnly {
			t.Fatalf("expected VisibilityClaudeOnly after one v, got %v", got)
		}
	})
}

func TestTKeyTogglesTrend(t *testing.T) {
	m := NewModel()
	if m.prefs.HideTrend {
		t.Fatalf("trend should start visible")
	}
	updated, _ := m.Update(tea.KeyPressMsg{Code: 't', Text: "t"})
	got := updated.(Model)
	if !got.prefs.HideTrend {
		t.Fatalf("expected 't' to hide the trend line")
	}
	updated, _ = got.Update(tea.KeyPressMsg{Code: 't', Text: "t"})
	if updated.(Model).prefs.HideTrend {
		t.Fatalf("expected second 't' to show the trend line")
	}
}

func TestTickSchedulesRefreshAndNextTick(t *testing.T) {
	model := NewModel()
	updated, cmd := model.Update(tickMsg(fixedNow))
	if cmd == nil {
		t.Fatal("expected tick command batch")
	}
	batch, ok := cmd().(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected tea.BatchMsg, got %T", cmd())
	}
	if len(batch) != 2 {
		t.Fatalf("expected refresh request and next tick commands, got %d", len(batch))
	}
	if msg := batch[0](); msg != (refreshRequestedMsg{}) {
		t.Fatalf("expected refresh request, got %T", msg)
	}
	if updated.(Model).refreshing {
		t.Fatal("tick should request refresh without marking in-flight until request message is handled")
	}
}

func dataResult(product sources.Product, kind sources.WindowKind, percent float64, now time.Time) sourceRefreshResult {
	return sourceRefreshResult{
		product: product,
		windows: []sources.Window{{
			Product:     product,
			Kind:        kind,
			Label:       string(product),
			UsedPercent: percent,
			ResetsAt:    now.Add(time.Hour),
			CapturedAt:  now,
		}},
	}
}

func barIndex(t *testing.T, product sources.Product, kind sources.WindowKind) int {
	t.Helper()
	for i, spec := range quotaRowSpecs {
		if spec.product == product && spec.kind == kind {
			return i
		}
	}
	t.Fatalf("no row spec for %s/%s", product, kind)
	return -1
}

func TestRefreshStartsAnimationForNewData(t *testing.T) {
	updated, _ := NewModel().Update(refreshMsg{
		results:   []sourceRefreshResult{dataResult(sources.ProductClaude, sources.WindowFiveHour, 60, fixedNow)},
		fetchedAt: fixedNow,
	})
	m := updated.(Model)
	i := barIndex(t, sources.ProductClaude, sources.WindowFiveHour)
	if !m.bars[i].IsAnimating() {
		t.Fatal("expected Claude 5h bar to animate from empty toward its target")
	}
}

func TestRefreshDoesNotReanimateUnchangedValue(t *testing.T) {
	first, _ := NewModel().Update(refreshMsg{
		results:   []sourceRefreshResult{dataResult(sources.ProductClaude, sources.WindowFiveHour, 60, fixedNow)},
		fetchedAt: fixedNow,
	})
	m := first.(Model)
	i := barIndex(t, sources.ProductClaude, sources.WindowFiveHour)
	// Settle the first animation so IsAnimating would only be true again on a change.
	m.bars[i] = settleBar(m.bars[i])

	second, _ := m.Update(refreshMsg{
		results:   []sourceRefreshResult{dataResult(sources.ProductClaude, sources.WindowFiveHour, 60, fixedNow.Add(time.Minute))},
		fetchedAt: fixedNow.Add(time.Minute),
	})
	m2 := second.(Model)
	if m2.bars[i].IsAnimating() {
		t.Fatal("expected no re-animation when the value is unchanged")
	}
}

func TestMissingRowBarDoesNotAnimate(t *testing.T) {
	updated, _ := NewModel().Update(refreshMsg{
		results:   []sourceRefreshResult{dataResult(sources.ProductClaude, sources.WindowFiveHour, 60, fixedNow)},
		fetchedAt: fixedNow,
	})
	m := updated.(Model)
	sonnet := barIndex(t, sources.ProductClaude, sources.WindowSonnetSevenDay)
	if m.bars[sonnet].IsAnimating() {
		t.Fatal("expected absent Sonnet bar to stay idle")
	}
}

func TestWindowSizeDoesNotAnimate(t *testing.T) {
	updated, _ := NewModel().Update(tea.WindowSizeMsg{Width: 50, Height: 12})
	m := updated.(Model)
	for i := range m.bars {
		if m.bars[i].IsAnimating() {
			t.Fatalf("bar %d should not animate on resize", i)
		}
	}
}

func TestMergeRefreshAppendsHistory(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	m := NewModel(WithClock(func() time.Time { return now }))

	reset := now.Add(2 * time.Hour)
	msg := refreshMsg{
		fetchedAt: now,
		results: []sourceRefreshResult{
			{product: sources.ProductClaude, windows: []sources.Window{
				{Product: sources.ProductClaude, Kind: sources.WindowFiveHour, Label: "Claude 5h",
					UsedPercent: 41, ResetsAt: reset, CapturedAt: now},
			}},
		},
	}
	m.mergeRefresh(msg)

	key := trend.Key(sources.ProductClaude, sources.WindowFiveHour)
	got := m.history.EpochSamples(key, reset)
	if len(got) != 1 || got[0].UsedPct != 41 {
		t.Fatalf("expected one appended sample at 41%%, got %+v", got)
	}
}

func TestNewModelHasEmptyHistoryWithoutStore(t *testing.T) {
	m := NewModel()
	if m.history == nil {
		t.Fatalf("history should be initialized even without a store")
	}
}

// settleBar advances a progress bar until it stops animating, feeding it the
// frame messages its own commands produce (FrameMsg has unexported fields, so
// we can only obtain valid ones from the bar's own commands).
func settleBar(b progress.Model) progress.Model {
	cmd := b.SetPercent(b.Percent())
	for b.IsAnimating() {
		msg := cmd()
		b, cmd = b.Update(msg)
		if cmd == nil {
			cmd = b.SetPercent(b.Percent())
		}
	}
	return b
}
