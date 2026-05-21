package snippet_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/snippet"
	"github.com/proggarapsody/bitbottle/pkg/cmd/snippet/create"
	"github.com/proggarapsody/bitbottle/pkg/cmd/snippet/delete"
	"github.com/proggarapsody/bitbottle/pkg/cmd/snippet/list"
	"github.com/proggarapsody/bitbottle/pkg/cmd/snippet/view"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// cloudConfig has user: alice so workspace defaults to "alice".
const cloudConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

func newFactory(t *testing.T, fake backend.Client) (*factory.Factory, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	f, out, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)
	return f, out, errOut
}

// noSnippetFake wraps backend.Client without implementing SnippetClient.
type noSnippetFake struct {
	backend.Client
}

// ---- snippet root command ----

func TestSnippet_HasSubcommands(t *testing.T) {
	t.Parallel()
	f, _, _ := newFactory(t, &testhelpers.FakeClient{T: t})
	cmd := snippet.NewCmdSnippet(f)
	names := make([]string, 0)
	for _, sub := range cmd.Commands() {
		names = append(names, sub.Name())
	}
	assert.Contains(t, names, "list")
	assert.Contains(t, names, "view")
	assert.Contains(t, names, "create")
	assert.Contains(t, names, "delete")
}

// ---- list ----

func TestList_PrintsSnippets(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListSnippetsFn: func(workspace string, limit int) ([]backend.Snippet, error) {
			assert.Equal(t, "alice", workspace, "default workspace should come from user config")
			assert.Equal(t, 30, limit)
			return []backend.Snippet{
				{ID: "abc", Title: "Hello world", Owner: "alice"},
				{ID: "def", Title: "Secret", Owner: "alice", IsPrivate: true},
			}, nil
		},
	}
	f, out, _ := newFactory(t, fake)
	cmd := list.NewCmdList(f)
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "abc")
	assert.Contains(t, got, "Hello world")
	assert.Contains(t, got, "def")
}

func TestList_ExplicitWorkspace(t *testing.T) {
	t.Parallel()
	var gotWS string
	fake := &testhelpers.FakeClient{
		T: t,
		ListSnippetsFn: func(workspace string, limit int) ([]backend.Snippet, error) {
			gotWS = workspace
			return nil, nil
		},
	}
	f, _, _ := newFactory(t, fake)
	cmd := list.NewCmdList(f)
	cmd.SetArgs([]string{"--workspace", "otherws"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "otherws", gotWS)
}

func TestList_UnsupportedBackend(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, noSnippetFake{Client: &testhelpers.FakeClient{T: t}})
	cmd := list.NewCmdList(f)
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Cloud only")
}

// ---- view ----

func TestView_PrintsSnippet(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetSnippetFn: func(workspace, id string) (backend.Snippet, error) {
			assert.Equal(t, "alice", workspace)
			assert.Equal(t, "Xqjyp1GV", id)
			return backend.Snippet{ID: "Xqjyp1GV", Title: "My snippet", Owner: "alice"}, nil
		},
	}
	f, out, _ := newFactory(t, fake)
	cmd := view.NewCmdView(f)
	cmd.SetArgs([]string{"Xqjyp1GV"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "Xqjyp1GV")
	assert.Contains(t, got, "My snippet")
}

func TestView_RequiresID(t *testing.T) {
	t.Parallel()
	f, _, _ := newFactory(t, &testhelpers.FakeClient{T: t})
	cmd := view.NewCmdView(f)
	cmd.SetArgs([]string{})
	require.Error(t, cmd.Execute())
}

// ---- create ----

func TestCreate_RequiresTitle(t *testing.T) {
	t.Parallel()
	f, _, _ := newFactory(t, &testhelpers.FakeClient{T: t})
	cmd := create.NewCmdCreate(f)
	cmd.SetArgs([]string{})
	require.Error(t, cmd.Execute(), "--title is required")
}

func TestCreate_PassesInputToBackend(t *testing.T) {
	t.Parallel()
	var gotWS string
	var gotIn backend.CreateSnippetInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreateSnippetFn: func(workspace string, in backend.CreateSnippetInput) (backend.Snippet, error) {
			gotWS = workspace
			gotIn = in
			return backend.Snippet{ID: "new1", Title: in.Title}, nil
		},
	}
	f, out, _ := newFactory(t, fake)
	cmd := create.NewCmdCreate(f)
	cmd.SetArgs([]string{"--title", "My snippet", "--private"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "alice", gotWS)
	assert.Equal(t, "My snippet", gotIn.Title)
	assert.True(t, gotIn.IsPrivate)
	assert.Contains(t, out.String(), "new1")
}

// ---- delete ----

func TestDelete_RequiresConfirmInNonTTY(t *testing.T) {
	t.Parallel()
	f, _, _ := newFactory(t, &testhelpers.FakeClient{T: t})
	// The default factory from factorytest has non-TTY stdout.
	cmd := delete.NewCmdDelete(f)
	cmd.SetArgs([]string{"abc"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestDelete_DeletesWithConfirmFlag(t *testing.T) {
	t.Parallel()
	deleted := false
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteSnippetFn: func(workspace, id string) error {
			assert.Equal(t, "alice", workspace)
			assert.Equal(t, "abc", id)
			deleted = true
			return nil
		},
	}
	f, out, _ := newFactory(t, fake)
	cmd := delete.NewCmdDelete(f)
	cmd.SetArgs([]string{"abc", "--confirm"})
	require.NoError(t, cmd.Execute())
	assert.True(t, deleted)
	assert.Contains(t, out.String(), "abc")
}
