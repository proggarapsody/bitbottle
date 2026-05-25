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

func TestListMirrorServers_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListMirrorServersFn: func(limit int) ([]backend.MirrorServer, error) {
			return []backend.MirrorServer{
				{ID: "mirror-1", Name: "Primary Mirror", BaseURL: "https://mirror.example.com", Enabled: true},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listMirrorServers(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertJSONContains(t, result, "mirror-1", "Primary Mirror")
}

func TestListMirrorServers_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	type noMirrorFake struct{ backend.Client }
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: singleHostConfig})
	factorytest.UseBackend(f, noMirrorFake{Client: &testhelpers.FakeClient{T: t}})
	h := newHandlers(f)
	result, err := h.listMirrorServers(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestViewMirrorServer_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetMirrorServerFn: func(id string) (backend.MirrorServer, error) {
			assert.Equal(t, "mirror-1", id)
			return backend.MirrorServer{ID: "mirror-1", Name: "Primary Mirror", BaseURL: "https://mirror.example.com", Enabled: true}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.viewMirrorServer(context.Background(), makeReq(map[string]any{
		"id": "mirror-1",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "Primary Mirror", "https://mirror.example.com")
}

func TestViewMirrorServer_MissingID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.viewMirrorServer(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "id")
}

func TestListMirroredRepos_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListMirroredReposFn: func(mirrorID string, limit int) ([]backend.MirroredRepo, error) {
			assert.Equal(t, "mirror-1", mirrorID)
			return []backend.MirroredRepo{
				{Slug: "test-repo", MirrorID: "mirror-1", Status: "AVAILABLE"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listMirroredRepos(context.Background(), makeReq(map[string]any{
		"mirror_id": "mirror-1",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "test-repo", "AVAILABLE")
}

func TestListMirroredRepos_MissingMirrorID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listMirroredRepos(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "mirror_id")
}
