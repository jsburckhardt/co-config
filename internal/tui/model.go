package tui

import (
"fmt"
"log/slog"
"sort"
"strings"

"github.com/charmbracelet/bubbles/key"
tea "github.com/charmbracelet/bubbletea"
"github.com/charmbracelet/lipgloss"
"github.com/jsburckhardt/co-config/internal/config"
"github.com/jsburckhardt/co-config/internal/copilot"
"github.com/jsburckhardt/co-config/internal/sensitive"
)

// Model is the main Bubbletea model for the two-panel TUI.
type Model struct {
cfg          *config.Config
schema       []copilot.SchemaField
envVars      []copilot.EnvVarInfo
version      string
configPath   string

state        State
listPanel    *ListPanel
detailPanel  DetailPanel
envPanel     *EnvVarsPanel
modelPickerPanel *ModelPickerPanel
keys         KeyMap

windowWidth  int
windowHeight int
err          error
saved        bool
}

// NewModel creates a new TUI model with two-panel layout.
func NewModel(cfg *config.Config, schema []copilot.SchemaField, envVars []copilot.EnvVarInfo, version, configPath string) *Model {
entries := buildEntries(cfg, schema)
lp := NewListPanel(entries)
dp := NewDetailPanel()
ep := NewEnvVarsPanel(envVars)

if item := lp.SelectedItem(); item != nil {
dp.SetField(item.Field, item.Value)
}

return &Model{
cfg:         cfg,
schema:      schema,
envVars:     envVars,
version:     version,
configPath:  configPath,
state:       StateBrowsing,
listPanel:   lp,
detailPanel: dp,
envPanel:    ep,
keys:        DefaultKeyMap(),
}
}

func buildEntries(cfg *config.Config, schema []copilot.SchemaField) []listEntry {
categories := map[string][]ConfigItem{
"Model & AI":         {},
"Display":            {},
"URLs & Permissions": {},
"General":            {},
"Sensitive":          {},
}

sorted := make([]copilot.SchemaField, len(schema))
copy(sorted, schema)
sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

for _, sf := range sorted {
value := cfg.Get(sf.Name)
item := ConfigItem{Field: sf, Value: value}

isSens := sensitive.IsSensitive(sf.Name)
var isToken bool
if s, ok := value.(string); ok {
isToken = sensitive.LooksLikeToken(s)
}

switch {
case isSens || isToken:
categories["Sensitive"] = append(categories["Sensitive"], item)
case isModelField(sf.Name):
categories["Model & AI"] = append(categories["Model & AI"], item)
case isURLField(sf.Name):
categories["URLs & Permissions"] = append(categories["URLs & Permissions"], item)
case isDisplayField(sf.Name):
categories["Display"] = append(categories["Display"], item)
default:
categories["General"] = append(categories["General"], item)
}
}

var entries []listEntry
for _, cat := range []string{"Model & AI", "Display", "URLs & Permissions", "General", "Sensitive"} {
items := categories[cat]
if len(items) > 0 {
entries = append(entries, listEntry{isHeader: true, header: cat})
for _, item := range items {
entries = append(entries, listEntry{item: item})
}
}
}
return entries
}

func isModelField(name string) bool {
for _, f := range []string{"model", "reasoning_effort", "parallel_tool_execution", "stream", "experimental"} {
if f == name {
return true
}
}
return false
}

func isURLField(name string) bool {
for _, f := range []string{"allowed_urls", "denied_urls", "trusted_folders"} {
if f == name {
return true
}
}
return name == "custom_agents" || (len(name) > 13 && name[:13] == "custom_agents")
}

func isDisplayField(name string) bool {
for _, f := range []string{"theme", "alt_screen", "render_markdown", "screen_reader", "banner", "beep", "update_terminal_title", "streamer_mode"} {
if f == name {
return true
}
}
return false
}

func isSensitiveItem(item ConfigItem) bool {
if sensitive.IsSensitive(item.Field.Name) {
return true
}
if s, ok := item.Value.(string); ok && sensitive.LooksLikeToken(s) {
return true
}
return false
}

