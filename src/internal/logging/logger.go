// Package logging provides a re-export of azd-core's logutil package.
// This allows azd-copilot code to use logging without changing all imports.
package logging

import (
	"github.com/jongio/azd-core/logutil"
)

// Re-export logutil functions for backward compatibility

// SetupLogger configures the global logger.
func SetupLogger(debug, structured bool) {
	logutil.SetupLogger(debug, structured)
}

// IsDebugEnabled returns true if debug logging is enabled.
func IsDebugEnabled() bool {
	return logutil.IsDebugEnabled()
}

// Debug logs a debug message with optional key-value pairs.
func Debug(msg string, args ...any) {
	logutil.Debug(msg, args...)
}

// Info logs an info message with optional key-value pairs.
func Info(msg string, args ...any) {
	logutil.Info(msg, args...)
}

// Warn logs a warning message with optional key-value pairs.
func Warn(msg string, args ...any) {
	logutil.Warn(msg, args...)
}

// Error logs an error message with optional key-value pairs.
func Error(msg string, args ...any) {
	logutil.Error(msg, args...)
}
