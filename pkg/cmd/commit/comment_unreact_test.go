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

// ── commit comment unreact ────────────────────────────────────────────────────

func TestCommitCommentUnreact_Success(t *testing.T) {
	t.Parallel()

	var gotNS, gotSlug, gotHash, gotEmoji string
	var gotCommentID int
	fake := &commitReactorFake{
		FakeClient: &testhelpers.FakeClient{T: t},
		RemoveCommitCommentReactionFn: func(ns, slug, hash string, commentID int, emoji string) error {
			gotNS = ns
			gotSlug = slug
			gotHash = hash
			gotCommentID = commentID
			gotEmoji = emoji
			return nil
		},
	}

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitCommentUnreact(f)
	cmd.SetArgs([]string{"myworkspace/my-repo", "deadbeef", "7", "--emoji", "heart"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "myworkspace", gotNS)
	assert.Equal(t, "my-repo", gotSlug)
	assert.Equal(t, "deadbeef", gotHash)
	assert.Equal(t, 7, gotCommentID)
	assert.Equal(t, "heart", gotEmoji)
	assert.Contains(t, out.String(), "reaction")
}

func TestCommitCommentUnreact_EmojiNormalised(t *testing.T) {
	t.Parallel()

	var gotEmoji string
	fake := &commitReactorFake{
		FakeClient: &testhelpers.FakeClient{T: t},
		RemoveCommitCommentReactionFn: func(ns, slug, hash string, commentID int, emoji string) error {
			gotEmoji = emoji
			return nil
		},
	}

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitCommentUnreact(f)
	cmd.SetArgs([]string{"myworkspace/my-repo", "deadbeef", "7", "--emoji", ":heart:"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "heart", gotEmoji)
}

func TestCommitCommentUnreact_MissingEmoji(t *testing.T) {
	t.Parallel()

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, &testhelpers.FakeClient{T: t})
	cmd := commit.NewCmdCommitCommentUnreact(f)
	cmd.SetArgs([]string{"myworkspace/my-repo", "deadbeef", "7"})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestCommitCommentUnreact_InvalidCommentID(t *testing.T) {
	t.Parallel()

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, &testhelpers.FakeClient{T: t})
	cmd := commit.NewCmdCommitCommentUnreact(f)
	cmd.SetArgs([]string{"myworkspace/my-repo", "deadbeef", "bad", "--emoji", "heart"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "COMMENT_ID")
}

func TestCommitCommentUnreact_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, &testhelpers.FakeClient{T: t})
	cmd := commit.NewCmdCommitCommentUnreact(f)
	cmd.SetArgs([]string{"myworkspace/my-repo", "deadbeef", "7", "--emoji", "heart"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "commit comment reactions are not supported")
}

func TestCommitCommentUnreact_BackendError(t *testing.T) {
	t.Parallel()

	fake := &commitReactorFake{
		FakeClient: &testhelpers.FakeClient{T: t},
		RemoveCommitCommentReactionFn: func(ns, slug, hash string, commentID int, emoji string) error {
			return errors.New("remove failed")
		},
	}

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitCommentUnreact(f)
	cmd.SetArgs([]string{"myworkspace/my-repo", "deadbeef", "7", "--emoji", "heart"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "remove failed")
}
