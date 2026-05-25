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

func TestListMilestones_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListMilestonesFn: func(ns, slug string, limit int) ([]backend.Milestone, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			return []backend.Milestone{
				{ID: 1, Name: "v1.0"},
				{ID: 2, Name: "v2.0"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listMilestones(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "name", "v1.0")
}

func TestListMilestones_MissingSlug(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listMilestones(context.Background(), makeReq(map[string]any{
		"project": "myws",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "slug")
}

func TestListMilestones_UnsupportedOnHost(t *testing.T) {
	t.Parallel()
	type noMilestoneFake struct{ backend.Client }
	const serverConfig = "git.example.com:\n  oauth_token: tok\n"
	base := &testhelpers.FakeClient{T: t}
	noM := &noMilestoneFake{Client: base}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, noM)
	h := newHandlers(f)
	result, err := h.listMilestones(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "unsupported")
}

func TestViewMilestone_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetMilestoneFn: func(ns, slug string, id int) (backend.Milestone, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, 1, id)
			return backend.Milestone{ID: 1, Name: "v1.0"}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.viewMilestone(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
		"id":      float64(1),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "name", "v1.0")
}

func TestViewMilestone_MissingID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.viewMilestone(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "id")
}
