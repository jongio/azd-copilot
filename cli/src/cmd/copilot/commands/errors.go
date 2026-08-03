package commands

import (
	"fmt"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
)

// Error codes reported to the azd host through azdext.LocalError.
//
// The host renders these as `ext.<category>.<code>` in telemetry, so the strings
// are effectively a public contract. They are lowercase snake_case, stable, and
// alphabetical.
const (
	// ErrCodeCopilotNotInstalled is reported when the GitHub Copilot CLI is not
	// on PATH. Every command in this extension shells out to it, so this is a
	// hard dependency failure rather than a usage mistake.
	ErrCodeCopilotNotInstalled = "copilot_not_installed"

	// ErrCodeMissingDescription is reported when `azd copilot build` is invoked
	// with neither a description nor an existing spec to fall back on.
	ErrCodeMissingDescription = "missing_build_description"

	// ErrCodeSessionNotFound is reported when a session ID is well formed but no
	// session with that ID exists on disk.
	ErrCodeSessionNotFound = "session_not_found"

	// ErrCodeSpecNotFound is reported when a command needs a generated spec and
	// none exists yet.
	ErrCodeSpecNotFound = "spec_not_found"

	// ErrCodeInvalidSessionID is reported when a session ID is empty or contains
	// characters that are unsafe in a filesystem path.
	ErrCodeInvalidSessionID = "invalid_session_id"
)

// NewCopilotNotInstalledError reports a missing GitHub Copilot CLI. It is
// exported because the root command in package main hits the same condition
// before it can start an interactive session.
func NewCopilotNotInstalledError() error {
	return &azdext.LocalError{
		Message:    "the GitHub Copilot CLI is not installed",
		Code:       ErrCodeCopilotNotInstalled,
		Category:   azdext.LocalErrorCategoryDependency,
		Suggestion: "Install it with `winget install GitHub.Copilot` or `npm install -g @github/copilot`, then run this command again.",
	}
}

// newSpecNotFoundError reports that no generated spec exists yet.
func newSpecNotFoundError() error {
	return &azdext.LocalError{
		Message:    "no spec found",
		Code:       ErrCodeSpecNotFound,
		Category:   azdext.LocalErrorCategoryUser,
		Suggestion: "Run `azd copilot build \"<description of what you want to build>\"` to generate a spec first.",
	}
}

// newMissingDescriptionError reports `azd copilot build` with nothing to build.
func newMissingDescriptionError() error {
	return &azdext.LocalError{
		Message:    "no description provided",
		Code:       ErrCodeMissingDescription,
		Category:   azdext.LocalErrorCategoryValidation,
		Suggestion: "Describe what you want to build, for example `azd copilot build \"a Go API with a Postgres database\"`.",
	}
}

// newInvalidSessionIDError reports an empty or unsafe session ID.
func newInvalidSessionIDError(id string) error {
	message := "session ID cannot be empty"
	if id != "" {
		message = fmt.Sprintf("invalid session ID: %q", id)
	}

	return &azdext.LocalError{
		Message:    message,
		Code:       ErrCodeInvalidSessionID,
		Category:   azdext.LocalErrorCategoryValidation,
		Suggestion: "A session ID may contain only letters, digits, hyphens, and underscores. Run `azd copilot sessions list` to see the valid IDs.",
	}
}

// newSessionNotFoundError reports a well-formed session ID that does not exist.
func newSessionNotFoundError(id string) error {
	return &azdext.LocalError{
		Message:    fmt.Sprintf("session not found: %s", id),
		Code:       ErrCodeSessionNotFound,
		Category:   azdext.LocalErrorCategoryUser,
		Suggestion: "Run `azd copilot sessions list` to see the sessions that are available.",
	}
}
