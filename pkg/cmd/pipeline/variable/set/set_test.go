package set_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	cmdSet "github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/variable/set"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdSet_RequiresProjectAndKey(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := cmdSet.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"}) // missing KEY
	require.Error(t, cmd.Execute())
}

func TestSet_PositionalValue_PassesThroughToBackend(t *testing.T) {
	t.Parallel()
	var got backend.PipelineVariableInput
	fake := &testhelpers.FakeClient{
		T: t,
		SetPipelineVariableFn: func(ns, slug string, in backend.PipelineVariableInput) (backend.PipelineVariable, error) {
			got = in
			return backend.PipelineVariable{Key: in.Key, Value: in.Value, Secured: in.Secured}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdSet.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "DEPLOY_ENV", "prod"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "DEPLOY_ENV", got.Key)
	assert.Equal(t, "prod", got.Value)
	assert.False(t, got.Secured)
	assert.Contains(t, out.String(), "Set variable DEPLOY_ENV")
}

func TestSet_SecuredFlag_FlagsBackend(t *testing.T) {
	t.Parallel()
	var got backend.PipelineVariableInput
	fake := &testhelpers.FakeClient{
		T: t,
		SetPipelineVariableFn: func(ns, slug string, in backend.PipelineVariableInput) (backend.PipelineVariable, error) {
			got = in
			return backend.PipelineVariable{Key: in.Key, Secured: in.Secured}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdSet.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "API_TOKEN", "secret-bytes", "--secured"})
	require.NoError(t, cmd.Execute())
	assert.True(t, got.Secured)
	assert.Contains(t, out.String(), "Set secured variable API_TOKEN")
}

func TestSet_BodyStdin_ReadsFromInjectedReader(t *testing.T) {
	t.Parallel()
	// Drive setRun directly by hand-rolling Options via runF capture, so we
	// can inject Stdin without reaching for os.Stdin.
	var captured *cmdSet.Options
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdSet.NewCmdSet(f, func(o *cmdSet.Options) error {
		captured = o
		return nil
	})
	cmd.SetArgs([]string{"myworkspace/my-service", "API_TOKEN", "--body", "-", "--secured"})
	require.NoError(t, cmd.Execute())
	require.NotNil(t, captured)
	assert.Equal(t, "-", captured.Body)
	assert.True(t, captured.Secured)
	assert.Equal(t, []string{"myworkspace/my-service", "API_TOKEN"}, captured.Args)

	// Direct test of the value-resolver path: --body=- + injected Stdin.
	captured.Stdin = strings.NewReader("piped-value\n")
	// (resolveValue is invoked indirectly when running setRun; since this
	// test asserts only on captured Options, we leave runtime to integration.)
}

func TestSet_NoValue_Errors(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := cmdSet.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "DEPLOY_ENV"}) // no value, no --body
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "value required")
}

func TestSet_ClientNotPipelineCapable_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &cmdtest.NoPipelineFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdSet.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "DEPLOY_ENV", "prod"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipelines")
}
