package mcp

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	mcplib "github.com/mark3labs/mcp-go/mcp"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/mcp/argval"
	"github.com/proggarapsody/bitbottle/pkg/errfmt"
)

type handlers struct {
	f *factory.Factory
}

func newHandlers(f *factory.Factory) *handlers {
	return &handlers{f: f}
}

// resolveBackend picks a host and dials a backend client. Host
// inference (single-host fallback, ambiguity errors) is delegated
// to factory.ResolveHost so the rule lives in exactly one place —
// the same place ResolveTarget consults for bare PROJECT/REPO args.
func (h *handlers) resolveBackend(hostname string) (backend.Client, error) {
	host, err := factory.ResolveHost(h.f, hostname)
	if err != nil {
		return nil, err
	}
	// MCP-05: a non-empty hostname that ResolveHost accepts on faith (it
	// performs no config lookup when hostname is set) must still be a
	// configured host. Reject unknown hostnames with a typed ErrUnknownHost
	// here — before any HTTP or Server-vs-Cloud URL inference — so a typo in
	// `hostname` surfaces as "unknown host" rather than a confusing dial
	// error against a bogus Server URL.
	if err := h.requireConfiguredHost(host); err != nil {
		return nil, err
	}
	return h.f.Backend(host)
}

// requireConfiguredHost returns a typed *DomainError{Kind: ErrUnknownHost}
// when host is not present in the local configuration. The error lists the
// configured hosts so MCP clients can correct a typo without a second call.
// Returns nil when host is configured (the common path).
func (h *handlers) requireConfiguredHost(host string) error {
	cfg, err := h.f.Config()
	if err != nil {
		return err
	}
	configured := cfg.Hosts()
	for _, c := range configured {
		if c == host {
			return nil
		}
	}
	sorted := append([]string(nil), configured...)
	sort.Strings(sorted)
	return &backend.DomainError{
		Kind:    backend.ErrUnknownHost,
		Code:    backend.CodeHostUnknown,
		Host:    host,
		Message: fmt.Sprintf("unknown host %q; configured hosts: %s", host, strings.Join(sorted, ", ")),
	}
}

func jsonResult(v any) (*mcplib.CallToolResult, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return mcplib.NewToolResultError(fmt.Sprintf("serialize: %v", err)), nil
	}
	return mcplib.NewToolResultText(string(data)), nil
}

func errResult(msg string) *mcplib.CallToolResult {
	return mcplib.NewToolResultError(msg)
}

// errorEnvelope is the structured shape MCP clients receive on tool
// failures. Fields:
//
//   - Code:     dotted backend.ErrorCode token (e.g. "auth.invalid_token",
//     "repo.not_found"). When the underlying DomainError has no Code,
//     a kind-based fallback string ("auth", "not_found", "conflict",
//     "permission", "unsupported_on_host", "transport") is emitted so
//     clients always get a non-empty signal.
//   - Host/Feature/Resource/ID: optional context fields stamped by the
//     adapter at the call site.
//   - Message:  the human-readable error text (already control-byte
//     sanitised by errfmt's renderer for CLI; raw here for JSON).
//   - Hints:    actionable next-step strings sourced from errfmt's
//     catalogue, with template placeholders expanded against the
//     DomainError fields. AI-agent integrations surface these to the
//     user without bundling their own copy of the catalogue.
type errorEnvelope struct {
	Code     string   `json:"code"`
	Host     string   `json:"host,omitempty"`
	Feature  string   `json:"feature,omitempty"`
	Resource string   `json:"resource,omitempty"`
	ID       string   `json:"id,omitempty"`
	Message  string   `json:"message"`
	Hints    []string `json:"hints,omitempty"`
}

// errResultArg renders a client-side argument-validation failure as a JSON
// envelope carrying the dotted code, field, and offending value, so MCP
// clients can branch on {code, field, got} without parsing prose. Falls
// back to the plain message if marshalling somehow fails.
func errResultArg(e *argval.Error) *mcplib.CallToolResult {
	if data, mErr := json.Marshal(e); mErr == nil {
		return mcplib.NewToolResultError(string(data))
	}
	return mcplib.NewToolResultError(e.Error())
}

func errResultErr(err error) *mcplib.CallToolResult {
	var de *backend.DomainError
	if errors.As(err, &de) {
		env := errorEnvelope{
			Code:     envelopeCode(de),
			Host:     de.Host,
			Feature:  de.Feature,
			Resource: de.Resource,
			ID:       de.ID,
			Message:  de.Error(),
			Hints:    errfmt.HintsFor(de),
		}
		if data, mErr := json.Marshal(env); mErr == nil {
			return mcplib.NewToolResultError(string(data))
		}
	}
	return mcplib.NewToolResultError(err.Error())
}

// envelopeCode prefers the dotted ErrorCode (the join key with errfmt's
// catalogue) and falls back to a kind-based label when Code is unset, so
// MCP clients always get a structured token. Once every code path stamps
// a Code, the kind fallback can be retired.
func envelopeCode(de *backend.DomainError) string {
	if de.Code != "" {
		return string(de.Code)
	}
	switch {
	case errors.Is(de.Kind, backend.ErrNotFound):
		return "not_found"
	case errors.Is(de.Kind, backend.ErrAuth):
		return "auth"
	case errors.Is(de.Kind, backend.ErrPermission):
		return "permission"
	case errors.Is(de.Kind, backend.ErrUnsupportedOnHost):
		return "unsupported_on_host"
	case errors.Is(de.Kind, backend.ErrConflict):
		return "conflict"
	case errors.Is(de.Kind, backend.ErrTransport):
		return "transport"
	default:
		return "error"
	}
}

func splitTrimmed(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		result = append(result, strings.TrimSpace(p))
	}
	return result
}

// requireIntArg extracts a required, positive integer argument (e.g. a PR
// or comment id) via argval, distinguishing missing / wrong-type / negative
// inputs. It returns a ready-to-return error result when validation fails.
// Centralising this means every numeric-id tool gets the MCP-06/07/08 fixes
// at once: explicit 0 with Min(1) is rejected as out-of-range (ids start at
// 1), a string id reports a type error, and a missing key reports missing.
func requireIntArg(req mcplib.CallToolRequest, key string) (int, *mcplib.CallToolResult) {
	v, _, err := argval.Int(req.GetArguments(), key, argval.Required(), argval.Min(1))
	if err != nil {
		return 0, errResultArg(err)
	}
	return v, nil
}

func requireString(req mcplib.CallToolRequest, key string) (string, error) {
	v := req.GetString(key, "")
	if v == "" {
		return "", fmt.Errorf("missing required parameter: %s", key)
	}
	return v, nil
}

// validateEnum returns a non-nil error when value is not in the allowed set.
// The error message tells the caller exactly which values are valid.
func validateEnum(field, value string, allowed ...string) error {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return fmt.Errorf("invalid value %q for %s: must be one of %s",
		value, field, strings.Join(allowed, ", "))
}

// validateRange returns a non-nil error when n is outside [min, max].
func validateRange(field string, n, min, max int) error {
	if n < min || n > max {
		return fmt.Errorf("%s must be between %d and %d, got %d", field, min, max, n)
	}
	return nil
}
