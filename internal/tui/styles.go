package tui

import "github.com/charmbracelet/lipgloss"

var (
primaryColor   = lipgloss.AdaptiveColor{Light: "#5A67D8", Dark: "#7C3AED"}
secondaryColor = lipgloss.AdaptiveColor{Light: "#4299E1", Dark: "#3B82F6"}
successColor   = lipgloss.AdaptiveColor{Light: "#38A169", Dark: "#10B981"}
errorColor     = lipgloss.AdaptiveColor{Light: "#E53E3E", Dark: "#EF4444"}
mutedColor     = lipgloss.AdaptiveColor{Light: "#718096", Dark: "#6B7280"}
borderColor    = lipgloss.AdaptiveColor{Light: "#CBD5E0", Dark: "#4B5563"}

outerFrameStyle = lipgloss.NewStyle().
Border(lipgloss.RoundedBorder()).
BorderForeground(borderColor)

headerStyle = lipgloss.NewStyle().
Bold(true).
Foreground(primaryColor)

versionStyle = lipgloss.NewStyle().
Foreground(mutedColor).
Italic(true)

savedStyle = lipgloss.NewStyle().
Foreground(successColor).
Bold(true)

panelStyle = lipgloss.NewStyle().
Border(lipgloss.RoundedBorder()).
BorderForeground(borderColor).
Padding(0, 1)

focusedPanelStyle = lipgloss.NewStyle().
Border(lipgloss.RoundedBorder()).
BorderForeground(primaryColor).
Padding(0, 1)

groupHeaderStyle = lipgloss.NewStyle().
Bold(true).
Foreground(secondaryColor)

itemStyle = lipgloss.NewStyle().
Foreground(lipgloss.AdaptiveColor{Light: "#1A202C", Dark: "#F7FAFC"})

selectedItemStyle = lipgloss.NewStyle().
Bold(true).
Foreground(primaryColor)

sensitiveItemStyle = lipgloss.NewStyle().
Foreground(mutedColor).
Italic(true)

valueStyle = lipgloss.NewStyle().
Foreground(mutedColor)

detailHeaderStyle = lipgloss.NewStyle().
Bold(true).
Foreground(primaryColor).
Underline(true)

detailLabelStyle = lipgloss.NewStyle().
Bold(true).
Foreground(secondaryColor)

detailDescStyle = lipgloss.NewStyle().
Foreground(lipgloss.AdaptiveColor{Light: "#2D3748", Dark: "#E2E8F0"})

detailValueStyle = lipgloss.NewStyle().
Foreground(lipgloss.AdaptiveColor{Light: "#1A202C", Dark: "#F7FAFC"})

detailNoteStyle = lipgloss.NewStyle().
Foreground(mutedColor).
Italic(true)

sensitiveValueStyle = lipgloss.NewStyle().
Foreground(errorColor).
Bold(true)

toggleOnStyle       = lipgloss.NewStyle().Foreground(successColor).Bold(true)
toggleOffStyle      = lipgloss.NewStyle().Foreground(mutedColor)
selectedOptionStyle = lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
optionStyle         = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#2D3748", Dark: "#E2E8F0"})
errorStyle          = lipgloss.NewStyle().Foreground(errorColor).Bold(true)

helpStyle = lipgloss.NewStyle().Foreground(mutedColor)
)
