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
type DomainError struct {
	Kind     error
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

// Is matches against either the Kind sentinel or the wrapped Cause, enabling
// errors.Is(err, backend.ErrNotFound) regardless of whether the caller built
// the DomainError directly or wrapped a transport error.
func (e *DomainError) Is(target error) bool {
	if e.Kind != nil && errors.Is(e.Kind, target) {
		return true
	}
	return false
}

// ClassifyHTTPError translates an HTTPError into a DomainError, picking a
// Kind sentinel from the response status code. Statuses outside the bounded
// domain set leave Kind unset; the original HTTPError is preserved as Cause
// in every case so adapters can attach further context (resource, ID, etc.).
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
