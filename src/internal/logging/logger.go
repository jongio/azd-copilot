// Package logging provides a structured logging abstraction built on top of slog.
package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// Level represents the logging level.
type Level int

const (
	// LevelDebug is for debug messages.
	LevelDebug Level = iota
	// LevelInfo is for informational messages.
	LevelInfo
	// LevelWarn is for warnings.
	LevelWarn
	// LevelError is for errors.
	LevelError
)

var (
	globalLogger *slog.Logger
	currentLevel = LevelInfo
	isStructured = false
	outputWriter io.Writer = os.Stderr
)

func init() {
	SetupLogger(false, false)
}

// SetupLogger configures the global logger.
func SetupLogger(debug, structured bool) {
	var level slog.Level
	if debug {
		level = slog.LevelDebug
		currentLevel = LevelDebug
	} else {
		level = slog.LevelInfo
		currentLevel = LevelInfo
	}

	isStructured = structured
	outputWriter = os.Stderr

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: level,
	}

	if structured {
		handler = slog.NewJSONHandler(outputWriter, opts)
	} else {
		handler = slog.NewTextHandler(outputWriter, opts)
	}

	globalLogger = slog.New(handler)
	slog.SetDefault(globalLogger)
}

// IsDebugEnabled returns true if debug logging is enabled.
func IsDebugEnabled() bool {
	return currentLevel == LevelDebug || os.Getenv("AZD_COPILOT_DEBUG") == "true"
}

// Debug logs a debug message with optional key-value pairs.
func Debug(msg string, args ...any) {
	if IsDebugEnabled() {
		globalLogger.Debug(msg, args...)
	}
}

// Info logs an info message with optional key-value pairs.
func Info(msg string, args ...any) {
	globalLogger.Info(msg, args...)
}

// Warn logs a warning message with optional key-value pairs.
func Warn(msg string, args ...any) {
	globalLogger.Warn(msg, args...)
}

// Error logs an error message with optional key-value pairs.
func Error(msg string, args ...any) {
	globalLogger.Error(msg, args...)
}

// ParseLevel parses a string into a Level.
func ParseLevel(s string) Level {
	switch strings.ToLower(s) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}
