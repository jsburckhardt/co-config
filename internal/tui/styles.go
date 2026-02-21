package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Color palette
	primaryColor   = lipgloss.AdaptiveColor{Light: "#5A67D8", Dark: "#7C3AED"}
	secondaryColor = lipgloss.AdaptiveColor{Light: "#4299E1", Dark: "#3B82F6"}
	successColor   = lipgloss.AdaptiveColor{Light: "#38A169", Dark: "#10B981"}
	errorColor     = lipgloss.AdaptiveColor{Light: "#E53E3E", Dark: "#EF4444"}
	mutedColor     = lipgloss.AdaptiveColor{Light: "#718096", Dark: "#6B7280"}
	borderColor    = lipgloss.AdaptiveColor{Light: "#CBD5E0", Dark: "#4B5563"}

	// Header styles
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Padding(0, 1)

	versionStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true).
			Padding(0, 1)

	// Panel borders
	listPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2)

	detailPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(borderColor).
				Padding(1, 2)

	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(1, 2)

	// List item styles
	groupHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(secondaryColor).
				Underline(true).
				MarginTop(1).
				MarginBottom(1)

	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1A202C", Dark: "#F7FAFC"})

	selectedItemStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(primaryColor)

	descStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	selectedDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#4A5568", Dark: "#A0AEC0"})

	sensitiveItemStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Italic(true)

	// Detail panel styles
	detailHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(primaryColor).
				Underline(true).
				MarginBottom(1)

	detailLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(secondaryColor)

	detailDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#2D3748", Dark: "#E2E8F0"})

	detailValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#1A202C", Dark: "#F7FAFC"}).
				Italic(true)

	detailNoteStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	sensitiveValueStyle = lipgloss.NewStyle().
				Foreground(errorColor).
				Bold(true)

	// Input widget styles
	toggleOnStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	toggleOffStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	selectedOptionStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	optionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#2D3748", Dark: "#E2E8F0"})

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Help bar style
	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Padding(1, 2)

	// Status messages
	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true).
			Padding(1, 2)

	errorMessageStyle = lipgloss.NewStyle().
				Foreground(errorColor).
				Bold(true).
				Padding(1, 2)
)
