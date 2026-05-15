package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestRepoVisibility_GetPrivate_ReturnsModeText(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		GetRepoFn: func(ns, slug string) (backend.Repository, error) {
			return backend.Repository{
				Slug:      slug,
				Namespace: ns,
				IsPrivate: true,
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.repoVisibility(context.Background(), makeReq(map[string]any{
		"repo": "MYPROJ/my-service",
	}))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
	assert.Equal(t, "private", extractText(t, result))
}

func TestRepoVisibility_GetPublic_ReturnsModeText(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		GetRepoFn: func(ns, slug string) (backend.Repository, error) {
			return backend.Repository{
				Slug:      slug,
				Namespace: ns,
				IsPrivate: false,
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.repoVisibility(context.Background(), makeReq(map[string]any{
		"repo": "MYPROJ/my-service",
	}))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
	assert.Equal(t, "public", extractText(t, result))
}

func TestRepoVisibility_SetPublic_CallsBackend(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug string
	var gotPrivate bool
	fake := &testhelpers.FakeClient{
		SetRepoVisibilityFn: func(ns, slug string, isPrivate bool) error {
			gotNS, gotSlug, gotPrivate = ns, slug, isPrivate
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.repoVisibility(context.Background(), makeReq(map[string]any{
		"repo":       "MYPROJ/my-service",
		"visibility": "public",
	}))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-service", gotSlug)
	assert.False(t, gotPrivate)
}

func TestRepoVisibility_InvalidVisibility_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.repoVisibility(context.Background(), makeReq(map[string]any{
		"repo":       "MYPROJ/my-service",
		"visibility": "internal",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "invalid visibility")
}
