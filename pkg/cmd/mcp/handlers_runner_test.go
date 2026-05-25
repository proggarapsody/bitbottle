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

// ── list_runners ──────────────────────────────────────────────────────────────

func TestListRunners_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListRunnersFn: func(workspace string) ([]backend.Runner, error) {
			assert.Equal(t, "myws", workspace)
			return []backend.Runner{
				{UUID: "runner-1", Name: "my-runner", State: "ONLINE",
					Platform: backend.RunnerPlatform{Operating: "LINUX", Arch: "AMD64"},
					Labels:   []string{"self.hosted", "linux"}},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listRunners(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "runner-1", "my-runner")
}

func TestListRunners_MissingWorkspace(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listRunners(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "workspace")
}

func TestListRunners_APIError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListRunnersFn: func(workspace string) ([]backend.Runner, error) {
			return nil, errors.New("403 Forbidden")
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listRunners(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

// ── create_runner ─────────────────────────────────────────────────────────────

func TestCreateRunner_Success(t *testing.T) {
	t.Parallel()
	var gotIn backend.CreateRunnerInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreateRunnerFn: func(workspace string, in backend.CreateRunnerInput) (backend.Runner, error) {
			assert.Equal(t, "myws", workspace)
			gotIn = in
			return backend.Runner{UUID: "new-uuid", Name: in.Name, State: "OFFLINE"}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.createRunner(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"name":      "my-runner",
		"platform":  "linux_amd64",
		"labels":    []any{"self.hosted", "linux"},
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "new-uuid", "")
	assert.Equal(t, "my-runner", gotIn.Name)
	assert.Equal(t, backend.RunnerPlatform{Operating: "LINUX", Arch: "AMD64"}, gotIn.Platform)
}

func TestCreateRunner_DefaultPlatform(t *testing.T) {
	t.Parallel()
	var gotIn backend.CreateRunnerInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreateRunnerFn: func(workspace string, in backend.CreateRunnerInput) (backend.Runner, error) {
			gotIn = in
			return backend.Runner{UUID: "u", Name: in.Name}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.createRunner(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"name":      "r",
	}))
	require.NoError(t, err)
	require.False(t, result.IsError)
	assert.Equal(t, backend.RunnerPlatform{Operating: "LINUX", Arch: "AMD64"}, gotIn.Platform)
}

func TestCreateRunner_MissingWorkspace(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.createRunner(context.Background(), makeReq(map[string]any{
		"name": "r",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "workspace")
}

func TestCreateRunner_MissingName(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.createRunner(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "name")
}

func TestCreateRunner_InvalidPlatform(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.createRunner(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"name":      "r",
		"platform":  "invalid_platform",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "invalid platform")
}

func TestCreateRunner_APIError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		CreateRunnerFn: func(workspace string, in backend.CreateRunnerInput) (backend.Runner, error) {
			return backend.Runner{}, errors.New("403 Forbidden")
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.createRunner(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"name":      "r",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

// ── delete_runner ─────────────────────────────────────────────────────────────

func TestDeleteRunner_Success(t *testing.T) {
	t.Parallel()
	deleted := false
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteRunnerFn: func(workspace, runnerUUID string) error {
			assert.Equal(t, "myws", workspace)
			assert.Equal(t, "runner-xyz", runnerUUID)
			deleted = true
			return nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.deleteRunner(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"uuid":      "runner-xyz",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "deleted", "")
	assert.True(t, deleted)
}

func TestDeleteRunner_MissingWorkspace(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.deleteRunner(context.Background(), makeReq(map[string]any{
		"uuid": "runner-xyz",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "workspace")
}

func TestDeleteRunner_MissingUUID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.deleteRunner(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "uuid")
}

func TestDeleteRunner_APIError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteRunnerFn: func(workspace, runnerUUID string) error {
			return errors.New("403 Forbidden")
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.deleteRunner(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"uuid":      "runner-xyz",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
