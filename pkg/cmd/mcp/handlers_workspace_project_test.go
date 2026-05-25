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

func TestCreateWorkspaceProject_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		CreateWorkspaceProjectFn: func(ws string, input backend.CreateWorkspaceProjectInput) (backend.WorkspaceProject, error) {
			assert.Equal(t, "myws", ws)
			assert.Equal(t, "MYPROJ", input.Key)
			assert.Equal(t, "My Project", input.Name)
			return backend.WorkspaceProject{Key: "MYPROJ", Name: "My Project"}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.createWorkspaceProject(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"key":       "MYPROJ",
		"name":      "My Project",
	}))
	require.NoError(t, err)
	text := extractText(t, result)
	assert.Contains(t, text, "Created")
	assert.Contains(t, text, "MYPROJ")
}

func TestCreateWorkspaceProject_MissingKey(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.createWorkspaceProject(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"name":      "My Project",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "key")
}

func TestCreateWorkspaceProject_UnsupportedOnHost(t *testing.T) {
	t.Parallel()
	type noProjectFake struct{ backend.Client }
	const serverConfig = "git.example.com:\n  oauth_token: tok\n"
	base := &testhelpers.FakeClient{T: t}
	noP := &noProjectFake{Client: base}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, noP)
	h := newHandlers(f)
	result, err := h.createWorkspaceProject(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"key":       "MYPROJ",
		"name":      "My Project",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "unsupported")
}

func TestViewWorkspaceProject_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetWorkspaceProjectFn: func(ws, key string) (backend.WorkspaceProject, error) {
			assert.Equal(t, "myws", ws)
			assert.Equal(t, "MYPROJ", key)
			return backend.WorkspaceProject{Key: "MYPROJ", Name: "My Project"}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.viewWorkspaceProject(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"key":       "MYPROJ",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "MYPROJ", "")
}

func TestViewWorkspaceProject_MissingKey(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.viewWorkspaceProject(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "key")
}

func TestEditWorkspaceProject_Success(t *testing.T) {
	t.Parallel()
	updatedName := "Updated Project"
	fake := &testhelpers.FakeClient{
		T: t,
		UpdateWorkspaceProjectFn: func(ws, key string, input backend.UpdateWorkspaceProjectInput) (backend.WorkspaceProject, error) {
			assert.Equal(t, "myws", ws)
			assert.Equal(t, "MYPROJ", key)
			return backend.WorkspaceProject{Key: "MYPROJ", Name: updatedName}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.editWorkspaceProject(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"key":       "MYPROJ",
		"name":      "Updated Project",
	}))
	require.NoError(t, err)
	text := extractText(t, result)
	assert.Contains(t, text, "Updated")
}

func TestEditWorkspaceProject_MissingWorkspace(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.editWorkspaceProject(context.Background(), makeReq(map[string]any{
		"key": "MYPROJ",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "workspace")
}

func TestDeleteWorkspaceProject_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteWorkspaceProjectFn: func(ws, key string) error {
			assert.Equal(t, "myws", ws)
			assert.Equal(t, "MYPROJ", key)
			return nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.deleteWorkspaceProject(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"key":       "MYPROJ",
	}))
	require.NoError(t, err)
	text := extractText(t, result)
	assert.Contains(t, text, "Deleted")
}

func TestDeleteWorkspaceProject_MissingKey(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.deleteWorkspaceProject(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "key")
}
