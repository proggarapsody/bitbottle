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

// commitReactorFake embeds FakeClient and implements CommitCommentReactor.
type commitReactorFake struct {
	*testhelpers.FakeClient
	ListCommitCommentReactionsFn  func(ns, slug, hash string, commentID int) ([]backend.CommentReaction, error)
	AddCommitCommentReactionFn    func(ns, slug, hash string, commentID int, emoji string) error
	RemoveCommitCommentReactionFn func(ns, slug, hash string, commentID int, emoji string) error
}

func (r *commitReactorFake) ListCommitCommentReactions(ns, slug, hash string, commentID int) ([]backend.CommentReaction, error) {
	if r.ListCommitCommentReactionsFn != nil {
		return r.ListCommitCommentReactionsFn(ns, slug, hash, commentID)
	}
	r.T.Fatalf("unexpected call to ListCommitCommentReactions")
	return nil, nil
}

func (r *commitReactorFake) AddCommitCommentReaction(ns, slug, hash string, commentID int, emoji string) error {
	if r.AddCommitCommentReactionFn != nil {
		return r.AddCommitCommentReactionFn(ns, slug, hash, commentID, emoji)
	}
	r.T.Fatalf("unexpected call to AddCommitCommentReaction")
	return nil
}

func (r *commitReactorFake) RemoveCommitCommentReaction(ns, slug, hash string, commentID int, emoji string) error {
	if r.RemoveCommitCommentReactionFn != nil {
		return r.RemoveCommitCommentReactionFn(ns, slug, hash, commentID, emoji)
	}
	r.T.Fatalf("unexpected call to RemoveCommitCommentReaction")
	return nil
}

// ── commit comment react ──────────────────────────────────────────────────────

func TestCommitCommentReact_Success(t *testing.T) {
	t.Parallel()

	var gotNS, gotSlug, gotHash, gotEmoji string
	var gotCommentID int
	fake := &commitReactorFake{
		FakeClient: &testhelpers.FakeClient{T: t},
		AddCommitCommentReactionFn: func(ns, slug, hash string, commentID int, emoji string) error {
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
	cmd := commit.NewCmdCommitCommentReact(f)
	cmd.SetArgs([]string{"myworkspace/my-repo", "deadbeef", "7", "--emoji", "thumbs_up"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "myworkspace", gotNS)
	assert.Equal(t, "my-repo", gotSlug)
	assert.Equal(t, "deadbeef", gotHash)
	assert.Equal(t, 7, gotCommentID)
	assert.Equal(t, "thumbs_up", gotEmoji)
	assert.Contains(t, out.String(), "reaction")
}

func TestCommitCommentReact_EmojiNormalised(t *testing.T) {
	t.Parallel()

	var gotEmoji string
	fake := &commitReactorFake{
		FakeClient: &testhelpers.FakeClient{T: t},
		AddCommitCommentReactionFn: func(ns, slug, hash string, commentID int, emoji string) error {
			gotEmoji = emoji
			return nil
		},
	}

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitCommentReact(f)
	cmd.SetArgs([]string{"myworkspace/my-repo", "deadbeef", "7", "--emoji", ":thumbsup:"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "thumbs_up", gotEmoji)
}

func TestCommitCommentReact_MissingEmoji(t *testing.T) {
	t.Parallel()

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, &testhelpers.FakeClient{T: t})
	cmd := commit.NewCmdCommitCommentReact(f)
	cmd.SetArgs([]string{"myworkspace/my-repo", "deadbeef", "7"})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestCommitCommentReact_InvalidCommentID(t *testing.T) {
	t.Parallel()

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, &testhelpers.FakeClient{T: t})
	cmd := commit.NewCmdCommitCommentReact(f)
	cmd.SetArgs([]string{"myworkspace/my-repo", "deadbeef", "notanid", "--emoji", "thumbs_up"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "COMMENT_ID")
}

func TestCommitCommentReact_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()

	// Plain FakeClient does NOT implement CommitCommentReactor.
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, &testhelpers.FakeClient{T: t})
	cmd := commit.NewCmdCommitCommentReact(f)
	cmd.SetArgs([]string{"myworkspace/my-repo", "deadbeef", "7", "--emoji", "thumbs_up"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "commit comment reactions are not supported")
}

func TestCommitCommentReact_BackendError(t *testing.T) {
	t.Parallel()

	fake := &commitReactorFake{
		FakeClient: &testhelpers.FakeClient{T: t},
		AddCommitCommentReactionFn: func(ns, slug, hash string, commentID int, emoji string) error {
			return errors.New("reaction failed")
		},
	}

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: commitConfig})
	factorytest.UseBackend(f, fake)
	cmd := commit.NewCmdCommitCommentReact(f)
	cmd.SetArgs([]string{"myworkspace/my-repo", "deadbeef", "7", "--emoji", "thumbs_up"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reaction failed")
}
