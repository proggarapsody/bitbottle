package list_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/deployment/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/deployment/list"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdList_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	assert.NotNil(t, cmd.Flag("limit"))
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("jq"))
}

func TestNewCmdList_LimitDefault(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	assert.Equal(t, "10", cmd.Flag("limit").DefValue)
}

func TestNewCmdList_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{})
	require.Error(t, cmd.Execute())
}

func TestList_PrintsDeployments(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListDeploymentsFn: func(ns, slug string, limit int) ([]backend.Deployment, error) {
			d1 := backend.Deployment{UUID: "abc-123", State: "COMPLETED"}
			d1.Environment.Name = "Production"
			d1.Release.Name = "v1.0"
			d1.Release.CommitHash = "deadbeef"
			d2 := backend.Deployment{UUID: "def-456", State: "FAILED"}
			d2.Environment.Name = "Staging"
			d2.Release.Name = "v0.9"
			d2.Release.CommitHash = "cafebabe"
			return []backend.Deployment{d1, d2}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "COMPLETED")
	assert.Contains(t, got, "FAILED")
	assert.Contains(t, got, "Production")
	assert.Contains(t, got, "Staging")
}

func TestList_JSON_Output(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListDeploymentsFn: func(ns, slug string, limit int) ([]backend.Deployment, error) {
			d := backend.Deployment{UUID: "abc-123", State: "COMPLETED"}
			d.Environment.Name = "Production"
			d.Release.Name = "v1.0"
			d.Release.CommitHash = "deadbeef"
			return []backend.Deployment{d}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--json"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, `"state":"COMPLETED"`)
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
