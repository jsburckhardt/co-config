package tui

import "github.com/charmbracelet/lipgloss"

var (
	// titleStyle is used for the main header
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			MarginBottom(1)

	// versionStyle is used for displaying the Copilot CLI version
	versionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

	// fieldLabelStyle is used for field labels
	fieldLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("99"))

	// sensitiveFieldStyle highlights sensitive fields
	sensitiveFieldStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true)

	// statusBarStyle is used for the bottom status bar
	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)
)
