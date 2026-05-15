package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestListPipelineCaches_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineCachesFn: func(ns, slug string) ([]backend.PipelineCache, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			return []backend.PipelineCache{
				{UUID: "cache-1", Name: "node_modules", Path: "/app/node_modules", FileSizeBytes: 12345678, CreatedOn: "2024-01-01T00:00:00.000Z"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listPipelineCaches(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "cache-1", "node_modules")
}

func TestListPipelineCaches_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listPipelineCaches(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}

func TestDeletePipelineCache_Success(t *testing.T) {
	t.Parallel()
	deleted := false
	fake := &testhelpers.FakeClient{
		T: t,
		DeletePipelineCacheFn: func(ns, slug, uuid string) error {
			assert.Equal(t, "cache-xyz", uuid)
			deleted = true
			return nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.deletePipelineCache(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"uuid": "cache-xyz",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "deleted", "")
	assert.True(t, deleted)
}

func TestDeletePipelineCache_MissingUUID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.deletePipelineCache(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "uuid")
}

func TestDeletePipelineCache_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.deletePipelineCache(context.Background(), makeReq(map[string]any{
		"uuid": "cache-xyz",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}
