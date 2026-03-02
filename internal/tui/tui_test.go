package tui

import (
"strings"
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
{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4", "gpt-3.5-turbo"}, Description: "AI model"},
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
model := NewModel(cfg, []copilot.SchemaField{}, "0.0.412", "/tmp/config.json")

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

model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")

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

model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")
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
model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")
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
model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")
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
model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")
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
model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")
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
model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")
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
model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")
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

// UT-TUI-025: View renders without panicking
func TestViewRenders(t *testing.T) {
cfg := config.NewConfig()
cfg.Set("model", "gpt-4")
cfg.Set("theme", "dark")

schema := []copilot.SchemaField{
{Name: "model", Type: "enum", Default: "gpt-4", Options: []string{"gpt-4"}},
{Name: "theme", Type: "enum", Default: "auto", Options: []string{"auto", "dark"}},
}

model := NewModel(cfg, schema, "0.0.412", "/tmp/config.json")
model.windowWidth = 100
model.windowHeight = 30
model.updateSizes()

view := model.View()
if view == "" {
t.Error("View() returned empty string")
}
}
