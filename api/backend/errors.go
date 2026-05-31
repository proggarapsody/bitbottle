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
	// ErrUnknownHost is returned when a caller names a hostname that is not
	// present in the local configuration (hosts.yml). Distinct from
	// ErrUnsupportedOnHost (host is known, but the feature isn't) and ErrAuth
	// (host is known, but the token is bad): the host itself is unrecognised,
	// so no HTTP should be attempted against it.
	ErrUnknownHost    = errors.New("unknown host")
	ErrConflict       = errors.New("conflict")
	ErrTransport      = errors.New("transport error")
	ErrInvalidRequest = errors.New("invalid request")
	// ErrEndpointDeprecated is returned when a Bitbucket API endpoint returns
	// HTTP 410 Gone indicating the endpoint has been removed (e.g. CHANGE-2770).
	ErrEndpointDeprecated = errors.New("endpoint deprecated")
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

	// repo cluster — repository-shaped failures
	CodeRepoNotFound ErrorCode = "repo.not_found"

	// pr cluster — pull-request lifecycle failures
	CodePRNotFound              ErrorCode = "pr.not_found"
	CodePRMergeConflict         ErrorCode = "pr.merge.conflict"
	CodePRMergeBehind           ErrorCode = "pr.merge.behind"
	CodePRCreateDuplicateBranch ErrorCode = "pr.create.duplicate_branch"
	CodePRReviewerUnknown       ErrorCode = "pr.reviewer.unknown"
	// CodePRAutoMergeBetaDisabled is returned by Bitbucket Cloud when the
	// auto-merge beta endpoint is unavailable because the workspace has not
	// opted into the feature.
	CodePRAutoMergeBetaDisabled ErrorCode = "pr.automerge.beta_disabled"

	// branch cluster — branch-protection / write-side failures
	CodeBranchProtected ErrorCode = "branch.protected"

	// host cluster — feature unavailable on the targeted Bitbucket flavour,
	// or the named host is not configured at all
	CodeHostUnsupported ErrorCode = "host.unsupported"
	CodeHostUnknown     ErrorCode = "host.unknown"

	// network cluster — pre-classify codes attached at the transport
	// layer before an HTTPError exists. ClassifyTransportError stamps
	// these.
	CodeNetworkTLSUnknownAuthority ErrorCode = "network.tls_unknown_authority"
	CodeTransportTimeout           ErrorCode = "transport.timeout"

	// request cluster — malformed or invalid input
	CodeInvalidRequest ErrorCode = "request.invalid"

	// endpoint cluster — API endpoint removed or deprecated
	CodeEndpointDeprecated ErrorCode = "endpoint.deprecated"
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
	CodeRepoNotFound,
	CodePRNotFound,
	CodePRMergeConflict,
	CodePRMergeBehind,
	CodePRCreateDuplicateBranch,
	CodePRReviewerUnknown,
	CodePRAutoMergeBetaDisabled,
	CodeBranchProtected,
	CodeHostUnsupported,
	CodeHostUnknown,
	CodeNetworkTLSUnknownAuthority,
	CodeTransportTimeout,
	CodeInvalidRequest,
	CodeEndpointDeprecated,
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

// HTTPStatus returns the HTTP status code that produced this DomainError, or
// 0 if the underlying cause is not an *HTTPError. Useful for adapters that
// need to refine call-site Code stamping based on status (e.g. 404 → not
// found, 409 → conflict). Walks the unwrap chain via errors.As.
func (e *DomainError) HTTPStatus() int {
	if e == nil {
		return 0
	}
	var he *HTTPError
	if errors.As(e.Cause, &he) {
		return he.StatusCode
	}
	return 0
}

// StampCode is an adapter convenience: when err is a *DomainError, sets the
// fields on a copy and returns the copy as an error; otherwise returns err
// unchanged. This lets call sites annotate post-classification errors with
// operation-specific codes (e.g. pr.merge.conflict vs the bare ErrConflict
// kind) without per-call boilerplate.
//
// Empty arguments are skipped so a caller can stamp only Code while leaving
// Resource/ID/Feature alone, or fill in just Resource+ID for a not-found path.
//
// The returned error preserves Kind, Host, Cause, and Message — only the
// specified fields are overwritten.
func StampCode(err error, code ErrorCode, resource, id, feature string) error {
	if err == nil {
		return nil
	}
	var de *DomainError
	if !errors.As(err, &de) {
		return err
	}
	out := *de
	if code != "" {
		out.Code = code
	}
	if resource != "" {
		out.Resource = resource
	}
	if id != "" {
		out.ID = id
	}
	if feature != "" {
		out.Feature = feature
	}
	return &out
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
	case 400, 422:
		de.Kind = ErrInvalidRequest
		de.Code = CodeInvalidRequest
	case 410:
		de.Kind = ErrEndpointDeprecated
		de.Code = CodeEndpointDeprecated
	default:
		if err.StatusCode >= 500 {
			de.Kind = ErrTransport
		}
	}
	return de
}
