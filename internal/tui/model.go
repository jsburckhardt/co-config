package tui

import (
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
version      string
configPath   string

state        State
listPanel    *ListPanel
detailPanel  DetailPanel
keys         KeyMap

windowWidth  int
windowHeight int
err          error
saved        bool
}

// NewModel creates a new TUI model with two-panel layout.
func NewModel(cfg *config.Config, schema []copilot.SchemaField, version, configPath string) *Model {
entries := buildEntries(cfg, schema)
lp := NewListPanel(entries)
dp := NewDetailPanel()

if item := lp.SelectedItem(); item != nil {
dp.SetField(item.Field, item.Value)
}

return &Model{
cfg:         cfg,
schema:      schema,
version:     version,
configPath:  configPath,
state:       StateBrowsing,
listPanel:   lp,
detailPanel: dp,
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
return m, nil
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
k := msg.String()

// Global keys
if k == "ctrl+c" {
slog.Info("user quit", "state", m.state)
return m, tea.Quit
}
if k == "ctrl+s" {
slog.Info("saving config", "path", m.configPath)
if err := config.SaveConfig(m.configPath, m.cfg); err != nil {
m.err = err
slog.Error("save failed", "error", err)
} else {
m.saved = true
slog.Info("config saved")
}
return m, nil
}

switch m.state {
case StateBrowsing:
switch k {
case "up", "k":
m.listPanel.Up()
m.syncDetailPanel()
case "down", "j":
m.listPanel.Down()
m.syncDetailPanel()
case "enter":
if item := m.listPanel.SelectedItem(); item != nil && !isSensitiveItem(*item) {
m.state = StateEditing
slog.Info("editing", "field", item.Field.Name)
return m, m.detailPanel.StartEditing()
}
}
case StateEditing:
if k == "esc" {
newValue := m.detailPanel.StopEditing()
if item := m.listPanel.SelectedItem(); item != nil {
slog.Info("field updated", "field", item.Field.Name)
m.cfg.Set(item.Field.Name, newValue)
m.listPanel.UpdateItemValue(item.Field.Name, newValue)
}
m.state = StateBrowsing
return m, nil
}
// All other keys go to detail panel
return m, m.detailPanel.Update(msg)
}

return m, nil
}

func (m *Model) syncDetailPanel() {
if item := m.listPanel.SelectedItem(); item != nil {
m.detailPanel.SetField(item.Field, item.Value)
}
}

func (m *Model) updateSizes() {
// Outer frame border takes 2 chars each direction
innerWidth := m.windowWidth - 2
innerHeight := m.windowHeight - 2

// Header(2) + blank(1) + help(1) = 4 lines overhead
panelHeight := innerHeight - 4
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
}

// View renders the model.
func (m *Model) View() string {
if m.windowWidth == 0 || m.windowHeight == 0 {
return "Loading..."
}

innerWidth := m.windowWidth - 2
innerHeight := m.windowHeight - 2
panelHeight := innerHeight - 4
if panelHeight < 3 {
panelHeight = 3
}
leftWidth := int(float64(innerWidth) * 0.40)
rightWidth := innerWidth - leftWidth - 1

// Header: two lines
title := headerStyle.Render("⚙  ccc — Copilot Config CLI")
version := ""
if m.version != "" {
version = versionStyle.Render("   Copilot CLI v" + m.version)
}
if m.saved {
version += "  " + savedStyle.Render("✓ Saved")
}
if m.err != nil {
version += "  " + errorStyle.Render("✗ "+m.err.Error())
}
header := title + "\n" + version

// Panels
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

panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, " ", rightPanel)

// Help bar
helpKeys := m.keys.ShortHelp(m.state)
var parts []string
for _, kb := range helpKeys {
h := kb.Help()
parts = append(parts, h.Key+" "+h.Desc)
}
helpBar := helpStyle.Render(strings.Join(parts, "  •  "))

// Assemble inner content
inner := lipgloss.JoinVertical(lipgloss.Left, header, "", panels, helpBar)

// Outer frame
return outerFrameStyle.
Width(innerWidth).
Height(innerHeight).
Render(inner)
}

// ShortHelp returns bindings for the help bar based on current state.
func (k KeyMap) ShortHelp(state State) []key.Binding {
switch state {
case StateBrowsing:
return []key.Binding{k.Up, k.Down, k.Enter, k.Save, k.Quit}
case StateEditing:
return []key.Binding{k.Escape, k.Save, k.Quit}
default:
return []key.Binding{k.Quit}
}
}
