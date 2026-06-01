package view_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	cmdView "github.com/proggarapsody/bitbottle/pkg/cmd/variable/view"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestView_RequiresAtLeastOneArg(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := cmdView.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{})
	require.Error(t, cmd.Execute())
}

func TestView_RepositoryScope_PrintsVariable(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineVariablesFn: func(ns, slug string) ([]backend.PipelineVariable, error) {
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "my-service", slug)
			return []backend.PipelineVariable{
				{UUID: "v1", Key: "DEPLOY_ENV", Value: "prod", Secured: false},
				{UUID: "v2", Key: "API_TOKEN", Secured: true},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdView.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "DEPLOY_ENV"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "DEPLOY_ENV")
	assert.Contains(t, got, "prod")
	assert.NotContains(t, got, "API_TOKEN")
}

func TestView_RepositoryScope_SecuredVariable_RedactsValue(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineVariablesFn: func(ns, slug string) ([]backend.PipelineVariable, error) {
			return []backend.PipelineVariable{
				{UUID: "v2", Key: "API_TOKEN", Value: "", Secured: true},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdView.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "API_TOKEN"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "<secured>")
}

func TestView_RepositoryScope_KeyNotFound_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineVariablesFn: func(ns, slug string) ([]backend.PipelineVariable, error) {
			return []backend.PipelineVariable{
				{UUID: "v1", Key: "OTHER_KEY", Value: "val"},
			}, nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdView.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "MISSING_KEY"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MISSING_KEY")
}

func TestView_WorkspaceScope_PrintsVariable(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspaceVariablesFn: func(ns string) ([]backend.PipelineVariable, error) {
			assert.Equal(t, "myworkspace", ns)
			return []backend.PipelineVariable{
				{UUID: "w1", Key: "WS_VAR", Value: "globalval", Secured: false},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdView.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "WS_VAR", "--scope", "workspace"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "WS_VAR")
	assert.Contains(t, got, "globalval")
}

func TestView_DeploymentScope_PrintsVariable(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListEnvVariablesFn: func(ns, slug, envUUID string) ([]backend.EnvVariable, error) {
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "my-service", slug)
			assert.Equal(t, "env-123", envUUID)
			return []backend.EnvVariable{
				{UUID: "ev1", Key: "PROD_DB_URL", Value: "postgres://...", Secured: false},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdView.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "PROD_DB_URL", "--scope", "deployment", "--env", "env-123"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "PROD_DB_URL")
	assert.Contains(t, got, "postgres://...")
}

func TestView_DeploymentScope_RequiresEnv(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := cmdView.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "MY_KEY", "--scope", "deployment"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--env ENV-UUID")
}

func TestView_UnknownScope_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := cmdView.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "MY_KEY", "--scope", "bogus"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown scope")
}

func TestView_RepositoryScope_JSON_RedactsSecuredValue(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineVariablesFn: func(ns, slug string) ([]backend.PipelineVariable, error) {
			return []backend.PipelineVariable{
				{UUID: "v1", Key: "SECRET", Value: "", Secured: true},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdView.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "SECRET", "--json"})
	require.NoError(t, cmd.Execute())
	var rows []map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &rows))
	require.Len(t, rows, 1)
	assert.Equal(t, "SECRET", rows[0]["key"])
	assert.Equal(t, "<secured>", rows[0]["value"])
	assert.Equal(t, true, rows[0]["secured"])
}

func TestView_UnsupportedHost_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &cmdtest.NoPipelineFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdView.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "MY_KEY"})
	err := cmd.Execute()
	require.Error(t, err)
}
