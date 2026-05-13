package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestListCommitFiles_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListCommitFilesFn: func(ns, slug, hash string) ([]backend.DiffStatEntry, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, "abc123", hash)
			return []backend.DiffStatEntry{
				{Path: "foo.go", Status: "modified", Additions: 5, Deletions: 2},
				{Path: "bar.go", Status: "added", Additions: 10, Deletions: 0},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listCommitFiles(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"hash": "abc123",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "foo.go", "modified")
	assertJSONContains(t, result, "bar.go", "added")
}

func TestListCommitFiles_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listCommitFiles(context.Background(), makeReq(map[string]any{
		"hash": "abc123",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}

func TestListCommitFiles_MissingHash(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listCommitFiles(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "hash")
}
