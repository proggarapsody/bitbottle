package list_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	cmdList "github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/variable/list"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdList_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := cmdList.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{})
	require.Error(t, cmd.Execute())
}

func TestList_PrintsKeysAndRedactsSecuredValuesOnTTY(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineVariablesFn: func(ns, slug string) ([]backend.PipelineVariable, error) {
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
	assert.Contains(t, got, "<secured>", "secured variable value must be redacted on TTY")
}

func TestList_JSON_RedactsSecuredValuesViaSameChokepoint(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineVariablesFn: func(ns, slug string) ([]backend.PipelineVariable, error) {
			return []backend.PipelineVariable{
				{Key: "API_TOKEN", Secured: true},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdList.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--json"})
	require.NoError(t, cmd.Execute())
	// Parse the JSON output to assert on semantic content rather than the
	// stdlib's HTML-escaped wire form (`<`/`>`).
	var rows []map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &rows))
	require.Len(t, rows, 1)
	assert.Equal(t, "API_TOKEN", rows[0]["key"])
	assert.Equal(t, true, rows[0]["secured"])
	assert.Equal(t, "<secured>", rows[0]["value"], "JSON path also routes through the redaction chokepoint")
}

func TestList_ClientNotPipelineCapable_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &cmdtest.NoPipelineFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdList.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipelines")
}
