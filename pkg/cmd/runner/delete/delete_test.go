package delete_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	cmddelete "github.com/proggarapsody/bitbottle/pkg/cmd/runner/delete"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const cloudConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

func TestNewCmdDelete_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := cmddelete.NewCmdDelete(f, nil)
	assert.NotNil(t, cmd.Flag("hostname"))
}

func TestRunnerDelete_DeletesRunner(t *testing.T) {
	t.Parallel()
	deleted := false
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteRunnerFn: func(workspace, runnerUUID string) error {
			assert.Equal(t, "myworkspace", workspace)
			assert.Equal(t, "runner-abc", runnerUUID)
			deleted = true
			return nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := cmddelete.NewCmdDelete(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace", "runner-abc"})
	require.NoError(t, cmd.Execute())
	assert.True(t, deleted)
	assert.Contains(t, out.String(), "runner-abc")
}

func TestRunnerDelete_WithUUIDOnly(t *testing.T) {
	t.Parallel()
	deleted := false
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteRunnerFn: func(workspace, runnerUUID string) error {
			assert.Equal(t, "runner-xyz", runnerUUID)
			deleted = true
			return nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := cmddelete.NewCmdDelete(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myws/my-repo", "runner-xyz"})
	// workspace arg is the first positional, UUID is second
	// but since we only have 2 args, workspace is taken as first
	require.NoError(t, cmd.Execute())
	assert.True(t, deleted)
}

func TestRunnerDelete_NoWorkspace_ReturnsError(t *testing.T) {
	t.Parallel()
	// Only one arg with no slash - treated as UUID, workspace fallback fails
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, &testhelpers.FakeClient{T: t})

	cmd := cmddelete.NewCmdDelete(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"runner-abc"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace required")
}

func TestRunnerDelete_ClientNotCapable_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &noRunnerFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := cmddelete.NewCmdDelete(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace", "runner-abc"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipeline runners")
}

// noRunnerFake wraps backend.Client without implementing backend.RunnerClient.
type noRunnerFake struct {
	backend.Client
}
