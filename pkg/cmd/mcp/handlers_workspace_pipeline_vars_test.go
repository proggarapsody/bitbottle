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

// ── listWorkspacePipelineVars ─────────────────────────────────────────────────

func TestListWorkspacePipelineVars_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspacePipelineVariablesFn: func(workspace string) ([]backend.PipelineVariable, error) {
			assert.Equal(t, "acme", workspace)
			return []backend.PipelineVariable{
				{UUID: "uuid-1", Key: "MY_VAR", Value: "hello", Secured: false},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listWorkspacePipelineVars(context.Background(), makeReq(map[string]any{
		"workspace": "acme",
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assertJSONContains(t, result, "MY_VAR", "")
}

func TestListWorkspacePipelineVars_SecuredValueRedacted(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspacePipelineVariablesFn: func(workspace string) ([]backend.PipelineVariable, error) {
			return []backend.PipelineVariable{
				{UUID: "uuid-2", Key: "SECRET", Value: "s3cr3t", Secured: true},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listWorkspacePipelineVars(context.Background(), makeReq(map[string]any{
		"workspace": "acme",
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assertJSONContains(t, result, "SECRET", "")
	// Secured value must NOT appear in output.
	for _, c := range result.Content {
		if tc, ok := c.(interface{ GetText() string }); ok {
			assert.NotContains(t, tc.GetText(), "s3cr3t")
		}
	}
}

func TestListWorkspacePipelineVars_MissingWorkspace(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listWorkspacePipelineVars(context.Background(), makeReq(map[string]any{}))
	require.NoError(t, err)
	assertErrorResult(t, result, "workspace")
}

func TestListWorkspacePipelineVars_APIError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspacePipelineVariablesFn: func(workspace string) ([]backend.PipelineVariable, error) {
			return nil, errors.New("500 internal server error")
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listWorkspacePipelineVars(context.Background(), makeReq(map[string]any{
		"workspace": "acme",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

// ── getWorkspacePipelineVar ───────────────────────────────────────────────────

func TestGetWorkspacePipelineVar_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspacePipelineVariablesFn: func(workspace string) ([]backend.PipelineVariable, error) {
			return []backend.PipelineVariable{
				{UUID: "uuid-1", Key: "MY_VAR", Value: "hello", Secured: false},
			}, nil
		},
		GetWorkspacePipelineVariableFn: func(workspace, uuid string) (backend.PipelineVariable, error) {
			assert.Equal(t, "acme", workspace)
			assert.Equal(t, "uuid-1", uuid)
			return backend.PipelineVariable{UUID: "uuid-1", Key: "MY_VAR", Value: "hello"}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.getWorkspacePipelineVar(context.Background(), makeReq(map[string]any{
		"workspace": "acme",
		"key":       "MY_VAR",
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assertJSONContains(t, result, "MY_VAR", "")
}

func TestGetWorkspacePipelineVar_NotFound(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspacePipelineVariablesFn: func(workspace string) ([]backend.PipelineVariable, error) {
			return []backend.PipelineVariable{}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.getWorkspacePipelineVar(context.Background(), makeReq(map[string]any{
		"workspace": "acme",
		"key":       "MISSING",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assertErrorResult(t, result, "not found")
}

func TestGetWorkspacePipelineVar_MissingKey(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.getWorkspacePipelineVar(context.Background(), makeReq(map[string]any{
		"workspace": "acme",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "key")
}

// ── setWorkspacePipelineVar ───────────────────────────────────────────────────

func TestSetWorkspacePipelineVar_Success(t *testing.T) {
	t.Parallel()
	var gotWorkspace string
	var gotInput backend.PipelineVariableInput
	fake := &testhelpers.FakeClient{
		T: t,
		SetWorkspacePipelineVariableFn: func(workspace string, in backend.PipelineVariableInput) (backend.PipelineVariable, error) {
			gotWorkspace = workspace
			gotInput = in
			return backend.PipelineVariable{UUID: "new-uuid", Key: in.Key, Value: in.Value, Secured: in.Secured}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.setWorkspacePipelineVar(context.Background(), makeReq(map[string]any{
		"workspace": "acme",
		"key":       "NEW_VAR",
		"value":     "newval",
		"secured":   false,
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "acme", gotWorkspace)
	assert.Equal(t, "NEW_VAR", gotInput.Key)
	assert.Equal(t, "newval", gotInput.Value)
	assertJSONContains(t, result, "NEW_VAR", "")
}

func TestSetWorkspacePipelineVar_MissingValue(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.setWorkspacePipelineVar(context.Background(), makeReq(map[string]any{
		"workspace": "acme",
		"key":       "MY_VAR",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "value")
}

func TestSetWorkspacePipelineVar_APIError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		SetWorkspacePipelineVariableFn: func(workspace string, in backend.PipelineVariableInput) (backend.PipelineVariable, error) {
			return backend.PipelineVariable{}, errors.New("422 unprocessable entity")
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.setWorkspacePipelineVar(context.Background(), makeReq(map[string]any{
		"workspace": "acme",
		"key":       "MY_VAR",
		"value":     "val",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

// ── deleteWorkspacePipelineVar ────────────────────────────────────────────────

func TestDeleteWorkspacePipelineVar_Success(t *testing.T) {
	t.Parallel()
	var gotWorkspace, gotUUID string
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspacePipelineVariablesFn: func(workspace string) ([]backend.PipelineVariable, error) {
			return []backend.PipelineVariable{
				{UUID: "uuid-del", Key: "DEL_VAR"},
			}, nil
		},
		DeleteWorkspacePipelineVariableFn: func(workspace, uuid string) error {
			gotWorkspace = workspace
			gotUUID = uuid
			return nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.deleteWorkspacePipelineVar(context.Background(), makeReq(map[string]any{
		"workspace": "acme",
		"key":       "DEL_VAR",
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "acme", gotWorkspace)
	assert.Equal(t, "uuid-del", gotUUID)
	assertJSONContains(t, result, "deleted", "")
}

func TestDeleteWorkspacePipelineVar_NotFound(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspacePipelineVariablesFn: func(workspace string) ([]backend.PipelineVariable, error) {
			return []backend.PipelineVariable{}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.deleteWorkspacePipelineVar(context.Background(), makeReq(map[string]any{
		"workspace": "acme",
		"key":       "MISSING",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assertErrorResult(t, result, "not found")
}

func TestDeleteWorkspacePipelineVar_MissingKey(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.deleteWorkspacePipelineVar(context.Background(), makeReq(map[string]any{
		"workspace": "acme",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "key")
}
