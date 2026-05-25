package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestListIssueAttachments_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListIssueAttachmentsFn: func(ns, slug string, id int) ([]backend.IssueAttachment, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "myrepo", slug)
			assert.Equal(t, 42, id)
			a := backend.IssueAttachment{Name: "screenshot.png", Size: 12345}
			a.Links.Self = "https://api.bitbucket.org/2.0/repositories/myws/myrepo/issues/42/attachments/screenshot.png"
			return []backend.IssueAttachment{a}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listIssueAttachments(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "myrepo",
		"id":      float64(42),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "screenshot.png", "12345")
}

func TestListIssueAttachments_MissingID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listIssueAttachments(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "myrepo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "id")
}

func TestDeleteIssueAttachment_Success(t *testing.T) {
	t.Parallel()
	called := false
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteIssueAttachmentFn: func(ns, slug string, id int, filename string) error {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "myrepo", slug)
			assert.Equal(t, 7, id)
			assert.Equal(t, "file.txt", filename)
			called = true
			return nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.deleteIssueAttachment(context.Background(), makeReq(map[string]any{
		"project":  "myws",
		"slug":     "myrepo",
		"id":       float64(7),
		"filename": "file.txt",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "deleted", "true")
	assert.True(t, called)
}

func TestVoteIssue_Success(t *testing.T) {
	t.Parallel()
	called := false
	fake := &testhelpers.FakeClient{
		T: t,
		VoteIssueFn: func(ns, slug string, id int) error {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "myrepo", slug)
			assert.Equal(t, 5, id)
			called = true
			return nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.voteIssue(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "myrepo",
		"id":      float64(5),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "voted", "true")
	assert.True(t, called)
}

func TestUnvoteIssue_Success(t *testing.T) {
	t.Parallel()
	called := false
	fake := &testhelpers.FakeClient{
		T: t,
		UnvoteIssueFn: func(ns, slug string, id int) error {
			called = true
			return nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.unvoteIssue(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "myrepo",
		"id":      float64(5),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "voted", "false")
	assert.True(t, called)
}

func TestWatchIssue_Success(t *testing.T) {
	t.Parallel()
	called := false
	fake := &testhelpers.FakeClient{
		T: t,
		WatchIssueFn: func(ns, slug string, id int) error {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "myrepo", slug)
			assert.Equal(t, 9, id)
			called = true
			return nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.watchIssue(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "myrepo",
		"id":      float64(9),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "watching", "true")
	assert.True(t, called)
}

func TestUnwatchIssue_Success(t *testing.T) {
	t.Parallel()
	called := false
	fake := &testhelpers.FakeClient{
		T: t,
		UnwatchIssueFn: func(ns, slug string, id int) error {
			called = true
			return nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.unwatchIssue(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "myrepo",
		"id":      float64(9),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "watching", "false")
	assert.True(t, called)
}

func TestWatchIssue_MissingID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.watchIssue(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "myrepo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "id")
}
