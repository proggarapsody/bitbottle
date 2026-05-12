package delete_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	cmdDelete "github.com/proggarapsody/bitbottle/pkg/cmd/variable/delete"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestDelete_RequiresTwoArgs(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := cmdDelete.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.Error(t, cmd.Execute())
}

func TestDelete_NonInteractive_RequiresConfirm(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	// factorytest sets TTY to false by default, so --confirm required.
	cmd := cmdDelete.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "MY_KEY"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestDelete_RepositoryScope_DeletesVariable(t *testing.T) {
	t.Parallel()
	var deletedKey string
	fake := &testhelpers.FakeClient{
		T: t,
		DeletePipelineVariableFn: func(ns, slug, key string) error {
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "my-service", slug)
			deletedKey = key
			return nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdDelete.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "MY_KEY", "--confirm"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "MY_KEY", deletedKey)
	assert.Contains(t, out.String(), "Deleted variable MY_KEY")
}

func TestDelete_WorkspaceScope_DeletesVariable(t *testing.T) {
	t.Parallel()
	var deletedKey string
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteWorkspaceVariableFn: func(ns, key string) error {
			assert.Equal(t, "myworkspace", ns)
			deletedKey = key
			return nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdDelete.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "WS_KEY", "--scope", "workspace", "--confirm"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "WS_KEY", deletedKey)
	assert.Contains(t, out.String(), "Deleted variable WS_KEY")
}

func TestDelete_DeploymentScope_RequiresEnv(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := cmdDelete.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "MY_KEY", "--scope", "deployment", "--confirm"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--env ENV-UUID")
}

func TestDelete_DeploymentScope_DeletesVariableByKey(t *testing.T) {
	t.Parallel()
	var deletedVarUUID string
	fake := &testhelpers.FakeClient{
		T: t,
		ListEnvVariablesFn: func(ns, slug, envUUID string) ([]backend.EnvVariable, error) {
			return []backend.EnvVariable{
				{UUID: "var-uuid-1", Key: "OTHER_KEY"},
				{UUID: "var-uuid-2", Key: "PROD_KEY"},
			}, nil
		},
		DeleteEnvVariableFn: func(ns, slug, envUUID, varUUID string) error {
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "my-service", slug)
			assert.Equal(t, "env-abc", envUUID)
			deletedVarUUID = varUUID
			return nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdDelete.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "PROD_KEY", "--scope", "deployment", "--env", "env-abc", "--confirm"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "var-uuid-2", deletedVarUUID)
	assert.Contains(t, out.String(), "Deleted variable PROD_KEY")
}

func TestDelete_DeploymentScope_KeyNotFound(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListEnvVariablesFn: func(ns, slug, envUUID string) ([]backend.EnvVariable, error) {
			return []backend.EnvVariable{
				{UUID: "var-uuid-1", Key: "OTHER_KEY"},
			}, nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdDelete.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "MISSING_KEY", "--scope", "deployment", "--env", "env-abc", "--confirm"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MISSING_KEY")
	assert.Contains(t, err.Error(), "not found")
}

func TestDelete_UnknownScope_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := cmdDelete.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "KEY", "--scope", "badscope", "--confirm"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown scope")
}
