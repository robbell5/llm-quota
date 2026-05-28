package tui

import (
	"fmt"

	"github.com/robbell5/llm-quota/internal/sources"
)

// Visibility selects which providers' rows are shown. The "hide both" state is
// unrepresentable by construction.
type Visibility int

const (
	VisibilityBoth Visibility = iota
	VisibilityClaudeOnly
	VisibilityCodexOnly
)

func (v Visibility) String() string {
	switch v {
	case VisibilityBoth:
		return "VisibilityBoth"
	case VisibilityClaudeOnly:
		return "VisibilityClaudeOnly"
	case VisibilityCodexOnly:
		return "VisibilityCodexOnly"
	default:
		return fmt.Sprintf("Visibility(%d)", int(v))
	}
}

func (v Visibility) next() Visibility {
	switch v {
	case VisibilityBoth:
		return VisibilityClaudeOnly
	case VisibilityClaudeOnly:
		return VisibilityCodexOnly
	default:
		return VisibilityBoth
	}
}

func (v Visibility) shows(product sources.Product) bool {
	switch v {
	case VisibilityClaudeOnly:
		return product == sources.ProductClaude
	case VisibilityCodexOnly:
		return product == sources.ProductCodex
	default:
		return true
	}
}

// DisplayPrefs holds user display preferences. The zero value is the default
// view: both providers visible, trend line shown, no Nerd Font icons.
type DisplayPrefs struct {
	Visibility Visibility
	HideTrend  bool
	Icons      bool // Nerd Font icon mode (default false = safe Unicode)
}

func (p DisplayPrefs) trendVisible() bool {
	return !p.HideTrend
}
