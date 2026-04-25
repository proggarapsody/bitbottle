package pipeline_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdPipelineRun_HasBranchFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := newPipelineFactory(t, &testhelpers.FakeClient{T: t}, newPipelineRunner())
	cmd := pipeline.NewCmdPipelineRun(f)
	assert.NotNil(t, cmd.Flag("branch"))
}

func TestNewCmdPipelineRun_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := newPipelineFactory(t, &testhelpers.FakeClient{T: t}, newPipelineRunner())
	cmd := pipeline.NewCmdPipelineRun(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestPipelineRun_PrintsBuildNumber(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		RunPipelineFn: func(ns, slug string, in backend.RunPipelineInput) (backend.Pipeline, error) {
			return backend.Pipeline{
				BuildNumber: 99,
				State:       "PENDING",
				RefName:     in.Branch,
			}, nil
		},
	}

	f, out, _ := newPipelineFactory(t, fake, newPipelineRunner())
	cmd := pipeline.NewCmdPipelineRun(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "--branch", "main"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "99")
}

func TestPipelineRun_PassesBranchToAPI(t *testing.T) {
	t.Parallel()

	var gotBranch string
	fake := &testhelpers.FakeClient{
		T: t,
		RunPipelineFn: func(ns, slug string, in backend.RunPipelineInput) (backend.Pipeline, error) {
			gotBranch = in.Branch
			return backend.Pipeline{BuildNumber: 1, State: "PENDING"}, nil
		},
	}

	f, _, _ := newPipelineFactory(t, fake, newPipelineRunner())
	cmd := pipeline.NewCmdPipelineRun(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "--branch", "feature/login"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "feature/login", gotBranch)
}

func TestPipelineRun_ClientNotPipelineCapable_ReturnsError(t *testing.T) {
	t.Parallel()

	fake := &noPipelineFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := newPipelineFactory(t, fake, newPipelineRunner())
	cmd := pipeline.NewCmdPipelineRun(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "--branch", "main"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipelines")
}
