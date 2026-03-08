package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Scope represents the configuration scope level.
type Scope int

const (
	ScopeUser         Scope = iota // ~/.copilot/config.json
	ScopeProject                   // <dir>/.copilot/settings.json
	ScopeProjectLocal              // <dir>/.copilot/settings.local.json
)

// String returns the CLI flag value for the scope.
func (s Scope) String() string {
	switch s {
	case ScopeUser:
		return "user"
	case ScopeProject:
		return "project"
	case ScopeProjectLocal:
		return "local"
	default:
		return "user"
	}
}

// Label returns a human-readable label for TUI header display.
func (s Scope) Label() string {
	switch s {
	case ScopeUser:
		return "User"
	case ScopeProject:
		return "Project"
	case ScopeProjectLocal:
		return "Project-Local"
	default:
		return "User"
	}
}

// ParseScope parses a string into a Scope value.
// Accepts "user", "project", or "local". Returns an error for unknown values.
func ParseScope(s string) (Scope, error) {
	switch s {
	case "user":
		return ScopeUser, nil
	case "project":
		return ScopeProject, nil
	case "local":
		return ScopeProjectLocal, nil
	default:
		return ScopeUser, fmt.Errorf("invalid scope %q: must be user, project, or local", s)
	}
}

// ProjectSettingsPath returns the project-level settings path for the given project directory.
func ProjectSettingsPath(projectDir string) string {
	return filepath.Join(projectDir, ".copilot", "settings.json")
}

// ProjectLocalSettingsPath returns the project-local settings path for the given project directory.
func ProjectLocalSettingsPath(projectDir string) string {
	return filepath.Join(projectDir, ".copilot", "settings.local.json")
}

// ScopePathFor returns the config file path for the given scope and project directory.
func ScopePathFor(scope Scope, projectDir string) string {
	switch scope {
	case ScopeProject:
		return ProjectSettingsPath(projectDir)
	case ScopeProjectLocal:
		return ProjectLocalSettingsPath(projectDir)
	default:
		return DefaultPath()
	}
}

// Config holds the copilot CLI configuration.
// All fields are stored in a single map to ensure round-trip fidelity.
type Config struct {
	data map[string]any
}

// NewConfig creates an empty Config.
func NewConfig() *Config {
	return &Config{data: make(map[string]any)}
}

// Get returns the value for a config key, or nil if not set.
func (c *Config) Get(key string) any {
	return c.data[key]
}

// Set sets a config key to the given value.
func (c *Config) Set(key string, value any) {
	c.data[key] = value
}

// Keys returns all config keys.
func (c *Config) Keys() []string {
	keys := make([]string, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}
	return keys
}

// Data returns the raw underlying map (for serialization).
func (c *Config) Data() map[string]any {
	return c.data
}

// DefaultPath returns the default config file path.
// Checks XDG_CONFIG_HOME first, then falls back to ~/.copilot/config.json.
// If home directory cannot be determined, uses temp directory as fallback.
func DefaultPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "copilot", "config.json")
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(os.TempDir(), "copilot", "config.json")
	}
	return filepath.Join(home, ".copilot", "config.json")
}

// LoadConfig reads and parses the config file at the given path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path is user-provided config file path, not attacker-controlled
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrConfigNotFound, path)
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrConfigInvalid, err)
	}

	return &Config{data: raw}, nil
}

// SaveConfig writes the config to the given path as JSON with 2-space indentation.
func SaveConfig(path string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg.data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}
	return nil
}
