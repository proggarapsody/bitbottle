package list_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/list"
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
	assert.Equal(t, "20", cmd.Flag("limit").DefValue)
}

func TestNewCmdList_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{})
	require.Error(t, cmd.Execute())
}

func TestList_PrintsBuildNumbers(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelinesFn: func(ns, slug string, limit int) ([]backend.Pipeline, error) {
			return []backend.Pipeline{
				{BuildNumber: 42, State: "SUCCESSFUL", RefName: "main"},
				{BuildNumber: 43, State: "IN_PROGRESS", RefName: "feature/x"},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "42")
	assert.Contains(t, got, "43")
}

func TestList_PrintsState(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelinesFn: func(ns, slug string, limit int) ([]backend.Pipeline, error) {
			return []backend.Pipeline{{BuildNumber: 1, State: "SUCCESSFUL", RefName: "main"}}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "SUCCESSFUL")
}

func TestList_PrintsRefName(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelinesFn: func(ns, slug string, limit int) ([]backend.Pipeline, error) {
			return []backend.Pipeline{{BuildNumber: 5, State: "SUCCESSFUL", RefName: "main"}}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "main")
}

func TestList_JSON_FieldsOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelinesFn: func(ns, slug string, limit int) ([]backend.Pipeline, error) {
			return []backend.Pipeline{{BuildNumber: 42, State: "SUCCESSFUL", RefName: "main"}}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--json"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, `"buildNumber":42`)
	assert.Contains(t, got, `"state":"SUCCESSFUL"`)
}

func TestList_JQ_FilterOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelinesFn: func(ns, slug string, limit int) ([]backend.Pipeline, error) {
			return []backend.Pipeline{
				{BuildNumber: 10, State: "SUCCESSFUL"},
				{BuildNumber: 20, State: "FAILED"},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--json", "--jq", ".[] | .buildNumber"})
	require.NoError(t, cmd.Execute())
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	assert.Equal(t, []string{"10", "20"}, lines)
}

func TestList_ClientNotPipelineCapable_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &cmdtest.NoPipelineFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipelines")
}
