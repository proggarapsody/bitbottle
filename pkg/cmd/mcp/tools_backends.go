package mcp

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	mcplib "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/bbinstance"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Backend allowlist tokens used in tool _meta.backends. They mirror the
// Server-vs-Cloud split the rest of the codebase reasons about (see
// internal/bbinstance and api/backend/features.go's CloudSupport/ServerSupport
// flags).
const (
	backendServer = "server"
	backendCloud  = "cloud"
)

// backendsBoth is the default allowlist for tools that work on either flavour.
var backendsBoth = []string{backendServer, backendCloud}

// toolBackends records the per-tool backend allowlist contributed at
// registration time via addGatedTool. It is the in-process source of truth
// the call-time gate consults and the tools/list _meta reflects. Tools
// registered through the plain s.AddTool path are absent from this map and
// are treated as backendsBoth (no gating).
var (
	backendsMu   sync.RWMutex
	toolBackends = map[string][]string{}
)

// recordToolBackends stores (a defensive copy of) the allowlist for name.
func recordToolBackends(name string, backends []string) {
	backendsMu.Lock()
	defer backendsMu.Unlock()
	cp := append([]string(nil), backends...)
	sort.Strings(cp)
	toolBackends[name] = cp
}

// backendsForTool returns the recorded allowlist for name, or backendsBoth
// when the tool was registered without explicit metadata.
func backendsForTool(name string) []string {
	backendsMu.RLock()
	defer backendsMu.RUnlock()
	if b, ok := toolBackends[name]; ok {
		return b
	}
	return backendsBoth
}

// withBackends is a mcplib.ToolOption that stamps _meta.backends onto a tool
// definition so the tools/list response advertises which Bitbucket flavours
// the tool supports. AI clients read this to avoid picking a Cloud-broken
// tool on a Server host (and vice versa) without parsing description prose.
func withBackends(backends []string) mcplib.ToolOption {
	cp := append([]string(nil), backends...)
	sort.Strings(cp)
	return func(t *mcplib.Tool) {
		if t.Meta == nil {
			t.Meta = &mcplib.Meta{}
		}
		if t.Meta.AdditionalFields == nil {
			t.Meta.AdditionalFields = map[string]any{}
		}
		// Marshal as []any so the JSON shape is a plain string array.
		arr := make([]any, len(cp))
		for i, v := range cp {
			arr[i] = v
		}
		t.Meta.AdditionalFields["backends"] = arr
	}
}

// addGatedTool registers a tool whose backend allowlist is both advertised
// (via _meta.backends on tools/list) and enforced (a pre-HTTP gate wrapping
// the handler). backends must be non-empty; pass backendsBoth for tools that
// run on either flavour.
//
// The gate resolves the target host, classifies it as server or cloud, and —
// when that flavour is not in the allowlist — returns a host.unsupported
// envelope listing the allowed backends *before* the handler dials anything.
// Unknown-host and host-resolution errors surface from resolveBackend as they
// would inside the handler.
func addGatedTool(
	s *mcpserver.MCPServer,
	h *handlers,
	tool mcplib.Tool,
	backends []string,
	handler mcpserver.ToolHandlerFunc,
) {
	recordToolBackends(tool.Name, backends)
	withBackends(backends)(&tool)
	name := tool.Name
	gated := func(ctx context.Context, req mcplib.CallToolRequest) (*mcplib.CallToolResult, error) {
		if res := h.gateBackend(name, req.GetString("hostname", "")); res != nil {
			return res, nil
		}
		return handler(ctx, req)
	}
	s.AddTool(tool, gated)
}

// gateBackend enforces a tool's backend allowlist before any HTTP. It returns
// a non-nil error result when the configured host's flavour isn't allowed (or
// when host resolution itself fails — e.g. an unknown hostname), and nil when
// the call may proceed.
func (h *handlers) gateBackend(toolName, hostname string) *mcplib.CallToolResult {
	allowed := backendsForTool(toolName)
	if len(allowed) >= 2 {
		// Allowed on both flavours — nothing to gate. Skip host resolution so
		// these tools keep their existing single-host inference inside the
		// handler.
		return nil
	}
	kind, err := h.hostBackendKind(hostname)
	if err != nil {
		return errResultErr(err)
	}
	for _, a := range allowed {
		if a == kind {
			return nil
		}
	}
	flavour := "Bitbucket Server / Data Center"
	if allowed[0] == backendCloud {
		flavour = "Bitbucket Cloud"
	}
	return errResultErr(&backend.DomainError{
		Kind:    backend.ErrUnsupportedOnHost,
		Code:    backend.CodeHostUnsupported,
		Feature: toolName,
		Message: fmt.Sprintf("tool %q is supported only on %s (allowed backends: %s); the configured host is %s",
			toolName, flavour, strings.Join(allowed, ", "), kind),
	})
}

// hostBackendKind resolves hostname to a configured host and classifies it as
// "server" or "cloud". It reuses ResolveHost (single-host inference + ambiguity
// errors) and the requireConfiguredHost unknown-host check so the gate fails
// the same way the handler's own resolveBackend would — before any HTTP.
func (h *handlers) hostBackendKind(hostname string) (string, error) {
	host, err := factory.ResolveHost(h.f, hostname)
	if err != nil {
		return "", err
	}
	if err := h.requireConfiguredHost(host); err != nil {
		return "", err
	}
	cfg, err := h.f.Config()
	if err != nil {
		return "", err
	}
	hc, _ := cfg.Get(host)
	if bbinstance.IsCloud(host, hc.BackendType) {
		return backendCloud, nil
	}
	return backendServer, nil
}
