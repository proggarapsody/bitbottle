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

func TestGetPRActivity_ReturnsEvents(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	var gotProject, gotSlug string
	var gotID, gotLimit int
	fake := &testhelpers.FakeClient{
		GetPRActivityFn: func(ns, slug string, id int, limit int) ([]backend.PRActivityEvent, error) {
			gotProject = ns
			gotSlug = slug
			gotID = id
			gotLimit = limit
			return []backend.PRActivityEvent{
				{Type: "approval", Actor: backend.User{Slug: "alice", DisplayName: "Alice"}, CreatedAt: now},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.getPRActivity(context.Background(), makeReq(map[string]any{
		"project": "myproj",
		"slug":    "my-repo",
		"pr_id":   float64(42),
		"limit":   float64(10),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "approval", "")
	assertJSONContains(t, result, "alice", "")
	assert.Equal(t, "MYPROJ", gotProject)
	assert.Equal(t, "my-repo", gotSlug)
	assert.Equal(t, 42, gotID)
	assert.Equal(t, 10, gotLimit)
}

func TestGetPRActivity_MissingProjectReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.getPRActivity(context.Background(), makeReq(map[string]any{
		"slug":  "my-repo",
		"pr_id": float64(42),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestGetPRActivity_BackendError_ReturnsErrorResult(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		GetPRActivityFn: func(ns, slug string, id int, limit int) ([]backend.PRActivityEvent, error) {
			return nil, errors.New("upstream failure")
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.getPRActivity(context.Background(), makeReq(map[string]any{
		"project": "myproj",
		"slug":    "my-repo",
		"pr_id":   float64(1),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "upstream failure")
}
