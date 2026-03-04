package copilot

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// SchemaField represents a configuration field in the copilot schema
type SchemaField struct {
	Name        string
	Type        string // one of "bool", "string", "enum", "list"
	Default     string
	Options     []string
	Description string
}

// DetectVersion runs `copilot version` and parses the version string
func DetectVersion() (string, error) {
	// First check if copilot binary exists
	if _, err := exec.LookPath("copilot"); err != nil {
		return "", ErrCopilotNotInstalled
	}
	
	cmd := exec.Command("copilot", "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to execute copilot version: %w", err)
	}
	return ParseVersion(string(output))
}

// ParseVersion parses the version string from copilot version output
func ParseVersion(output string) (string, error) {
	// Expected format: "GitHub Copilot CLI 0.0.412\n\nYou are running the latest version.\n"
	re := regexp.MustCompile(`GitHub Copilot CLI (\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) < 2 {
		return "", ErrVersionParseFailed
	}
	return matches[1], nil
}

// DetectSchema runs `copilot help config` and parses all settings into SchemaField structs
func DetectSchema() ([]SchemaField, error) {
	// First check if copilot binary exists
	if _, err := exec.LookPath("copilot"); err != nil {
		return nil, ErrCopilotNotInstalled
	}
	
	cmd := exec.Command("copilot", "help", "config")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute copilot help config: %w", err)
	}
	return ParseSchema(string(output))
}

// ParseSchema parses the schema from copilot help config output
func ParseSchema(output string) ([]SchemaField, error) {
	var fields []SchemaField
	
	lines := strings.Split(output, "\n")
	
	// Pattern to match field name: `field_name`:
	fieldPattern := regexp.MustCompile(`^\s*` + "`" + `([^` + "`" + `]+)` + "`" + `:(.*)$`)
	// Pattern to match default value: defaults to `value` or defaults to "value"
	defaultPattern := regexp.MustCompile(`defaults to (?:` + "`" + `([^` + "`" + `]*)` + "`" + `|"([^"]*)")`)
	// Pattern to match enum options: - "option"
	optionPattern := regexp.MustCompile(`^\s*-\s+"([^"]+)"`)
	
	var currentField *SchemaField
	
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		
		// Check if this is a new field
		if matches := fieldPattern.FindStringSubmatch(line); matches != nil {
			// Save the previous field if exists
			if currentField != nil {
				fields = append(fields, *currentField)
			}
			
			// Start a new field
			fieldName := matches[1]
			descriptionStart := strings.TrimSpace(matches[2])
			
			currentField = &SchemaField{
				Name:        fieldName,
				Description: descriptionStart,
				Options:     []string{},
			}
			
			// Extract default value from the first line if present
			if defaultMatches := defaultPattern.FindStringSubmatch(descriptionStart); defaultMatches != nil {
				// Group 1 is for backtick defaults, group 2 is for quote defaults
				if defaultMatches[1] != "" {
					currentField.Default = defaultMatches[1]
				} else if defaultMatches[2] != "" {
					currentField.Default = defaultMatches[2]
				}
			}
			
			// Determine initial type based on description
			lowerDesc := strings.ToLower(descriptionStart)
			if strings.HasPrefix(lowerDesc, "list of") {
				currentField.Type = "list"
			} else if strings.Contains(lowerDesc, "whether to") {
				currentField.Type = "bool"
			}
		} else if currentField != nil {
			// Continue processing the current field
			trimmedLine := strings.TrimSpace(line)
			
			// Check for enum options
			if optionMatches := optionPattern.FindStringSubmatch(line); optionMatches != nil {
				currentField.Options = append(currentField.Options, optionMatches[1])
				// If we have options, this is an enum field
				if currentField.Type == "" || currentField.Type == "string" {
					currentField.Type = "enum"
				}
			} else if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "-") {
				// Append continuation of description
				if currentField.Description != "" && !strings.HasSuffix(currentField.Description, " ") {
					currentField.Description += " "
				}
				currentField.Description += trimmedLine
				
				// Check for default value in continuation lines
				if currentField.Default == "" {
					if defaultMatches := defaultPattern.FindStringSubmatch(trimmedLine); defaultMatches != nil {
						// Group 1 is for backtick defaults, group 2 is for quote defaults
						if defaultMatches[1] != "" {
							currentField.Default = defaultMatches[1]
						} else if defaultMatches[2] != "" {
							currentField.Default = defaultMatches[2]
						}
					}
				}
			}
		}
	}
	
	// Save the last field
	if currentField != nil {
		fields = append(fields, *currentField)
	}
	
	// Post-process fields to finalize types
	for i := range fields {
		field := &fields[i]
		
		// If type is still not set, default to string
		if field.Type == "" {
			field.Type = "string"
		}
		
		// If we detected enum options but type wasn't set, it's an enum
		if len(field.Options) > 0 && field.Type != "enum" {
			field.Type = "enum"
		}
	}
	
	if len(fields) == 0 {
		return nil, ErrSchemaParseFailed
	}
	
	return fields, nil
}

