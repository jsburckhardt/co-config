package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jsburckhardt/co-config/internal/copilot"
	"github.com/jsburckhardt/co-config/internal/sensitive"
)

// ConfigItem represents an item in the config list
type ConfigItem struct {
	Field copilot.SchemaField
	Value any
}

func (i ConfigItem) FilterValue() string { return i.Field.Name }
func (i ConfigItem) Title() string       { return i.Field.Name }
func (i ConfigItem) Description() string {
	// Show truncated current value
	valueStr := formatValue(i.Value, 40)
	if sensitive.IsSensitive(i.Field.Name) {
		return fmt.Sprintf("%s 🔒 (read-only)", sensitive.MaskValue(i.Value))
	}
	// Check if value looks like a token
	if strVal, ok := i.Value.(string); ok && sensitive.LooksLikeToken(strVal) {
		return fmt.Sprintf("%s 🔒 (read-only)", sensitive.MaskValue(i.Value))
	}
	return valueStr
}

// GroupHeader represents a category header in the list
type GroupHeader struct {
	Name string
}

func (h GroupHeader) FilterValue() string { return "" }
func (h GroupHeader) Title() string       { return h.Name }
func (h GroupHeader) Description() string { return "" }

// ItemDelegate customizes the rendering of list items
type ItemDelegate struct {
	list.DefaultDelegate
}

func NewItemDelegate() ItemDelegate {
	d := list.NewDefaultDelegate()
	return ItemDelegate{DefaultDelegate: d}
}

func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	var s string

	switch i := item.(type) {
	case GroupHeader:
		// Render as a category header
		s = groupHeaderStyle.Render(i.Name)
	case ConfigItem:
		title := i.Title()
		desc := i.Description()
		
		// Determine if selected
		isSelected := index == m.Index()
		
		if isSelected {
			title = selectedItemStyle.Render("▶ " + title)
			desc = selectedDescStyle.Render("  " + desc)
		} else {
			if sensitive.IsSensitive(i.Field.Name) {
				title = sensitiveItemStyle.Render("  " + title)
			} else {
				// Check if value looks like a token
				if strVal, ok := i.Value.(string); ok && sensitive.LooksLikeToken(strVal) {
					title = sensitiveItemStyle.Render("  " + title)
				} else {
					title = itemStyle.Render("  " + title)
				}
			}
			desc = descStyle.Render("  " + desc)
		}
		
		s = lipgloss.JoinVertical(lipgloss.Left, title, desc)
	default:
		// Fallback to default rendering
		d.DefaultDelegate.Render(w, m, index, item)
		return
	}

	fmt.Fprint(w, s)
}

func (d ItemDelegate) Height() int {
	return 2
}

func (d ItemDelegate) Spacing() int {
	return 0
}

func (d ItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

// formatValue formats a value for display, truncating if necessary
func formatValue(val any, maxLen int) string {
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
			s = "[]"
		} else {
			strs := make([]string, 0, len(v))
			for _, item := range v {
				if str, ok := item.(string); ok {
					strs = append(strs, str)
				}
			}
			s = fmt.Sprintf("[%s]", strings.Join(strs, ", "))
		}
	case map[string]any:
		s = fmt.Sprintf("{%d keys}", len(v))
	case nil:
		s = ""
	default:
		s = fmt.Sprintf("%v", v)
	}

	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}
