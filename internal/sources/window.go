package sources

import "time"

type Product string

const (
	ProductClaude Product = "claude"
	ProductCodex  Product = "codex"
)

type WindowKind string

const (
	WindowFiveHour WindowKind = "five_hour"
	WindowSevenDay WindowKind = "seven_day"
)

type Metadata map[string]string

type Window struct {
	Product     Product
	Kind        WindowKind
	Label       string
	UsedPercent float64
	ResetsAt    time.Time
	CapturedAt  time.Time
	Stale       bool
	StaleAge    time.Duration
	Metadata    Metadata
}

type ErrorCategory string

const (
	ErrorMissing       ErrorCategory = "missing"
	ErrorMalformed     ErrorCategory = "malformed"
	ErrorNoUsableEvent ErrorCategory = "no_usable_event"
	ErrorRead          ErrorCategory = "read_error"
)

type SourceError struct {
	Source   Product
	Category ErrorCategory
	Err      error
}

func (e SourceError) Error() string {
	if e.Err == nil {
		return string(e.Source) + " source " + string(e.Category)
	}

	return string(e.Source) + " source " + string(e.Category) + ": " + e.Err.Error()
}

func (e SourceError) Unwrap() error {
	return e.Err
}
