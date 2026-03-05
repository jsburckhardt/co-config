package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jsburckhardt/co-config/internal/config"
	"github.com/jsburckhardt/co-config/internal/copilot"
)

// UT-TUI-001: NewModel creates a valid model with two-panel layout
func TestNewModel(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4")

	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4", "gpt-3.5-turbo"}, Description: "AI model"},
	}

	model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")

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

	model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")

	if model.state != StateBrowsing {
		t.Errorf("Expected initial state Browsing, got %v", model.state)
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
		{Name: "test_enum", Type: "enum", Default: "option1", Options: []string{"option1", "option2"}},
		{Name: "test_list", Type: "list", Default: "", Description: "List field"},
	}

	entries := buildEntries(cfg, schema)

	configItemCount := 0
	for _, e := range entries {
		if !e.isHeader {
			configItemCount++
		}
	}

	if configItemCount != 4 {
		t.Errorf("Expected 4 config items, got %d", configItemCount)
	}
}

// UT-TUI-004: Sensitive fields are categorized under Sensitive group
func TestSensitiveFieldsInList(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("copilot_tokens", map[string]any{"token": "secret"})
	cfg.Set("model", "gpt-4")

	schema := []copilot.SchemaField{
		{Name: "copilot_tokens", Type: "string", Default: "", Description: "Tokens"},
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}, Description: "Model"},
	}

	entries := buildEntries(cfg, schema)

	var foundSensitive, foundModel bool
	for _, e := range entries {
		if !e.isHeader {
			if e.item.Field.Name == "copilot_tokens" {
				foundSensitive = true
			}
			if e.item.Field.Name == "model" {
				foundModel = true
			}
		}
	}

	if !foundSensitive {
		t.Error("Sensitive field not found in entries")
	}
	if !foundModel {
		t.Error("Model field not found in entries")
	}
}

// UT-TUI-005: Alt-screen compatibility
func TestAltScreenCompatibility(t *testing.T) {
	cfg := config.NewConfig()
	model := NewModel(cfg, []copilot.SchemaField{}, nil, "0.0.412", "/tmp/config.json")

	cmd := model.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

// UT-TUI-006: Window resize updates panel sizes
func TestWindowResize(t *testing.T) {
	cfg := config.NewConfig()
	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
	}

	model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")

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

	model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")
	model.windowWidth = 100
	model.windowHeight = 30
	model.updateSizes()

	// Cursor starts on first ConfigItem (skips GroupHeader)
	if item := model.listPanel.SelectedItem(); item == nil {
		t.Fatal("No item selected initially")
	}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)

	m := newModel.(*Model)
	if m.state != StateEditing {
		t.Errorf("Expected Editing state after Enter, got %v", m.state)
	}
}

// UT-TUI-008: State transition from Editing to Browsing
func TestEditingToBrowsingTransition(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4")

	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4", "gpt-3.5-turbo"}},
	}

	model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")
	model.state = StateEditing

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ := model.Update(msg)

	m := newModel.(*Model)
	if m.state != StateBrowsing {
		t.Errorf("Expected Browsing state after Esc, got %v", m.state)
	}
}

// UT-TUI-009: Token-like values are treated as sensitive
func TestTokenLikeValueTreatedAsSensitive(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("custom_field", "ghp_abc123secrettoken")

	schema := []copilot.SchemaField{
		{Name: "custom_field", Type: "string", Default: "", Description: "Custom field"},
	}

	entries := buildEntries(cfg, schema)

	// Should be in Sensitive category
	var inSensitive bool
	var sawSensitiveHeader bool
	for _, e := range entries {
		if e.isHeader && e.header == "Sensitive" {
			sawSensitiveHeader = true
		}
		if sawSensitiveHeader && !e.isHeader && e.item.Field.Name == "custom_field" {
			inSensitive = true
			break
		}
	}

	if !inSensitive {
		t.Error("Token-like value field should be in Sensitive category")
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

	entries := buildEntries(cfg, schema)

	headerCount := 0
	for _, e := range entries {
		if e.isHeader {
			headerCount++
		}
	}

	if headerCount < 2 {
		t.Errorf("Expected at least 2 group headers, got %d", headerCount)
	}
}

// UT-TUI-011: DetailPanel renders field information
func TestDetailPanelRender(t *testing.T) {
	detail := NewDetailPanel()
	detail.SetSize(50, 20)

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

// UT-TUI-012: formatValueCompact handles different value types
func TestFormatValueCompact(t *testing.T) {
	tests := []struct {
		name       string
		value      any
		defaultVal string
		maxLen     int
		want       string
	}{
		{"string", "test", "", 10, "test"},
		{"bool true", true, "", 10, "true"},
		{"bool false", false, "", 10, "false"},
		{"empty list", []any{}, "", 10, "(empty)"},
		{"list", []any{"a", "b"}, "", 20, "(2 items)"},
		{"truncated", "very long string that exceeds max length", "", 10, "very lo..."},
		{"nil", nil, "", 10, "(not set)"},
		// New test cases for default-value display (WI-0004)
		{"nil with default", nil, "auto", 20, "auto (default)"},
		{"nil with bool default", nil, "false", 20, "false (default)"},
		{"nil with long default truncated", nil, "very-long-default", 10, "very-lo..."},
		{"nil no default", nil, "", 10, "(not set)"},
		{"non-nil ignores default", "custom", "auto", 20, "custom"},
		{"bool false ignores default", false, "false", 20, "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatValueCompact(tt.value, tt.defaultVal, tt.maxLen)
			if got != tt.want {
				t.Errorf("formatValueCompact(%v, %q, %d) = %q, want %q", tt.value, tt.defaultVal, tt.maxLen, got, tt.want)
			}
		})
	}
}

// UT-TUI-013: ListPanel skips group headers during navigation
func TestListPanelSkipsHeaders(t *testing.T) {
	entries := []listEntry{
		{isHeader: true, header: "Group A"},
		{item: ConfigItem{Field: copilot.SchemaField{Name: "field1"}}},
		{item: ConfigItem{Field: copilot.SchemaField{Name: "field2"}}},
		{isHeader: true, header: "Group B"},
		{item: ConfigItem{Field: copilot.SchemaField{Name: "field3"}}},
	}

	lp := NewListPanel(entries)

	// Should start on field1
	if item := lp.SelectedItem(); item == nil || item.Field.Name != "field1" {
		t.Fatalf("Expected cursor on field1, got %v", lp.SelectedItem())
	}

	lp.Down()
	if item := lp.SelectedItem(); item == nil || item.Field.Name != "field2" {
		t.Errorf("Expected cursor on field2 after Down, got %v", lp.SelectedItem())
	}

	// Down again should skip header and land on field3
	lp.Down()
	if item := lp.SelectedItem(); item == nil || item.Field.Name != "field3" {
		t.Errorf("Expected cursor on field3 after Down (skipping header), got %v", lp.SelectedItem())
	}

	// Up should skip header and land on field2
	lp.Up()
	if item := lp.SelectedItem(); item == nil || item.Field.Name != "field2" {
		t.Errorf("Expected cursor on field2 after Up (skipping header), got %v", lp.SelectedItem())
	}
}

// UT-TUI-014: Copilot icon constant structure
func TestCopilotIconConstant(t *testing.T) {
	lines := strings.Split(copilotIcon, "\n")
	if len(lines) != 4 {
		t.Errorf("copilotIcon should have 4 lines, got %d", len(lines))
	}
	for i, line := range lines {
		w := len([]rune(line))
		if w != 6 {
			t.Errorf("copilotIcon line %d width: got %d runes, want 6 (line=%q)", i, w, line)
		}
	}
}

// UT-TUI-015: View no longer contains gear emoji
func TestViewNoGearEmoji(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4")
	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
	}
	model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")
	model.windowWidth = 100
	model.windowHeight = 30
	model.updateSizes()

	view := model.View()
	if strings.Contains(view, "⚙") {
		t.Error("View() should not contain the gear emoji after icon replacement")
	}
}

