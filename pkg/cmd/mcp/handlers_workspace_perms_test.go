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

func TestListWorkspacePerms_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspaceMemberPermsFn: func(ws string, limit int) ([]backend.WorkspaceMemberPerm, error) {
			assert.Equal(t, "myworkspace", ws)
			return []backend.WorkspaceMemberPerm{
				{User: "alice", Permission: "member"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listWorkspacePerms(context.Background(), makeReq(map[string]any{
		"workspace": "myworkspace",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "alice", "member")
}

func TestListWorkspacePerms_MissingWorkspace(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listWorkspacePerms(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "workspace")
}

func TestListWorkspacePerms_UnsupportedOnServer(t *testing.T) {
	t.Parallel()
	type noPermsFake struct{ backend.Client }
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: singleHostConfig})
	factorytest.UseBackend(f, noPermsFake{Client: &testhelpers.FakeClient{T: t}})
	h := newHandlers(f)
	result, err := h.listWorkspacePerms(context.Background(), makeReq(map[string]any{
		"workspace": "myworkspace",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestListWorkspaceRepoPerms_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspaceRepoPermsFn: func(ws string, limit int) ([]backend.WorkspaceRepoPerm, error) {
			return []backend.WorkspaceRepoPerm{
				{Repo: "my-repo", User: "alice", Permission: "write"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listWorkspaceRepoPerms(context.Background(), makeReq(map[string]any{
		"workspace": "myworkspace",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "my-repo", "alice")
}

func TestListWorkspaceRepoPerms_MissingWorkspace(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listWorkspaceRepoPerms(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "workspace")
}

func TestGrantWorkspacePerm_Success(t *testing.T) {
	t.Parallel()
	var gotWS, gotUser, gotPerm string
	fake := &testhelpers.FakeClient{
		T: t,
		GrantWorkspacePermFn: func(ws, user, permission string) error {
			gotWS, gotUser, gotPerm = ws, user, permission
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.grantWorkspacePerm(context.Background(), makeReq(map[string]any{
		"workspace":  "myworkspace",
		"user":       "alice",
		"permission": "member",
	}))
	require.NoError(t, err)
	assert.Equal(t, "myworkspace", gotWS)
	assert.Equal(t, "alice", gotUser)
	assert.Equal(t, "member", gotPerm)
	assertJSONContains(t, result, "granted", "alice")
}

func TestGrantWorkspacePerm_MissingUser(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.grantWorkspacePerm(context.Background(), makeReq(map[string]any{
		"workspace":  "myworkspace",
		"permission": "member",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "user")
}

func TestGrantWorkspacePerm_MissingPermission(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.grantWorkspacePerm(context.Background(), makeReq(map[string]any{
		"workspace": "myworkspace",
		"user":      "alice",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "permission")
}

func TestRevokeWorkspacePerm_Success(t *testing.T) {
	t.Parallel()
	var gotWS, gotUser string
	fake := &testhelpers.FakeClient{
		T: t,
		RevokeWorkspacePermFn: func(ws, user string) error {
			gotWS, gotUser = ws, user
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.revokeWorkspacePerm(context.Background(), makeReq(map[string]any{
		"workspace": "myworkspace",
		"user":      "alice",
	}))
	require.NoError(t, err)
	assert.Equal(t, "myworkspace", gotWS)
	assert.Equal(t, "alice", gotUser)
	assertJSONContains(t, result, "revoked", "alice")
}

func TestRevokeWorkspacePerm_MissingUser(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.revokeWorkspacePerm(context.Background(), makeReq(map[string]any{
		"workspace": "myworkspace",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "user")
}
