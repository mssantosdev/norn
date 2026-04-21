package components

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Select   key.Binding
	Confirm  key.Binding
	Cancel   key.Binding
	Quit     key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Select:   key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "select")),
		Confirm:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		Cancel:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Tab:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next")),
		ShiftTab: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev")),
	}
}
