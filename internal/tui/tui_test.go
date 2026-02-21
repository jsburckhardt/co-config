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
name   string
value  any
maxLen int
want   string
}{
{"string", "test", 10, "test"},
{"bool true", true, 10, "true"},
{"bool false", false, 10, "false"},
{"empty list", []any{}, 10, "(empty)"},
{"list", []any{"a", "b"}, 20, "(2 items)"},
{"truncated", "very long string that exceeds max length", 10, "very lo..."},
{"nil", nil, 10, "(not set)"},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
got := formatValueCompact(tt.value, tt.maxLen)
if got != tt.want {
t.Errorf("formatValueCompact(%v, %d) = %q, want %q", tt.value, tt.maxLen, got, tt.want)
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

// UT-TUI-014: View renders without panicking
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
