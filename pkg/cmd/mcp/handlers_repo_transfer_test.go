package mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestTransferRepo_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		TransferRepoFn: func(ns, slug, target string) (backend.Repository, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, "newws", target)
			return backend.Repository{Slug: slug, Name: slug, Namespace: target}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.transferRepo(context.Background(), makeReq(map[string]any{
		"repo":   "myws/my-repo",
		"target": "newws",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "my-repo", "newws")
}

func TestTransferRepo_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.transferRepo(context.Background(), makeReq(map[string]any{
		"target": "newws",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}

func TestTransferRepo_MissingTarget(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.transferRepo(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "target")
}

func TestTransferRepo_APIError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		TransferRepoFn: func(ns, slug, target string) (backend.Repository, error) {
			return backend.Repository{}, errors.New("403 forbidden")
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.transferRepo(context.Background(), makeReq(map[string]any{
		"repo":   "myws/my-repo",
		"target": "newws",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "403")
}