// Init initializes the model.
func (m *Model) Init() tea.Cmd {
return nil
}

// Update handles messages and updates the model.
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
// Non-key messages (e.g. blink timers for text input)
if m.state == StateEditing {
return m, m.detailPanel.Update(msg)
}
if m.state == StateModelPicker && m.modelPickerPanel != nil {
return m, m.modelPickerPanel.Update(msg)
}
return m, nil
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
k := msg.String()

// Global keys
if k == "ctrl+c" {
slog.Info("user quit", "state", m.state)
return m, tea.Quit
}

switch m.state {
case StateBrowsing:
switch k {
case "ctrl+s":
m.saveConfig()
case "up", "k":
m.listPanel.Up()
m.syncDetailPanel()
case "down", "j":
m.listPanel.Down()
m.syncDetailPanel()
case "enter":
if item := m.listPanel.SelectedItem(); item != nil && !isSensitiveItem(*item) {
// Route large enums to the filterable model picker
if item.Field.Type == "enum" && len(item.Field.Options) >= 5 {
current := ""
if s, ok := item.Value.(string); ok {
current = s
}
picker := NewModelPickerPanel(item.Field.Options, current)
m.modelPickerPanel = &picker
m.updateSizes()
m.state = StateModelPicker
slog.Info("model picker opened", "field", item.Field.Name)
return m, nil
}
m.state = StateEditing
slog.Info("editing", "field", item.Field.Name)
return m, m.detailPanel.StartEditing()
}
case "right", "l", "tab":
m.state = StateEnvVars
slog.Info("switched to env vars view")
}
case StateEditing:
switch k {
case "ctrl+s":
m.saveConfig()
case "esc":
m.commitAndReturnToBrowsing()
return m, nil
case "enter":
if m.detailPanel.CurrentFieldType() != "list" {
m.commitAndReturnToBrowsing()
return m, nil
}
// For list fields, fall through to detail panel
return m, m.detailPanel.Update(msg)
default:
// All other keys go to detail panel
return m, m.detailPanel.Update(msg)
}
case StateModelPicker:
if m.modelPickerPanel == nil {
m.state = StateBrowsing
return m, nil
}
switch k {
case "ctrl+s":
m.saveConfig()
case "enter":
newValue := m.modelPickerPanel.SelectedValue()
if item := m.listPanel.SelectedItem(); item != nil {
m.cfg.Set(item.Field.Name, newValue)
m.listPanel.UpdateItemValue(item.Field.Name, newValue)
m.detailPanel.SetField(item.Field, newValue)
slog.Info("model picker confirmed", "field", item.Field.Name, "value", newValue)
}
m.modelPickerPanel = nil
m.state = StateBrowsing
return m, nil
case "esc":
newValue := m.modelPickerPanel.SelectedValue()
if item := m.listPanel.SelectedItem(); item != nil {
m.cfg.Set(item.Field.Name, newValue)
m.listPanel.UpdateItemValue(item.Field.Name, newValue)
m.detailPanel.SetField(item.Field, newValue)
slog.Info("model picker confirmed via esc", "field", item.Field.Name, "value", newValue)
}
m.modelPickerPanel = nil
m.state = StateBrowsing
return m, nil
default:
// Forward all other keys to the picker
return m, m.modelPickerPanel.Update(msg)
}
case StateEnvVars:
switch k {
case "left", "h", "tab":
m.state = StateBrowsing
slog.Info("switched to config view")
case "up", "k":
m.envPanel.Up()
case "down", "j":
m.envPanel.Down()
}
}

return m, nil
}

func (m *Model) syncDetailPanel() {
if item := m.listPanel.SelectedItem(); item != nil {
m.detailPanel.SetField(item.Field, item.Value)
}
}

