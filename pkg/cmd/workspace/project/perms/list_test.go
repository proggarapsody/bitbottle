package perms_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/workspace/project/perms"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const cloudConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

func TestNewCmdList_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := perms.NewCmdList(f, nil)
	assert.NotNil(t, cmd.Flag("hostname"))
	assert.NotNil(t, cmd.Flag("json")) // registered via format.RegisterOutputFlags inside the command
}

func TestNewCmdList_RequiresTwoArgs(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := perms.NewCmdList(f, nil)
	cmd.SetArgs([]string{"myws"})
	require.Error(t, cmd.Execute())
}

func TestList_PrintsPermissions(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspaceProjectPermsFn: func(workspace, projectKey string) ([]backend.WorkspaceProjectPerm, error) {
			assert.Equal(t, "myws", workspace)
			assert.Equal(t, "PROJ", projectKey)
			alice := &backend.User{Slug: "alice", DisplayName: "Alice"}
			return []backend.WorkspaceProjectPerm{
				{Permission: "write", User: alice},
				{Permission: "read", Group: &backend.WorkspaceGroup{Slug: "devs", Name: "Developers"}},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := perms.NewCmdList(f, nil)
	cmd.SetArgs([]string{"myws", "PROJ"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "alice")
	assert.Contains(t, got, "write")
	assert.Contains(t, got, "devs")
	assert.Contains(t, got, "read")
}

func TestList_JSONOutput(t *testing.T) {
	t.Parallel()
	alice := &backend.User{Slug: "alice", DisplayName: "Alice"}
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspaceProjectPermsFn: func(workspace, projectKey string) ([]backend.WorkspaceProjectPerm, error) {
			return []backend.WorkspaceProjectPerm{
				{Permission: "admin", User: alice},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := perms.NewCmdList(f, nil)
	cmd.SetArgs([]string{"myws", "PROJ", "--json"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "alice")
	assert.Contains(t, got, "admin")
}

// noProjectPermsFake wraps backend.Client without satisfying WorkspaceProjectPermsClient.
type noProjectPermsFake struct {
	backend.Client
}

func TestList_ServerBackend_ReturnsUnsupportedError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, noProjectPermsFake{Client: &testhelpers.FakeClient{T: t}})

	cmd := perms.NewCmdList(f, nil)
	cmd.SetArgs([]string{"myws", "PROJ"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Cloud only")
}
