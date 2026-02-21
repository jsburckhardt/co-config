package tui

import (
"fmt"
"strings"

"github.com/jsburckhardt/co-config/internal/copilot"
"github.com/jsburckhardt/co-config/internal/sensitive"
)

// ConfigItem represents a config field in the list.
type ConfigItem struct {
Field copilot.SchemaField
Value any
}

type listEntry struct {
isHeader bool
header   string
item     ConfigItem
}

// ListPanel is a custom scrollable list with group headers.
type ListPanel struct {
entries []listEntry
cursor  int
offset  int
width   int
height  int
}

// NewListPanel creates a list panel with cursor on the first ConfigItem.
func NewListPanel(entries []listEntry) *ListPanel {
lp := &ListPanel{entries: entries, cursor: -1}
for i, e := range entries {
if !e.isHeader {
lp.cursor = i
break
}
}
return lp
}

func (l *ListPanel) SetSize(w, h int) {
l.width = w
l.height = h
l.ensureVisible()
}

func (l *ListPanel) Up() {
for i := l.cursor - 1; i >= 0; i-- {
if !l.entries[i].isHeader {
l.cursor = i
l.ensureVisible()
return
}
}
}

func (l *ListPanel) Down() {
for i := l.cursor + 1; i < len(l.entries); i++ {
if !l.entries[i].isHeader {
l.cursor = i
l.ensureVisible()
return
}
}
}

func (l *ListPanel) SelectedItem() *ConfigItem {
if l.cursor >= 0 && l.cursor < len(l.entries) && !l.entries[l.cursor].isHeader {
item := l.entries[l.cursor].item
return &item
}
return nil
}

func (l *ListPanel) UpdateItemValue(fieldName string, newValue any) {
for i, e := range l.entries {
if !e.isHeader && e.item.Field.Name == fieldName {
l.entries[i].item.Value = newValue
break
}
}
}

func (l *ListPanel) ensureVisible() {
if l.height <= 0 || l.cursor < 0 {
return
}
if l.cursor < l.offset {
l.offset = l.cursor
if l.offset > 0 && l.entries[l.offset-1].isHeader {
l.offset--
}
}
if l.cursor >= l.offset+l.height {
l.offset = l.cursor - l.height + 1
}
}

func (l *ListPanel) View() string {
if l.width <= 0 || l.height <= 0 {
return ""
}

var lines []string
end := l.offset + l.height
if end > len(l.entries) {
end = len(l.entries)
}

for i := l.offset; i < end && len(lines) < l.height; i++ {
e := l.entries[i]
if e.isHeader {
label := "── " + e.header + " "
pad := l.width - len(label)
if pad > 0 {
label += strings.Repeat("─", pad)
} else if len(label) > l.width {
label = label[:l.width]
}
lines = append(lines, groupHeaderStyle.Render(label))
} else {
lines = append(lines, l.renderItem(e.item, i == l.cursor))
}
}

for len(lines) < l.height {
lines = append(lines, "")
}

return strings.Join(lines, "\n")
}

func (l *ListPanel) renderItem(item ConfigItem, selected bool) string {
name := item.Field.Name
isSens := sensitive.IsSensitive(name)
var isToken bool
if s, ok := item.Value.(string); ok {
isToken = sensitive.LooksLikeToken(s)
}

nameWidth := 20
if l.width < 30 {
nameWidth = l.width / 2
}
if len(name) > nameWidth {
name = name[:nameWidth-1] + "…"
}
var val string
if isSens || isToken {
val = "🔒"
} else {
valWidth := l.width - nameWidth - 4
if valWidth < 3 {
valWidth = 3
}
val = formatValueCompact(item.Value, valWidth)
}

line := fmt.Sprintf("%-*s %s", nameWidth, name, val)
if len(line) > l.width-2 {
line = line[:l.width-3] + "…"
}

if selected {
return selectedItemStyle.Render("▶ " + line)
}
if isSens || isToken {
return sensitiveItemStyle.Render("  " + line)
}
return itemStyle.Render("  " + line)
}

func formatValueCompact(val any, maxLen int) string {
if maxLen < 3 {
maxLen = 3
}
var s string
switch v := val.(type) {
case string:
s = v
case bool:
if v {
s = "true"
} else {
s = "false"
}
case []any:
if len(v) == 0 {
s = "(empty)"
} else {
s = fmt.Sprintf("(%d items)", len(v))
}
case map[string]any:
s = fmt.Sprintf("{%d keys}", len(v))
case nil:
s = "(not set)"
default:
s = fmt.Sprintf("%v", v)
}
if len(s) > maxLen {
return s[:maxLen-3] + "..."
}
return s
}
