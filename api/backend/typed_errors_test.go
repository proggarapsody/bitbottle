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
// not fit the bounded domain set still produce a DomainError but with Kind=nil
// so callers do not match it against a sentinel.
func TestClassifyHTTPError_UnknownStatusHasNoKind(t *testing.T) {
	t.Parallel()
	// 418 is not a mapped status — Kind must remain nil.
	got := backend.ClassifyHTTPError("h.example", &backend.HTTPError{StatusCode: 418, Message: "I'm a teapot"})
	require.NotNil(t, got)
	assert.Nil(t, got.Kind)
	assert.NotErrorIs(t, got, backend.ErrNotFound)
}

// TestClassifyHTTPError_400_SetsInvalidRequest verifies that HTTP 400 is
// classified as ErrInvalidRequest with CodeInvalidRequest.
func TestClassifyHTTPError_400_SetsInvalidRequest(t *testing.T) {
	t.Parallel()
	he := &backend.HTTPError{StatusCode: 400, Message: "Bad Request"}
	de := backend.ClassifyHTTPError("bb.example.com", he)
	require.NotNil(t, de)
	assert.ErrorIs(t, de, backend.ErrInvalidRequest)
	assert.Equal(t, backend.CodeInvalidRequest, de.Code)
}

// TestClassifyHTTPError_422_SetsInvalidRequest verifies that HTTP 422 is
// classified as ErrInvalidRequest with CodeInvalidRequest.
func TestClassifyHTTPError_422_SetsInvalidRequest(t *testing.T) {
	t.Parallel()
	he := &backend.HTTPError{StatusCode: 422, Message: "Unprocessable Entity"}
	de := backend.ClassifyHTTPError("bb.example.com", he)
	require.NotNil(t, de)
	assert.ErrorIs(t, de, backend.ErrInvalidRequest)
	assert.Equal(t, backend.CodeInvalidRequest, de.Code)
}

// TestClassifyHTTPError_AttachesAuthInvalidTokenCode verifies that 401 not
// only sets Kind=ErrAuth but also stamps the dotted error code that errfmt
// uses to look up the user-facing hint.
//
// 401 is the one HTTP status with an unambiguous user remedy (re-login),
// so it is safe to auto-attach the code at the classifier layer. Other
// statuses (403, 404, 409) require op-specific context to pick the right
// code and are stamped by the adapters themselves.
func TestClassifyHTTPError_AttachesAuthInvalidTokenCode(t *testing.T) {
	t.Parallel()
	got := backend.ClassifyHTTPError("h.example", &backend.HTTPError{StatusCode: 401, Message: "Unauthorized"})
	require.NotNil(t, got)
	assert.Equal(t, backend.CodeAuthInvalidToken, got.Code)
}
