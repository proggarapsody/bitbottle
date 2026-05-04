package cmdutil_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

// TestExplainError_Conflict_FriendlyMessage verifies that a DomainError with
// Kind=ErrConflict is rendered as a user-actionable message on ErrOut, rather
// than the raw HTTP status string. This is the tracer bullet for the typed-
// error → friendly-message pipeline: it proves the pattern matches the
// classification the API layer already produced.
func TestExplainError_Conflict_FriendlyMessage(t *testing.T) {
	t.Parallel()

	err := &backend.DomainError{
		Kind:     backend.ErrConflict,
		Resource: "pull-request",
		Cause:    errors.New("HTTP 409"),
	}

	ios := iostreams.Test()
	cmdutil.ExplainError(ios, err)

	got := errOut(ios)
	assert.Contains(t, got, "already exists",
		"ErrConflict should produce a user-actionable 'already exists' message; got: %q", got)
	assert.NotContains(t, got, "HTTP 409",
		"raw HTTP detail should not appear in the user-facing message; got: %q", got)
}

// TestExplainError_NotFound_NamesTheResource verifies that ErrNotFound
// surfaces both the kind and the structured Resource/ID fields so users
// know what wasn't found, not just that something wasn't.
func TestExplainError_NotFound_NamesTheResource(t *testing.T) {
	t.Parallel()

	err := &backend.DomainError{
		Kind:     backend.ErrNotFound,
		Resource: "pull-request",
		ID:       "42",
	}

	ios := iostreams.Test()
	cmdutil.ExplainError(ios, err)

	got := errOut(ios)
	assert.Contains(t, got, "not found", "should mention not found; got: %q", got)
	assert.Contains(t, got, "pull-request", "should name the resource kind; got: %q", got)
	assert.Contains(t, got, "42", "should include the ID; got: %q", got)
}

// TestExplainError_Permission_TellsUserAboutAuth verifies the message
// for ErrPermission points the user toward fixing it (auth/scope), since a
// raw "permission denied" leaves a CLI user wondering what to do next.
func TestExplainError_Permission_TellsUserAboutAuth(t *testing.T) {
	t.Parallel()
	err := &backend.DomainError{Kind: backend.ErrPermission, Host: "git.example.com"}

	ios := iostreams.Test()
	cmdutil.ExplainError(ios, err)

	got := strings.ToLower(errOut(ios))
	assert.Contains(t, got, "permission",
		"should reference permission; got: %q", got)
	assert.Contains(t, got, "auth",
		"should hint at re-authentication or scope check; got: %q", got)
}

// TestExplainError_UnsupportedOnHost_NamesFeatureAndPlatform verifies
// that when a user tries a Cloud-only feature on Server (or vice-versa) the
// message tells them WHICH feature isn't available — not just "unsupported."
func TestExplainError_UnsupportedOnHost_NamesFeatureAndPlatform(t *testing.T) {
	t.Parallel()
	err := &backend.DomainError{
		Kind:    backend.ErrUnsupportedOnHost,
		Host:    "git.example.com",
		Feature: "pipelines",
	}

	ios := iostreams.Test()
	cmdutil.ExplainError(ios, err)

	got := errOut(ios)
	assert.Contains(t, got, "pipelines", "should name the unsupported feature; got: %q", got)
	assert.Contains(t, got, "git.example.com", "should name the host; got: %q", got)
}

// TestExplainError_UnrecognizedError_PassesThrough verifies that errors
// outside the typed-error taxonomy are written verbatim. Friendly messages
// shouldn't swallow information for failure modes we don't understand —
// the user (or future contributor) needs to see what happened.
func TestExplainError_UnrecognizedError_PassesThrough(t *testing.T) {
	t.Parallel()
	err := errors.New("some unexpected wire failure: connection reset")

	ios := iostreams.Test()
	cmdutil.ExplainError(ios, err)

	got := errOut(ios)
	assert.Contains(t, got, "some unexpected wire failure: connection reset",
		"unknown errors should pass through verbatim; got: %q", got)
}

// TestExplainError_NilError_NoOp verifies the function tolerates a nil
// error gracefully so callers can use it unconditionally.
func TestExplainError_NilError_NoOp(t *testing.T) {
	t.Parallel()
	ios := iostreams.Test()
	cmdutil.ExplainError(ios, nil)
	assert.Empty(t, errOut(ios), "nil error must produce no output")
}

// errOut extracts the bytes written to ErrOut for a Test() IOStreams.
func errOut(ios *iostreams.IOStreams) string {
	type stringer interface{ String() string }
	if s, ok := ios.ErrOut.(stringer); ok {
		return s.String()
	}
	return ""
}
