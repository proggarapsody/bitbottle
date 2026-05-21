package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestListSnippets_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListSnippetsFn: func(workspace string, limit int) ([]backend.Snippet, error) {
			assert.Equal(t, "myws", workspace)
			return []backend.Snippet{
				{ID: "abc123", Title: "Test snippet", Owner: "myws"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listSnippets(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "abc123", "Test snippet")
}

func TestListSnippets_MissingWorkspace(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listSnippets(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "workspace")
}

func TestListSnippets_UnsupportedBackend(t *testing.T) {
	t.Parallel()
	type noSnippetFake struct{ backend.Client }
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: singleCloudConfig})
	factorytest.UseBackend(f, noSnippetFake{Client: &testhelpers.FakeClient{T: t}})
	h := newHandlers(f)
	result, err := h.listSnippets(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestViewSnippet_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetSnippetFn: func(workspace, id string) (backend.Snippet, error) {
			assert.Equal(t, "myws", workspace)
			assert.Equal(t, "Xqjyp1GV", id)
			return backend.Snippet{
				ID:        "Xqjyp1GV",
				Title:     "My snippet",
				Owner:     "myws",
				IsPrivate: false,
				CreatedOn: time.Date(2026, 1, 10, 8, 0, 0, 0, time.UTC),
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.viewSnippet(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"id":        "Xqjyp1GV",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "Xqjyp1GV", "My snippet")
}

func TestViewSnippet_MissingID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.viewSnippet(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "id")
}

func TestCreateSnippet_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		CreateSnippetFn: func(workspace string, in backend.CreateSnippetInput) (backend.Snippet, error) {
			assert.Equal(t, "myws", workspace)
			assert.Equal(t, "My new snippet", in.Title)
			assert.True(t, in.IsPrivate)
			return backend.Snippet{
				ID:    "newsnip",
				Title: in.Title,
				Owner: workspace,
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.createSnippet(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"title":     "My new snippet",
		"private":   true,
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "newsnip", "My new snippet")
}

func TestCreateSnippet_MissingTitle(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.createSnippet(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "title")
}

func TestDeleteSnippet_Success(t *testing.T) {
	t.Parallel()
	deleted := false
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteSnippetFn: func(workspace, id string) error {
			assert.Equal(t, "myws", workspace)
			assert.Equal(t, "Xqjyp1GV", id)
			deleted = true
			return nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.deleteSnippet(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"id":        "Xqjyp1GV",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "deleted", "")
	assert.True(t, deleted)
}

func TestDeleteSnippet_MissingID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.deleteSnippet(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "id")
}
