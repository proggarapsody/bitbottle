package list_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/project/list"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const cloudConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

func TestNewCmdList_RequiresWorkspaceArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := list.NewCmdList(f, nil)
	cmd.SetArgs([]string{})
	require.Error(t, cmd.Execute(), "missing WORKSPACE arg must error")
}

func TestNewCmdList_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := list.NewCmdList(f, nil)
	assert.NotNil(t, cmd.Flag("limit"))
	assert.NotNil(t, cmd.Flag("hostname"))
	// --json / --jq are persistent flags on the root in OUT2; they
	// resolve at execution time via format.ConfigFromCmd rather than as
	// local subcommand flags.
}

func TestList_ForwardsWorkspaceArg(t *testing.T) {
	t.Parallel()
	var gotWorkspace string
	fake := &testhelpers.FakeClient{
		T: t,
		ListProjectsFn: func(workspace string, limit int) ([]backend.Project, error) {
			gotWorkspace = workspace
			return nil, nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := list.NewCmdList(f, nil)
	cmd.SetArgs([]string{"acme"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "acme", gotWorkspace)
}

func TestList_PrintsKeyAndName(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListProjectsFn: func(workspace string, limit int) ([]backend.Project, error) {
			return []backend.Project{
				{Key: "ALPHA", Name: "Project Alpha", UUID: "u-1"},
				{Key: "BETA", Name: "Project Beta", UUID: "u-2"},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := list.NewCmdList(f, nil)
	cmd.SetArgs([]string{"acme"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "ALPHA")
	assert.Contains(t, got, "Project Alpha")
	assert.Contains(t, got, "BETA")
}

type noWorkspaceFake struct {
	backend.Client
}

func TestList_ServerBackend_ReturnsUnsupportedError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, noWorkspaceFake{Client: &testhelpers.FakeClient{T: t}})

	cmd := list.NewCmdList(f, nil)
	cmd.SetArgs([]string{"acme"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Cloud only")
}
