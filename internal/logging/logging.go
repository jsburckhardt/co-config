package logging

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

var logFile *os.File

// Init initializes the global slog logger to write to the given file path at the given level.
func Init(level slog.Level, logPath string) error {
	// Close existing log file if open
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}

	// Create parent directories if they don't exist
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating log directory: %w", err)
	}

	// Open log file with restricted permissions (owner read/write only)
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("opening log file: %w", err)
	}
	
	logFile = f
	handler := slog.NewTextHandler(f, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
	return nil
}

// ParseLevel converts a string log level to slog.Level.
func ParseLevel(s string) slog.Level {
	// Normalize input: trim whitespace and convert to lowercase
	normalized := strings.ToLower(strings.TrimSpace(s))
	
	switch normalized {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelWarn
	}
}

// Shutdown closes the log file if it's open
func Shutdown() error {
	if logFile != nil {
		err := logFile.Close()
		logFile = nil
		return err
	}
	return nil
}
