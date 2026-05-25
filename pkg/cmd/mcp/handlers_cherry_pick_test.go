package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestCherryPickCommit_CallsClientWithCorrectParams(t *testing.T) {
	t.Parallel()
	called := false
	fake := &testhelpers.FakeClient{
		T: t,
		CherryPickCommitFn: func(ns, slug string, in backend.CherryPickInput) (backend.Commit, error) {
			assert.Equal(t, "MYPROJ", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, "abc123", in.SourceHash)
			assert.Equal(t, "main", in.TargetBranch)
			assert.Equal(t, "cherry-pick msg", in.Message)
			called = true
			return backend.Commit{Hash: "newdef456", Message: "Fix thing"}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.cherryPickCommit(context.Background(), makeReq(map[string]any{
		"repo":          "MYPROJ/my-repo",
		"commit_hash":   "abc123",
		"target_branch": "main",
		"message":       "cherry-pick msg",
	}))
	require.NoError(t, err)
	assert.True(t, called)
	assertJSONContains(t, result, "newdef456", "")
}

func TestCherryPickCommit_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.cherryPickCommit(context.Background(), makeReq(map[string]any{
		"commit_hash":   "abc123",
		"target_branch": "main",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}

func TestCherryPickCommit_MissingCommitHash(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.cherryPickCommit(context.Background(), makeReq(map[string]any{
		"repo":          "MYPROJ/my-repo",
		"target_branch": "main",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "commit_hash")
}
