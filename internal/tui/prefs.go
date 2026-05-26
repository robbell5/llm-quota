package tui

import (
	"fmt"

	"github.com/robbell5/llm-quota/internal/sources"
)

// BarStyle selects how progress bars are filled.
type BarStyle int

const (
	BarSegmented BarStyle = iota // half-block '▌' fill (default)
	BarSolid                     // full-block '█' fill
)

func (s BarStyle) String() string {
	switch s {
	case BarSegmented:
		return "BarSegmented"
	case BarSolid:
		return "BarSolid"
	default:
		return fmt.Sprintf("BarStyle(%d)", int(s))
	}
}

func (s BarStyle) toggled() BarStyle {
	if s == BarSolid {
		return BarSegmented
	}
	return BarSolid
}

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
// view: segmented bars, both providers visible, trend line shown.
type DisplayPrefs struct {
	BarStyle   BarStyle
	Visibility Visibility
	HideTrend  bool
}

func (p DisplayPrefs) trendVisible() bool {
	return !p.HideTrend
}