// UT-TUI-016: View renders with ASCII icon and title
func TestViewRendersWithIcon(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4")
	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
	}
	model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")
	model.windowWidth = 100
	model.windowHeight = 30
	model.updateSizes()

	view := model.View()
	if !strings.Contains(view, "╭─╮╭─╮") {
		t.Error("View() should contain the first line of the copilot icon")
	}
	if !strings.Contains(view, "ccc") {
		t.Error("View() should contain the title text")
	}
	if !strings.Contains(view, "0.0.412") {
		t.Error("View() should contain the version string")
	}
}

// UT-TUI-017: View renders framed header with border characters
func TestViewRendersFramedHeader(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4")
	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
	}
	model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")
	model.windowWidth = 100
	model.windowHeight = 30
	model.updateSizes()

	view := model.View()
	// outer frame + header frame + help bar frame + 2 panels = 5 top-left corners minimum
	topLeftCount := strings.Count(view, "╭")
	if topLeftCount < 5 {
		t.Errorf("Expected at least 5 top-left corners for framed sections, got %d", topLeftCount)
	}
}

// UT-TUI-018: View renders at 80x24 without panic
func TestViewRendersAt80x24(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4")
	cfg.Set("theme", "dark")
	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
		{Name: "theme", Type: "enum", Default: "auto", Options: []string{"auto", "dark"}},
	}
	model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")
	model.windowWidth = 80
	model.windowHeight = 24
	model.updateSizes()

	view := model.View()
	if view == "" {
		t.Error("View() at 80x24 returned empty string")
	}
}

// UT-TUI-019: Panel height overhead arithmetic at various sizes
func TestPanelHeightOverhead(t *testing.T) {
	tests := []struct {
		name          string
		width, height int
	}{
		{"80x24", 80, 24},
		{"120x40", 120, 40},
		{"100x30", 100, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.NewConfig()
			schema := []copilot.SchemaField{
				{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
			}
			model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")
			msg := tea.WindowSizeMsg{Width: tt.width, Height: tt.height}
			newModel, _ := model.Update(msg)
			m := newModel.(*Model)

			view := m.View()
			if view == "" {
				t.Errorf("View() at %dx%d returned empty string", tt.width, tt.height)
			}
		})
	}
}

// UT-TUI-020: Small terminal floor guard prevents panic
func TestSmallTerminalFloor(t *testing.T) {
	tests := []struct {
		name          string
		width, height int
	}{
		{"40x15 - at floor", 40, 15},
		{"30x12 - below floor", 30, 12},
		{"20x10 - very small", 20, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.NewConfig()
			schema := []copilot.SchemaField{
				{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
			}
			model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")
			msg := tea.WindowSizeMsg{Width: tt.width, Height: tt.height}
			newModel, _ := model.Update(msg)
			m := newModel.(*Model)

			view := m.View()
			if view == "" {
				t.Errorf("View() at %dx%d returned empty string", tt.width, tt.height)
			}
		})
	}
}

// UT-TUI-021: Detail panel shows default annotation for unset field with default
func TestDetailPanelRenderUnsetWithDefault(t *testing.T) {
	detail := NewDetailPanel()
	detail.SetSize(50, 20)

	field := copilot.SchemaField{
		Name:    "theme",
		Type:    "enum",
		Default: "auto",
		Options: []string{"auto", "dark", "light"},
	}

	detail.SetField(field, nil)
	view := detail.View()

	if !strings.Contains(view, "auto") {
		t.Error("Expected detail panel to show default value 'auto' for unset field")
	}
	if !strings.Contains(view, "default") {
		t.Error("Expected detail panel to show '(default)' annotation for unset field")
	}
}

// UT-TUI-022: Detail panel shows "(not set)" for unset field without default
func TestDetailPanelRenderUnsetNoDefault(t *testing.T) {
	detail := NewDetailPanel()
	detail.SetSize(50, 20)

	field := copilot.SchemaField{
		Name:    "model",
		Type:    "enum",
		Default: "",
		Options: []string{"gpt-4", "claude-sonnet-4"},
	}

	detail.SetField(field, nil)
	view := detail.View()

	if !strings.Contains(view, "not set") {
		t.Error("Expected detail panel to show '(not set)' for unset field without default")
	}
	if strings.Contains(view, "(default)") {
		t.Error("Detail panel should NOT show '(default)' annotation when no default exists")
	}
}

// UT-TUI-023: Detail panel shows set value without default annotation
func TestDetailPanelRenderSetValueIgnoresDefault(t *testing.T) {
	detail := NewDetailPanel()
	detail.SetSize(50, 20)

	field := copilot.SchemaField{
		Name:    "theme",
		Type:    "enum",
		Default: "auto",
		Options: []string{"auto", "dark", "light"},
	}

	detail.SetField(field, "dark")
	view := detail.View()

	if !strings.Contains(view, "dark") {
		t.Error("Expected detail panel to show set value 'dark'")
	}
	if strings.Contains(view, "(default)") {
		t.Error("Detail panel should NOT show '(default)' when value is explicitly set")
	}
}

// UT-TUI-024: Detail panel with nil field does not panic
func TestDetailPanelNilFieldNoPanic(t *testing.T) {
	detail := NewDetailPanel()
	detail.SetSize(50, 20)

	// Do NOT call SetField — field is nil
	view := detail.View()
	if !strings.Contains(view, "Select a field") {
		t.Error("Expected placeholder text when no field is selected")
	}
}

// UT-TUI-026: StateEnvVars String returns "EnvVars"
func TestStateEnvVarsString(t *testing.T) {
	got := StateEnvVars.String()
	if got != "EnvVars" {
		t.Errorf("StateEnvVars.String() = %q, want %q", got, "EnvVars")
	}
}

