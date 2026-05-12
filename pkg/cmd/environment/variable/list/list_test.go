package list_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/environment/variable/list"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdList_RequiresTwoArgs(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.Error(t, cmd.Execute())
}

func TestList_PrintsVariables(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListEnvVariablesFn: func(ns, slug, envUUID string) ([]backend.EnvVariable, error) {
			return []backend.EnvVariable{
				{UUID: "var-1", Key: "API_KEY", Value: "secret", Secured: false},
				{UUID: "var-2", Key: "DB_PASS", Value: "", Secured: true},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "env-uuid"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "API_KEY")
	assert.Contains(t, got, "DB_PASS")
	assert.Contains(t, got, "<secured>")
}

func TestList_JSON_Output(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListEnvVariablesFn: func(ns, slug, envUUID string) ([]backend.EnvVariable, error) {
			return []backend.EnvVariable{
				{UUID: "var-1", Key: "MY_VAR", Value: "hello", Secured: false},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "env-uuid", "--json"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), `"key":"MY_VAR"`)
}
