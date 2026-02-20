package logging

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// UT-LOG-001: Init creates log file at given path
func TestInit_CreatesLogFile(t *testing.T) {
	// Use t.TempDir() for test isolation
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	// Save original default logger and restore after test
	originalLogger := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
	})

	err := Init(slog.LevelWarn, logPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Verify log file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("Log file was not created at %s", logPath)
	}
}

// UT-LOG-002: slog.Warn writes entry at warn level
func TestInit_WarnLevelWritesWarnEntry(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	originalLogger := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
	})

	err := Init(slog.LevelWarn, logPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Write a warn log entry
	slog.Warn("test warning message", "key", "value")

	// Read the log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "test warning message") {
		t.Errorf("Log file does not contain warn message. Content: %s", logContent)
	}
	if !strings.Contains(logContent, "level=WARN") {
		t.Errorf("Log file does not contain WARN level. Content: %s", logContent)
	}
	if !strings.Contains(logContent, "key=value") {
		t.Errorf("Log file does not contain structured fields. Content: %s", logContent)
	}
}

// UT-LOG-003: slog.Info is filtered at warn level (not written)
func TestInit_InfoFilteredAtWarnLevel(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	originalLogger := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
	})

	err := Init(slog.LevelWarn, logPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Write an info log entry (should be filtered out)
	slog.Info("test info message")

	// Read the log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	if strings.Contains(logContent, "test info message") {
		t.Errorf("Info message should be filtered at warn level. Content: %s", logContent)
	}
}

// UT-LOG-004: slog.Debug writes at debug level
func TestInit_DebugLevelWritesDebugEntry(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	originalLogger := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
	})

	err := Init(slog.LevelDebug, logPath)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Write a debug log entry
	slog.Debug("test debug message", "debug_key", "debug_value")

	// Read the log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	if !strings.Contains(logContent, "test debug message") {
		t.Errorf("Log file does not contain debug message. Content: %s", logContent)
	}
	if !strings.Contains(logContent, "level=DEBUG") {
		t.Errorf("Log file does not contain DEBUG level. Content: %s", logContent)
	}
	if !strings.Contains(logContent, "debug_key=debug_value") {
		t.Errorf("Log file does not contain structured fields. Content: %s", logContent)
	}
}

// UT-LOG-005: Init with invalid path returns error
func TestInit_InvalidPathReturnsError(t *testing.T) {
	originalLogger := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
	})

	// Try to create log file in a non-existent directory
	invalidPath := "/nonexistent/directory/test.log"

	err := Init(slog.LevelWarn, invalidPath)
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}

// Test ParseLevel for all cases
func TestParseLevel_AllCases(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"unknown", slog.LevelWarn},  // default
		{"", slog.LevelWarn},          // default
		{"DEBUG", slog.LevelWarn},     // case-sensitive, returns default
		{"Info", slog.LevelWarn},      // case-sensitive, returns default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseLevel(tt.input)
			if result != tt.expected {
				t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
