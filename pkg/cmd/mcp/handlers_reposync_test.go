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

func TestSyncRepo_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		SyncRepoFn: func(ns, slug, branch string) (backend.SyncResult, error) {
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "my-fork", slug)
			assert.Equal(t, "main", branch)
			return backend.SyncResult{Behind: 3, CommitsMerged: 3}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.syncRepo(context.Background(), makeReq(map[string]any{
		"ns":     "myworkspace",
		"slug":   "my-fork",
		"branch": "main",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "commits_merged", "3")
}

func TestSyncRepo_AlreadyUpToDate(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		SyncRepoFn: func(ns, slug, branch string) (backend.SyncResult, error) {
			return backend.SyncResult{Behind: 0, CommitsMerged: 0}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.syncRepo(context.Background(), makeReq(map[string]any{
		"ns":   "myworkspace",
		"slug": "my-fork",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "commits_merged", "0")
}

func TestSyncRepo_MissingNS(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.syncRepo(context.Background(), makeReq(map[string]any{
		"slug": "my-fork",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "ns")
}

func TestSyncRepo_MissingSlug(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.syncRepo(context.Background(), makeReq(map[string]any{
		"ns": "myworkspace",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "slug")
}

func TestSyncRepo_APIError_ReturnsErrorResult(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		SyncRepoFn: func(ns, slug, branch string) (backend.SyncResult, error) {
			return backend.SyncResult{}, errors.New("upstream not configured")
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.syncRepo(context.Background(), makeReq(map[string]any{
		"ns":   "myworkspace",
		"slug": "my-fork",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestSyncRepo_UnsupportedOnServer(t *testing.T) {
	t.Parallel()
	// FakeClient does not satisfy RepoSyncer when SyncRepoFn is nil.
	// With SyncRepoFn set it does satisfy it, so test via AsRepoSyncer
	// returning typed error by using a client without the capability.
	// We rely on the host.unsupported DomainError path in handlers.
	fake := &testhelpers.FakeClient{T: t}
	// Leave SyncRepoFn nil — FakeClient still satisfies RepoSyncer
	// (via the compile-time assertion), so AsRepoSyncer will succeed.
	// We call directly to verify the result is not an error.
	fake.SyncRepoFn = func(ns, slug, branch string) (backend.SyncResult, error) {
		return backend.SyncResult{}, &backend.DomainError{
			Kind: backend.ErrUnsupportedOnHost,
			Code: backend.CodeHostUnsupported,
		}
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.syncRepo(context.Background(), makeReq(map[string]any{
		"ns":   "myworkspace",
		"slug": "my-fork",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
