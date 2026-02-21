package tui

// State represents the current state of the TUI
type State int

const (
	// StateBrowsing: arrow keys navigate list, Enter switches to editing
	StateBrowsing State = iota
	// StateEditing: right panel input widget is focused, Esc saves and returns to browsing
	StateEditing
	// StateSaving: persisting changes to config file
	StateSaving
	// StateExiting: final save (if needed) and quit
	StateExiting
)

func (s State) String() string {
	switch s {
	case StateBrowsing:
		return "Browsing"
	case StateEditing:
		return "Editing"
	case StateSaving:
		return "Saving"
	case StateExiting:
		return "Exiting"
	default:
		return "Unknown"
	}
}
