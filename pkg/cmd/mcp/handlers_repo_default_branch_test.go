package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestSetRepoDefaultBranch_SetsCorrectly(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug, gotBranch string
	fake := &testhelpers.FakeClient{
		SetRepoDefaultBranchFn: func(ns, slug, branch string) error {
			gotNS, gotSlug, gotBranch = ns, slug, branch
			return nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.handleSetRepoDefaultBranch(context.Background(), makeReq(map[string]any{
		"repo":   "MYPROJ/my-service",
		"branch": "main",
	}))
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-service", gotSlug)
	assert.Equal(t, "main", gotBranch)
	assert.Contains(t, extractText(t, result), "main")
}

func TestSetRepoDefaultBranch_MissingRepo_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.handleSetRepoDefaultBranch(context.Background(), makeReq(map[string]any{
		"branch": "main",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "missing required parameter")
}

func TestSetRepoDefaultBranch_MissingBranch_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.handleSetRepoDefaultBranch(context.Background(), makeReq(map[string]any{
		"repo": "MYPROJ/my-service",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "missing required parameter")
}

func TestSetRepoDefaultBranch_InvalidRepo_ReturnsError(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.handleSetRepoDefaultBranch(context.Background(), makeReq(map[string]any{
		"repo":   "not-a-valid-repo",
		"branch": "main",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "")
}
