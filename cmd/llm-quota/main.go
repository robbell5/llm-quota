package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/rob/llm-quota/internal/tui"
)

func main() {
	if len(os.Args) > 1 {
		fmt.Fprintf(os.Stderr, "llm-quota: unknown argument: %s\n", os.Args[1])
		os.Exit(2)
	}

	program := tea.NewProgram(tui.NewModel())
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "llm-quota: %v\n", err)
		os.Exit(1)
	}
}
