package mcp

import (
	"context"
	"testing"

	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// ── MCP-05: unknown-host rejection ───────────────────────────────────────────

// TestResolveBackend_UnknownHost_RejectedBeforeHTTP verifies that a non-empty
// hostname that isn't in hosts.yml is rejected with a typed unknown-host error
// listing the configured hosts — before any HTTP/URL inference. We assert this
// at the resolveBackend seam (the single point every handler dials through).
func TestResolveBackend_UnknownHost_RejectedBeforeHTTP(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, multiHostConfig, &testhelpers.FakeClient{T: t})
	_, err := h.resolveBackend("not-a-real-host.example")
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.ErrorIs(t, de.Kind, backend.ErrUnknownHost)
	assert.Equal(t, backend.CodeHostUnknown, de.Code)
	// Lists the configured hosts so the client can correct a typo.
	assert.Contains(t, de.Error(), "git.example.com")
	assert.Contains(t, de.Error(), "git.other.com")
}

// TestUnknownHost_SurfacesViaEnvelope checks the MCP error envelope carries the
// dotted host.unknown code for an unknown hostname.
func TestUnknownHost_SurfacesViaEnvelope(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listPRCommits(context.Background(), makeReq(map[string]any{
		"hostname": "typo.example",
		"project":  "MYPROJ",
		"slug":     "my-repo",
		"pr_id":    float64(42),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "host.unknown")
}

// TestKnownHost_NotRejected confirms a configured host passes the gate.
func TestKnownHost_NotRejected(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	_, err := h.resolveBackend("git.example.com")
	require.NoError(t, err)
}

// ── MCP-16: _meta.backends + pre-HTTP backend gating ─────────────────────────

// serverToolFor builds an MCP server over a single cloud host and returns the
// registered (gated) ServerTool for name.
func serverToolFor(t *testing.T, cfg, name string, fake backend.Client) mcplib.Tool {
	t.Helper()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cfg})
	if fake != nil {
		factorytest.UseBackend(f, fake)
	}
	s := newMCPServer(f)
	tools := s.ListTools()
	st, ok := tools[name]
	require.True(t, ok, "tool %q should be registered", name)
	return st.Tool
}

// TestMetaBackends_ServerOnlyTool asserts a Server-only tool advertises
// backends: ["server"] in its _meta.
func TestMetaBackends_ServerOnlyTool(t *testing.T) {
	t.Parallel()
	tool := serverToolFor(t, singleHostConfig, "get_repo_pr_settings", nil)
	require.NotNil(t, tool.Meta, "tool should carry _meta")
	require.NotNil(t, tool.Meta.AdditionalFields)
	backends, ok := tool.Meta.AdditionalFields["backends"].([]any)
	require.True(t, ok, "_meta.backends should be a string array")
	require.Len(t, backends, 1)
	assert.Equal(t, "server", backends[0])
}

// TestMetaBackends_BothBackendsTool asserts a both-backend tool advertises
// backends: ["cloud","server"].
func TestMetaBackends_BothBackendsTool(t *testing.T) {
	t.Parallel()
	tool := serverToolFor(t, singleHostConfig, "compare_refs", nil)
	require.NotNil(t, tool.Meta)
	backends, ok := tool.Meta.AdditionalFields["backends"].([]any)
	require.True(t, ok)
	assert.ElementsMatch(t, []any{"cloud", "server"}, backends)
}

// TestBackendGating_ServerOnlyTool_OnCloudHost_RejectedBeforeHTTP is the MCP-16
// adapter test: a Server-only tool invoked against a Cloud host returns
// host.unsupported *before* any HTTP. We prove "before HTTP" by wiring a Cloud
// host whose FakeClient does NOT implement RepoPRSettingsClient and would be
// dialed by the handler — the gate must short-circuit first. We invoke through
// the server-registered (gated) handler, not the bare method.
func TestBackendGating_ServerOnlyTool_OnCloudHost_RejectedBeforeHTTP(t *testing.T) {
	t.Parallel()
	var dialed bool
	fake := &testhelpers.FakeClient{
		T: t,
		GetRepoPRSettingsFn: func(ns, slug string) (backend.RepoPRSettings, error) {
			dialed = true
			return backend.RepoPRSettings{}, nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: singleCloudConfig})
	factorytest.UseBackend(f, fake)
	s := newMCPServer(f)
	handler := s.ListTools()["get_repo_pr_settings"].Handler
	require.NotNil(t, handler)

	result, err := handler(context.Background(), makeReq(map[string]any{
		"hostname": "bitbucket.org",
		"project":  "MYWS",
		"slug":     "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "host.unsupported")
	assert.False(t, dialed, "the backend must not be dialed when the gate rejects the host")
}

// TestBackendGating_ServerOnlyTool_OnServerHost_Allowed confirms the gate lets
// a Server-only tool through on a Server host.
func TestBackendGating_ServerOnlyTool_OnServerHost_Allowed(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetRepoPRSettingsFn: func(ns, slug string) (backend.RepoPRSettings, error) {
			return backend.RepoPRSettings{RequiredApprovers: 2}, nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: singleHostConfig})
	factorytest.UseBackend(f, fake)
	s := newMCPServer(f)
	handler := s.ListTools()["get_repo_pr_settings"].Handler
	result, err := handler(context.Background(), makeReq(map[string]any{
		"hostname": "git.example.com",
		"project":  "MYPROJ",
		"slug":     "my-repo",
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError, "server-only tool should run on a server host")
}

// ── MCP-04: arg-shape regression (allowlist of the 5 migrated tools) ─────────

// TestArgShapeRegression_MigratedToolsUseProjectSlug asserts the five MCP-04
// tools accept {project, slug} and do NOT require the legacy {repo} /
// {project, repo} shape. This is an allowlist, not a global ban: many other
// tools intentionally keep {repo}.
func TestArgShapeRegression_MigratedToolsUseProjectSlug(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: singleHostConfig})
	s := newMCPServer(f)
	tools := s.ListTools()

	migrated := []string{
		"compare_refs",
		"list_pr_commits",
		"list_pr_files",
		"get_repo_pr_settings",
		"set_repo_pr_settings",
	}
	for _, name := range migrated {
		st, ok := tools[name]
		require.True(t, ok, "tool %q should be registered", name)
		props := st.Tool.InputSchema.Properties
		require.Contains(t, props, "project", "%s must expose project", name)
		require.Contains(t, props, "slug", "%s must expose slug", name)
		// The legacy repo arg must NOT be required anymore.
		for _, req := range st.Tool.InputSchema.Required {
			assert.NotEqual(t, "repo", req, "%s must not require the legacy repo arg", name)
		}
	}
}
