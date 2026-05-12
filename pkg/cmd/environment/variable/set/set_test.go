package set_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/environment/variable/set"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdSet_RequiresFourArgs(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := set.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "env-uuid", "KEY"})
	require.Error(t, cmd.Execute())
}

func TestSet_CreatesVariable(t *testing.T) {
	t.Parallel()
	var captured backend.EnvVariableInput
	fake := &testhelpers.FakeClient{
		T: t,
		SetEnvVariableFn: func(ns, slug, envUUID string, in backend.EnvVariableInput) (backend.EnvVariable, error) {
			captured = in
			return backend.EnvVariable{UUID: "new-var-uuid", Key: in.Key, Value: in.Value, Secured: in.Secured}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := set.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "env-uuid", "MY_KEY", "my-value"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "MY_KEY", captured.Key)
	assert.Equal(t, "my-value", captured.Value)
	assert.False(t, captured.Secured)
	assert.Contains(t, out.String(), "MY_KEY")
}

func TestSet_SecuredFlag(t *testing.T) {
	t.Parallel()
	var captured backend.EnvVariableInput
	fake := &testhelpers.FakeClient{
		T: t,
		SetEnvVariableFn: func(ns, slug, envUUID string, in backend.EnvVariableInput) (backend.EnvVariable, error) {
			captured = in
			return backend.EnvVariable{UUID: "var-uuid", Key: in.Key, Secured: in.Secured}, nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := set.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "env-uuid", "SECRET", "s3cr3t", "--secured"})
	require.NoError(t, cmd.Execute())
	assert.True(t, captured.Secured)
}
