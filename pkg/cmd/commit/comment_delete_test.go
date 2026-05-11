package commit_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/commit"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestCommitCommentDelete_Success(t *testing.T) {
	t.Parallel()

	var gotID int
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteCommitCommentFn: func(ns, slug, hash string, commentID int) error {
			gotID = commentID
			return nil
		},
	}

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitCommentDelete(f)
	cmd.SetArgs([]string{"myworkspace/my-repo", "deadbeef", "5"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, 5, gotID)
}

func TestCommitCommentDelete_BackendError(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		DeleteCommitCommentFn: func(ns, slug, hash string, commentID int) error {
			return errors.New("delete denied")
		},
	}

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitCommentDelete(f)
	cmd.SetArgs([]string{"myworkspace/my-repo", "deadbeef", "5"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete denied")
}
