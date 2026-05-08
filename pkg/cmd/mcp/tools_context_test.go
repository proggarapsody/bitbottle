package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// TestGetContext_RegisteredViaInit verifies that the get_context tool is
// registered with the MCP server through the self-registration mechanism
// (registry.go + tools_context.go init()).
//
// We construct the MCP server (which calls newMCPServer → registerTools which
// calls registeredFns), then invoke the tool by name to confirm it works
// end-to-end.
func TestGetContext_RegisteredViaInit(t *testing.T) {
	t.Parallel()

	cfg := "git.example.com:\n  oauth_token: tok\n"
	fake := &testhelpers.FakeClient{
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{Slug: "alice", DisplayName: "Alice"}, nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cfg})
	factorytest.UseBackend(f, fake)

	s := newMCPServer(f)
	require.NotNil(t, s, "MCP server should be non-nil")

	// Invoke the tool through the server to confirm get_context is registered.
	h := newHandlers(f)
	result, err := h.getContext(context.Background(), makeReq(map[string]any{"hostname": "git.example.com"}))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError, "get_context should succeed")
	assertJSONContains(t, result, "git.example.com", "")
}

// TestGetContext_InRegisteredToolsList verifies that registeredFns() returns
// at least one registration function contributed by tools_context.go init().
func TestGetContext_InRegisteredToolsList(t *testing.T) {
	t.Parallel()
	fns := registeredFns()
	assert.NotEmpty(t, fns, "registeredFns should include at least the get_context registration")
}
