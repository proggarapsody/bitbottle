package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestListRepoLabels_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListRepoLabelsFn: func(ns, slug string) ([]backend.RepoLabel, error) {
			assert.Equal(t, "PROJ", ns)
			assert.Equal(t, "myrepo", slug)
			return []backend.RepoLabel{
				{ID: 1, Name: "bug", Color: "#ff0000"},
				{ID: 2, Name: "enhancement", Color: "#00ff00"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listRepoLabels(context.Background(), makeReq(map[string]any{
		"project": "PROJ",
		"slug":    "myrepo",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "bug", "#ff0000")
}

func TestListRepoLabels_MissingProject(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listRepoLabels(context.Background(), makeReq(map[string]any{
		"slug": "myrepo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestCreateRepoLabel_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		CreateRepoLabelFn: func(ns, slug string, in backend.CreateRepoLabelInput) (backend.RepoLabel, error) {
			assert.Equal(t, "PROJ", ns)
			assert.Equal(t, "myrepo", slug)
			assert.Equal(t, "bug", in.Name)
			assert.Equal(t, "#ff0000", in.Color)
			return backend.RepoLabel{ID: 3, Name: "bug", Color: "#ff0000"}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.createRepoLabel(context.Background(), makeReq(map[string]any{
		"project": "PROJ",
		"slug":    "myrepo",
		"name":    "bug",
		"color":   "#ff0000",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "bug", "#ff0000")
}

func TestCreateRepoLabel_MissingName(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.createRepoLabel(context.Background(), makeReq(map[string]any{
		"project": "PROJ",
		"slug":    "myrepo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "name")
}

func TestUpdateRepoLabel_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		UpdateRepoLabelFn: func(ns, slug string, id int, in backend.UpdateRepoLabelInput) (backend.RepoLabel, error) {
			assert.Equal(t, 3, id)
			assert.Equal(t, "bug-fixed", in.Name)
			return backend.RepoLabel{ID: 3, Name: "bug-fixed", Color: "#ff0000"}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.updateRepoLabel(context.Background(), makeReq(map[string]any{
		"project": "PROJ",
		"slug":    "myrepo",
		"id":      float64(3),
		"name":    "bug-fixed",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "bug-fixed", "")
}

func TestDeleteRepoLabel_Success(t *testing.T) {
	t.Parallel()
	called := false
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteRepoLabelFn: func(ns, slug string, id int) error {
			assert.Equal(t, 5, id)
			called = true
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.deleteRepoLabel(context.Background(), makeReq(map[string]any{
		"project": "PROJ",
		"slug":    "myrepo",
		"id":      float64(5),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "deleted", "true")
	assert.True(t, called)
}