// EnvVarInfo represents a single environment variable entry from `copilot help environment`
type EnvVarInfo struct {
	Names       []string // primary name is Names[0]; aliases follow
	Description string
	Qualifier   string // e.g. "in order of precedence", may be empty
}

// DetectEnvVars runs `copilot help environment` and parses the output into EnvVarInfo structs
func DetectEnvVars() ([]EnvVarInfo, error) {
	if _, err := exec.LookPath("copilot"); err != nil {
		return nil, ErrCopilotNotInstalled
	}

	cmd := exec.Command("copilot", "help", "environment")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute copilot help environment: %w", err)
	}
	return ParseEnvVars(string(output))
}

// ParseEnvVars parses the output of `copilot help environment` into EnvVarInfo structs
func ParseEnvVars(output string) ([]EnvVarInfo, error) {
	if output == "" {
		return nil, nil
	}

	// Pattern to match the start of an entry: indented backtick-quoted names
	namePattern := regexp.MustCompile("`([^`]+)`")
	// Pattern to match a parenthesised qualifier
	qualifierPattern := regexp.MustCompile(`\(([^)]+)\)`)

	var entries []EnvVarInfo
	lines := strings.Split(output, "\n")

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and header lines (no backtick-quoted names)
		if trimmed == "" || !strings.Contains(trimmed, "`") {
			continue
		}

		// Check if this line starts with a backtick-quoted name followed by a colon
		// The format is: `NAME1`, `NAME2` (qualifier): description
		// Find the colon that separates names from description
		colonIdx := -1
		// Find the colon after the last backtick-quoted name (and optional qualifier)
		// We need to find the colon that is NOT inside backticks
		inBacktick := false
		inParen := false
		for j := 0; j < len(trimmed); j++ {
			if trimmed[j] == '`' {
				inBacktick = !inBacktick
			} else if trimmed[j] == '(' && !inBacktick {
				inParen = true
			} else if trimmed[j] == ')' && !inBacktick {
				inParen = false
			} else if trimmed[j] == ':' && !inBacktick && !inParen {
				colonIdx = j
				break
			}
		}

		if colonIdx < 0 {
			continue
		}

		namesPart := trimmed[:colonIdx]
		descPart := strings.TrimSpace(trimmed[colonIdx+1:])

		// Extract names from backtick-quoted segments
		nameMatches := namePattern.FindAllStringSubmatch(namesPart, -1)
		if len(nameMatches) == 0 {
			continue
		}

		var names []string
		for _, m := range nameMatches {
			names = append(names, m[1])
		}

		// Extract qualifier from parentheses in the names part
		var qualifier string
		if qMatch := qualifierPattern.FindStringSubmatch(namesPart); qMatch != nil {
			qualifier = qMatch[1]
		}

		// Collect continuation lines (indented, non-empty, no backtick-quoted start)
		for i+1 < len(lines) {
			nextLine := lines[i+1]
			nextTrimmed := strings.TrimSpace(nextLine)
			// Stop if we hit a blank line or a new entry (starts with backtick)
			if nextTrimmed == "" {
				break
			}
			// Check if next line looks like a new entry (starts with backtick-quoted name)
			if strings.HasPrefix(nextTrimmed, "`") {
				break
			}
			// It's a continuation line
			if descPart != "" {
				descPart += " "
			}
			descPart += nextTrimmed
			i++
		}

		entries = append(entries, EnvVarInfo{
			Names:       names,
			Description: descPart,
			Qualifier:   qualifier,
		})
	}

	if len(entries) == 0 {
		return nil, ErrEnvVarsParseFailed
	}

	return entries, nil
}
