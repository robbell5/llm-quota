package tui

import (
	"time"

	"github.com/charmbracelet/harmonica"

	"github.com/robbell5/llm-quota/internal/cost"
	"github.com/robbell5/llm-quota/internal/sources"
	"github.com/robbell5/llm-quota/internal/trend"
)

const (
	animFPS           = 15
	highlightDuration = 900 * time.Millisecond
	// springSettleEpsilon is the fraction-unit tolerance below which a bar is
	// considered visually settled (≈0.1%), distinct from numerical epsilon.
	springSettleEpsilon = 0.001
)

// barAnim is one row's spring-animated fill fraction. target is -1 until the
// first real value so the bar animates up from empty on first data.
type barAnim struct {
	pos    float64
	vel    float64
	target float64
}

type SourceReader interface {
	Fetch(now time.Time) ([]sources.Window, error)
}

type CostReader interface {
	WindowCosts(now time.Time, windows []sources.Window) map[sources.WindowKind]cost.WindowCost
}

type Model struct {
	width  int
	height int

	claudeReader        SourceReader
	codexReader         SourceReader
	claudeCost          CostReader
	codexCost           CostReader
	costs               map[sources.Product]map[sources.WindowKind]cost.WindowCost
	now                 func() time.Time
	refreshEvery        time.Duration
	staleAfter          time.Duration
	refreshing          bool
	windows             map[sources.Product][]sources.Window
	errors              map[sources.Product]sources.SourceError
	claudeHookInstalled bool

	bars   []barAnim
	spring harmonica.Spring

	animPhase      int
	animRunning    bool
	highlightUntil []time.Time

	prefs DisplayPrefs

	history *trend.History
	store   *trend.Store
}

type Option func(*Model)

func NewModel(options ...Option) Model {
	m := Model{
		now:          time.Now,
		refreshEvery: 30 * time.Second,
		staleAfter:   time.Hour,
		windows: map[sources.Product][]sources.Window{
			sources.ProductClaude: nil,
			sources.ProductCodex:  nil,
		},
		errors: make(map[sources.Product]sources.SourceError),
	}

	for _, option := range options {
		option(&m)
	}

	if m.now == nil {
		m.now = time.Now
	}
	if m.refreshEvery <= 0 {
		m.refreshEvery = 30 * time.Second
	}
	if m.staleAfter <= 0 {
		m.staleAfter = time.Hour
	}
	if m.windows == nil {
		m.windows = make(map[sources.Product][]sources.Window)
	}
	if m.errors == nil {
		m.errors = make(map[sources.Product]sources.SourceError)
	}

	m.spring = harmonica.NewSpring(harmonica.FPS(animFPS), 12.0, 1.0)
	m.bars = make([]barAnim, len(quotaRowSpecs))
	m.highlightUntil = make([]time.Time, len(quotaRowSpecs))
	for i := range m.bars {
		m.bars[i] = barAnim{target: -1}
	}

	if m.store != nil {
		m.history = m.store.Load()
	}
	if m.history == nil {
		m.history = trend.NewHistory()
	}

	return m
}

func WithReaders(claude SourceReader, codex SourceReader) Option {
	return func(m *Model) {
		m.claudeReader = claude
		m.codexReader = codex
	}
}

func WithCostReaders(claude CostReader, codex CostReader) Option {
	return func(m *Model) {
		m.claudeCost = claude
		m.codexCost = codex
	}
}

// WithCosts injects a precomputed costs map (used by tests to render without
// touching the filesystem).
func WithCosts(costs map[sources.Product]map[sources.WindowKind]cost.WindowCost) Option {
	return func(m *Model) {
		m.costs = costs
	}
}

func WithClock(now func() time.Time) Option {
	return func(m *Model) {
		m.now = now
	}
}

func WithRefreshEvery(interval time.Duration) Option {
	return func(m *Model) {
		m.refreshEvery = interval
	}
}

func WithClaudeHookInstalled(installed bool) Option {
	return func(m *Model) {
		m.claudeHookInstalled = installed
	}
}

func WithDisplayPrefs(prefs DisplayPrefs) Option {
	return func(m *Model) {
		m.prefs = prefs
	}
}

func WithHistoryStore(store *trend.Store) Option {
	return func(m *Model) {
		m.store = store
	}
}

// costActive reports whether the value clusters and consolidated freshness line
// should render: cost must be visible AND there must be at least one value.
func (m Model) costActive() bool {
	if !m.prefs.costVisible() {
		return false
	}
	for _, byKind := range m.costs {
		if len(byKind) > 0 {
			return true
		}
	}
	return false
}
