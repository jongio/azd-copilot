// Package logging provides a re-export of azd-core's logutil package.
// This allows azd-copilot code to use logging without changing all imports.
package logging

import "github.com/jongio/azd-core/logutil"

// Re-export logutil functions for backward compatibility.
var (
	SetupLogger    = logutil.SetupLogger
	IsDebugEnabled = logutil.IsDebugEnabled
	Debug          = logutil.Debug
	Info           = logutil.Info
	Warn           = logutil.Warn
	Error          = logutil.Error
	NewLogger      = logutil.NewLogger
)

// Logger is an alias for logutil.ComponentLogger for component-scoped logging.
type Logger = logutil.ComponentLogger
