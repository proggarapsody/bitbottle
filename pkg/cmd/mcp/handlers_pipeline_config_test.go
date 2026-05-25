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

func TestGetPipelineConfig_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelinesConfigFn: func(ws, slug string) (backend.PipelineConfig, error) {
			assert.Equal(t, "myws", ws)
			assert.Equal(t, "my-repo", slug)
			return backend.PipelineConfig{Enabled: true}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.getPipelineConfig(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "enabled", "true")
}

func TestGetPipelineConfig_MissingProject(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.getPipelineConfig(context.Background(), makeReq(map[string]any{
		"slug": "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestGetPipelineConfig_MissingSlug(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.getPipelineConfig(context.Background(), makeReq(map[string]any{
		"project": "myws",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "slug")
}

func TestGetPipelineConfig_UnsupportedOnHost(t *testing.T) {
	t.Parallel()
	type noPipelineConfigFake struct {
		backend.Client
	}
	const serverConfig = "git.example.com:\n  oauth_token: tok\n"
	base := &testhelpers.FakeClient{T: t}
	noPC := &noPipelineConfigFake{Client: base}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, noPC)
	h := newHandlers(f)
	result, err := h.getPipelineConfig(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "unsupported")
}

func TestEnablePipelines_Success(t *testing.T) {
	t.Parallel()
	var calledWith backend.PipelineConfig
	fake := &testhelpers.FakeClient{
		T: t,
		UpdatePipelinesConfigFn: func(ws, slug string, in backend.PipelineConfig) (backend.PipelineConfig, error) {
			assert.Equal(t, "myws", ws)
			assert.Equal(t, "my-repo", slug)
			calledWith = in
			return backend.PipelineConfig{Enabled: true}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.enablePipelines(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assert.True(t, calledWith.Enabled)
	text := extractText(t, result)
	assert.Contains(t, text, "enabled")
}

func TestDisablePipelines_Success(t *testing.T) {
	t.Parallel()
	var calledWith backend.PipelineConfig
	fake := &testhelpers.FakeClient{
		T: t,
		UpdatePipelinesConfigFn: func(ws, slug string, in backend.PipelineConfig) (backend.PipelineConfig, error) {
			assert.Equal(t, "myws", ws)
			assert.Equal(t, "my-repo", slug)
			calledWith = in
			return backend.PipelineConfig{Enabled: false}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.disablePipelines(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assert.False(t, calledWith.Enabled)
	text := extractText(t, result)
	assert.Contains(t, text, "disabled")
}
