package ipallowlist_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/workspace/ipallowlist"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestDelete_CallsBackendWithCorrectArgs(t *testing.T) {
	t.Parallel()
	var gotWorkspace, gotUUID string
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteIPAllowlistFn: func(workspace, uuid string) error {
			gotWorkspace = workspace
			gotUUID = uuid
			return nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := ipallowlist.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"acme", "entry-uuid-1", "--confirm"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "acme", gotWorkspace)
	assert.Equal(t, "entry-uuid-1", gotUUID)
	assert.Contains(t, out.String(), "Deleted IP allowlist entry entry-uuid-1")
}

func TestDelete_UUIDOnlyArg_InfersWorkspaceFromRepo(t *testing.T) {
	t.Parallel()
	// With only one arg (UUID), workspace must come from factory.
	// factorytest without a BaseRepo override returns an error for f.BaseRepo().
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, &testhelpers.FakeClient{T: t})

	cmd := ipallowlist.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"only-a-uuid", "--confirm"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace required")
}

func TestDelete_NoArgs_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := ipallowlist.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{})
	require.Error(t, cmd.Execute())
}

func TestDelete_TooManyArgs_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := ipallowlist.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"ws", "uuid", "extra"})
	require.Error(t, cmd.Execute())
}

func TestDelete_NonTTY_RequiresConfirmFlag(t *testing.T) {
	t.Parallel()
	// factorytest always sets IsStdoutTTY=false, so omitting --confirm triggers the guard.
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, &testhelpers.FakeClient{T: t})

	cmd := ipallowlist.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"acme", "entry-uuid-1"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm required")
}
