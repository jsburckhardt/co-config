package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/jsburckhardt/co-config/internal/config"
	"github.com/jsburckhardt/co-config/internal/copilot"
	"github.com/jsburckhardt/co-config/internal/sensitive"
)

// FormResult holds the edited values from the form.
type FormResult struct {
	Values map[string]any
}

// BuildForm creates a Huh form from config and schema.
// Groups fields into: General, Model & AI, URLs & Permissions, Display, Sensitive (read-only info).
// Returns the form and a FormResult that will be populated when the form completes.
func BuildForm(cfg *config.Config, schema []copilot.SchemaField) (*huh.Form, *FormResult) {
	result := &FormResult{Values: make(map[string]any)}

	// Categorize fields
	var generalFields, modelFields, urlFields, displayFields []huh.Field
	var sensitiveNotes []huh.Field

	// Sort schema by name for consistent ordering
	sorted := make([]copilot.SchemaField, len(schema))
	copy(sorted, schema)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	for _, sf := range sorted {
		if sensitive.IsSensitive(sf.Name) {
			// Show as read-only note
			val := cfg.Get(sf.Name)
			masked := sensitive.MaskValue(val)
			note := huh.NewNote().
				Title(sf.Name).
				Description(fmt.Sprintf("Value: %s (read-only)", masked))
			sensitiveNotes = append(sensitiveNotes, note)
			continue
		}

		field := buildField(sf, cfg, result)
		if field == nil {
			continue
		}

		// Categorize
		switch {
		case sf.Name == "model" || sf.Name == "reasoning_effort" || sf.Name == "parallel_tool_execution" || sf.Name == "stream" || sf.Name == "experimental":
			modelFields = append(modelFields, field)
		case sf.Name == "allowed_urls" || sf.Name == "denied_urls" || sf.Name == "trusted_folders" || strings.HasPrefix(sf.Name, "custom_agents"):
			urlFields = append(urlFields, field)
		case sf.Name == "theme" || sf.Name == "alt_screen" || sf.Name == "render_markdown" || sf.Name == "screen_reader" || sf.Name == "banner" || sf.Name == "beep" || sf.Name == "update_terminal_title" || sf.Name == "streamer_mode":
			displayFields = append(displayFields, field)
		default:
			generalFields = append(generalFields, field)
		}
	}

	// Also add any config keys not in schema (undocumented fields)
	schemaNames := make(map[string]bool)
	for _, sf := range schema {
		schemaNames[sf.Name] = true
	}
	for _, key := range cfg.Keys() {
		if !schemaNames[key] && !sensitive.IsSensitive(key) {
			val := cfg.Get(key)
			if strVal, ok := val.(string); ok {
				// Store a pointer to a new string variable
				ptr := new(string)
				*ptr = strVal
				result.Values[key] = ptr
				input := huh.NewInput().
					Title(key).
					Value(ptr).
					Description("(undocumented)")
				generalFields = append(generalFields, input)
			}
		}
	}

	// Build groups
	var groups []*huh.Group
	if len(generalFields) > 0 {
		groups = append(groups, huh.NewGroup(generalFields...).Title("General"))
	}
	if len(modelFields) > 0 {
		groups = append(groups, huh.NewGroup(modelFields...).Title("Model & AI"))
	}
	if len(urlFields) > 0 {
		groups = append(groups, huh.NewGroup(urlFields...).Title("URLs & Permissions"))
	}
	if len(displayFields) > 0 {
		groups = append(groups, huh.NewGroup(displayFields...).Title("Display"))
	}
	if len(sensitiveNotes) > 0 {
		groups = append(groups, huh.NewGroup(sensitiveNotes...).Title("Sensitive (Read-Only)"))
	}

	form := huh.NewForm(groups...)
	return form, result
}

func buildField(sf copilot.SchemaField, cfg *config.Config, result *FormResult) huh.Field {
	currentVal := cfg.Get(sf.Name)

	switch sf.Type {
	case "bool":
		val := false
		if b, ok := currentVal.(bool); ok {
			val = b
		} else if sf.Default == "true" {
			val = true
		}
		// Store pointer to bool
		ptr := new(bool)
		*ptr = val
		result.Values[sf.Name] = ptr
		return huh.NewConfirm().
			Title(sf.Name).
			Description(sf.Description).
			Value(ptr).
			Affirmative("Yes").
			Negative("No")

	case "enum":
		val := sf.Default
		if s, ok := currentVal.(string); ok {
			val = s
		}
		// Store pointer to string
		ptr := new(string)
		*ptr = val
		result.Values[sf.Name] = ptr
		options := make([]huh.Option[string], 0, len(sf.Options))
		for _, opt := range sf.Options {
			options = append(options, huh.NewOption(opt, opt))
		}
		return huh.NewSelect[string]().
			Title(sf.Name).
			Description(sf.Description).
			Options(options...).
			Value(ptr)

	case "list":
		val := ""
		if arr, ok := currentVal.([]any); ok {
			strs := make([]string, 0, len(arr))
			for _, item := range arr {
				if s, ok := item.(string); ok {
					strs = append(strs, s)
				}
			}
			val = strings.Join(strs, "\n")
		}
		// Store pointer to string
		ptr := new(string)
		*ptr = val
		result.Values[sf.Name] = ptr
		return huh.NewText().
			Title(sf.Name + " (one per line)").
			Description(sf.Description).
			Value(ptr).
			Lines(5)

	case "string":
		val := sf.Default
		if s, ok := currentVal.(string); ok {
			val = s
		}
		// Store pointer to string
		ptr := new(string)
		*ptr = val
		result.Values[sf.Name] = ptr
		return huh.NewInput().
			Title(sf.Name).
			Description(sf.Description).
			Value(ptr)

	default:
		return nil
	}
}
