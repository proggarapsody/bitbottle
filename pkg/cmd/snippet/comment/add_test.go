package comment_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/snippet/comment"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestAddCmd_Success(t *testing.T) {
	t.Parallel()
	var called bool
	fake := &testhelpers.FakeClient{
		T: t,
		AddSnippetCommentFn: func(workspace, snippetID, body string) (backend.SnippetComment, error) {
			assert.Equal(t, "testuser", workspace)
			assert.Equal(t, "abc123", snippetID)
			assert.Equal(t, "Great work!", body)
			called = true
			return backend.SnippetComment{ID: 42, Body: body}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{
		InitialConfig: cloudHosts,
		BackendType:   "cloud",
	})
	factorytest.UseBackend(f, fake)

	cmd := comment.NewCmdAdd(f)
	cmd.SetArgs([]string{"abc123", "--body", "Great work!"})
	require.NoError(t, cmd.Execute())
	assert.True(t, called)
	assert.Contains(t, out.String(), "42")
	assert.Contains(t, out.String(), "abc123")
}

func TestAddCmd_WithExplicitWorkspace(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		AddSnippetCommentFn: func(workspace, snippetID, body string) (backend.SnippetComment, error) {
			assert.Equal(t, "myws", workspace)
			return backend.SnippetComment{ID: 7}, nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{
		InitialConfig: cloudHosts,
		BackendType:   "cloud",
	})
	factorytest.UseBackend(f, fake)

	cmd := comment.NewCmdAdd(f)
	cmd.SetArgs([]string{"abc123", "myws", "--body", "hi"})
	require.NoError(t, cmd.Execute())
}

func TestAddCmd_RunnerOverride(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudHosts})

	var called bool
	cmd := comment.NewCmdAdd(f, func(opts *comment.AddOptions) error {
		called = true
		assert.Equal(t, "abc123", opts.SnippetID)
		assert.Equal(t, "hello", opts.Body)
		return nil
	})
	cmd.SetArgs([]string{"abc123", "--body", "hello"})
	require.NoError(t, cmd.Execute())
	assert.True(t, called)
}
