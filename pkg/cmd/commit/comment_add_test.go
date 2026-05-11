package commit_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/commit"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestCommitCommentAdd_Success(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		AddCommitCommentFn: func(ns, slug, hash string, in backend.AddCommitCommentInput) (backend.CommitComment, error) {
			return backend.CommitComment{ID: 42, Author: backend.User{Slug: "alice"}, Body: in.Body}, nil
		},
	}

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitCommentAdd(f)
	cmd.SetArgs([]string{"myworkspace/my-repo", "deadbeef", "--body", "great work"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "42")
}

func TestCommitCommentAdd_MissingBody(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{T: t}

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitCommentAdd(f)
	cmd.SetArgs([]string{"myworkspace/my-repo", "deadbeef"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}

func TestCommitCommentAdd_BackendError(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		AddCommitCommentFn: func(ns, slug, hash string, in backend.AddCommitCommentInput) (backend.CommitComment, error) {
			return backend.CommitComment{}, errors.New("server rejected comment")
		},
	}

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitCommentAdd(f)
	cmd.SetArgs([]string{"myworkspace/my-repo", "deadbeef", "--body", "hi"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "server rejected comment")
}