// UT-TUI-027: StateEnvVars has distinct value from other states
func TestStateEnvVarsDistinct(t *testing.T) {
	if StateEnvVars == StateBrowsing {
		t.Error("StateEnvVars should differ from StateBrowsing")
	}
	if StateEnvVars == StateEditing {
		t.Error("StateEnvVars should differ from StateEditing")
	}
	if StateEnvVars == StateSaving {
		t.Error("StateEnvVars should differ from StateSaving")
	}
	if StateEnvVars == StateExiting {
		t.Error("StateEnvVars should differ from StateExiting")
	}
}

// UT-TUI-025: View renders without panicking
func TestViewRenders(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4")
	cfg.Set("theme", "dark")

	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
		{Name: "theme", Type: "enum", Default: "auto", Options: []string{"auto", "dark"}},
	}

	model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")
	model.windowWidth = 100
	model.windowHeight = 30
	model.updateSizes()

	view := model.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
}

// UT-TUI-028: DefaultKeyMap Left binding has correct keys
func TestDefaultKeyMapLeftBinding(t *testing.T) {
	km := DefaultKeyMap()

	// Test that Left responds to "left" arrow key
	leftMsg := tea.KeyMsg{Type: tea.KeyLeft}
	if !key.Matches(leftMsg, km.Left) {
		t.Error("Left binding should match tea.KeyLeft")
	}

	// Test that Left responds to "h" key
	hMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	if !key.Matches(hMsg, km.Left) {
		t.Error("Left binding should match 'h' key")
	}

	// Verify help text
	if km.Left.Help().Desc != "config" {
		t.Errorf("Left help desc = %q, want %q", km.Left.Help().Desc, "config")
	}
}

// UT-TUI-029: DefaultKeyMap Right binding has correct keys
func TestDefaultKeyMapRightBinding(t *testing.T) {
	km := DefaultKeyMap()

	// Test that Right responds to "right" arrow key
	rightMsg := tea.KeyMsg{Type: tea.KeyRight}
	if !key.Matches(rightMsg, km.Right) {
		t.Error("Right binding should match tea.KeyRight")
	}

	// Test that Right responds to "l" key
	lMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	if !key.Matches(lMsg, km.Right) {
		t.Error("Right binding should match 'l' key")
	}

	// Verify help text
	if km.Right.Help().Desc != "env vars" {
		t.Errorf("Right help desc = %q, want %q", km.Right.Help().Desc, "env vars")
	}
}

// UT-TUI-030: DefaultKeyMap Tab help text is "switch view"
func TestDefaultKeyMapTabHelpText(t *testing.T) {
	km := DefaultKeyMap()

	if km.Tab.Help().Desc != "switch view" {
		t.Errorf("Tab help desc = %q, want %q", km.Tab.Help().Desc, "switch view")
	}
}

// UT-TUI-031: NewEnvVarsPanel with non-empty slice starts cursor at 0
func TestNewEnvVarsPanelCursorStart(t *testing.T) {
	envVars := []copilot.EnvVarInfo{
		{Names: []string{"VAR_A"}, Description: "First"},
		{Names: []string{"VAR_B"}, Description: "Second"},
		{Names: []string{"VAR_C"}, Description: "Third"},
	}
	panel := NewEnvVarsPanel(envVars)
	if panel.Cursor() != 0 {
		t.Errorf("Expected cursor at 0, got %d", panel.Cursor())
	}
}

// UT-TUI-032: NewEnvVarsPanel with nil/empty slice does not panic
func TestNewEnvVarsPanelEmptyNoPanic(t *testing.T) {
	// nil slice
	panel1 := NewEnvVarsPanel(nil)
	if panel1 == nil {
		t.Error("NewEnvVarsPanel(nil) returned nil")
	}

	// empty slice
	panel2 := NewEnvVarsPanel([]copilot.EnvVarInfo{})
	if panel2 == nil {
		t.Error("NewEnvVarsPanel(empty) returned nil")
	}
}

// UT-TUI-033: EnvVarsPanel View renders primary name for each entry
func TestEnvVarsPanelViewPrimaryNames(t *testing.T) {
	envVars := []copilot.EnvVarInfo{
		{Names: []string{"COPILOT_MODEL"}, Description: "Model setting"},
		{Names: []string{"XDG_CONFIG_HOME"}, Description: "Config home"},
	}
	panel := NewEnvVarsPanel(envVars)
	panel.SetSize(80, 40)

	view := panel.View()
	if !strings.Contains(view, "COPILOT_MODEL") {
		t.Error("View should contain COPILOT_MODEL")
	}
	if !strings.Contains(view, "XDG_CONFIG_HOME") {
		t.Error("View should contain XDG_CONFIG_HOME")
	}
}

// UT-TUI-034: EnvVarsPanel View renders alias names for multi-name entries
func TestEnvVarsPanelViewAliasNames(t *testing.T) {
	envVars := []copilot.EnvVarInfo{
		{Names: []string{"COPILOT_EDITOR", "VISUAL", "EDITOR"}, Description: "Editor"},
	}
	panel := NewEnvVarsPanel(envVars)
	panel.SetSize(80, 40)

	view := panel.View()
	if !strings.Contains(view, "COPILOT_EDITOR") {
		t.Error("View should contain primary name COPILOT_EDITOR")
	}
	if !strings.Contains(view, "VISUAL") {
		t.Error("View should contain alias VISUAL")
	}
	if !strings.Contains(view, "EDITOR") {
		t.Error("View should contain alias EDITOR")
	}
}

// UT-TUI-035: EnvVarsPanel View shows masked value for sensitive set env var
func TestEnvVarsPanelViewSensitiveSet(t *testing.T) {
	envVars := []copilot.EnvVarInfo{
		{Names: []string{"COPILOT_GITHUB_TOKEN"}, Description: "GitHub token"},
	}
	t.Setenv("COPILOT_GITHUB_TOKEN", "ghp_secret123")

	panel := NewEnvVarsPanel(envVars)
	panel.SetSize(80, 40)

	view := panel.View()
	if !strings.Contains(view, "🔒") {
		t.Error("View should contain lock emoji for sensitive set env var")
	}
	if strings.Contains(view, "ghp_secret123") {
		t.Error("View should NOT contain the raw token value")
	}
}

// UT-TUI-036: EnvVarsPanel View shows "(not set)" for unset env var
func TestEnvVarsPanelViewUnset(t *testing.T) {
	envVars := []copilot.EnvVarInfo{
		{Names: []string{"COPILOT_MODEL"}, Description: "Model"},
	}
	// Ensure COPILOT_MODEL is not set
	t.Setenv("COPILOT_MODEL", "")
	// os.Unsetenv within t.Setenv scope: re-unset to guarantee empty
	// Actually t.Setenv sets it to "", os.Getenv returns "" which is treated as unset
	// in our logic (we check for non-empty value)

	panel := NewEnvVarsPanel(envVars)
	panel.SetSize(80, 40)

	view := panel.View()
	if !strings.Contains(view, "not set") {
		t.Error("View should contain 'not set' for unset env var")
	}
}

