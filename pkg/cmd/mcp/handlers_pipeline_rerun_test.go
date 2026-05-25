package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestRerunPipeline_CallsClientWithCorrectParams(t *testing.T) {
	t.Parallel()
	called := false
	fake := &testhelpers.FakeClient{
		T: t,
		RerunPipelineFn: func(ns, slug, srcUUID string) (backend.Pipeline, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, "abc-123", srcUUID)
			called = true
			return backend.Pipeline{
				UUID:        "newuuid",
				BuildNumber: 42,
				WebURL:      "https://bitbucket.org/myws/my-repo/pipelines/results/42",
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.rerunPipeline(context.Background(), makeReq(map[string]any{
		"repo":          "myws/my-repo",
		"pipeline_uuid": "abc-123",
	}))
	require.NoError(t, err)
	assert.True(t, called)
	assertJSONContains(t, result, "42", "")
}

func TestRerunPipeline_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.rerunPipeline(context.Background(), makeReq(map[string]any{
		"pipeline_uuid": "abc-123",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}

func TestRerunPipeline_MissingPipelineUUID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.rerunPipeline(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "pipeline_uuid")
}
