package hook_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/workspace/hook"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestDelete_CallsBackendWithCorrectArgs(t *testing.T) {
	t.Parallel()
	var gotWorkspace, gotUUID string
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteWorkspaceWebhookFn: func(workspace, uuid string) error {
			gotWorkspace = workspace
			gotUUID = uuid
			return nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := hook.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"acme", "some-uuid-1"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "acme", gotWorkspace)
	assert.Equal(t, "some-uuid-1", gotUUID)
	assert.Contains(t, out.String(), "Deleted workspace webhook some-uuid-1")
}

func TestDelete_UUIDOnlyArg_InfersWorkspaceFromRepo(t *testing.T) {
	t.Parallel()
	// With only one arg (the UUID), the workspace must come from the factory.
	// Since factorytest without a BaseRepo override returns an error for f.BaseRepo(),
	// we expect the "workspace required" error in this test setup.
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, &testhelpers.FakeClient{T: t})

	cmd := hook.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"only-a-uuid"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace required")
}

func TestDelete_NoArgs_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := hook.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{})
	require.Error(t, cmd.Execute())
}

func TestDelete_TooManyArgs_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := hook.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"ws", "uuid", "extra"})
	require.Error(t, cmd.Execute())
}