// UT-TUI-037: EnvVarsPanel View shows value for non-sensitive set env var
func TestEnvVarsPanelViewNonSensitiveSet(t *testing.T) {
	envVars := []copilot.EnvVarInfo{
		{Names: []string{"COPILOT_MODEL"}, Description: "Model"},
	}
	t.Setenv("COPILOT_MODEL", "gpt-4")

	panel := NewEnvVarsPanel(envVars)
	panel.SetSize(80, 40)

	view := panel.View()
	if !strings.Contains(view, "gpt-4") {
		t.Error("View should contain the actual value 'gpt-4' for non-sensitive set env var")
	}
}

// UT-TUI-038: EnvVarsPanel Down advances cursor and Up retreats cursor
func TestEnvVarsPanelCursorNavigation(t *testing.T) {
	envVars := []copilot.EnvVarInfo{
		{Names: []string{"VAR_A"}, Description: "A"},
		{Names: []string{"VAR_B"}, Description: "B"},
		{Names: []string{"VAR_C"}, Description: "C"},
	}
	panel := NewEnvVarsPanel(envVars)
	panel.SetSize(80, 40)

	// Initial cursor at 0
	if panel.Cursor() != 0 {
		t.Fatalf("Expected initial cursor 0, got %d", panel.Cursor())
	}

	// Down → 1
	panel.Down()
	if panel.Cursor() != 1 {
		t.Errorf("After Down, expected cursor 1, got %d", panel.Cursor())
	}

	// Down → 2
	panel.Down()
	if panel.Cursor() != 2 {
		t.Errorf("After second Down, expected cursor 2, got %d", panel.Cursor())
	}

	// Down at end → stays at 2
	panel.Down()
	if panel.Cursor() != 2 {
		t.Errorf("Down at end should stay at 2, got %d", panel.Cursor())
	}

	// Up → 1
	panel.Up()
	if panel.Cursor() != 1 {
		t.Errorf("After Up, expected cursor 1, got %d", panel.Cursor())
	}

	// Up → 0
	panel.Up()
	if panel.Cursor() != 0 {
		t.Errorf("After second Up, expected cursor 0, got %d", panel.Cursor())
	}

	// Up at start → stays at 0
	panel.Up()
	if panel.Cursor() != 0 {
		t.Errorf("Up at start should stay at 0, got %d", panel.Cursor())
	}
}

// UT-TUI-039: EnvVarsPanel View renders qualifier text
func TestEnvVarsPanelViewQualifier(t *testing.T) {
	envVars := []copilot.EnvVarInfo{
		{
			Names:       []string{"COPILOT_GITHUB_TOKEN", "GH_TOKEN"},
			Description: "Token for auth",
			Qualifier:   "in order of precedence",
		},
	}
	panel := NewEnvVarsPanel(envVars)
	panel.SetSize(80, 40)

	view := panel.View()
	if !strings.Contains(view, "in order of precedence") {
		t.Error("View should contain qualifier text 'in order of precedence'")
	}
}

// UT-TUI-040: Empty EnvVarsPanel renders without panic
func TestEnvVarsPanelEmptyRender(t *testing.T) {
	panel := NewEnvVarsPanel(nil)
	panel.SetSize(80, 40)

	view := panel.View()
	if view == "" {
		t.Error("Empty panel View() should return a non-empty string (placeholder message)")
	}
}

// UT-TUI-041: StateBrowsing + right key transitions to StateEnvVars
func TestBrowsingRightToEnvVars(t *testing.T) {
	cfg := config.NewConfig()
	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
	}
	envVars := []copilot.EnvVarInfo{
		{Names: []string{"COPILOT_MODEL"}, Description: "test"},
	}
	model := NewModel(cfg, schema, envVars, "0.0.412", "/tmp/config.json")
	model.windowWidth = 100
	model.windowHeight = 30
	model.updateSizes()

	msg := tea.KeyMsg{Type: tea.KeyRight}
	newModel, _ := model.Update(msg)
	m := newModel.(*Model)
	if m.state != StateEnvVars {
		t.Errorf("Expected StateEnvVars after right key, got %v", m.state)
	}
}

// UT-TUI-042: StateBrowsing + "l" key transitions to StateEnvVars
func TestBrowsingLToEnvVars(t *testing.T) {
	cfg := config.NewConfig()
	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
	}
	envVars := []copilot.EnvVarInfo{
		{Names: []string{"COPILOT_MODEL"}, Description: "test"},
	}
	model := NewModel(cfg, schema, envVars, "0.0.412", "/tmp/config.json")

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	newModel, _ := model.Update(msg)
	m := newModel.(*Model)
	if m.state != StateEnvVars {
		t.Errorf("Expected StateEnvVars after 'l' key, got %v", m.state)
	}
}

// UT-TUI-043: StateBrowsing + tab key transitions to StateEnvVars
func TestBrowsingTabToEnvVars(t *testing.T) {
	cfg := config.NewConfig()
	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
	}
	envVars := []copilot.EnvVarInfo{
		{Names: []string{"COPILOT_MODEL"}, Description: "test"},
	}
	model := NewModel(cfg, schema, envVars, "0.0.412", "/tmp/config.json")

	msg := tea.KeyMsg{Type: tea.KeyTab}
	newModel, _ := model.Update(msg)
	m := newModel.(*Model)
	if m.state != StateEnvVars {
		t.Errorf("Expected StateEnvVars after tab key, got %v", m.state)
	}
}

// UT-TUI-044: StateEnvVars + left key transitions to StateBrowsing
func TestEnvVarsLeftToBrowsing(t *testing.T) {
	cfg := config.NewConfig()
	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
	}
	envVars := []copilot.EnvVarInfo{
		{Names: []string{"COPILOT_MODEL"}, Description: "test"},
	}
	model := NewModel(cfg, schema, envVars, "0.0.412", "/tmp/config.json")
	model.state = StateEnvVars

	msg := tea.KeyMsg{Type: tea.KeyLeft}
	newModel, _ := model.Update(msg)
	m := newModel.(*Model)
	if m.state != StateBrowsing {
		t.Errorf("Expected StateBrowsing after left key, got %v", m.state)
	}
}

// UT-TUI-045: StateEnvVars + "h" key transitions to StateBrowsing
func TestEnvVarsHToBrowsing(t *testing.T) {
	cfg := config.NewConfig()
	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
	}
	envVars := []copilot.EnvVarInfo{
		{Names: []string{"COPILOT_MODEL"}, Description: "test"},
	}
	model := NewModel(cfg, schema, envVars, "0.0.412", "/tmp/config.json")
	model.state = StateEnvVars

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}
	newModel, _ := model.Update(msg)
	m := newModel.(*Model)
	if m.state != StateBrowsing {
		t.Errorf("Expected StateBrowsing after 'h' key, got %v", m.state)
	}
}

