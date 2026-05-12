package create_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/environment/create"
	"github.com/proggarapsody/bitbottle/pkg/cmd/environment/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdCreate_RequiresArgs(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := create.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{})
	require.Error(t, cmd.Execute())
}

func TestNewCmdCreate_RequiresName(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := create.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "--type", "Test"})
	require.Error(t, cmd.Execute())
}

func TestNewCmdCreate_RequiresType(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := create.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "--name", "test-env"})
	require.Error(t, cmd.Execute())
}

func TestNewCmdCreate_InvalidType(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := create.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "--name", "test-env", "--type", "Invalid"})
	require.Error(t, cmd.Execute())
}

func TestCreate_PrintsResult(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		CreateEnvironmentFn: func(ns, slug string, in backend.CreateEnvironmentInput) (backend.Environment, error) {
			return backend.Environment{UUID: "new-env-uuid", Name: in.Name, Type: in.Type, Rank: in.Rank}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := create.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "--name", "QA", "--type", "Test"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "QA")
	assert.Contains(t, got, "new-env-uuid")
}
