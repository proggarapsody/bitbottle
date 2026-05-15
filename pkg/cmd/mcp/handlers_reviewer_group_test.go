package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestListReviewerGroups_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListReviewerGroupsFn: func(ns, slug string) ([]backend.ReviewerGroup, error) {
			assert.Equal(t, "PROJ", ns)
			assert.Equal(t, "my-repo", slug)
			return []backend.ReviewerGroup{
				{ID: 1, Name: "team-a", RequiredApprovals: 1, Reviewers: []backend.User{{Slug: "alice"}}},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listReviewerGroups(context.Background(), makeReq(map[string]any{
		"repo": "PROJ/my-repo",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "team-a", "alice")
}

func TestListReviewerGroups_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listReviewerGroups(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}

func TestAddReviewerGroup_Success(t *testing.T) {
	t.Parallel()
	var gotInput backend.CreateReviewerGroupInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreateReviewerGroupFn: func(ns, slug string, in backend.CreateReviewerGroupInput) (backend.ReviewerGroup, error) {
			gotInput = in
			return backend.ReviewerGroup{
				ID:                5,
				Name:              in.Name,
				RequiredApprovals: in.RequiredApprovals,
				Reviewers:         []backend.User{{Slug: "alice"}},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.addReviewerGroup(context.Background(), makeReq(map[string]any{
		"repo":  "PROJ/my-repo",
		"name":  "team-a",
		"users": "alice,bob",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "team-a", "alice")
	assert.Equal(t, "team-a", gotInput.Name)
	assert.Equal(t, []string{"alice", "bob"}, gotInput.UserSlugs)
	assert.Equal(t, 1, gotInput.RequiredApprovals)
}

func TestAddReviewerGroup_MissingName(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.addReviewerGroup(context.Background(), makeReq(map[string]any{
		"repo":  "PROJ/my-repo",
		"users": "alice",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "name")
}

func TestAddReviewerGroup_MissingUsers(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.addReviewerGroup(context.Background(), makeReq(map[string]any{
		"repo": "PROJ/my-repo",
		"name": "team-a",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "users")
}

func TestRemoveReviewerGroup_Success(t *testing.T) {
	t.Parallel()
	var gotID int
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteReviewerGroupFn: func(ns, slug string, id int) error {
			gotID = id
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.removeReviewerGroup(context.Background(), makeReq(map[string]any{
		"repo": "PROJ/my-repo",
		"id":   float64(7),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "removed", "")
	assert.Equal(t, 7, gotID)
}

func TestRemoveReviewerGroup_MissingID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.removeReviewerGroup(context.Background(), makeReq(map[string]any{
		"repo": "PROJ/my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "id")
}
