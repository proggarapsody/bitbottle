package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestGetDiff_Success(t *testing.T) {
	t.Parallel()
	const diffText = "--- a/foo.go\n+++ b/foo.go\n@@ -1 +1 @@\n-old\n+new\n"
	fake := &testhelpers.FakeClient{
		T: t,
		GetDiffFn: func(ns, slug, from, to string) (string, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, "main", from)
			assert.Equal(t, "feature", to)
			return diffText, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.getDiff(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"from": "main",
		"to":   "feature",
	}))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError, "expected success result, got error")
	text := extractText(t, result)
	assert.Equal(t, diffText, text)
}

func TestGetDiff_MissingFrom(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.getDiff(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"to":   "feature",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "from")
}

func TestGetDiff_MissingTo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.getDiff(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"from": "main",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "to")
}

func TestGetDiff_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.getDiff(context.Background(), makeReq(map[string]any{
		"from": "main",
		"to":   "feature",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}

func TestGetDiffStat_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetDiffStatFn: func(ns, slug, from, to string) (backend.DiffStat, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, "main", from)
			assert.Equal(t, "feature", to)
			return backend.DiffStat{
				FilesChanged: 1,
				Additions:    5,
				Deletions:    2,
				Files: []backend.DiffStatEntry{
					{Path: "api/foo.go", Status: "modified", Additions: 5, Deletions: 2},
				},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.getDiffStat(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"from": "main",
		"to":   "feature",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "FilesChanged", "")
	assertJSONContains(t, result, "api/foo.go", "")
}

func TestGetDiffStat_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.getDiffStat(context.Background(), makeReq(map[string]any{
		"from": "main",
		"to":   "feature",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}