// UT-TUI-046: StateEnvVars + tab key transitions to StateBrowsing
func TestEnvVarsTabToBrowsing(t *testing.T) {
	cfg := config.NewConfig()
	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
	}
	envVars := []copilot.EnvVarInfo{
		{Names: []string{"COPILOT_MODEL"}, Description: "test"},
	}
	model := NewModel(cfg, schema, envVars, "0.0.412", "/tmp/config.json")
	model.state = StateEnvVars

	msg := tea.KeyMsg{Type: tea.KeyTab}
	newModel, _ := model.Update(msg)
	m := newModel.(*Model)
	if m.state != StateBrowsing {
		t.Errorf("Expected StateBrowsing after tab key in EnvVars, got %v", m.state)
	}
}

// UT-TUI-047: StateEditing + right key does NOT transition to StateEnvVars
func TestEditingRightNoTransition(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4")
	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4", "gpt-3.5-turbo"}},
	}
	model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")
	model.state = StateEditing

	msg := tea.KeyMsg{Type: tea.KeyRight}
	newModel, _ := model.Update(msg)
	m := newModel.(*Model)
	if m.state != StateEditing {
		t.Errorf("Expected StateEditing after right key in editing, got %v", m.state)
	}
}

// UT-TUI-048: StateEditing + left key does NOT transition
func TestEditingLeftNoTransition(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4")
	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4", "gpt-3.5-turbo"}},
	}
	model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")
	model.state = StateEditing

	msg := tea.KeyMsg{Type: tea.KeyLeft}
	newModel, _ := model.Update(msg)
	m := newModel.(*Model)
	if m.state != StateEditing {
		t.Errorf("Expected StateEditing after left key in editing, got %v", m.state)
	}
}

// UT-TUI-049: ctrl+s in StateEnvVars does not trigger save
func TestEnvVarsCtrlSNoSave(t *testing.T) {
	cfg := config.NewConfig()
	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
	}
	envVars := []copilot.EnvVarInfo{
		{Names: []string{"COPILOT_MODEL"}, Description: "test"},
	}
	model := NewModel(cfg, schema, envVars, "0.0.412", "/tmp/config.json")
	model.state = StateEnvVars
	model.saved = false

	ctrlS := tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ := model.Update(ctrlS)
	m := newModel.(*Model)
	if m.saved {
		t.Error("ctrl+s in StateEnvVars should NOT trigger save")
	}
}

// UT-TUI-050: View in StateEnvVars renders env panel content
func TestViewInEnvVarsState(t *testing.T) {
	cfg := config.NewConfig()
	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
	}
	envVars := []copilot.EnvVarInfo{
		{Names: []string{"COPILOT_MODEL"}, Description: "optionally set the agent model"},
	}
	model := NewModel(cfg, schema, envVars, "0.0.412", "/tmp/config.json")
	model.windowWidth = 100
	model.windowHeight = 30
	model.updateSizes()
	model.state = StateEnvVars

	view := model.View()
	if !strings.Contains(view, "COPILOT_MODEL") {
		t.Error("View in StateEnvVars should contain env var name COPILOT_MODEL")
	}
}

// UT-TUI-051: ShortHelp for StateBrowsing includes right/tab binding
func TestShortHelpBrowsingIncludesRight(t *testing.T) {
	km := DefaultKeyMap()
	bindings := km.ShortHelp(StateBrowsing, "")
	found := false
	for _, b := range bindings {
		if b.Help().Desc == "env vars" {
			found = true
			break
		}
	}
	if !found {
		t.Error("ShortHelp(StateBrowsing) should include a binding with desc 'env vars'")
	}
}

// UT-TUI-052: ShortHelp for StateEnvVars includes left/tab and omits enter/save
func TestShortHelpEnvVarsBindings(t *testing.T) {
	km := DefaultKeyMap()
	bindings := km.ShortHelp(StateEnvVars, "")
	var foundConfig, foundEdit, foundSave bool
	for _, b := range bindings {
		desc := b.Help().Desc
		if desc == "config" {
			foundConfig = true
		}
		if desc == "edit" {
			foundEdit = true
		}
		if desc == "save" {
			foundSave = true
		}
	}
	if !foundConfig {
		t.Error("ShortHelp(StateEnvVars) should include binding with desc 'config'")
	}
	if foundEdit {
		t.Error("ShortHelp(StateEnvVars) should NOT include 'edit' binding")
	}
	if foundSave {
		t.Error("ShortHelp(StateEnvVars) should NOT include 'save' binding")
	}
}

// UT-TUI-053: ShortHelp for StateEditing remains unchanged
func TestShortHelpEditingUnchanged(t *testing.T) {
	km := DefaultKeyMap()
	bindings := km.ShortHelp(StateEditing, "")
	descs := make(map[string]bool)
	for _, b := range bindings {
		descs[b.Help().Desc] = true
	}
	if !descs["done"] {
		t.Error("ShortHelp(StateEditing) should include 'done' (Escape)")
	}
	if !descs["save"] {
		t.Error("ShortHelp(StateEditing) should include 'save'")
	}
	if !descs["quit"] {
		t.Error("ShortHelp(StateEditing) should include 'quit'")
	}
}

// UT-TUI-054: NewModel with nil envVars does not panic
func TestNewModelNilEnvVars(t *testing.T) {
	cfg := config.NewConfig()
	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
	}
	model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")
	if model == nil {
		t.Fatal("NewModel with nil envVars should not return nil")
	}
	if model.envPanel == nil {
		t.Fatal("NewModel with nil envVars should still create envPanel")
	}
}

// UT-TUI-055: Enter commits enum field and returns to Browsing
func TestEnterCommitsEnumField(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4")

	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4", "gpt-3.5-turbo"}},
	}

	model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")
	model.windowWidth = 100
	model.windowHeight = 30
	model.updateSizes()

	// Enter editing
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := newModel.(*Model)
	if m.state != StateEditing {
		t.Fatal("Expected StateEditing after Enter")
	}

	// Press Enter to commit (enum field, not list)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(*Model)
	if m.state != StateBrowsing {
		t.Errorf("Expected StateBrowsing after Enter on enum field, got %v", m.state)
	}
}

// UT-TUI-056: Enter commits string field and returns to Browsing
func TestEnterCommitsStringField(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("theme", "dark")

	schema := []copilot.SchemaField{
		{Name: "theme", Type: "string", Default: ""},
	}

	model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")
	model.windowWidth = 100
	model.windowHeight = 30
	model.updateSizes()

	// Enter editing
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := newModel.(*Model)
	if m.state != StateEditing {
		t.Fatal("Expected StateEditing after Enter")
	}

	// Press Enter to commit (string field)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(*Model)
	if m.state != StateBrowsing {
		t.Errorf("Expected StateBrowsing after Enter on string field, got %v", m.state)
	}
}

