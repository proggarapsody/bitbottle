package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestVariableList_Repository(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineVariablesFn: func(ns, slug string) ([]backend.PipelineVariable, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			return []backend.PipelineVariable{{Key: "API_KEY", Value: "secret"}}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.variableList(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "API_KEY", "")
}

func TestVariableList_Workspace(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspaceVariablesFn: func(ns string) ([]backend.PipelineVariable, error) {
			assert.Equal(t, "myws", ns)
			return []backend.PipelineVariable{{Key: "WS_VAR", Value: "wsval"}}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.variableList(context.Background(), makeReq(map[string]any{
		"repo":  "myws/my-repo",
		"scope": "workspace",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "WS_VAR", "")
}

func TestVariableList_DeploymentMissingEnv(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.variableList(context.Background(), makeReq(map[string]any{
		"repo":  "myws/my-repo",
		"scope": "deployment",
		// no env_uuid
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assertErrorResult(t, result, "env_uuid is required")
}

func TestVariableList_UnknownScope(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.variableList(context.Background(), makeReq(map[string]any{
		"repo":  "myws/my-repo",
		"scope": "bogus",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestVariableView_Repository_Found(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineVariablesFn: func(ns, slug string) ([]backend.PipelineVariable, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			return []backend.PipelineVariable{
				{UUID: "v1", Key: "API_KEY", Value: "secret"},
				{UUID: "v2", Key: "OTHER", Value: "val"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.variableView(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"key":  "API_KEY",
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assertJSONContains(t, result, "API_KEY", "")
}

func TestVariableView_Repository_NotFound(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineVariablesFn: func(ns, slug string) ([]backend.PipelineVariable, error) {
			return []backend.PipelineVariable{
				{UUID: "v1", Key: "OTHER", Value: "val"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.variableView(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"key":  "MISSING",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assertErrorResult(t, result, "not found")
}

func TestVariableView_DeploymentMissingEnv(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.variableView(context.Background(), makeReq(map[string]any{
		"repo":  "myws/my-repo",
		"key":   "MY_KEY",
		"scope": "deployment",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assertErrorResult(t, result, "env_uuid is required")
}

func TestVariableView_UnknownScope(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.variableView(context.Background(), makeReq(map[string]any{
		"repo":  "myws/my-repo",
		"key":   "MY_KEY",
		"scope": "bogus",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestVariableSet_Repository(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		SetPipelineVariableFn: func(ns, slug string, in backend.PipelineVariableInput) (backend.PipelineVariable, error) {
			return backend.PipelineVariable{Key: in.Key, Value: in.Value}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.variableSet(context.Background(), makeReq(map[string]any{
		"repo":  "myws/my-repo",
		"key":   "MY_KEY",
		"value": "MY_VAL",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "MY_KEY", "")
}

func TestVariableSet_DeploymentMissingEnv(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.variableSet(context.Background(), makeReq(map[string]any{
		"repo":  "myws/my-repo",
		"key":   "MY_KEY",
		"value": "MY_VAL",
		"scope": "deployment",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assertErrorResult(t, result, "env_uuid is required")
}

func TestVariableDelete_Repository(t *testing.T) {
	t.Parallel()
	deleted := false
	fake := &testhelpers.FakeClient{
		T: t,
		DeletePipelineVariableFn: func(ns, slug, key string) error {
			deleted = true
			assert.Equal(t, "MY_KEY", key)
			return nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.variableDelete(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"key":  "MY_KEY",
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.True(t, deleted)
}

func TestVariableDelete_DeploymentMissingEnv(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.variableDelete(context.Background(), makeReq(map[string]any{
		"repo":  "myws/my-repo",
		"key":   "MY_KEY",
		"scope": "deployment",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assertErrorResult(t, result, "env_uuid is required")
}
