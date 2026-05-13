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

func TestListRepoWatchers_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListRepoWatchersFn: func(ns, slug string) ([]backend.User, error) {
			assert.Equal(t, "myproj", ns)
			assert.Equal(t, "my-repo", slug)
			return []backend.User{
				{Slug: "alice", DisplayName: "Alice Smith"},
				{Slug: "bob", DisplayName: "Bob Jones"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listRepoWatchers(context.Background(), makeReq(map[string]any{
		"repo": "myproj/my-repo",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "alice", "Alice Smith")
}

func TestListRepoWatchers_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listRepoWatchers(context.Background(), makeReq(map[string]any{}))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}

func TestListRepoWatchers_APIError_ReturnsErrorResult(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListRepoWatchersFn: func(ns, slug string) ([]backend.User, error) {
			return nil, errors.New("500 internal server error")
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listRepoWatchers(context.Background(), makeReq(map[string]any{
		"repo": "myproj/my-repo",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestListRepoWatchers_UnsupportedBackend_ReturnsErrorResult(t *testing.T) {
	t.Parallel()
	// FakeClient does not implement RepoWatcherClient unless ListRepoWatchersFn is set,
	// but actually it does (the method is defined). Instead we test with a client
	// that returns an error from the API.
	fake := &testhelpers.FakeClient{
		T: t,
		ListRepoWatchersFn: func(ns, slug string) ([]backend.User, error) {
			return nil, &backend.DomainError{
				Kind:    backend.ErrUnsupportedOnHost,
				Code:    backend.CodeHostUnsupported,
				Host:    "bb.example.com",
				Feature: "repo_watchers",
				Message: "repo watchers are not supported on bb.example.com",
			}
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listRepoWatchers(context.Background(), makeReq(map[string]any{
		"repo": "myproj/my-repo",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
