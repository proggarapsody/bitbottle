package list_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/variable/internal/cmdtest"
	cmdList "github.com/proggarapsody/bitbottle/pkg/cmd/variable/list"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestList_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := cmdList.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{})
	require.Error(t, cmd.Execute())
}

func TestList_RepositoryScope_PrintsVariables(t *testing.T) {
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
	cmd := cmdList.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "DEPLOY_ENV")
	assert.Contains(t, got, "prod")
	assert.Contains(t, got, "API_TOKEN")
	assert.Contains(t, got, "<secured>")
}

func TestList_WorkspaceScope_PrintsVariables(t *testing.T) {
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
	cmd := cmdList.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--scope", "workspace"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "WS_VAR")
	assert.Contains(t, got, "globalval")
}

func TestList_DeploymentScope_RequiresEnv(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := cmdList.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--scope", "deployment"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--env ENV-UUID")
}

func TestList_DeploymentScope_PrintsEnvVariables(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListEnvVariablesFn: func(ns, slug, envUUID string) ([]backend.EnvVariable, error) {
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "my-service", slug)
			assert.Equal(t, "env-123", envUUID)
			return []backend.EnvVariable{
				{UUID: "ev1", Key: "PROD_DB_URL", Value: "postgres://...", Secured: false},
				{UUID: "ev2", Key: "SECRET_KEY", Secured: true},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdList.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--scope", "deployment", "--env", "env-123"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "PROD_DB_URL")
	assert.Contains(t, got, "postgres://...")
	assert.Contains(t, got, "SECRET_KEY")
	assert.Contains(t, got, "<secured>")
}

func TestList_UnknownScope_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := cmdList.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--scope", "invalid"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown scope")
}

func TestList_WorkspaceScope_UnsupportedHost_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &cmdtest.NoWorkspaceVarFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdList.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--scope", "workspace"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace variables")
}

func TestList_RepositoryScope_JSON_RedactsSecuredValues(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineVariablesFn: func(ns, slug string) ([]backend.PipelineVariable, error) {
			return []backend.PipelineVariable{
				{Key: "SECRET", Secured: true},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdList.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--json"})
	require.NoError(t, cmd.Execute())
	var rows []map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &rows))
	require.Len(t, rows, 1)
	assert.Equal(t, "SECRET", rows[0]["key"])
	assert.Equal(t, "<secured>", rows[0]["value"])
	assert.Equal(t, true, rows[0]["secured"])
}
