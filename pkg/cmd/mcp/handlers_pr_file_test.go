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

func TestListPRFiles_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRFilesFn: func(ns, slug string, prID int) ([]backend.DiffStatEntry, error) {
			assert.Equal(t, "MYPROJ", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, 42, prID)
			return []backend.DiffStatEntry{
				{
					Path:      "foo.go",
					Status:    "added",
					Additions: 10,
					Deletions: 0,
				},
				{
					Path:      "bar.go",
					Status:    "modified",
					Additions: 3,
					Deletions: 2,
				},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listPRFiles(context.Background(), makeReq(map[string]any{
		"project": "myproj",
		"slug":    "my-repo",
		"pr_id":   float64(42),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "foo.go", "added")
}

// MCP-04: legacy {repo} shape still works, with a deprecation note prepended.
func TestListPRFiles_LegacyRepoShape_StillWorksWithDeprecation(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRFilesFn: func(ns, slug string, prID int) ([]backend.DiffStatEntry, error) {
			assert.Equal(t, "MYPROJ", ns)
			assert.Equal(t, "my-repo", slug)
			return []backend.DiffStatEntry{{Path: "foo.go", Status: "added"}}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listPRFiles(context.Background(), makeReq(map[string]any{
		"repo":  "myproj/my-repo",
		"pr_id": float64(42),
	}))
	require.NoError(t, err)
	require.False(t, result.IsError)
	require.Len(t, result.Content, 2)
	assert.Contains(t, extractTextAt(t, result, 0), "DEPRECATION")
}

func TestListPRFiles_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listPRFiles(context.Background(), makeReq(map[string]any{
		"pr_id": float64(42),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestListPRFiles_MissingPRID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listPRFiles(context.Background(), makeReq(map[string]any{
		"repo": "myproj/my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "pr_id")
}

func TestListPRFiles_APIError_ReturnsErrorResult(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRFilesFn: func(ns, slug string, prID int) ([]backend.DiffStatEntry, error) {
			return nil, errors.New("500 internal server error")
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listPRFiles(context.Background(), makeReq(map[string]any{
		"repo":  "myproj/my-repo",
		"pr_id": float64(42),
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