// commitAndReturnToBrowsing stops editing, commits the value, and returns to browsing.
func (m *Model) commitAndReturnToBrowsing() {
newValue := m.detailPanel.StopEditing()
if item := m.listPanel.SelectedItem(); item != nil {
slog.Info("field updated", "field", item.Field.Name)
m.cfg.Set(item.Field.Name, newValue)
m.listPanel.UpdateItemValue(item.Field.Name, newValue)
}
m.saved = false
m.err = nil
m.state = StateBrowsing
m.syncDetailPanel()
}

// saveConfig persists config to disk, reloads to verify round-trip, and clears modified flags.
func (m *Model) saveConfig() {
slog.Info("saving config", "path", m.configPath)
if err := config.SaveConfig(m.configPath, m.cfg); err != nil {
m.err = err
slog.Error("save failed", "error", err)
return
}
m.saved = true
m.err = nil
slog.Info("config saved")

// Post-save reload from disk
reloaded, err := config.LoadConfig(m.configPath)
if err != nil {
m.err = fmt.Errorf("saved but reload failed: %w", err)
slog.Error("post-save reload failed", "error", err)
} else {
// Save cursor position by field name
var cursorFieldName string
if item := m.listPanel.SelectedItem(); item != nil {
cursorFieldName = item.Field.Name
}

// Replace config and rebuild entries
m.cfg = reloaded
entries := buildEntries(m.cfg, m.schema)
m.listPanel = NewListPanel(entries)
m.listPanel.SetSize(m.listPanelWidth(), m.listPanelHeight())

// Restore cursor to same field name
if cursorFieldName != "" {
m.selectFieldByName(cursorFieldName)
}

m.syncDetailPanel()
}

// Clear modified flags
m.listPanel.ClearAllModified()
}

// listPanelWidth returns the content width of the list panel.
func (m *Model) listPanelWidth() int {
innerWidth := m.windowWidth - 2
leftWidth := int(float64(innerWidth) * 0.40)
w := leftWidth - 4
if w < 1 {
w = 1
}
return w
}

// listPanelHeight returns the content height of the list panel.
func (m *Model) listPanelHeight() int {
innerHeight := m.windowHeight - 2
panelHeight := innerHeight - 10
if panelHeight < 3 {
panelHeight = 3
}
h := panelHeight - 2
if h < 1 {
h = 1
}
return h
}

// selectFieldByName moves the cursor to the entry with the given field name.
func (m *Model) selectFieldByName(name string) {
for i, e := range m.listPanel.entries {
if !e.isHeader && e.item.Field.Name == name {
m.listPanel.cursor = i
m.listPanel.ensureVisible()
return
}
}
}

func (m *Model) updateSizes() {
// Outer frame border takes 2 chars each direction
innerWidth := m.windowWidth - 2
innerHeight := m.windowHeight - 2

// framedHeader(6) + blank(1) + framedHelpBar(3) = 10 lines overhead
panelHeight := innerHeight - 10
if panelHeight < 3 {
panelHeight = 3
}

// 35% left, 65% right, 1 char gap between panels
leftWidth := int(float64(innerWidth) * 0.40)
rightWidth := innerWidth - leftWidth - 1

// Panel border(2) + padding(2) = 4 chars width overhead, border(2) height overhead
listContentW := leftWidth - 4
listContentH := panelHeight - 2
detailContentW := rightWidth - 4
detailContentH := panelHeight - 2

if listContentW < 1 {
listContentW = 1
}
if detailContentW < 1 {
detailContentW = 1
}
if listContentH < 1 {
listContentH = 1
}

m.listPanel.SetSize(listContentW, listContentH)
m.detailPanel.SetSize(detailContentW, detailContentH)

// Env panel sizing
envPanelW := innerWidth - 4
envPanelH := panelHeight - 2
if envPanelW < 1 {
envPanelW = 1
}
if envPanelH < 1 {
envPanelH = 1
}
m.envPanel.SetSize(envPanelW, envPanelH)

// Model picker sizing
if m.modelPickerPanel != nil {
pickerW := innerWidth - 4
pickerH := panelHeight - 2
if pickerW < 1 {
pickerW = 1
}
if pickerH < 1 {
pickerH = 1
}
m.modelPickerPanel.SetSize(pickerW, pickerH)
}
}

