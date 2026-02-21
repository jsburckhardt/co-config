package tui

import (
	"log/slog"
	"sort"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jsburckhardt/co-config/internal/config"
	"github.com/jsburckhardt/co-config/internal/copilot"
	"github.com/jsburckhardt/co-config/internal/sensitive"
)

// Model is the main Bubbletea model for the two-panel TUI
type Model struct {
	cfg          *config.Config
	schema       []copilot.SchemaField
	version      string
	configPath   string
	
	state        State
	list         list.Model
	detailPanel  DetailPanel
	help         help.Model
	keys         KeyMap
	
	windowWidth  int
	windowHeight int
	err          error
	saved        bool
}

// NewModel creates a new TUI model with two-panel layout
func NewModel(cfg *config.Config, schema []copilot.SchemaField, version, configPath string) *Model {
	// Create list items from schema
	items := buildListItems(cfg, schema)
	
	// Create list model
	delegate := NewItemDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()
	
	// Create detail panel
	detail := NewDetailPanel(0, 0)
	
	// Set initial field if list is not empty
	if len(items) > 0 {
		if item, ok := items[0].(ConfigItem); ok {
			detail.SetField(item.Field, item.Value)
		}
	}
	
	return &Model{
		cfg:         cfg,
		schema:      schema,
		version:     version,
		configPath:  configPath,
		state:       StateBrowsing,
		list:        l,
		detailPanel: detail,
		help:        help.New(),
		keys:        DefaultKeyMap(),
	}
}

// buildListItems creates list items from config and schema
func buildListItems(cfg *config.Config, schema []copilot.SchemaField) []list.Item {
	// Categorize fields
	categories := map[string][]ConfigItem{
		"Model & AI":          {},
		"URLs & Permissions":  {},
		"Display":            {},
		"General":            {},
		"Sensitive":          {},
	}
	
	// Sort schema by name for consistent ordering
	sorted := make([]copilot.SchemaField, len(schema))
	copy(sorted, schema)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
	
	for _, sf := range sorted {
		value := cfg.Get(sf.Name)
		item := ConfigItem{Field: sf, Value: value}
		
		// Categorize
		isSensitive := sensitive.IsSensitive(sf.Name)
		var isTokenLike bool
		if strVal, ok := value.(string); ok {
			isTokenLike = sensitive.LooksLikeToken(strVal)
		}
		
		if isSensitive || isTokenLike {
			categories["Sensitive"] = append(categories["Sensitive"], item)
		} else if isModelField(sf.Name) {
			categories["Model & AI"] = append(categories["Model & AI"], item)
		} else if isURLField(sf.Name) {
			categories["URLs & Permissions"] = append(categories["URLs & Permissions"], item)
		} else if isDisplayField(sf.Name) {
			categories["Display"] = append(categories["Display"], item)
		} else {
			categories["General"] = append(categories["General"], item)
		}
	}
	
	// Build final list with group headers
	var items []list.Item
	categoryOrder := []string{"Model & AI", "Display", "URLs & Permissions", "General", "Sensitive"}
	
	for _, catName := range categoryOrder {
		catItems := categories[catName]
		if len(catItems) > 0 {
			items = append(items, GroupHeader{Name: catName})
			for _, item := range catItems {
				items = append(items, item)
			}
		}
	}
	
	return items
}

func isModelField(name string) bool {
	modelFields := []string{"model", "reasoning_effort", "parallel_tool_execution", "stream", "experimental"}
	for _, f := range modelFields {
		if f == name {
			return true
		}
	}
	return false
}

func isURLField(name string) bool {
	urlFields := []string{"allowed_urls", "denied_urls", "trusted_folders"}
	for _, f := range urlFields {
		if f == name {
			return true
		}
	}
	return name == "custom_agents" || len(name) > 13 && name[:13] == "custom_agents"
}

