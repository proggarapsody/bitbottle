package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestListDefaultReviewers_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListDefaultReviewersFn: func(ns, slug string) ([]backend.DefaultReviewer, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			return []backend.DefaultReviewer{
				{UserSlug: "alice", DisplayName: "Alice Smith", EmailAddress: "alice@co.com"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listDefaultReviewers(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "alice", "Alice Smith")
}

func TestListDefaultReviewers_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listDefaultReviewers(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}

func TestAddDefaultReviewer_Success(t *testing.T) {
	t.Parallel()
	var gotUser string
	fake := &testhelpers.FakeClient{
		T: t,
		AddDefaultReviewerFn: func(ns, slug, userSlug string) error {
			gotUser = userSlug
			return nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.addDefaultReviewer(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"user": "alice",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "added", "alice")
	assert.Equal(t, "alice", gotUser)
}

func TestAddDefaultReviewer_MissingUser(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.addDefaultReviewer(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "user")
}

func TestRemoveDefaultReviewer_Success(t *testing.T) {
	t.Parallel()
	removed := false
	fake := &testhelpers.FakeClient{
		T: t,
		RemoveDefaultReviewerFn: func(ns, slug, userSlug string) error {
			assert.Equal(t, "alice", userSlug)
			removed = true
			return nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.removeDefaultReviewer(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"user": "alice",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "removed", "alice")
	assert.True(t, removed)
}

func TestRemoveDefaultReviewer_MissingUser(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.removeDefaultReviewer(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "user")
}
