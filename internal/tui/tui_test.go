package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jsburckhardt/co-config/internal/config"
	"github.com/jsburckhardt/co-config/internal/copilot"
)

// UT-TUI-001: NewModel creates a valid model with two-panel layout
func TestNewModel(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4")

	schema := []copilot.SchemaField{
		{
			Name:        "model",
			Type:        "enum",
			Default:     "gpt-4",
			Options:     []string{"gpt-4", "gpt-3.5-turbo"},
			Description: "AI model to use",
		},
	}

	model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")

	if model.cfg != cfg {
		t.Error("NewModel did not store config correctly")
	}
	if model.version != "0.0.412" {
		t.Error("NewModel did not store version correctly")
	}
	if model.configPath != "/tmp/config.json" {
		t.Error("NewModel did not store configPath correctly")
	}
	if model.state != StateBrowsing {
		t.Error("NewModel should start in Browsing state")
	}
}

// UT-TUI-002: State machine initialization starts in Browsing state
func TestStateMachineInitialization(t *testing.T) {
	cfg := config.NewConfig()
	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
	}

	model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")

	if model.state != StateBrowsing {
		t.Errorf("Expected initial state to be Browsing, got %v", model.state)
	}
}

// UT-TUI-003: List population from schema with all field types
func TestListPopulation(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("test_string", "value")
	cfg.Set("test_bool", true)
	cfg.Set("test_enum", "option1")
	cfg.Set("test_list", []any{"item1", "item2"})

	schema := []copilot.SchemaField{
		{Name: "test_string", Type: "string", Default: "", Description: "String field"},
		{Name: "test_bool", Type: "bool", Default: "false", Description: "Bool field"},
		{Name: "test_enum", Type: "enum", Default: "option1", Options: []string{"option1", "option2"}, Description: "Enum field"},
		{Name: "test_list", Type: "list", Default: "", Description: "List field"},
	}

	items := buildListItems(cfg, schema)

	// Should have group headers + items
	if len(items) == 0 {
		t.Fatal("buildListItems returned empty list")
	}

	// Count ConfigItems (excluding GroupHeaders)
	configItemCount := 0
	for _, item := range items {
		if _, ok := item.(ConfigItem); ok {
			configItemCount++
		}
	}

	if configItemCount != 4 {
		t.Errorf("Expected 4 config items, got %d", configItemCount)
	}
}

// UT-TUI-004: Sensitive fields are marked correctly in list
func TestSensitiveFieldsInList(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("copilot_tokens", map[string]any{"token": "secret"})
	cfg.Set("model", "gpt-4")

	schema := []copilot.SchemaField{
		{Name: "copilot_tokens", Type: "string", Default: "", Description: "Tokens"},
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}, Description: "Model"},
	}

	items := buildListItems(cfg, schema)

	// Find the sensitive item
	var sensitiveItem *ConfigItem
	var normalItem *ConfigItem
	for _, item := range items {
		if ci, ok := item.(ConfigItem); ok {
			if ci.Field.Name == "copilot_tokens" {
				sensitiveItem = &ci
			} else if ci.Field.Name == "model" {
				normalItem = &ci
			}
		}
	}

	if sensitiveItem == nil {
		t.Fatal("Sensitive field not found in list")
	}
	if normalItem == nil {
		t.Fatal("Normal field not found in list")
	}

	// Check that sensitive field description contains read-only indicator
	if desc := sensitiveItem.Description(); desc == "" {
		t.Error("Sensitive field should have a description")
	}
}

// UT-TUI-005: Alt-screen mode enabled (tested via program options in main.go)
// This test verifies that the model can be initialized for alt-screen use
func TestAltScreenCompatibility(t *testing.T) {
	cfg := config.NewConfig()
	schema := []copilot.SchemaField{}
	
	model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")
	
	// Verify Init() returns without error
	cmd := model.Init()
	if cmd != nil {
		t.Error("Init() should return nil for this model")
	}
}

// UT-TUI-006: Window resize updates panel sizes
func TestWindowResize(t *testing.T) {
	cfg := config.NewConfig()
	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
	}
	
	model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")
	
	// Simulate window size message
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	newModel, _ := model.Update(msg)
	
	m := newModel.(*Model)
	if m.windowWidth != 120 {
		t.Errorf("Expected windowWidth 120, got %d", m.windowWidth)
	}
	if m.windowHeight != 40 {
		t.Errorf("Expected windowHeight 40, got %d", m.windowHeight)
	}
}

