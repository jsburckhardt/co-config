package tui

import (
"fmt"
"strings"

"github.com/charmbracelet/bubbles/textarea"
"github.com/charmbracelet/bubbles/textinput"
tea "github.com/charmbracelet/bubbletea"
"github.com/jsburckhardt/co-config/internal/copilot"
"github.com/jsburckhardt/co-config/internal/sensitive"
)

// DetailPanel represents the right panel showing field details and editing widgets.
type DetailPanel struct {
field         *copilot.SchemaField
value         any
isEditing     bool
textInput     textinput.Model
textArea      textarea.Model
toggleValue   bool
selectIndex   int
validationErr string
width         int
height        int
}

// NewDetailPanel creates a new detail panel.
func NewDetailPanel() DetailPanel {
ti := textinput.New()
ti.Placeholder = "Enter value..."
ti.CharLimit = 500

ta := textarea.New()
ta.Placeholder = "Enter values, one per line..."
ta.CharLimit = 5000
ta.SetHeight(5)

return DetailPanel{
textInput: ti,
textArea:  ta,
}
}

// SetField sets the current field to display/edit.
func (d *DetailPanel) SetField(field copilot.SchemaField, value any) {
d.field = &field
d.value = value
d.validationErr = ""

switch field.Type {
case "string":
if str, ok := value.(string); ok {
d.textInput.SetValue(str)
} else {
d.textInput.SetValue(field.Default)
}
if d.width > 4 {
d.textInput.Width = d.width - 4
}
case "list":
var lines []string
if arr, ok := value.([]any); ok {
for _, item := range arr {
if s, ok := item.(string); ok {
lines = append(lines, s)
}
}
}
d.textArea.SetValue(strings.Join(lines, "\n"))
if d.width > 4 {
d.textArea.SetWidth(d.width - 4)
}
case "bool":
if b, ok := value.(bool); ok {
d.toggleValue = b
} else {
d.toggleValue = field.Default == "true"
}
case "enum":
selectedValue := field.Default
if s, ok := value.(string); ok {
selectedValue = s
}
d.selectIndex = 0
for i, opt := range field.Options {
if opt == selectedValue {
d.selectIndex = i
break
}
}
}
}

// StartEditing enables edit mode.
func (d *DetailPanel) StartEditing() tea.Cmd {
d.isEditing = true
d.validationErr = ""

switch d.field.Type {
case "string":
d.textInput.Focus()
return textinput.Blink
case "list":
d.textArea.Focus()
return textarea.Blink
}
return nil
}

// StopEditing disables edit mode and returns the new value.
func (d *DetailPanel) StopEditing() any {
d.isEditing = false
d.textInput.Blur()
d.textArea.Blur()
return d.GetValue()
}

// GetValue returns the current value from the editing widget.
func (d *DetailPanel) GetValue() any {
if d.field == nil {
return nil
}
switch d.field.Type {
case "string":
return d.textInput.Value()
case "bool":
return d.toggleValue
case "enum":
if d.selectIndex >= 0 && d.selectIndex < len(d.field.Options) {
return d.field.Options[d.selectIndex]
}
return d.field.Default
case "list":
lines := strings.Split(d.textArea.Value(), "\n")
var result []any
for _, line := range lines {
line = strings.TrimSpace(line)
if line != "" {
result = append(result, line)
}
}
return result
}
return nil
}

// Update handles messages for the detail panel.
func (d *DetailPanel) Update(msg tea.Msg) tea.Cmd {
if !d.isEditing || d.field == nil {
return nil
}

if keyMsg, ok := msg.(tea.KeyMsg); ok {
switch d.field.Type {
case "bool":
if keyMsg.String() == " " || keyMsg.String() == "enter" {
d.toggleValue = !d.toggleValue
}
return nil
case "enum":
switch keyMsg.String() {
case "up", "k":
if d.selectIndex > 0 {
d.selectIndex--
}
case "down", "j":
if d.selectIndex < len(d.field.Options)-1 {
d.selectIndex++
}
}
return nil
}
}

var cmd tea.Cmd
switch d.field.Type {
case "string":
d.textInput, cmd = d.textInput.Update(msg)
case "list":
d.textArea, cmd = d.textArea.Update(msg)
}
return cmd
}

