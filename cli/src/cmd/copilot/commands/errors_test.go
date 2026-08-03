package commands

import (
	"errors"
	"fmt"
	"testing"

	"github.com/azure/azure-dev/cli/azd/pkg/azdext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCopilotNotInstalledError(t *testing.T) {
	err := NewCopilotNotInstalledError()

	var localErr *azdext.LocalError
	require.True(t, errors.As(err, &localErr))
	assert.Equal(t, ErrCodeCopilotNotInstalled, localErr.Code)
	assert.Equal(t, azdext.LocalErrorCategoryDependency, localErr.Category)
	assert.Contains(t, localErr.Suggestion, "winget install GitHub.Copilot")
	assert.Contains(t, localErr.Suggestion, "npm install -g @github/copilot")
}

func TestNewSpecNotFoundError(t *testing.T) {
	err := newSpecNotFoundError()

	var localErr *azdext.LocalError
	require.True(t, errors.As(err, &localErr))
	assert.Equal(t, ErrCodeSpecNotFound, localErr.Code)
	assert.Equal(t, azdext.LocalErrorCategoryUser, localErr.Category)
	assert.Contains(t, localErr.Suggestion, "azd copilot build")
}

func TestNewMissingDescriptionError(t *testing.T) {
	err := newMissingDescriptionError()

	var localErr *azdext.LocalError
	require.True(t, errors.As(err, &localErr))
	assert.Equal(t, ErrCodeMissingDescription, localErr.Code)
	assert.Equal(t, azdext.LocalErrorCategoryValidation, localErr.Category)
	assert.NotEmpty(t, localErr.Suggestion)
}

func TestNewInvalidSessionIDError(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		wantMessage string
	}{
		{name: "empty", id: "", wantMessage: "session ID cannot be empty"},
		{name: "unsafe", id: "../etc", wantMessage: `invalid session ID: "../etc"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := newInvalidSessionIDError(tt.id)

			var localErr *azdext.LocalError
			require.True(t, errors.As(err, &localErr))
			assert.Equal(t, ErrCodeInvalidSessionID, localErr.Code)
			assert.Equal(t, azdext.LocalErrorCategoryValidation, localErr.Category)
			assert.Equal(t, tt.wantMessage, localErr.Error())
		})
	}
}

func TestNewSessionNotFoundError(t *testing.T) {
	err := newSessionNotFoundError("abc-123")

	var localErr *azdext.LocalError
	require.True(t, errors.As(err, &localErr))
	assert.Equal(t, ErrCodeSessionNotFound, localErr.Code)
	assert.Equal(t, azdext.LocalErrorCategoryUser, localErr.Category)
	assert.Contains(t, localErr.Error(), "abc-123")
}

// TestStructuredErrorsSurviveWrapping guards the property the whole design rests
// on: the azd host locates these errors with errors.As, so a caller that adds
// context with fmt.Errorf must not hide the code, category, or suggestion.
func TestStructuredErrorsSurviveWrapping(t *testing.T) {
	wrapped := fmt.Errorf("while starting a session: %w", NewCopilotNotInstalledError())

	var localErr *azdext.LocalError
	require.True(t, errors.As(wrapped, &localErr))
	assert.Equal(t, ErrCodeCopilotNotInstalled, localErr.Code)
	assert.NotEmpty(t, azdext.ErrorSuggestion(wrapped))
}

// TestErrorCodesAreDistinct keeps a copy-paste from collapsing two codes into
// one, which would silently merge unrelated failures in host telemetry.
func TestErrorCodesAreDistinct(t *testing.T) {
	codes := []string{
		ErrCodeCopilotNotInstalled,
		ErrCodeMissingDescription,
		ErrCodeSessionNotFound,
		ErrCodeSpecNotFound,
		ErrCodeInvalidSessionID,
	}

	seen := make(map[string]bool, len(codes))
	for _, code := range codes {
		assert.NotEmpty(t, code)
		assert.False(t, seen[code], "duplicate error code %q", code)
		seen[code] = true
	}
}
