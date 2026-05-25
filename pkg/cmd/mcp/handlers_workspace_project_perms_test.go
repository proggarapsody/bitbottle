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

func TestListWorkspaceProjectPerms_Success(t *testing.T) {
	t.Parallel()
	alice := &backend.User{Slug: "alice", DisplayName: "Alice"}
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspaceProjectPermsFn: func(workspace, projectKey string) ([]backend.WorkspaceProjectPerm, error) {
			assert.Equal(t, "myworkspace", workspace)
			assert.Equal(t, "PROJ", projectKey)
			return []backend.WorkspaceProjectPerm{
				{Permission: "write", User: alice},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listWorkspaceProjectPerms(context.Background(), makeReq(map[string]any{
		"workspace":   "myworkspace",
		"project_key": "PROJ",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "alice", "write")
}

func TestListWorkspaceProjectPerms_MissingWorkspace(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listWorkspaceProjectPerms(context.Background(), makeReq(map[string]any{
		"project_key": "PROJ",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "workspace")
}

func TestListWorkspaceProjectPerms_MissingProjectKey(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listWorkspaceProjectPerms(context.Background(), makeReq(map[string]any{
		"workspace": "myworkspace",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project_key")
}

func TestListWorkspaceProjectPerms_UnsupportedOnServer(t *testing.T) {
	t.Parallel()
	type noPermsProjectFake struct{ backend.Client }
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: singleHostConfig})
	factorytest.UseBackend(f, noPermsProjectFake{Client: &testhelpers.FakeClient{T: t}})
	h := newHandlers(f)
	result, err := h.listWorkspaceProjectPerms(context.Background(), makeReq(map[string]any{
		"workspace":   "myworkspace",
		"project_key": "PROJ",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestGrantWorkspaceProjectPerm_UserSuccess(t *testing.T) {
	t.Parallel()
	var gotWS, gotKey string
	var gotIn backend.WorkspaceProjectPermInput
	fake := &testhelpers.FakeClient{
		T: t,
		GrantWorkspaceProjectPermFn: func(workspace, projectKey string, in backend.WorkspaceProjectPermInput) error {
			gotWS, gotKey, gotIn = workspace, projectKey, in
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.grantWorkspaceProjectPerm(context.Background(), makeReq(map[string]any{
		"workspace":   "myworkspace",
		"project_key": "PROJ",
		"permission":  "write",
		"user_slug":   "alice",
	}))
	require.NoError(t, err)
	assert.Equal(t, "myworkspace", gotWS)
	assert.Equal(t, "PROJ", gotKey)
	assert.Equal(t, "alice", gotIn.UserSlug)
	assert.Equal(t, "write", gotIn.Permission)
	assertJSONContains(t, result, "granted", "alice")
}

func TestGrantWorkspaceProjectPerm_GroupSuccess(t *testing.T) {
	t.Parallel()
	var gotIn backend.WorkspaceProjectPermInput
	fake := &testhelpers.FakeClient{
		T: t,
		GrantWorkspaceProjectPermFn: func(workspace, projectKey string, in backend.WorkspaceProjectPermInput) error {
			gotIn = in
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.grantWorkspaceProjectPerm(context.Background(), makeReq(map[string]any{
		"workspace":   "myworkspace",
		"project_key": "PROJ",
		"permission":  "read",
		"group_slug":  "devs",
	}))
	require.NoError(t, err)
	assert.Equal(t, "devs", gotIn.GroupSlug)
	assertJSONContains(t, result, "granted", "devs")
}

func TestGrantWorkspaceProjectPerm_MissingSubject(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.grantWorkspaceProjectPerm(context.Background(), makeReq(map[string]any{
		"workspace":   "myworkspace",
		"project_key": "PROJ",
		"permission":  "read",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "user_slug")
}

func TestGrantWorkspaceProjectPerm_MissingPermission(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.grantWorkspaceProjectPerm(context.Background(), makeReq(map[string]any{
		"workspace":   "myworkspace",
		"project_key": "PROJ",
		"user_slug":   "alice",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "permission")
}

func TestRevokeWorkspaceProjectPerm_UserSuccess(t *testing.T) {
	t.Parallel()
	var gotWS, gotKey, gotSlug string
	var gotIsGroup bool
	fake := &testhelpers.FakeClient{
		T: t,
		RevokeWorkspaceProjectPermFn: func(workspace, projectKey, subjectSlug string, isGroup bool) error {
			gotWS, gotKey, gotSlug, gotIsGroup = workspace, projectKey, subjectSlug, isGroup
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.revokeWorkspaceProjectPerm(context.Background(), makeReq(map[string]any{
		"workspace":    "myworkspace",
		"project_key":  "PROJ",
		"subject_slug": "alice",
	}))
	require.NoError(t, err)
	assert.Equal(t, "myworkspace", gotWS)
	assert.Equal(t, "PROJ", gotKey)
	assert.Equal(t, "alice", gotSlug)
	assert.False(t, gotIsGroup)
	assertJSONContains(t, result, "revoked", "alice")
}

func TestRevokeWorkspaceProjectPerm_GroupSuccess(t *testing.T) {
	t.Parallel()
	var gotSlug string
	var gotIsGroup bool
	fake := &testhelpers.FakeClient{
		T: t,
		RevokeWorkspaceProjectPermFn: func(workspace, projectKey, subjectSlug string, isGroup bool) error {
			gotSlug, gotIsGroup = subjectSlug, isGroup
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	_, err := h.revokeWorkspaceProjectPerm(context.Background(), makeReq(map[string]any{
		"workspace":    "myworkspace",
		"project_key":  "PROJ",
		"subject_slug": "devs",
		"is_group":     true,
	}))
	require.NoError(t, err)
	assert.Equal(t, "devs", gotSlug)
	assert.True(t, gotIsGroup)
}

func TestRevokeWorkspaceProjectPerm_MissingSubjectSlug(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.revokeWorkspaceProjectPerm(context.Background(), makeReq(map[string]any{
		"workspace":   "myworkspace",
		"project_key": "PROJ",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "subject_slug")
}
