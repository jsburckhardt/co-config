package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/jsburckhardt/co-config/internal/config"
	"github.com/jsburckhardt/co-config/internal/copilot"
)

// Model is the Bubbletea model wrapping the Huh form.
type Model struct {
	form       *huh.Form
	result     *FormResult
	cfg        *config.Config
	schema     []copilot.SchemaField
	version    string
	configPath string
	saved      bool
	err        error
}

// NewModel creates a new TUI model.
func NewModel(cfg *config.Config, schema []copilot.SchemaField, version, configPath string) Model {
	form, result := BuildForm(cfg, schema)
	return Model{
		form:       form,
		result:     result,
		cfg:        cfg,
		schema:     schema,
		version:    version,
		configPath: configPath,
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle quit
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	// Delegate to form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	// Check if form completed
	if m.form.State == huh.StateCompleted {
		// Apply results back to config
		m.applyResults()
		// Save
		if err := config.SaveConfig(m.configPath, m.cfg); err != nil {
			m.err = err
		} else {
			m.saved = true
		}
		return m, tea.Quit
	}

	return m, cmd
}

// View renders the model.
func (m Model) View() string {
	if m.saved {
		return titleStyle.Render("✅ Configuration saved to " + m.configPath)
	}
	if m.err != nil {
		return titleStyle.Render("❌ Error: " + m.err.Error())
	}

	header := titleStyle.Render("ccc — Copilot Config CLI") + "\n"
	if m.version != "" {
		header += versionStyle.Render("Copilot CLI v" + m.version) + "\n"
	}
	header += "\n"

	return header + m.form.View()
}

func (m *Model) applyResults() {
	for key, valPtr := range m.result.Values {
		switch v := valPtr.(type) {
		case *bool:
			m.cfg.Set(key, *v)
		case *string:
			// Check if this was a list field - split by newline
			for _, sf := range m.schema {
				if sf.Name == key && sf.Type == "list" {
					lines := strings.Split(*v, "\n")
					var items []any
					for _, line := range lines {
						line = strings.TrimSpace(line)
						if line != "" {
							items = append(items, line)
						}
					}
					m.cfg.Set(key, items)
					goto next
				}
			}
			m.cfg.Set(key, *v)
		next:
		}
	}
}