// View renders the model.
func (m *Model) View() string {
if m.windowWidth == 0 || m.windowHeight == 0 {
return "Loading..."
}

innerWidth := m.windowWidth - 2
innerHeight := m.windowHeight - 2
panelHeight := innerHeight - 10
if panelHeight < 3 {
panelHeight = 3
}
leftWidth := int(float64(innerWidth) * 0.40)
rightWidth := innerWidth - leftWidth - 1

// Header: icon (4 lines) + title block, joined horizontally
iconBlock := lipgloss.NewStyle().Foreground(primaryColor).Render(copilotIcon)
title := headerStyle.Render("ccc — Copilot Config CLI")
version := ""
if m.version != "" {
version = versionStyle.Render("Copilot CLI v" + m.version)
}
if m.saved {
version += "  " + savedStyle.Render("✓ Saved")
}
if m.err != nil {
version += "  " + errorStyle.Render("✗ "+m.err.Error())
}
titleBlock := lipgloss.JoinVertical(lipgloss.Left, title, version)
headerContent := lipgloss.JoinHorizontal(lipgloss.Center, iconBlock, "  ", titleBlock)

// Panels
var panels string
if m.state == StateEnvVars {
envContent := m.envPanel.View()
envPanelRendered := focusedPanelStyle.
Width(innerWidth - 4).
Height(panelHeight - 2).
Render(envContent)
panels = envPanelRendered
} else if m.state == StateModelPicker && m.modelPickerPanel != nil {
pickerContent := m.modelPickerPanel.View()
panels = focusedPanelStyle.
Width(innerWidth - 4).
Height(panelHeight - 2).
Render(pickerContent)
} else {
listContent := m.listPanel.View()
detailContent := m.detailPanel.View()

var leftStyle, rightStyle lipgloss.Style
if m.state == StateBrowsing {
leftStyle = focusedPanelStyle
rightStyle = panelStyle
} else {
leftStyle = panelStyle
rightStyle = focusedPanelStyle
}

leftPanel := leftStyle.
Width(leftWidth - 4).
Height(panelHeight - 2).
Render(listContent)
rightPanel := rightStyle.
Width(rightWidth - 4).
Height(panelHeight - 2).
Render(detailContent)

panels = lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, " ", rightPanel)
}

// Help bar
fieldType := m.detailPanel.CurrentFieldType()
helpKeys := m.keys.ShortHelp(m.state, fieldType)
var parts []string
for _, kb := range helpKeys {
h := kb.Help()
parts = append(parts, h.Key+" "+h.Desc)
}
helpBar := helpStyle.Render(strings.Join(parts, "  •  "))

// Assemble inner content
framedHeader := headerFrameStyle.Width(innerWidth - 2).Render(headerContent)
framedHelpBar := helpBarFrameStyle.Width(innerWidth - 2).Render(helpBar)
inner := lipgloss.JoinVertical(lipgloss.Left, framedHeader, "", panels, framedHelpBar)

// Outer frame
return outerFrameStyle.
Width(innerWidth).
Height(innerHeight).
Render(inner)
}

// ShortHelp returns bindings for the help bar based on current state.
func (k KeyMap) ShortHelp(state State, fieldType string) []key.Binding {
switch state {
case StateBrowsing:
return []key.Binding{k.Up, k.Down, k.Enter, k.Right, k.Tab, k.Save, k.Quit}
case StateEditing:
if fieldType != "list" {
return []key.Binding{k.Confirm, k.Escape, k.Save, k.Quit}
}
return []key.Binding{k.Escape, k.Save, k.Quit}
case StateEnvVars:
return []key.Binding{k.Up, k.Down, k.Left, k.Tab, k.Quit}
case StateModelPicker:
return []key.Binding{k.Filter, k.Enter, k.Escape, k.Save, k.Quit}
default:
return []key.Binding{k.Quit}
}
}
