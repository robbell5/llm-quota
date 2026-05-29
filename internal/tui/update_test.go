package tui

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/robbell5/llm-quota/internal/cost"
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

		// The request batches the (synchronous) refresh command with the animation
		// tick; pull the refreshMsg out of the batch without blocking on the timer.
		msg := firstRefreshMsg(t, cmd)
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
	t.Run("v cycles visibility", func(t *testing.T) {
		m := NewModel()
		updated, _ := m.Update(tea.KeyPressMsg{Text: "v", Code: 'v'})
		if got := updated.(Model).prefs.Visibility; got != VisibilityClaudeOnly {
			t.Fatalf("expected VisibilityClaudeOnly after one v, got %v", got)
		}
	})
}

func TestIKeyTogglesIcons(t *testing.T) {
	var model tea.Model = NewModel()
	model, _ = model.Update(tea.KeyPressMsg{Text: "i", Code: 'i'})
	if !model.(Model).prefs.Icons {
		t.Fatalf("i should enable icons")
	}
	model, _ = model.Update(tea.KeyPressMsg{Text: "i", Code: 'i'})
	if model.(Model).prefs.Icons {
		t.Fatalf("i should toggle icons back off")
	}
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
	// New data sets the bar target away from its empty start, so the model now
	// reports it needs frames. (The frame loop itself is kicked by
	// refreshRequestedMsg via ensureAnim; mergeRefresh only updates targets.)
	updated, _ := NewModel().Update(refreshMsg{
		results:   []sourceRefreshResult{dataResult(sources.ProductClaude, sources.WindowFiveHour, 60, fixedNow)},
		fetchedAt: fixedNow,
	})
	m := updated.(Model)
	i := barIndex(t, sources.ProductClaude, sources.WindowFiveHour)
	if m.bars[i].target != progressFraction(60) {
		t.Fatalf("expected Claude 5h bar target %v, got %v", progressFraction(60), m.bars[i].target)
	}
	if !m.animating() {
		t.Fatal("expected model to be animating from empty toward its target")
	}
}

func TestRefreshRequestKicksAnimationLoop(t *testing.T) {
	claude := &fakeReader{windows: []sources.Window{window(sources.ProductClaude, sources.WindowFiveHour, fixedNow)}}
	codex := &fakeReader{windows: []sources.Window{window(sources.ProductCodex, sources.WindowFiveHour, fixedNow)}}
	model := NewModel(WithReaders(claude, codex), WithClock(func() time.Time { return fixedNow }))

	updated, cmd := model.Update(refreshRequestedMsg{})
	if cmd == nil {
		t.Fatal("expected refresh request to return a command batch")
	}
	if !updated.(Model).animRunning {
		t.Fatal("expected refresh request to start the animation loop via ensureAnim")
	}
}

func TestRefreshHighlightsChangedValueButNotFirstLoad(t *testing.T) {
	i := barIndex(t, sources.ProductClaude, sources.WindowFiveHour)

	first, _ := NewModel(WithClock(func() time.Time { return fixedNow })).Update(refreshMsg{
		results:   []sourceRefreshResult{dataResult(sources.ProductClaude, sources.WindowFiveHour, 60, fixedNow)},
		fetchedAt: fixedNow,
	})
	m := first.(Model)
	if !m.highlightUntil[i].IsZero() {
		t.Fatalf("first load should not arm a value-change highlight, got %v", m.highlightUntil[i])
	}

	second, _ := m.Update(refreshMsg{
		results:   []sourceRefreshResult{dataResult(sources.ProductClaude, sources.WindowFiveHour, 80, fixedNow.Add(time.Minute))},
		fetchedAt: fixedNow.Add(time.Minute),
	})
	m2 := second.(Model)
	if m2.highlightUntil[i].IsZero() {
		t.Fatal("a changed value should arm the highlight timer")
	}
}

func TestRefreshDoesNotRehighlightUnchangedValue(t *testing.T) {
	i := barIndex(t, sources.ProductClaude, sources.WindowFiveHour)
	first, _ := NewModel(WithClock(func() time.Time { return fixedNow })).Update(refreshMsg{
		results:   []sourceRefreshResult{dataResult(sources.ProductClaude, sources.WindowFiveHour, 60, fixedNow)},
		fetchedAt: fixedNow,
	})
	m := first.(Model)

	second, _ := m.Update(refreshMsg{
		results:   []sourceRefreshResult{dataResult(sources.ProductClaude, sources.WindowFiveHour, 60, fixedNow.Add(time.Minute))},
		fetchedAt: fixedNow.Add(time.Minute),
	})
	m2 := second.(Model)
	if !m2.highlightUntil[i].IsZero() {
		t.Fatal("expected no highlight when the value is unchanged")
	}
}