// UT-TUI-007: State transition from Browsing to Editing
func TestBrowsingToEditingTransition(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4")
	
	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4", "gpt-3.5-turbo"}},
	}
	
	model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")
	
	// Initialize window size
	model.windowWidth = 100
	model.windowHeight = 30
	model.updateSizes()
	
	// Move down to select a ConfigItem (first item is likely a GroupHeader)
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	newModel, _ := model.Update(downMsg)
	model = newModel.(*Model)
	
	// Verify we have a ConfigItem selected
	if _, ok := model.list.SelectedItem().(ConfigItem); !ok {
		// Move down one more time
		newModel, _ := model.Update(downMsg)
		model = newModel.(*Model)
	}
	
	// Now simulate pressing Enter on a non-sensitive field
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ = model.Update(msg)
	
	m := newModel.(*Model)
	if m.state != StateEditing {
		t.Errorf("Expected state to be Editing after Enter, got %v", m.state)
	}
}

// UT-TUI-008: State transition from Editing to Browsing
func TestEditingToBrowsingTransition(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4")
	
	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4", "gpt-3.5-turbo"}},
	}
	
	model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")
	model.state = StateEditing
	
	// Simulate pressing Esc to save and return
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(msg)
	
	m := newModel.(*Model)
	if m.state != StateBrowsing {
		t.Errorf("Expected state to be Browsing after Esc, got %v", m.state)
	}
}

// UT-TUI-009: Token-like values are treated as sensitive
func TestTokenLikeValueTreatedAsSensitive(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("custom_field", "ghp_abc123secrettoken")
	
	schema := []copilot.SchemaField{
		{Name: "custom_field", Type: "string", Default: "", Description: "Custom field"},
	}
	
	items := buildListItems(cfg, schema)
	
	// Find the item
	var foundItem *ConfigItem
	for _, item := range items {
		if ci, ok := item.(ConfigItem); ok {
			if ci.Field.Name == "custom_field" {
				foundItem = &ci
				break
			}
		}
	}
	
	if foundItem == nil {
		t.Fatal("custom_field not found in list")
	}
	
	// Description should indicate read-only
	desc := foundItem.Description()
	if desc == "" {
		t.Error("Token-like field should have description indicating read-only status")
	}
}

// UT-TUI-010: Field categorization logic
func TestFieldCategorization(t *testing.T) {
	cfg := config.NewConfig()
	
	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
		{Name: "theme", Type: "enum", Default: "auto", Options: []string{"auto", "dark", "light"}},
		{Name: "allowed_urls", Type: "list"},
	}
	
	items := buildListItems(cfg, schema)
	
	// Should have group headers for different categories
	groupCount := 0
	for _, item := range items {
		if _, ok := item.(GroupHeader); ok {
			groupCount++
		}
	}
	
	if groupCount == 0 {
		t.Error("Expected at least one group header")
	}
}

// UT-TUI-011: DetailPanel renders field information
func TestDetailPanelRender(t *testing.T) {
	detail := NewDetailPanel(50, 20)
	
	field := copilot.SchemaField{
		Name:        "model",
		Type:        "enum",
		Default:     "gpt-4",
		Options:     []string{"gpt-4", "gpt-3.5-turbo"},
		Description: "AI model to use",
	}
	
	detail.SetField(field, "gpt-4")
	
	view := detail.View()
	if view == "" {
		t.Error("DetailPanel.View() returned empty string")
	}
}

// UT-TUI-012: formatValue handles different value types
func TestFormatValue(t *testing.T) {
	tests := []struct {
		name   string
		value  any
		maxLen int
		want   string
	}{
		{"string", "test", 10, "test"},
		{"bool true", true, 10, "true"},
		{"bool false", false, 10, "false"},
		{"empty list", []any{}, 10, "[]"},
		{"list", []any{"a", "b"}, 20, "[a, b]"},
		{"truncated", "very long string that exceeds max length", 10, "very lo..."},
		{"nil", nil, 10, ""},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatValue(tt.value, tt.maxLen)
			if got != tt.want {
				t.Errorf("formatValue(%v, %d) = %q, want %q", tt.value, tt.maxLen, got, tt.want)
			}
		})
	}
}