// SetSize updates the panel dimensions.
func (d *DetailPanel) SetSize(width, height int) {
d.width = width
d.height = height
if width > 4 {
d.textInput.Width = width - 4
d.textArea.SetWidth(width - 4)
}
}

// View renders the detail panel content (no border — the caller applies the border).
func (d *DetailPanel) View() string {
if d.field == nil {
return detailNoteStyle.Render("Select a field to view details")
}

var b strings.Builder

b.WriteString(detailHeaderStyle.Render(d.field.Name))
b.WriteString("\n\n")

// Type
b.WriteString(detailLabelStyle.Render("Type: "))
b.WriteString(d.field.Type)
b.WriteString("\n\n")

// Description
if d.field.Description != "" {
b.WriteString(detailDescStyle.Render(d.field.Description))
b.WriteString("\n\n")
}

isSensitive := sensitive.IsSensitive(d.field.Name)
var isTokenLike bool
if strVal, ok := d.value.(string); ok {
isTokenLike = sensitive.LooksLikeToken(strVal)
}

if isSensitive || isTokenLike {
b.WriteString(detailLabelStyle.Render("Value (read-only):"))
b.WriteString("\n")
b.WriteString(sensitiveValueStyle.Render(sensitive.MaskValue(d.value)))
b.WriteString("\n\n")
b.WriteString(detailNoteStyle.Render("🔒 This field contains sensitive data and cannot be edited."))
} else if d.isEditing {
b.WriteString(detailLabelStyle.Render("Edit value:"))
b.WriteString("\n")
b.WriteString(d.renderEditWidget())
if d.validationErr != "" {
b.WriteString("\n")
b.WriteString(errorStyle.Render("⚠ " + d.validationErr))
}
} else {
b.WriteString(detailLabelStyle.Render("Current value:"))
b.WriteString("\n")
b.WriteString(d.renderCurrentValue())

// Show options for enum fields
if d.field.Type == "enum" && len(d.field.Options) > 0 {
b.WriteString("\n\n")
b.WriteString(detailLabelStyle.Render("Options: "))
b.WriteString(strings.Join(d.field.Options, ", "))
}

b.WriteString("\n\n")
b.WriteString(detailNoteStyle.Render("Press Enter to edit"))
}

return b.String()
}

func (d *DetailPanel) renderCurrentValue() string {
return detailValueStyle.Render(formatValueDetail(d.value))
}

func (d *DetailPanel) renderEditWidget() string {
switch d.field.Type {
case "string":
return d.textInput.View()
case "bool":
if d.toggleValue {
return toggleOnStyle.Render("✓ Yes") + "  " + detailNoteStyle.Render("(space/enter to toggle)")
}
return toggleOffStyle.Render("✗ No") + "  " + detailNoteStyle.Render("(space/enter to toggle)")
case "enum":
var opts strings.Builder
for i, opt := range d.field.Options {
if i == d.selectIndex {
opts.WriteString(selectedOptionStyle.Render("▶ " + opt))
} else {
opts.WriteString(optionStyle.Render("  " + opt))
}
opts.WriteString("\n")
}
return opts.String()
case "list":
return d.textArea.View()
}
return ""
}

func formatValueDetail(val any) string {
switch v := val.(type) {
case string:
return v
case bool:
if v {
return "true"
}
return "false"
case []any:
if len(v) == 0 {
return "(empty)"
}
var strs []string
for _, item := range v {
if s, ok := item.(string); ok {
strs = append(strs, s)
}
}
return strings.Join(strs, "\n")
case map[string]any:
return fmt.Sprintf("{%d keys}", len(v))
case nil:
return "(not set)"
default:
return fmt.Sprintf("%v", v)
}
}
