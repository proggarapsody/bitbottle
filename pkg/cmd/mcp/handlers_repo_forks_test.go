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

func TestListRepoForks_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListRepoForksFn: func(ns, slug string, limit int) ([]backend.Repository, error) {
			assert.Equal(t, "myproj", ns)
			assert.Equal(t, "my-repo", slug)
			return []backend.Repository{
				{Namespace: "teamA", Slug: "my-repo-fork", Name: "my-repo-fork"},
				{Namespace: "teamB", Slug: "another-fork", Name: "another-fork"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listRepoForks(context.Background(), makeReq(map[string]any{
		"repo": "myproj/my-repo",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "teamA", "my-repo-fork")
}

func TestListRepoForks_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listRepoForks(context.Background(), makeReq(map[string]any{}))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}

func TestListRepoForks_APIError_ReturnsErrorResult(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListRepoForksFn: func(ns, slug string, limit int) ([]backend.Repository, error) {
			return nil, errors.New("500 internal server error")
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listRepoForks(context.Background(), makeReq(map[string]any{
		"repo": "myproj/my-repo",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
