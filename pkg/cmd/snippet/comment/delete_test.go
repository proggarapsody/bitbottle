package comment_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/snippet/comment"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestDeleteCmd_Success(t *testing.T) {
	t.Parallel()
	deleted := false
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteSnippetCommentFn: func(workspace, snippetID string, commentID int) error {
			assert.Equal(t, "testuser", workspace)
			assert.Equal(t, "abc123", snippetID)
			assert.Equal(t, 17, commentID)
			deleted = true
			return nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{
		InitialConfig: cloudHosts,
		BackendType:   "cloud",
	})
	// Disable TTY check so --confirm isn't required.
	outBuf := &bytes.Buffer{}
	ios := iostreams.Test()
	ios.Out = outBuf
	f.IOStreams = ios
	factorytest.UseBackend(f, fake)

	cmd := comment.NewCmdDelete(f)
	cmd.SetArgs([]string{"abc123", "17", "--confirm"})
	require.NoError(t, cmd.Execute())
	assert.True(t, deleted)
	assert.Contains(t, outBuf.String(), "17")
}

func TestDeleteCmd_RequiresConfirmInNonTTY(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{
		InitialConfig: cloudHosts,
		BackendType:   "cloud",
	})
	f.IOStreams = iostreams.Test()
	factorytest.UseBackend(f, &testhelpers.FakeClient{T: t})

	cmd := comment.NewCmdDelete(f)
	cmd.SetArgs([]string{"abc123", "17"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestDeleteCmd_WithExplicitWorkspace(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteSnippetCommentFn: func(workspace, snippetID string, commentID int) error {
			assert.Equal(t, "myws", workspace)
			return nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{
		InitialConfig: cloudHosts,
		BackendType:   "cloud",
	})
	f.IOStreams = iostreams.Test()
	factorytest.UseBackend(f, fake)

	cmd := comment.NewCmdDelete(f)
	cmd.SetArgs([]string{"abc123", "17", "myws", "--confirm"})
	require.NoError(t, cmd.Execute())
}

func TestDeleteCmd_InvalidCommentID(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudHosts})
	cmd := comment.NewCmdDelete(f)
	cmd.SetArgs([]string{"abc123", "notanint", "--confirm"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid comment ID")
}

func TestDeleteCmd_RunnerOverride(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudHosts})

	var called bool
	cmd := comment.NewCmdDelete(f, func(opts *comment.DeleteOptions) error {
		called = true
		assert.Equal(t, "abc123", opts.SnippetID)
		assert.Equal(t, 17, opts.CommentID)
		return nil
	})
	cmd.SetArgs([]string{"abc123", "17", "--confirm"})
	require.NoError(t, cmd.Execute())
	assert.True(t, called)
}
