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

func TestCommitCommentEdit_Success(t *testing.T) {
	t.Parallel()

	var gotID int
	fake := &testhelpers.FakeClient{
		T: t,
		EditCommitCommentFn: func(ns, slug, hash string, commentID int, body string) (backend.CommitComment, error) {
			gotID = commentID
			return backend.CommitComment{ID: commentID, Body: body}, nil
		},
	}

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitCommentEdit(f)
	cmd.SetArgs([]string{"myworkspace/my-repo", "deadbeef", "7", "--body", "updated text"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, 7, gotID)
}

func TestCommitCommentEdit_MissingBody(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{T: t}

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitCommentEdit(f)
	cmd.SetArgs([]string{"myworkspace/my-repo", "deadbeef", "7"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}

func TestCommitCommentEdit_BackendError(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		EditCommitCommentFn: func(ns, slug, hash string, commentID int, body string) (backend.CommitComment, error) {
			return backend.CommitComment{}, errors.New("edit failed")
		},
	}

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitCommentEdit(f)
	cmd.SetArgs([]string{"myworkspace/my-repo", "deadbeef", "7", "--body", "new body"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "edit failed")
}
