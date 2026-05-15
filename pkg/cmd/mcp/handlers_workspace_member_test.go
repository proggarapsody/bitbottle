package mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestListWorkspaceMembers_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspaceMembersFn: func(workspace string, limit int) ([]backend.WorkspaceMember, error) {
			assert.Equal(t, "acme", workspace)
			return []backend.WorkspaceMember{
				{User: backend.User{Slug: "alice", DisplayName: "Alice Smith"}, Workspace: "acme"},
				{User: backend.User{Slug: "bob", DisplayName: "Bob Jones"}, Workspace: "acme"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listWorkspaceMembers(context.Background(), makeReq(map[string]any{
		"workspace": "acme",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "alice", "")
	assertJSONContains(t, result, "Alice Smith", "")
}

func TestListWorkspaceMembers_MissingWorkspace(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listWorkspaceMembers(context.Background(), makeReq(map[string]any{}))
	require.NoError(t, err)
	assertErrorResult(t, result, "workspace")
}

func TestListWorkspaceMembers_APIError_ReturnsErrorResult(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspaceMembersFn: func(workspace string, limit int) ([]backend.WorkspaceMember, error) {
			return nil, errors.New("500 internal server error")
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listWorkspaceMembers(context.Background(), makeReq(map[string]any{
		"workspace": "acme",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