// UT-TUI-057: Enter commits bool field and returns to Browsing
func TestEnterCommitsBoolField(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("stream", true)

	schema := []copilot.SchemaField{
		{Name: "stream", Type: "bool", Default: "true"},
	}

	model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")
	model.windowWidth = 100
	model.windowHeight = 30
	model.updateSizes()

	// Enter editing
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := newModel.(*Model)
	if m.state != StateEditing {
		t.Fatal("Expected StateEditing after Enter")
	}

	// Press Enter to commit (bool field)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(*Model)
	if m.state != StateBrowsing {
		t.Errorf("Expected StateBrowsing after Enter on bool field, got %v", m.state)
	}
}

// UT-TUI-058: Enter on list field stays in Editing
func TestEnterOnListFieldStaysEditing(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("allowed_urls", []any{"https://example.com"})

	schema := []copilot.SchemaField{
		{Name: "allowed_urls", Type: "list"},
	}

	model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")
	model.windowWidth = 100
	model.windowHeight = 30
	model.updateSizes()

	// Enter editing
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := newModel.(*Model)
	if m.state != StateEditing {
		t.Fatal("Expected StateEditing after Enter")
	}

	// Press Enter on list field — should stay in editing (newline in textarea)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(*Model)
	if m.state != StateEditing {
		t.Errorf("Expected to remain in StateEditing for list field, got %v", m.state)
	}
}

// UT-TUI-059: Modified flag default and after UpdateItemValue
func TestModifiedFlagDefaultAndAfterUpdate(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4")
	cfg.Set("theme", "dark")

	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4", "gpt-3.5-turbo"}},
		{Name: "theme", Type: "string", Default: "auto"},
	}

	entries := buildEntries(cfg, schema)
	lp := NewListPanel(entries)

	for _, e := range lp.entries {
		if !e.isHeader && e.item.Modified {
			t.Errorf("Entry %q should have Modified == false initially", e.item.Field.Name)
		}
	}

	lp.UpdateItemValue("model", "gpt-3.5-turbo")

	for _, e := range lp.entries {
		if !e.isHeader && e.item.Field.Name == "model" {
			if !e.item.Modified {
				t.Error("Entry 'model' should have Modified == true after UpdateItemValue")
			}
		}
		if !e.isHeader && e.item.Field.Name == "theme" {
			if e.item.Modified {
				t.Error("Entry 'theme' should still have Modified == false")
			}
		}
	}
}

// UT-TUI-060: renderItem appends (not-saved) when Modified is true
func TestRenderItemNotSavedIndicator(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4")
	cfg.Set("theme", "dark")

	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4", "gpt-3.5-turbo"}},
		{Name: "theme", Type: "string", Default: "auto"},
	}

	entries := buildEntries(cfg, schema)
	lp := NewListPanel(entries)
	lp.SetSize(60, 20)

	lp.UpdateItemValue("model", "gpt-3.5-turbo")

	view := lp.View()

	if !strings.Contains(view, "(not-saved)") {
		t.Error("Expected '(not-saved)' indicator in rendered output for modified field")
	}

	lines := strings.Split(view, "\n")
	for _, line := range lines {
		if strings.Contains(line, "theme") && strings.Contains(line, "(not-saved)") {
			t.Error("Unmodified field 'theme' should NOT have '(not-saved)' indicator")
		}
	}
}

// UT-TUI-061: ClearAllModified resets all Modified flags
func TestClearAllModified(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4")
	cfg.Set("theme", "dark")

	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4", "gpt-3.5-turbo"}},
		{Name: "theme", Type: "string", Default: "auto"},
	}

	entries := buildEntries(cfg, schema)
	lp := NewListPanel(entries)

	lp.UpdateItemValue("model", "gpt-3.5-turbo")
	lp.UpdateItemValue("theme", "light")

	modCount := 0
	for _, e := range lp.entries {
		if !e.isHeader && e.item.Modified {
			modCount++
		}
	}
	if modCount < 2 {
		t.Fatalf("Expected at least 2 modified entries, got %d", modCount)
	}

	lp.ClearAllModified()

	for _, e := range lp.entries {
		if !e.isHeader && e.item.Modified {
			t.Errorf("Entry %q should have Modified == false after ClearAllModified", e.item.Field.Name)
		}
	}
}

// UT-TUI-062: Saved flag cleared after commit
func TestSavedFlagClearedAfterCommit(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4")

	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4", "gpt-3.5-turbo"}},
	}

	t.Run("Esc commit clears saved", func(t *testing.T) {
		model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")
		model.windowWidth = 100
		model.windowHeight = 30
		model.updateSizes()
		model.saved = true

		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m := newModel.(*Model)

		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		m = newModel.(*Model)
		if m.saved {
			t.Error("saved should be false after Esc commit")
		}
		if m.state != StateBrowsing {
			t.Error("should be in StateBrowsing after Esc")
		}
	})

	t.Run("Enter commit clears saved", func(t *testing.T) {
		model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")
		model.windowWidth = 100
		model.windowHeight = 30
		model.updateSizes()
		model.saved = true

		newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m := newModel.(*Model)

		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m = newModel.(*Model)
		if m.saved {
			t.Error("saved should be false after Enter commit")
		}
		if m.state != StateBrowsing {
			t.Error("should be in StateBrowsing after Enter")
		}
	})
}

// UT-TUI-063: Help bar shows enter confirm for non-list editing
func TestHelpBarEnterConfirmNonList(t *testing.T) {
	keys := DefaultKeyMap()
	bindings := keys.ShortHelp(StateEditing, "string")

	found := false
	for _, b := range bindings {
		h := b.Help()
		if h.Key == "enter" && strings.Contains(h.Desc, "confirm") {
			found = true
		}
	}
	if !found {
		t.Error("Expected 'enter' + 'confirm' binding in editing help for non-list field")
	}
}

// UT-TUI-064: Help bar omits enter confirm for list editing
func TestHelpBarNoEnterConfirmList(t *testing.T) {
	keys := DefaultKeyMap()
	bindings := keys.ShortHelp(StateEditing, "list")

	for _, b := range bindings {
		h := b.Help()
		if h.Key == "enter" && strings.Contains(h.Desc, "confirm") {
			t.Error("List editing should NOT show 'enter confirm' binding")
		}
	}
}

