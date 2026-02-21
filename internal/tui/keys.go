package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the key bindings for the TUI.
type KeyMap struct {
Up     key.Binding
Down   key.Binding
Enter  key.Binding
Escape key.Binding
Save   key.Binding
Quit   key.Binding
Tab    key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
return KeyMap{
Up: key.NewBinding(
key.WithKeys("up", "k"),
key.WithHelp("↑/k", "up"),
),
Down: key.NewBinding(
key.WithKeys("down", "j"),
key.WithHelp("↓/j", "down"),
),
Enter: key.NewBinding(
key.WithKeys("enter"),
key.WithHelp("enter", "edit"),
),
Escape: key.NewBinding(
key.WithKeys("esc"),
key.WithHelp("esc", "done"),
),
Save: key.NewBinding(
key.WithKeys("ctrl+s"),
key.WithHelp("ctrl+s", "save"),
),
Quit: key.NewBinding(
key.WithKeys("ctrl+c"),
key.WithHelp("ctrl+c", "quit"),
),
Tab: key.NewBinding(
key.WithKeys("tab"),
key.WithHelp("tab", "switch"),
),
}
}
