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

func TestGrant_GrantsUserPermission(t *testing.T) {
	t.Parallel()
	var gotWS, gotKey string
	var gotIn backend.WorkspaceProjectPermInput
	fake := &testhelpers.FakeClient{
		T: t,
		GrantWorkspaceProjectPermFn: func(workspace, projectKey string, in backend.WorkspaceProjectPermInput) error {
			gotWS, gotKey, gotIn = workspace, projectKey, in
			return nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := perms.NewCmdGrant(f, nil)
	cmd.SetArgs([]string{"myws", "PROJ", "--user", "alice", "--permission", "write"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "myws", gotWS)
	assert.Equal(t, "PROJ", gotKey)
	assert.Equal(t, "alice", gotIn.UserSlug)
	assert.Equal(t, "write", gotIn.Permission)
	assert.Contains(t, out.String(), "Granted")
}

func TestGrant_GrantsGroupPermission(t *testing.T) {
	t.Parallel()
	var gotIn backend.WorkspaceProjectPermInput
	fake := &testhelpers.FakeClient{
		T: t,
		GrantWorkspaceProjectPermFn: func(workspace, projectKey string, in backend.WorkspaceProjectPermInput) error {
			gotIn = in
			return nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := perms.NewCmdGrant(f, nil)
	cmd.SetArgs([]string{"myws", "PROJ", "--group", "devs", "--permission", "read"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "devs", gotIn.GroupSlug)
	assert.Equal(t, "read", gotIn.Permission)
}

func TestGrant_RequiresUserOrGroup(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := perms.NewCmdGrant(f, nil)
	cmd.SetArgs([]string{"myws", "PROJ", "--permission", "read"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--user or --group")
}

func TestGrant_RejectsBothUserAndGroup(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := perms.NewCmdGrant(f, nil)
	cmd.SetArgs([]string{"myws", "PROJ", "--user", "alice", "--group", "devs", "--permission", "read"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestGrant_RejectsInvalidPermission(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := perms.NewCmdGrant(f, nil)
	cmd.SetArgs([]string{"myws", "PROJ", "--user", "alice", "--permission", "superadmin"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid permission")
}
