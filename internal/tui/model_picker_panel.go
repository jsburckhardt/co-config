package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// modelItem implements list.Item for model options in the picker.
type modelItem struct {
	name string
}

func (i modelItem) Title() string       { return i.name }
func (i modelItem) Description() string { return "" }
func (i modelItem) FilterValue() string { return i.name }

// ModelPickerPanel wraps a bubbles list.Model to provide a fuzzy-filterable
// model selection experience.
type ModelPickerPanel struct {
	list     list.Model
	selected string
}

// NewModelPickerPanel creates a new model picker panel with the given options
// and pre-selects the current value.
func NewModelPickerPanel(options []string, current string) ModelPickerPanel {
	items := make([]list.Item, len(options))
	for i, opt := range options {
		items[i] = modelItem{name: opt}
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetHeight(1)
	delegate.SetSpacing(0)
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(primaryColor).
		BorderForeground(primaryColor)
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Foreground(lipgloss.AdaptiveColor{Light: "#1A202C", Dark: "#F7FAFC"})

	l := list.New(items, delegate, 0, 0)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.Title = "Select Model"
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
	l.FilterInput.Prompt = "Search: "
	l.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(primaryColor)

	// Pre-select the current value
	for i, opt := range options {
		if opt == current {
			l.Select(i)
			break
		}
	}

	return ModelPickerPanel{
		list:     l,
		selected: current,
	}
}

// SetSize updates the dimensions of the underlying list.
func (p *ModelPickerPanel) SetSize(w, h int) {
	p.list.SetSize(w, h)
}

// Update forwards a tea.Msg to the underlying list and returns any command.
func (p *ModelPickerPanel) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	p.list, cmd = p.list.Update(msg)
	return cmd
}

// SelectedValue returns the name of the currently highlighted item,
// or the fallback selected value if no item is highlighted.
func (p *ModelPickerPanel) SelectedValue() string {
	if item := p.list.SelectedItem(); item != nil {
		if mi, ok := item.(modelItem); ok {
			return mi.name
		}
	}
	return p.selected
}

// View renders the picker list.
func (p *ModelPickerPanel) View() string {
	return p.list.View()
}

// FilterState returns the current filter state of the underlying list.
func (p *ModelPickerPanel) FilterState() list.FilterState {
	return p.list.FilterState()
}
