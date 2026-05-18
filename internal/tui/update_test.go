package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

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

func TestInitHasNoCommand(t *testing.T) {
	if cmd := NewModel().Init(); cmd != nil {
		t.Fatalf("expected nil init command, got %T", cmd())
	}
}
