package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/jsburckhardt/co-config/internal/copilot"
	"github.com/jsburckhardt/co-config/internal/sensitive"
)

// EnvVarsPanel displays environment variable entries in a scrollable, read-only view.
type EnvVarsPanel struct {
	envVars []copilot.EnvVarInfo
	cursor  int
	offset  int
	width   int
	height  int
}

// NewEnvVarsPanel creates a new env vars panel.
func NewEnvVarsPanel(envVars []copilot.EnvVarInfo) *EnvVarsPanel {
	return &EnvVarsPanel{
		envVars: envVars,
		cursor:  0,
	}
}

// SetSize updates the panel content dimensions.
func (p *EnvVarsPanel) SetSize(w, h int) {
	p.width = w
	p.height = h
	p.ensureVisible()
}

// Up moves cursor up one entry.
func (p *EnvVarsPanel) Up() {
	if p.cursor > 0 {
		p.cursor--
		p.ensureVisible()
	}
}

// Down moves cursor down one entry.
func (p *EnvVarsPanel) Down() {
	if p.cursor < len(p.envVars)-1 {
		p.cursor++
		p.ensureVisible()
	}
}

// Cursor returns the current cursor position.
func (p *EnvVarsPanel) Cursor() int {
	return p.cursor
}

// linesPerEntry is the number of rendered lines each env var entry occupies.
const linesPerEntry = 4

// ensureVisible adjusts offset so the cursor is within the visible viewport.
func (p *EnvVarsPanel) ensureVisible() {
	if p.height <= 0 || len(p.envVars) == 0 {
		return
	}
	visible := p.height / linesPerEntry
	if visible < 1 {
		visible = 1
	}
	if p.cursor < p.offset {
		p.offset = p.cursor
	}
	if p.cursor >= p.offset+visible {
		p.offset = p.cursor - visible + 1
	}
}

// View renders the panel content.
func (p *EnvVarsPanel) View() string {
	if p.width <= 0 || p.height <= 0 {
		return ""
	}

	if len(p.envVars) == 0 {
		return envVarDescStyle.Render("No environment variables detected")
	}

	visible := p.height / linesPerEntry
	if visible < 1 {
		visible = 1
	}

	var lines []string
	end := p.offset + visible
	if end > len(p.envVars) {
		end = len(p.envVars)
	}

	for i := p.offset; i < end; i++ {
		entry := p.envVars[i]
		selected := i == p.cursor

		lines = append(lines, p.renderEntry(entry, selected)...)
	}

	// Pad to fill the panel height
	for len(lines) < p.height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// renderEntry renders a single env var entry as multiple lines.
func (p *EnvVarsPanel) renderEntry(entry copilot.EnvVarInfo, selected bool) []string {
	var lines []string

	// Resolve value: iterate names, use first non-empty env var value
	var resolvedValue string
	var isSensitive bool
	for _, name := range entry.Names {
		if sensitive.IsEnvVarSensitive(name) {
			isSensitive = true
		}
		if v := os.Getenv(name); v != "" && resolvedValue == "" {
			resolvedValue = v
		}
	}
	// Also check if the value itself looks like a token
	if resolvedValue != "" && sensitive.LooksLikeToken(resolvedValue) {
		isSensitive = true
	}

	// Line 1: primary name + value status
	primaryName := entry.Names[0]
	var valueDisplay string
	if resolvedValue != "" {
		if isSensitive {
			valueDisplay = envVarSensitiveStyle.Render("🔒 set")
		} else {
			display := resolvedValue
			if len(display) > 30 {
				display = display[:27] + "..."
			}
			valueDisplay = envVarValueSetStyle.Render(display)
		}
	} else {
		valueDisplay = envVarValueUnsetStyle.Render("(not set)")
	}

	var nameLine string
	if selected {
		nameLine = fmt.Sprintf("▶ %s  %s", envVarNameStyle.Render(primaryName), valueDisplay)
	} else {
		nameLine = fmt.Sprintf("  %s  %s", envVarNameStyle.Render(primaryName), valueDisplay)
	}
	lines = append(lines, truncateLine(nameLine, p.width))

	// Line 2: alias names and/or qualifier
	var line2Parts []string
	if len(entry.Names) > 1 {
		aliases := strings.Join(entry.Names[1:], ", ")
		line2Parts = append(line2Parts, envVarAliasStyle.Render("  aliases: "+aliases))
	}
	if entry.Qualifier != "" {
		line2Parts = append(line2Parts, envVarQualifierStyle.Render("  ("+entry.Qualifier+")"))
	}
	if len(line2Parts) > 0 {
		lines = append(lines, strings.Join(line2Parts, " "))
	} else {
		lines = append(lines, "")
	}

	// Line 3: description
	if entry.Description != "" {
		desc := entry.Description
		if len(desc) > p.width-4 && p.width > 7 {
			desc = desc[:p.width-7] + "..."
		}
		lines = append(lines, envVarDescStyle.Render("  "+desc))
	} else {
		lines = append(lines, "")
	}

	// Line 4: blank separator
	lines = append(lines, "")

	return lines
}

// truncateLine truncates a rendered line if it exceeds width.
// Note: this is a simple byte-length truncation; styled strings may have
// ANSI codes that make byte length differ from visible width.
func truncateLine(line string, _ int) string {
	return line
}