// UT-TUI-065: CurrentFieldType accessor returns correct type
func TestCurrentFieldType(t *testing.T) {
	dp := NewDetailPanel()
	dp.SetSize(50, 20)

	tests := []struct {
		name     string
		field    copilot.SchemaField
		value    any
		wantType string
	}{
		{"string", copilot.SchemaField{Name: "theme", Type: "string", Default: ""}, "dark", "string"},
		{"bool", copilot.SchemaField{Name: "stream", Type: "bool", Default: "true"}, true, "bool"},
		{"enum", copilot.SchemaField{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4", "gpt-3.5-turbo"}}, "gpt-4", "enum"},
		{"list", copilot.SchemaField{Name: "allowed_urls", Type: "list"}, []any{"https://example.com"}, "list"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dp.SetField(tt.field, tt.value)
			got := dp.CurrentFieldType()
			if got != tt.wantType {
				t.Errorf("CurrentFieldType() = %q, want %q", got, tt.wantType)
			}
		})
	}
}

// UT-TUI-066: CurrentFieldType returns empty string for nil field
func TestCurrentFieldTypeNilField(t *testing.T) {
	dp := NewDetailPanel()
	got := dp.CurrentFieldType()
	if got != "" {
		t.Errorf("CurrentFieldType() on fresh panel = %q, want empty string", got)
	}
}

// UT-TUI-067: (not-saved) not shown when Modified is false
func TestNotSavedNotShownWhenUnmodified(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4")
	cfg.Set("theme", "dark")

	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4", "gpt-3.5-turbo"}},
		{Name: "theme", Type: "string", Default: "auto"},
	}

	entries := buildEntries(cfg, schema)
	lp := NewListPanel(entries)
	lp.SetSize(60, 20)

	view := lp.View()

	if strings.Contains(view, "(not-saved)") {
		t.Error("No '(not-saved)' indicator should appear when no fields are modified")
	}
}

// UT-TUI-068: Narrow terminal with (not-saved) does not panic
func TestNarrowTerminalNotSavedNoPanic(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4")

	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4", "gpt-3.5-turbo"}},
	}

	entries := buildEntries(cfg, schema)
	lp := NewListPanel(entries)
	lp.SetSize(30, 10)

	lp.UpdateItemValue("model", "gpt-3.5-turbo")

	view := lp.View()
	if view == "" {
		t.Error("View() at narrow width returned empty string")
	}
}

// UT-TUI-069: Post-save reload preserves cursor by field name
func TestPostSaveReloadPreservesCursor(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.json")

	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4")
	cfg.Set("theme", "dark")
	cfg.Set("stream", true)

	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4", "gpt-3.5-turbo"}},
		{Name: "stream", Type: "bool", Default: "true"},
		{Name: "theme", Type: "string", Default: "auto"},
	}

	if err := config.SaveConfig(tmpFile, cfg); err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	model := NewModel(cfg, schema, nil, "0.0.412", tmpFile)
	model.windowWidth = 100
	model.windowHeight = 30
	model.updateSizes()

	model.listPanel.Down()
	selectedBefore := model.listPanel.SelectedItem()
	if selectedBefore == nil {
		t.Fatal("Expected a selected item after Down")
	}
	fieldNameBefore := selectedBefore.Field.Name

	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m := newModel.(*Model)

	if !m.saved {
		t.Error("Expected saved == true after Ctrl+S")
	}
	if m.err != nil {
		t.Errorf("Expected no error after save, got: %v", m.err)
	}

	selectedAfter := m.listPanel.SelectedItem()
	if selectedAfter == nil {
		t.Fatal("Expected a selected item after save+reload")
	}
	if selectedAfter.Field.Name != fieldNameBefore {
		t.Errorf("Expected cursor on %q after save+reload, got %q", fieldNameBefore, selectedAfter.Field.Name)
	}
}

// UT-TUI-070: Modified flags cleared after successful save
func TestModifiedFlagsClearedAfterSave(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.json")

	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4")

	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4", "gpt-3.5-turbo"}},
	}

	if err := config.SaveConfig(tmpFile, cfg); err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	model := NewModel(cfg, schema, nil, "0.0.412", tmpFile)
	model.windowWidth = 100
	model.windowHeight = 30
	model.updateSizes()

	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := newModel.(*Model)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = newModel.(*Model)

	hasModified := false
	for _, e := range m.listPanel.entries {
		if !e.isHeader && e.item.Modified {
			hasModified = true
			break
		}
	}
	if !hasModified {
		t.Fatal("Expected at least one modified entry before save")
	}

	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = newModel.(*Model)

	if !m.saved {
		t.Error("Expected saved == true after Ctrl+S")
	}

	for _, e := range m.listPanel.entries {
		if !e.isHeader && e.item.Modified {
			t.Errorf("Entry %q should have Modified == false after save", e.item.Field.Name)
		}
	}

	view := m.listPanel.View()
	if strings.Contains(view, "(not-saved)") {
		t.Error("No '(not-saved)' text should appear after successful save")
	}
}

// UT-TUI-071: Save failure does not clear Modified flags
func TestSaveFailureKeepsModifiedFlags(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4")

	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4", "gpt-3.5-turbo"}},
	}

	invalidPath := filepath.Join(string(os.PathSeparator), "nonexistent", "deep", "path", "config.json")

	model := NewModel(cfg, schema, nil, "0.0.412", invalidPath)
	model.windowWidth = 100
	model.windowHeight = 30
	model.updateSizes()

	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := newModel.(*Model)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = newModel.(*Model)

	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = newModel.(*Model)

	if m.err == nil {
		t.Error("Expected error after save to invalid path")
	}
	if m.saved {
		t.Error("Expected saved == false after failed save")
	}

	hasModified := false
	for _, e := range m.listPanel.entries {
		if !e.isHeader && e.item.Modified {
			hasModified = true
			break
		}
	}
	if !hasModified {
		t.Error("Modified flags should be preserved when save fails")
	}
}

// UT-TUI-072: Enter on large enum transitions to StateModelPicker
func TestEnterOnLargeEnumTransitionsToModelPicker(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4.1")

	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "",
			Options: []string{
				"claude-sonnet-4.6", "claude-sonnet-4.5", "claude-haiku-4.5",
				"claude-opus-4.6", "claude-opus-4.6-fast", "gpt-4.1",
			},
			Description: "AI model"},
	}

	model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")
	model.windowWidth = 100
	model.windowHeight = 30
	model.updateSizes()

	if model.state != StateBrowsing {
		t.Fatalf("Expected initial state Browsing, got %v", model.state)
	}

	item := model.listPanel.SelectedItem()
	if item == nil || item.Field.Name != "model" {
		t.Fatalf("Expected selected item to be 'model', got %v", item)
	}

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	m := newModel.(*Model)

	if m.state != StateModelPicker {
		t.Errorf("Expected StateModelPicker after Enter on large enum, got %v", m.state)
	}
	if m.modelPickerPanel == nil {
		t.Error("Expected modelPickerPanel to be non-nil")
	}
	if m.modelPickerPanel != nil && m.modelPickerPanel.SelectedValue() != "gpt-4.1" {
		t.Errorf("Expected pre-selected value 'gpt-4.1', got %q", m.modelPickerPanel.SelectedValue())
	}
}

// UT-TUI-073: Enter on small enum transitions to StateEditing
func TestEnterOnSmallEnumTransitionsToEditing(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("theme", "dark")

	schema := []copilot.SchemaField{
		{Name: "theme", Type: "enum", Default: "auto",
			Options:     []string{"auto", "dark", "light"},
			Description: "Color theme"},
	}

	model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")
	model.windowWidth = 100
	model.windowHeight = 30
	model.updateSizes()

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(msg)
	m := newModel.(*Model)

	if m.state != StateEditing {
		t.Errorf("Expected StateEditing after Enter on small enum, got %v", m.state)
	}
	if m.modelPickerPanel != nil {
		t.Error("Expected modelPickerPanel to be nil for small enum")
	}
}

