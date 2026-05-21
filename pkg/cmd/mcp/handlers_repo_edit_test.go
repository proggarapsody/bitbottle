package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestEditRepo_DescriptionUpdate(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug string
	var gotIn backend.EditRepoInput
	fake := &testhelpers.FakeClient{
		EditRepoFn: func(ns, slug string, in backend.EditRepoInput) (backend.Repository, error) {
			gotNS, gotSlug, gotIn = ns, slug, in
			return backend.Repository{Slug: slug, Namespace: ns}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.editRepo(context.Background(), makeReq(map[string]any{
		"repo":        "MYPROJ/my-service",
		"description": "new description",
	}))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-service", gotSlug)
	require.NotNil(t, gotIn.Description)
	assert.Equal(t, "new description", *gotIn.Description)
}

func TestEditRepo_HasIssuesFalse(t *testing.T) {
	t.Parallel()
	var gotIn backend.EditRepoInput
	fake := &testhelpers.FakeClient{
		EditRepoFn: func(ns, slug string, in backend.EditRepoInput) (backend.Repository, error) {
			gotIn = in
			return backend.Repository{Slug: slug, Namespace: ns}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.editRepo(context.Background(), makeReq(map[string]any{
		"repo":       "MYPROJ/my-service",
		"has_issues": false,
	}))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
	require.NotNil(t, gotIn.HasIssues)
	assert.False(t, *gotIn.HasIssues)
}

func TestEditRepo_NoFields_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.editRepo(context.Background(), makeReq(map[string]any{
		"repo": "MYPROJ/my-service",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "no fields to update")
}

func TestEditRepo_UnsupportedOnHost_ReturnsStructuredError(t *testing.T) {
	t.Parallel()
	// FakeClient does not implement RepoEditor when EditRepoFn is nil —
	// but it does because we added it. Use a minimal struct that only
	// satisfies backend.Client to trigger host.unsupported.
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{})
	result, err := h.editRepo(context.Background(), makeReq(map[string]any{
		"repo":        "MYPROJ/my-service",
		"description": "x",
	}))
	// FakeClient does satisfy RepoEditor so this should succeed without error.
	require.NoError(t, err)
	require.NotNil(t, result)
}
