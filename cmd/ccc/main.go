package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/jsburckhardt/co-config/internal/config"
	"github.com/jsburckhardt/co-config/internal/copilot"
	"github.com/jsburckhardt/co-config/internal/logging"
	"github.com/jsburckhardt/co-config/internal/tui"
)

var version = "0.1.0"

func main() {
	rootCmd := &cobra.Command{
		Use:   "ccc",
		Short: "Copilot Config CLI — interactive TUI for GitHub Copilot CLI settings",
		Long:  "ccc reads ~/.copilot/config.json, auto-detects the installed Copilot CLI version and available config keys, and presents them in an interactive terminal UI for editing.",
		RunE:  run,
	}

	rootCmd.Version = version
	rootCmd.PersistentFlags().String("log-level", "warn", "Log level (debug, info, warn, error)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// Initialize logging
	logLevel, _ := cmd.Flags().GetString("log-level")
	logDir := filepath.Join(config.DefaultPath(), "..")
	logPath := filepath.Join(logDir, "ccc.log")
	if err := logging.Init(logging.ParseLevel(logLevel), logPath); err != nil {
		slog.Warn("failed to initialize logging", "error", err)
	}
	slog.Info("ccc starting", "version", version)

	// Detect copilot version
	copilotVersion, err := copilot.DetectVersion()
	if err != nil {
		if errors.Is(err, copilot.ErrCopilotNotInstalled) {
			fmt.Fprintln(os.Stderr, "Error: GitHub Copilot CLI is not installed.")
			fmt.Fprintln(os.Stderr, "Install it from: https://github.com/github/gh-copilot")
			return err
		}
		slog.Warn("failed to detect copilot version", "error", err)
	}
	slog.Info("detected copilot version", "version", copilotVersion)

	// Detect config schema
	schema, err := copilot.DetectSchema()
	if err != nil {
		slog.Warn("failed to detect config schema, using empty schema", "error", err)
		schema = []copilot.SchemaField{}
	}
	slog.Info("detected config schema", "fields", len(schema))

	// Load config
	configPath := config.DefaultPath()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		if errors.Is(err, config.ErrConfigNotFound) {
			slog.Info("config file not found, starting with empty config", "path", configPath)
			cfg = config.NewConfig()
		} else {
			return fmt.Errorf("loading config: %w", err)
		}
	}
	slog.Info("loaded config", "path", configPath, "keys", len(cfg.Keys()))

	// Build and run TUI
	model := tui.NewModel(cfg, schema, copilotVersion, configPath)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("running TUI: %w", err)
	}

	return nil
}
