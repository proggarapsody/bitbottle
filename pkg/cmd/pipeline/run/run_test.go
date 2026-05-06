package run_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/run"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdRun_HasBranchFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := run.NewCmdRun(f, nil)
	assert.NotNil(t, cmd.Flag("branch"))
}

func TestNewCmdRun_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := run.NewCmdRun(f, nil)
	cmd.SetArgs([]string{})
	require.Error(t, cmd.Execute())
}

func TestRun_PrintsBuildNumber(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		RunPipelineFn: func(ns, slug string, in backend.RunPipelineInput) (backend.Pipeline, error) {
			return backend.Pipeline{BuildNumber: 99, State: "PENDING", RefName: in.Branch}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := run.NewCmdRun(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "--branch", "main"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "99")
}

func TestRun_PassesBranchToAPI(t *testing.T) {
	t.Parallel()
	var gotBranch string
	fake := &testhelpers.FakeClient{
		T: t,
		RunPipelineFn: func(ns, slug string, in backend.RunPipelineInput) (backend.Pipeline, error) {
			gotBranch = in.Branch
			return backend.Pipeline{BuildNumber: 1, State: "PENDING"}, nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := run.NewCmdRun(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "--branch", "feature/login"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "feature/login", gotBranch)
}

func TestRun_ClientNotPipelineCapable_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &cmdtest.NoPipelineFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := run.NewCmdRun(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "--branch", "main"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipelines")
}
