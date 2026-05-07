package backend

import (
	"errors"
	"fmt"
)

// HTTPError represents an error HTTP response from the Bitbucket API.
type HTTPError struct {
	StatusCode int
	Message    string
	RequestURL string
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// Sentinel errors describing the bounded set of domain-level failure modes.
// Adapters classify wire errors into one of these so commands and the MCP
// surface can branch deterministically via errors.Is.
var (
	ErrNotFound          = errors.New("not found")
	ErrAuth              = errors.New("authentication required")
	ErrPermission        = errors.New("permission denied")
	ErrUnsupportedOnHost = errors.New("operation unsupported on this host")
	ErrConflict          = errors.New("conflict")
	ErrTransport         = errors.New("transport error")
)

// ErrorCode is a stable, dotted token identifying a specific user-visible
// failure mode (e.g. "auth.invalid_token"). Codes are the join key between
// the API layer (which classifies wire errors) and the errfmt renderer
// (which looks up titles + hints). Treat them as part of the public API:
// add new codes freely, never repurpose existing ones.
type ErrorCode string

// Catalogue of error codes. Group by cluster prefix; keep alphabetical
// inside each cluster. errfmt has a matching catalogue of titles + hints.
const (
	// auth cluster — credentials missing, expired, or insufficient
	CodeAuthNoToken       ErrorCode = "auth.no_token"
	CodeAuthInvalidToken  ErrorCode = "auth.invalid_token"
	CodePermWriteRequired ErrorCode = "perm.write_required"
)

// AllCodes lists every published ErrorCode. The errfmt test suite iterates
// this slice and asserts each code has a catalogue entry, so adding a new
// constant above without a matching errfmt entry fails the build.
//
// When adding a new code:
//  1. Append it to the const block above.
//  2. Append it here.
//  3. Add the matching catalogue entry in pkg/errfmt/errfmt.go.
var AllCodes = []ErrorCode{
	CodeAuthNoToken,
	CodeAuthInvalidToken,
	CodePermWriteRequired,
}

// DomainError wraps an underlying cause with structured context for renderers
// (CLI plain-text, MCP structured payload). Kind is one of the package-level
// sentinels; errors.Is(err, backend.ErrXxx) walks Kind, enabling deterministic
// branching without parsing prose.
//
// Optional fields populated when known:
//   - Host:     the hostname the request was directed at
//   - Feature:  capability name, populated for ErrUnsupportedOnHost
//   - Resource: domain kind ("pull-request", "branch", "repository", ...)
//   - ID:       resource identifier ("42", "feat/x", "ws/repo", ...)
//   - Code:     dotted ErrorCode token used by errfmt to look up the
//     user-facing title + hint. Auto-attached for HTTP 401 by
//     ClassifyHTTPError; otherwise stamped by adapters at the call site
//     (where they have op-specific context). May be empty — Render then
//     falls back to a Kind-based humanisation.
type DomainError struct {
	Kind     error
	Code     ErrorCode
	Host     string
	Feature  string
	Resource string
	ID       string
	Message  string
	Cause    error
}

// Error renders a single-line human-readable form. Structured emission (e.g.
// MCP) should read the fields directly rather than parsing this string.
func (e *DomainError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Kind != nil {
		return e.Kind.Error()
	}
	return "domain error"
}

// Unwrap exposes the underlying cause for errors.Is / errors.As walks.
func (e *DomainError) Unwrap() error { return e.Cause }

// Is matches against the Kind sentinel. The wrapped Cause is reached via
// Unwrap by errors.Is's default walk — intentionally not duplicated here.
func (e *DomainError) Is(target error) bool {
	return e.Kind != nil && errors.Is(e.Kind, target)
}

// ClassifyHTTPError translates an HTTPError into a DomainError, picking a
// Kind sentinel from the response status code. Statuses outside the bounded
// domain set leave Kind unset; the original HTTPError is preserved as Cause
// in every case so adapters can attach further context (resource, ID, etc.).
//
// Only HTTP 401 receives an automatic Code (CodeAuthInvalidToken) — re-login
// is an unambiguous remedy that does not need adapter context. 403/404/409
// each map to multiple user-facing situations (which resource? which write
// op?) and so the adapter that issued the request stamps the appropriate
// Code at the call site, where the necessary context lives.
func ClassifyHTTPError(host string, err *HTTPError) *DomainError {
	if err == nil {
		return nil
	}
	de := &DomainError{
		Host:    host,
		Cause:   err,
		Message: err.Error(),
	}
	switch err.StatusCode {
	case 401:
		de.Kind = ErrAuth
		de.Code = CodeAuthInvalidToken
	case 403:
		de.Kind = ErrPermission
	case 404:
		de.Kind = ErrNotFound
	case 409:
		de.Kind = ErrConflict
	default:
		if err.StatusCode >= 500 {
			de.Kind = ErrTransport
		}
	}
	return de
}