func TestMissingRowBarStaysAtSentinel(t *testing.T) {
	updated, _ := NewModel().Update(refreshMsg{
		results:   []sourceRefreshResult{dataResult(sources.ProductClaude, sources.WindowFiveHour, 60, fixedNow)},
		fetchedAt: fixedNow,
	})
	m := updated.(Model)
	sonnet := barIndex(t, sources.ProductClaude, sources.WindowSonnetSevenDay)
	if m.bars[sonnet].target != -1 {
		t.Fatalf("expected absent Sonnet bar to keep sentinel target -1, got %v", m.bars[sonnet].target)
	}
}

func TestWindowSizeDoesNotStartAnimation(t *testing.T) {
	updated, cmd := NewModel().Update(tea.WindowSizeMsg{Width: 50, Height: 12})
	m := updated.(Model)
	if m.animRunning {
		t.Fatal("resize should not start the animation loop")
	}
	if cmd != nil {
		t.Fatalf("resize should return a nil command, got %T", cmd())
	}
	for i, b := range m.bars {
		if b.target != -1 {
			t.Fatalf("bar %d should keep sentinel target after resize, got %v", i, b.target)
		}
	}
}

func TestAnimTickAdvancesSpringTowardTarget(t *testing.T) {
	m := NewModel()
	m.bars[0].target = 1.0
	var model tea.Model = m
	for i := 0; i < 60; i++ { // ~4s of frames at 15fps
		model, _ = model.Update(animTickMsg(time.Now()))
	}
	if got := model.(Model).bars[0].pos; got < 0.95 {
		t.Fatalf("spring did not settle toward target: pos=%v", got)
	}
}

func TestAnimatingFalseWhenIdle(t *testing.T) {
	m := NewModel()
	for i := range m.bars {
		m.bars[i].target = 0
		m.bars[i].pos = 0
	}
	if m.animating() {
		t.Fatalf("idle model should not be animating")
	}
}

func TestAnimatingTrueWhileRefreshing(t *testing.T) {
	m := NewModel()
	m.refreshing = true
	if !m.animating() {
		t.Fatalf("refreshing model should be animating")
	}
}

func TestAnimTickStopsWhenSettled(t *testing.T) {
	m := NewModel()
	for i := range m.bars {
		m.bars[i].target = 0
		m.bars[i].pos = 0
	}
	m.animRunning = true
	updated, cmd := m.Update(animTickMsg(time.Now()))
	if cmd != nil {
		t.Fatalf("settled animation should self-suspend with a nil command, got %T", cmd())
	}
	if updated.(Model).animRunning {
		t.Fatal("settled animation should clear animRunning so the loop self-suspends")
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

func TestCKeyTogglesCost(t *testing.T) {
	m := NewModel()
	updated, _ := m.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	if !updated.(Model).prefs.HideCost {
		t.Fatalf("c should set HideCost=true")
	}
	again, _ := updated.(Model).Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	if again.(Model).prefs.HideCost {
		t.Fatalf("c should toggle HideCost back to false")
	}
}

func TestMergeRefreshStoresCosts(t *testing.T) {
	m := NewModel()
	msg := refreshMsg{
		fetchedAt: time.Unix(1_780_000_000, 0),
		costs: map[sources.Product]map[sources.WindowKind]cost.WindowCost{
			sources.ProductClaude: {sources.WindowFiveHour: {Amount: 3.2}},
		},
	}
	m.mergeRefresh(msg)
	if m.costs[sources.ProductClaude][sources.WindowFiveHour].Amount != 3.2 {
		t.Fatalf("mergeRefresh did not store costs")
	}
}

// firstRefreshMsg runs cmd (a single command or a tea.Batch) and returns the
// refreshMsg it produces. Each leaf command runs in its own goroutine so a
// batched, blocking tea.Tick command (the animation loop) never stalls the test.
func firstRefreshMsg(t *testing.T, cmd tea.Cmd) refreshMsg {
	t.Helper()
	if cmd == nil {
		t.Fatal("expected a command producing refreshMsg, got nil")
	}

	results := make(chan tea.Msg, 8)
	var run func(c tea.Cmd)
	run = func(c tea.Cmd) {
		if c == nil {
			return
		}
		go func() {
			msg := c()
			if batch, ok := msg.(tea.BatchMsg); ok {
				for _, inner := range batch {
					run(inner)
				}
				return
			}
			results <- msg
		}()
	}
	run(cmd)

	timeout := time.After(2 * time.Second)
	for {
		select {
		case msg := <-results:
			if rm, ok := msg.(refreshMsg); ok {
				return rm
			}
		case <-timeout:
			t.Fatal("timed out waiting for refreshMsg from command batch")
		}
	}
}