// UT-TUI-074: ModelPickerPanel creation and pre-selection
func TestModelPickerPanelCreation(t *testing.T) {
	options := []string{
		"claude-sonnet-4.6", "claude-sonnet-4.5", "claude-haiku-4.5",
		"claude-opus-4.6", "claude-opus-4.6-fast", "claude-opus-4.5",
		"claude-sonnet-4", "gemini-3-pro-preview",
		"gpt-5.3-codex", "gpt-5.2-codex", "gpt-5.2",
		"gpt-5.1-codex-max", "gpt-5.1-codex", "gpt-5.1",
		"gpt-5.1-codex-mini", "gpt-5-mini", "gpt-4.1",
	}
	current := "claude-sonnet-4.5"

	picker := NewModelPickerPanel(options, current)
	picker.SetSize(60, 20)

	if picker.SelectedValue() != "claude-sonnet-4.5" {
		t.Errorf("Expected pre-selected value 'claude-sonnet-4.5', got %q", picker.SelectedValue())
	}
}

// UT-TUI-075: ModelPickerPanel.View() renders non-empty string
func TestModelPickerPanelView(t *testing.T) {
	options := []string{
		"claude-sonnet-4.6", "claude-sonnet-4.5", "claude-haiku-4.5",
		"gpt-5.1-codex", "gpt-4.1",
	}
	picker := NewModelPickerPanel(options, "gpt-4.1")
	picker.SetSize(60, 20)

	view := picker.View()
	if view == "" {
		t.Error("ModelPickerPanel.View() returned empty string")
	}
	if !strings.Contains(view, "Select Model") {
		t.Error("ModelPickerPanel.View() should contain 'Select Model' title")
	}
}

// UT-TUI-076: Esc from StateModelPicker confirms and returns to StateBrowsing
func TestEscFromModelPickerConfirmsAndReturns(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4.1")

	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "",
			Options: []string{
				"claude-sonnet-4.6", "claude-sonnet-4.5", "claude-haiku-4.5",
				"claude-opus-4.6", "claude-opus-4.6-fast", "gpt-4.1",
			},
			Description: "AI model"},
	}

	model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")
	model.windowWidth = 100
	model.windowHeight = 30
	model.updateSizes()

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(enterMsg)
	m := newModel.(*Model)

	if m.state != StateModelPicker {
		t.Fatalf("Expected StateModelPicker, got %v", m.state)
	}

	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, _ = m.Update(escMsg)
	m = newModel.(*Model)

	if m.state != StateBrowsing {
		t.Errorf("Expected StateBrowsing after Esc, got %v", m.state)
	}
	if m.modelPickerPanel != nil {
		t.Error("Expected modelPickerPanel to be nil after Esc")
	}
	if cfg.Get("model") != "gpt-4.1" {
		t.Errorf("Expected config model to be 'gpt-4.1', got %v", cfg.Get("model"))
	}
}

// UT-TUI-077: Enter from StateModelPicker confirms and returns
func TestEnterFromModelPickerConfirmsAndReturns(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4.1")

	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "",
			Options: []string{
				"claude-sonnet-4.6", "claude-sonnet-4.5", "claude-haiku-4.5",
				"claude-opus-4.6", "claude-opus-4.6-fast", "gpt-4.1",
			},
			Description: "AI model"},
	}

	model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")
	model.windowWidth = 100
	model.windowHeight = 30
	model.updateSizes()

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(enterMsg)
	m := newModel.(*Model)

	if m.state != StateModelPicker {
		t.Fatalf("Expected StateModelPicker, got %v", m.state)
	}

	newModel, _ = m.Update(enterMsg)
	m = newModel.(*Model)

	if m.state != StateBrowsing {
		t.Errorf("Expected StateBrowsing after Enter confirm, got %v", m.state)
	}
	if m.modelPickerPanel != nil {
		t.Error("Expected modelPickerPanel to be nil after Enter confirm")
	}
}

// UT-TUI-078: View() in StateModelPicker at various sizes renders without panic
func TestViewInModelPickerAtVariousSizes(t *testing.T) {
	sizes := []struct{ width, height int }{
		{80, 24},
		{120, 40},
		{100, 30},
	}

	for _, sz := range sizes {
		t.Run(fmt.Sprintf("%dx%d", sz.width, sz.height), func(t *testing.T) {
			cfg := config.NewConfig()
			cfg.Set("model", "gpt-4.1")

			schema := []copilot.SchemaField{
				{Name: "model", Type: "enum", Default: "",
					Options: []string{
						"claude-sonnet-4.6", "claude-sonnet-4.5", "claude-haiku-4.5",
						"claude-opus-4.6", "claude-opus-4.6-fast", "gpt-4.1",
					},
					Description: "AI model"},
			}

			model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/config.json")
			sizeMsg := tea.WindowSizeMsg{Width: sz.width, Height: sz.height}
			newModel, _ := model.Update(sizeMsg)
			m := newModel.(*Model)

			enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
			newModel, _ = m.Update(enterMsg)
			m = newModel.(*Model)

			if m.state != StateModelPicker {
				t.Fatalf("Expected StateModelPicker, got %v", m.state)
			}

			view := m.View()
			if view == "" {
				t.Errorf("View() at %dx%d in StateModelPicker returned empty string", sz.width, sz.height)
			}
			if !strings.Contains(view, "╭") {
				t.Errorf("View() at %dx%d should contain outer frame border", sz.width, sz.height)
			}
		})
	}
}

// UT-TUI-079: ctrl+s save works from StateModelPicker
func TestCtrlSSaveFromModelPicker(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Set("model", "gpt-4.1")

	schema := []copilot.SchemaField{
		{Name: "model", Type: "enum", Default: "",
			Options: []string{
				"claude-sonnet-4.6", "claude-sonnet-4.5", "claude-haiku-4.5",
				"claude-opus-4.6", "claude-opus-4.6-fast", "gpt-4.1",
			},
			Description: "AI model"},
	}

	model := NewModel(cfg, schema, nil, "0.0.412", "/tmp/nonexistent/config.json")
	model.windowWidth = 100
	model.windowHeight = 30
	model.updateSizes()

	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := model.Update(enterMsg)
	m := newModel.(*Model)

	if m.state != StateModelPicker {
		t.Fatalf("Expected StateModelPicker, got %v", m.state)
	}

	saveMsg := tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ = m.Update(saveMsg)
	m = newModel.(*Model)

	if m.state != StateModelPicker {
		t.Errorf("Expected state to remain StateModelPicker after ctrl+s, got %v", m.state)
	}
	// Save handler was reached (either succeeded or failed with error)
	if !m.saved && m.err == nil {
		t.Error("Expected save handler to be executed (either saved=true or err!=nil)")
	}
}
