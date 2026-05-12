package view_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/deployment/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/deployment/view"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdView_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := view.NewCmdView(f, nil)
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("jq"))
}

func TestNewCmdView_RequiresTwoArgs(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := view.NewCmdView(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.Error(t, cmd.Execute())
}

func TestView_PrintsDeployment(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetDeploymentFn: func(ns, slug, uuid string) (backend.Deployment, error) {
			d := backend.Deployment{UUID: "abc-123", State: "COMPLETED"}
			d.Environment = backend.Environment{Name: "Production", Type: "Production"}
			d.Release.Name = "v1.0"
			d.Release.CommitHash = "deadbeef1234567"
			return d, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := view.NewCmdView(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "abc-123"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "abc-123")
	assert.Contains(t, got, "COMPLETED")
	assert.Contains(t, got, "Production")
	assert.Contains(t, got, "deadbee") // first 7 chars of the hash
}

func TestView_JSON_Output(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetDeploymentFn: func(ns, slug, uuid string) (backend.Deployment, error) {
			d := backend.Deployment{UUID: "abc-123", State: "FAILED"}
			d.Environment = backend.Environment{Name: "Staging", Type: "Staging"}
			return d, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := view.NewCmdView(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "abc-123", "--json", "state"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), `"state":"FAILED"`)
}
