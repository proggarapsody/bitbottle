package comment_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/snippet/comment"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const cloudHosts = "bitbucket.org:\n  oauth_token: tok\n  user: testuser\n"

func TestListCmd_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListSnippetCommentsFn: func(workspace, snippetID string, limit int) ([]backend.SnippetComment, error) {
			assert.Equal(t, "testuser", workspace)
			assert.Equal(t, "abc123", snippetID)
			return []backend.SnippetComment{
				{ID: 17, Author: "Alice", Body: "Nice snippet!", CreatedOn: "2024-01-01T12:00:00Z"},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{
		InitialConfig: cloudHosts,
		BackendType:   "cloud",
	})
	factorytest.UseBackend(f, fake)

	var called bool
	cmd := comment.NewCmdList(f, func(opts *comment.ListOptions) error {
		called = true
		assert.Equal(t, "abc123", opts.SnippetID)
		assert.Equal(t, 50, opts.Limit)
		return nil
	})
	cmd.SetArgs([]string{"abc123"})
	require.NoError(t, cmd.Execute())
	assert.True(t, called)
	_ = out
}

func TestListCmd_JSONFlag(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListSnippetCommentsFn: func(workspace, snippetID string, limit int) ([]backend.SnippetComment, error) {
			return []backend.SnippetComment{
				{ID: 17, Author: "Alice", Body: "Nice snippet!"},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{
		InitialConfig: cloudHosts,
		BackendType:   "cloud",
	})
	factorytest.UseBackend(f, fake)

	cmd := comment.NewCmdList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"abc123", "--json", "id,author,body"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "17")
	assert.Contains(t, out.String(), "Alice")
}

func TestListCmd_WithExplicitWorkspace(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListSnippetCommentsFn: func(workspace, snippetID string, limit int) ([]backend.SnippetComment, error) {
			assert.Equal(t, "myws", workspace)
			return nil, nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{
		InitialConfig: cloudHosts,
		BackendType:   "cloud",
	})
	factorytest.UseBackend(f, fake)

	cmd := comment.NewCmdList(f)
	cmd.SetArgs([]string{"abc123", "myws"})
	require.NoError(t, cmd.Execute())
}

func TestListCmd_LimitFlag(t *testing.T) {
	t.Parallel()
	var gotLimit int
	fake := &testhelpers.FakeClient{
		T: t,
		ListSnippetCommentsFn: func(workspace, snippetID string, limit int) ([]backend.SnippetComment, error) {
			gotLimit = limit
			return nil, nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{
		InitialConfig: cloudHosts,
		BackendType:   "cloud",
	})
	factorytest.UseBackend(f, fake)

	cmd := comment.NewCmdList(f)
	cmd.SetArgs([]string{"abc123", "--limit", "10"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, 10, gotLimit)
}
