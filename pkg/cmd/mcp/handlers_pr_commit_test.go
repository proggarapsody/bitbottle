package mcp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestListPRCommits_Success(t *testing.T) {
	t.Parallel()
	ts := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRCommitsFn: func(ns, slug string, prID int) ([]backend.Commit, error) {
			assert.Equal(t, "MYPROJ", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, 42, prID)
			return []backend.Commit{
				{
					Hash:      "abc1234def456abc1234def456abc1234def456ab",
					Message:   "Fix null pointer in auth",
					Author:    backend.User{Slug: "alice", DisplayName: "Alice Smith"},
					Timestamp: ts,
				},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listPRCommits(context.Background(), makeReq(map[string]any{
		"repo":  "myproj/my-repo",
		"pr_id": float64(42),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "abc1234", "Fix null pointer")
}

func TestListPRCommits_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listPRCommits(context.Background(), makeReq(map[string]any{
		"pr_id": float64(42),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}

func TestListPRCommits_MissingPRID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listPRCommits(context.Background(), makeReq(map[string]any{
		"repo": "myproj/my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "pr_id")
}

func TestListPRCommits_APIError_ReturnsErrorResult(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRCommitsFn: func(ns, slug string, prID int) ([]backend.Commit, error) {
			return nil, errors.New("500 internal server error")
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listPRCommits(context.Background(), makeReq(map[string]any{
		"repo":  "myproj/my-repo",
		"pr_id": float64(42),
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
