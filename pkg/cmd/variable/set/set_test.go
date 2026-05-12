package set_test

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	cmdSet "github.com/proggarapsody/bitbottle/pkg/cmd/variable/set"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestSet_RequiresKeyArg(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := cmdSet.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"}) // missing KEY
	require.Error(t, cmd.Execute())
}

func TestSet_RepositoryScope_SetsVariable(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		SetPipelineVariableFn: func(ns, slug string, in backend.PipelineVariableInput) (backend.PipelineVariable, error) {
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "my-service", slug)
			assert.Equal(t, "MY_KEY", in.Key)
			assert.Equal(t, "myvalue", in.Value)
			assert.False(t, in.Secured)
			return backend.PipelineVariable{Key: in.Key, Value: in.Value, Secured: false}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdSet.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "MY_KEY", "myvalue"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Set variable MY_KEY")
}

func TestSet_RepositoryScope_SecuredVariable(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		SetPipelineVariableFn: func(ns, slug string, in backend.PipelineVariableInput) (backend.PipelineVariable, error) {
			assert.True(t, in.Secured)
			return backend.PipelineVariable{Key: in.Key, Secured: true}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdSet.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "SECRET", "val", "--secured"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Set secured variable SECRET")
}

func TestSet_WorkspaceScope_SetsVariable(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		SetWorkspaceVariableFn: func(ns string, in backend.PipelineVariableInput) (backend.PipelineVariable, error) {
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "WS_KEY", in.Key)
			return backend.PipelineVariable{Key: in.Key, Value: in.Value}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdSet.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "WS_KEY", "wsval", "--scope", "workspace"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Set variable WS_KEY")
}

func TestSet_DeploymentScope_RequiresEnv(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := cmdSet.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "MY_KEY", "val", "--scope", "deployment"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--env ENV-UUID")
}

func TestSet_DeploymentScope_SetsVariable(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		SetEnvVariableFn: func(ns, slug, envUUID string, in backend.EnvVariableInput) (backend.EnvVariable, error) {
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "my-service", slug)
			assert.Equal(t, "env-abc", envUUID)
			assert.Equal(t, "PROD_KEY", in.Key)
			return backend.EnvVariable{Key: in.Key, Value: in.Value, Secured: false}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdSet.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "PROD_KEY", "prodval", "--scope", "deployment", "--env", "env-abc"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Set variable PROD_KEY")
}

func TestSet_ValueFromStdin(t *testing.T) {
	t.Parallel()
	var capturedValue string
	fake := &testhelpers.FakeClient{
		T: t,
		SetPipelineVariableFn: func(ns, slug string, in backend.PipelineVariableInput) (backend.PipelineVariable, error) {
			capturedValue = in.Value
			return backend.PipelineVariable{Key: in.Key, Value: in.Value}, nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	// Inject stdin via IOStreams.In — setRun falls back to it when opts.Stdin is nil.
	f.IOStreams.In = io.NopCloser(strings.NewReader("stdinval\n"))
	cmd := cmdSet.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "MY_KEY", "--body", "-"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "stdinval", capturedValue)
}

func TestSet_UnknownScope_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := cmdSet.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "KEY", "val", "--scope", "badscope"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown scope")
}
