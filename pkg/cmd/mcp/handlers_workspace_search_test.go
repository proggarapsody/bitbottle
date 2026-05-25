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

// noWorkspaceSearchClient embeds backend.Client without satisfying
// WorkspaceSearcher, so AsWorkspaceSearcher type-assertion fails.
type noWorkspaceSearchClient struct {
	backend.Client
}

func TestSearchWorkspaces_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		SearchWorkspacesFn: func(opts backend.WorkspaceSearchOpts) ([]backend.Workspace, error) {
			assert.Equal(t, "myws", opts.Query)
			assert.Equal(t, "owner", opts.Role)
			return []backend.Workspace{
				{Slug: "myws-team", Name: "MyWS Team", UUID: "u-1", WebURL: "https://bitbucket.org/myws-team/"},
				{Slug: "myws-dev", Name: "MyWS Dev", UUID: "u-2", WebURL: "https://bitbucket.org/myws-dev/"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.searchWorkspaces(context.Background(), makeReq(map[string]any{
		"query": "myws",
		"role":  "owner",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "myws-team", "")
	assertJSONContains(t, result, "myws-dev", "")
}

func TestSearchWorkspaces_UnsupportedOnServer(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: singleCloudConfig})
	factorytest.UseBackend(f, noWorkspaceSearchClient{Client: &testhelpers.FakeClient{T: t}})
	h := newHandlers(f)
	result, err := h.searchWorkspaces(context.Background(), makeReq(map[string]any{}))
	require.NoError(t, err)
	assertErrorResult(t, result, "host.unsupported")
}

func TestSearchWorkspaces_EmptyQuery_ReturnsAll(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		SearchWorkspacesFn: func(opts backend.WorkspaceSearchOpts) ([]backend.Workspace, error) {
			assert.Equal(t, "", opts.Query)
			assert.Equal(t, "", opts.Role)
			return []backend.Workspace{
				{Slug: "acme", Name: "Acme Inc"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.searchWorkspaces(context.Background(), makeReq(map[string]any{}))
	require.NoError(t, err)
	assertJSONContains(t, result, "acme", "")
}
