package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestListSnippetComments_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListSnippetCommentsFn: func(workspace, snippetID string, limit int) ([]backend.SnippetComment, error) {
			assert.Equal(t, "myws", workspace)
			assert.Equal(t, "abc123", snippetID)
			return []backend.SnippetComment{
				{ID: 17, Author: "Alice", Body: "Nice snippet!"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listSnippetComments(context.Background(), makeReq(map[string]any{
		"workspace":  "myws",
		"snippet_id": "abc123",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "17", "Nice snippet!")
}

func TestListSnippetComments_MissingSnippetID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listSnippetComments(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "snippet_id")
}

func TestListSnippetComments_MissingWorkspace(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listSnippetComments(context.Background(), makeReq(map[string]any{
		"snippet_id": "abc123",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "workspace")
}

func TestListSnippetComments_UnsupportedBackend(t *testing.T) {
	t.Parallel()
	type noSnippetFake struct{ backend.Client }
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: singleCloudConfig})
	factorytest.UseBackend(f, noSnippetFake{Client: &testhelpers.FakeClient{T: t}})
	h := newHandlers(f)
	result, err := h.listSnippetComments(context.Background(), makeReq(map[string]any{
		"workspace":  "myws",
		"snippet_id": "abc123",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestAddSnippetComment_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		AddSnippetCommentFn: func(workspace, snippetID, body string) (backend.SnippetComment, error) {
			assert.Equal(t, "myws", workspace)
			assert.Equal(t, "abc123", snippetID)
			assert.Equal(t, "Great work!", body)
			return backend.SnippetComment{ID: 42, Body: body, Author: "Alice"}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.addSnippetComment(context.Background(), makeReq(map[string]any{
		"workspace":  "myws",
		"snippet_id": "abc123",
		"body":       "Great work!",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "42", "Great work!")
}

func TestAddSnippetComment_MissingBody(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.addSnippetComment(context.Background(), makeReq(map[string]any{
		"workspace":  "myws",
		"snippet_id": "abc123",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "body")
}

func TestAddSnippetComment_MissingSnippetID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.addSnippetComment(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"body":      "hi",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "snippet_id")
}

func TestDeleteSnippetComment_Success(t *testing.T) {
	t.Parallel()
	deleted := false
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteSnippetCommentFn: func(workspace, snippetID string, commentID int) error {
			assert.Equal(t, "myws", workspace)
			assert.Equal(t, "abc123", snippetID)
			assert.Equal(t, 17, commentID)
			deleted = true
			return nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.deleteSnippetComment(context.Background(), makeReq(map[string]any{
		"workspace":  "myws",
		"snippet_id": "abc123",
		"comment_id": float64(17),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "deleted", "")
	assert.True(t, deleted)
}

func TestDeleteSnippetComment_MissingCommentID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.deleteSnippetComment(context.Background(), makeReq(map[string]any{
		"workspace":  "myws",
		"snippet_id": "abc123",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "comment_id")
}

func TestDeleteSnippetComment_MissingSnippetID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.deleteSnippetComment(context.Background(), makeReq(map[string]any{
		"workspace":  "myws",
		"comment_id": float64(17),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "snippet_id")
}