func isDisplayField(name string) bool {
	displayFields := []string{"theme", "alt_screen", "render_markdown", "screen_reader", 
		"banner", "beep", "update_terminal_title", "streamer_mode"}
	for _, f := range displayFields {
		if f == name {
			return true
		}
	}
	return false
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		m.updateSizes()
		return m, nil
		
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}
	
	// Update active component based on state
	var cmd tea.Cmd
	switch m.state {
	case StateBrowsing:
		m.list, cmd = m.list.Update(msg)
		// Update detail panel when selection changes
		m.updateDetailPanel()
	case StateEditing:
		cmd = m.detailPanel.Update(msg)
	}
	
	return m, cmd
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		slog.Info("user quit with ctrl+c", "state", m.state)
		return m, tea.Quit
		
	case "enter":
		if m.state == StateBrowsing {
			// Check if selected item is editable
			if item, ok := m.list.SelectedItem().(ConfigItem); ok {
				// Check if sensitive or token-like
				isSensitive := sensitive.IsSensitive(item.Field.Name)
				var isTokenLike bool
				if strVal, ok := item.Value.(string); ok {
					isTokenLike = sensitive.LooksLikeToken(strVal)
				}
				
				if !isSensitive && !isTokenLike {
					// Start editing
					m.state = StateEditing
					slog.Info("entering edit mode", "field", item.Field.Name)
					return m, m.detailPanel.StartEditing()
				}
			}
		}
		
	case "esc":
		if m.state == StateEditing {
			// Save and return to browsing
			newValue := m.detailPanel.StopEditing()
			if item, ok := m.list.SelectedItem().(ConfigItem); ok {
				slog.Info("saving field value", "field", item.Field.Name, "value", newValue)
				m.cfg.Set(item.Field.Name, newValue)
				// Update the list item value
				m.updateListItemValue(item.Field.Name, newValue)
			}
			m.state = StateBrowsing
			slog.Info("returned to browsing mode")
		}
		
	case "ctrl+s":
		// Save config to file
		slog.Info("saving config", "path", m.configPath)
		if err := config.SaveConfig(m.configPath, m.cfg); err != nil {
			m.err = err
			slog.Error("failed to save config", "error", err)
		} else {
			m.saved = true
			slog.Info("config saved successfully")
		}
	}
	
	// Let active component handle the key if we didn't handle it
	var cmd tea.Cmd
	switch m.state {
	case StateBrowsing:
		m.list, cmd = m.list.Update(msg)
		m.updateDetailPanel()
	case StateEditing:
		cmd = m.detailPanel.Update(msg)
	}
	
	return m, cmd
}

func (m *Model) updateDetailPanel() {
	if item, ok := m.list.SelectedItem().(ConfigItem); ok {
		m.detailPanel.SetField(item.Field, item.Value)
	}
}

func (m *Model) updateListItemValue(fieldName string, newValue any) {
	items := m.list.Items()
	for i, item := range items {
		if configItem, ok := item.(ConfigItem); ok {
			if configItem.Field.Name == fieldName {
				configItem.Value = newValue
				m.list.SetItem(i, configItem)
				break
			}
		}
	}
}

func (m *Model) updateSizes() {
	// Header takes 3 lines, help bar takes 2 lines
	availableHeight := m.windowHeight - 5
	
	// 30% left panel, 70% right panel
	leftWidth := int(float64(m.windowWidth) * 0.3)
	rightWidth := m.windowWidth - leftWidth - 2
	
	// Account for borders and padding
	listWidth := leftWidth - 6
	listHeight := availableHeight - 4
	
	detailWidth := rightWidth - 6
	detailHeight := availableHeight - 4
	
	m.list.SetSize(listWidth, listHeight)
	m.detailPanel.SetSize(detailWidth, detailHeight)
}

// View renders the model
func (m *Model) View() string {
	if m.saved && m.state != StateEditing && m.state != StateBrowsing {
		return successStyle.Render("✅ Configuration saved to " + m.configPath + "\nPress Ctrl+C to exit.")
	}
	if m.err != nil {
		return errorMessageStyle.Render("❌ Error: " + m.err.Error() + "\nPress Ctrl+C to exit.")
	}
	
	// Header
	header := headerStyle.Render("🤖 ccc — Copilot Config CLI")
	if m.version != "" {
		header += " " + versionStyle.Render("(Copilot CLI v"+m.version+")")
	}
	
	// Calculate panel dimensions
	availableHeight := m.windowHeight - 5
	leftWidth := int(float64(m.windowWidth) * 0.3)
	rightWidth := m.windowWidth - leftWidth - 2
	
	// Render panels
	listPanel := m.list.View()
	detailPanelView := m.detailPanel.View()
	
	// Apply panel styles with focus
	if m.state == StateBrowsing {
		listPanel = focusedBorderStyle.Width(leftWidth-4).Height(availableHeight-2).Render(listPanel)
		detailPanelView = detailPanelStyle.Width(rightWidth-4).Height(availableHeight-2).Render(detailPanelView)
	} else {
		listPanel = listPanelStyle.Width(leftWidth-4).Height(availableHeight-2).Render(listPanel)
		detailPanelView = focusedBorderStyle.Width(rightWidth-4).Height(availableHeight-2).Render(detailPanelView)
	}
	
	// Combine panels horizontally
	panels := lipgloss.JoinHorizontal(lipgloss.Top, listPanel, detailPanelView)
	
	// Help bar
	helpKeys := m.keys.ShortHelp(m.state)
	helpView := m.help.ShortHelpView(helpKeys)
	helpBar := helpStyle.Width(m.windowWidth).Render(helpView)
	
	// Combine everything vertically
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		panels,
		helpBar,
	)
}
