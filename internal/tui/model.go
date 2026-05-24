package tui

import (
	"time"

	"charm.land/bubbles/v2/progress"

	"github.com/robbell5/llm-quota/internal/sources"
)

type SourceReader interface {
	Fetch(now time.Time) ([]sources.Window, error)
}

type Model struct {
	width  int
	height int

	claudeReader        SourceReader
	codexReader         SourceReader
	now                 func() time.Time
	refreshEvery        time.Duration
	staleAfter          time.Duration
	refreshing          bool
	windows             map[sources.Product][]sources.Window
	errors              map[sources.Product]sources.SourceError
	claudeHookInstalled bool

	bars []progress.Model
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

	m.bars = make([]progress.Model, len(quotaRowSpecs))
	for i := range m.bars {
		p := progress.New(progress.WithoutPercentage())
		p.EmptyColor = mochaSurface0
		// Spring tuning for the Phase 9 animation; harmless until SetPercent is called.
		p.SetSpringOptions(12.0, 1.0)
		m.bars[i] = p
	}

	return m
}

func WithReaders(claude SourceReader, codex SourceReader) Option {
	return func(m *Model) {
		m.claudeReader = claude
		m.codexReader = codex
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
