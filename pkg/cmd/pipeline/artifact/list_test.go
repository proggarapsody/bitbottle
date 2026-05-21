package artifact_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/artifact"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdList_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := artifact.NewCmdList(f, nil)
	assert.NotNil(t, cmd.Flag("step"))
	assert.NotNil(t, cmd.Flag("limit"))
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("jq"))
}

func TestNewCmdList_RequiresPipelineUUID(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := artifact.NewCmdList(f, nil)
	cmd.SetArgs([]string{"--step", "step-uuid"})
	require.Error(t, cmd.Execute())
}

func TestNewCmdList_RequiresStep(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := artifact.NewCmdList(f, nil)
	cmd.SetArgs([]string{"pipe-uuid", "myws/repo"})
	require.Error(t, cmd.Execute())
}

func TestList_PrintsArtifacts(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineArtifactsFn: func(ws, slug, pipelineUUID, stepUUID string, limit int) ([]backend.PipelineArtifact, error) {
			return []backend.PipelineArtifact{
				{Name: "build.tar.gz", SizeBytes: 1048576, URL: "https://dl/build.tar.gz"},
				{Name: "test.xml", SizeBytes: 2048, URL: "https://dl/test.xml"},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := artifact.NewCmdList(f, nil)
	cmd.SetArgs([]string{"pipe-uuid", "myws/repo", "--step", "step-uuid"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "build.tar.gz")
	assert.Contains(t, got, "test.xml")
}

func TestList_JSON_Output(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineArtifactsFn: func(ws, slug, pipelineUUID, stepUUID string, limit int) ([]backend.PipelineArtifact, error) {
			return []backend.PipelineArtifact{
				{Name: "build.tar.gz", SizeBytes: 1024, URL: "https://dl/build.tar.gz"},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := artifact.NewCmdList(f, nil)
	cmd.SetArgs([]string{"pipe-uuid", "myws/repo", "--step", "step-uuid", "--json"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, `"name"`)
	assert.Contains(t, got, "build.tar.gz")
	assert.Contains(t, got, `"size_bytes"`)
}

func TestList_NotCloudCapable_ReturnsError(t *testing.T) {
	t.Parallel()
	type serverOnlyFake struct{ backend.Client }
	fake := &serverOnlyFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := artifact.NewCmdList(f, nil)
	cmd.SetArgs([]string{"pipe-uuid", "myws/repo", "--step", "step-uuid"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipeline artifacts")
}

func TestList_InvalidLimit_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := artifact.NewCmdList(f, nil)
	cmd.SetArgs([]string{"pipe-uuid", "myws/repo", "--step", "step-uuid", "--limit", "0"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--limit")
}

func TestList_PassesPipelineAndStepUUID(t *testing.T) {
	t.Parallel()
	var gotPipelineUUID, gotStepUUID string
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineArtifactsFn: func(ws, slug, pipelineUUID, stepUUID string, limit int) ([]backend.PipelineArtifact, error) {
			gotPipelineUUID = pipelineUUID
			gotStepUUID = stepUUID
			return nil, nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := artifact.NewCmdList(f, nil)
	cmd.SetArgs([]string{"my-pipe-uuid", "myws/repo", "--step", "my-step-uuid"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "my-pipe-uuid", gotPipelineUUID)
	assert.Equal(t, "my-step-uuid", gotStepUUID)
}
