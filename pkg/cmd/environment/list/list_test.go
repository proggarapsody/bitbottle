package list_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/environment/list"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdList_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("jq"))
}

func TestNewCmdList_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{})
	require.Error(t, cmd.Execute())
}

func TestList_PrintsEnvironments(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListEnvironmentsFn: func(ns, slug string) ([]backend.Environment, error) {
			return []backend.Environment{
				{UUID: "env-1", Name: "Production", Type: "Production", Rank: 1},
				{UUID: "env-2", Name: "Staging", Type: "Staging", Rank: 2},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "Production")
	assert.Contains(t, got, "Staging")
}

func TestList_JSON_Output(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListEnvironmentsFn: func(ns, slug string) ([]backend.Environment, error) {
			return []backend.Environment{
				{UUID: "env-1", Name: "Production", Type: "Production", Rank: 1},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--json"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), `"name":"Production"`)
}

func TestList_ClientNotDeploymentCapable_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &cmdtest.NoDeploymentFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deployments")
}
