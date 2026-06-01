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

func TestSearchCommits_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		SearchCommitsFn: func(ns, slug string, opts backend.CommitSearchOpts) ([]backend.Commit, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, "fix", opts.Query)
			assert.Equal(t, "alice", opts.Author)
			return []backend.Commit{
				{Hash: "abc1234", Message: "fix: null pointer", Author: backend.User{Slug: "alice"}},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.searchCommits(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
		"query":   "fix",
		"author":  "alice",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "abc1234", "fix: null pointer")
}

func TestSearchCommits_MissingProject(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.searchCommits(context.Background(), makeReq(map[string]any{
		"slug": "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestSearchCommits_MissingSlug(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.searchCommits(context.Background(), makeReq(map[string]any{
		"project": "myws",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "slug")
}

func TestSearchCommits_BackendError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		SearchCommitsFn: func(ns, slug string, opts backend.CommitSearchOpts) ([]backend.Commit, error) {
			return nil, errors.New("search API error")
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.searchCommits(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "search API error")
}

func TestSearchCommits_OptionalParams(t *testing.T) {
	t.Parallel()
	var capturedOpts backend.CommitSearchOpts
	fake := &testhelpers.FakeClient{
		T: t,
		SearchCommitsFn: func(ns, slug string, opts backend.CommitSearchOpts) ([]backend.Commit, error) {
			capturedOpts = opts
			return nil, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.searchCommits(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
		"since":   "2026-01-01",
		"until":   "2026-06-01",
		"limit":   float64(5),
	}))
	require.NoError(t, err)
	require.False(t, result.IsError)
	assert.Equal(t, "2026-01-01", capturedOpts.Since)
	assert.Equal(t, "2026-06-01", capturedOpts.Until)
	assert.Equal(t, 5, capturedOpts.Limit)
}
