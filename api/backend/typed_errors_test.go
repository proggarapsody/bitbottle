package backend_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// TestDomainError_IsMatchesKind verifies that errors.Is matches a DomainError
// against its Kind sentinel — the contract callers depend on for branching.
func TestDomainError_IsMatchesKind(t *testing.T) {
	t.Parallel()
	err := &backend.DomainError{
		Kind:    backend.ErrNotFound,
		Host:    "git.moscow.alfaintra.net",
		Message: "pull request 42 not found",
	}
	require.ErrorIs(t, err, backend.ErrNotFound)
	assert.NotErrorIs(t, err, backend.ErrAuth)
}

// TestClassifyHTTPError covers the status-code → sentinel mapping that every
// adapter shares. Each row asserts the Kind chosen and that the original
// HTTPError is preserved as the Cause for downstream introspection.
func TestClassifyHTTPError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		status int
		want   error
	}{
		{"401 maps to auth", 401, backend.ErrAuth},
		{"403 maps to permission", 403, backend.ErrPermission},
		{"404 maps to not-found", 404, backend.ErrNotFound},
		{"409 maps to conflict", 409, backend.ErrConflict},
		{"500 maps to transport", 500, backend.ErrTransport},
		{"503 maps to transport", 503, backend.ErrTransport},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			httpErr := &backend.HTTPError{StatusCode: tt.status, Message: "oops"}
			got := backend.ClassifyHTTPError("h.example", httpErr)
			require.NotNil(t, got)
			assert.ErrorIs(t, got, tt.want)
			assert.Equal(t, "h.example", got.Host)
			assert.Same(t, httpErr, got.Cause,
				"original HTTPError must remain reachable via errors.As / Unwrap")
		})
	}
}

// TestClassifyHTTPError_UnknownStatusHasNoKind verifies that statuses that do
// not fit the bounded domain set (e.g. 400 validation) still produce a
// DomainError but with Kind=nil so callers do not match it against a sentinel.
func TestClassifyHTTPError_UnknownStatusHasNoKind(t *testing.T) {
	t.Parallel()
	got := backend.ClassifyHTTPError("h.example", &backend.HTTPError{StatusCode: 400, Message: "bad"})
	require.NotNil(t, got)
	assert.Nil(t, got.Kind)
	assert.NotErrorIs(t, got, backend.ErrNotFound)
}
