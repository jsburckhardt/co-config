package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jsburckhardt/co-config/internal/copilot"
	"github.com/jsburckhardt/co-config/internal/sensitive"
)

// DetailPanel represents the right panel showing field details and editing widgets
type DetailPanel struct {
	field        *copilot.SchemaField
	value        any
	isEditing    bool
	textInput    textinput.Model
	textArea     textarea.Model
	toggleValue  bool
	selectIndex  int
	validationErr string
	width        int
	height       int
}

// NewDetailPanel creates a new detail panel
func NewDetailPanel(width, height int) DetailPanel {
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
		width:     width,
		height:    height,
	}
}

// SetField sets the current field to display/edit
func (d *DetailPanel) SetField(field copilot.SchemaField, value any) {
	d.field = &field
	d.value = value
	d.validationErr = ""
	
	// Initialize widgets based on field type
	switch field.Type {
	case "string":
		if str, ok := value.(string); ok {
			d.textInput.SetValue(str)
		} else {
			d.textInput.SetValue(field.Default)
		}
		d.textInput.Width = d.width - 4
		
	case "list":
		lines := []string{}
		if arr, ok := value.([]any); ok {
			for _, item := range arr {
				if s, ok := item.(string); ok {
					lines = append(lines, s)
				}
			}
		}
		d.textArea.SetValue(strings.Join(lines, "\n"))
		d.textArea.SetWidth(d.width - 4)
		
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

// StartEditing enables edit mode
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

// StopEditing disables edit mode and returns the new value
func (d *DetailPanel) StopEditing() any {
	d.isEditing = false
	d.textInput.Blur()
	d.textArea.Blur()
	
	return d.GetValue()
}

// GetValue returns the current value from the editing widget
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
		result := []any{}
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

// Update handles messages for the detail panel
func (d *DetailPanel) Update(msg tea.Msg) tea.Cmd {
	if !d.isEditing || d.field == nil {
		return nil
	}
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch d.field.Type {
		case "bool":
			switch msg.String() {
			case " ", "enter":
				d.toggleValue = !d.toggleValue
			}
		case "enum":
			switch msg.String() {
			case "up", "k":
				if d.selectIndex > 0 {
					d.selectIndex--
				}
			case "down", "j":
				if d.selectIndex < len(d.field.Options)-1 {
					d.selectIndex++
				}
			}
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

// View renders the detail panel
func (d *DetailPanel) View() string {
	if d.field == nil {
		return detailPanelStyle.
			Width(d.width).
			Height(d.height).
			Render("Select a field to view details")
	}
	
	var content strings.Builder
	
	// Field name header
	content.WriteString(detailHeaderStyle.Render(d.field.Name))
	content.WriteString("\n\n")
	
	// Description
	if d.field.Description != "" {
		wrapped := lipgloss.NewStyle().Width(d.width - 4).Render(d.field.Description)
		content.WriteString(detailDescStyle.Render(wrapped))
		content.WriteString("\n\n")
	}
	
	// Check if sensitive
	isSensitive := sensitive.IsSensitive(d.field.Name)
	var isTokenLike bool
	if strVal, ok := d.value.(string); ok {
		isTokenLike = sensitive.LooksLikeToken(strVal)
	}
	
	if isSensitive || isTokenLike {
		// Read-only sensitive field
		content.WriteString(detailLabelStyle.Render("Current value (read-only):"))
		content.WriteString("\n")
		content.WriteString(sensitiveValueStyle.Render(sensitive.MaskValue(d.value)))
		content.WriteString("\n\n")
		content.WriteString(detailNoteStyle.Render("🔒 This field contains sensitive data and cannot be edited."))
	} else if d.isEditing {
		// Show edit widget
		content.WriteString(detailLabelStyle.Render("Edit value:"))
		content.WriteString("\n")
		content.WriteString(d.renderEditWidget())
		
		if d.validationErr != "" {
			content.WriteString("\n")
			content.WriteString(errorStyle.Render("⚠ " + d.validationErr))
		}
	} else {
		// Show current value
		content.WriteString(detailLabelStyle.Render("Current value:"))
		content.WriteString("\n")
		content.WriteString(d.renderValue())
		content.WriteString("\n\n")
		content.WriteString(detailNoteStyle.Render("Press Enter to edit"))
	}
	
	return detailPanelStyle.
		Width(d.width).
		Height(d.height).
		Render(content.String())
}

func (d *DetailPanel) renderValue() string {
	return detailValueStyle.Render(formatValue(d.value, d.width-4))
}

func (d *DetailPanel) renderEditWidget() string {
	switch d.field.Type {
	case "string":
		return d.textInput.View()
	case "bool":
		if d.toggleValue {
			return toggleOnStyle.Render("✓ Yes")
		}
		return toggleOffStyle.Render("✗ No")
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

// SetSize updates the panel dimensions
func (d *DetailPanel) SetSize(width, height int) {
	d.width = width
	d.height = height
	d.textInput.Width = width - 4
	d.textArea.SetWidth(width - 4)
}
