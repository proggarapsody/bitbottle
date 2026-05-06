package steps_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/steps"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdSteps_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := steps.NewCmdSteps(f, nil)
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("jq"))
	assert.NotNil(t, cmd.Flag("hostname"))
}

func TestNewCmdSteps_RequiresTwoArgs(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := steps.NewCmdSteps(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"}) // missing UUID
	require.Error(t, cmd.Execute())
}

func TestSteps_PrintsStepNamesAndStates(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineStepsFn: func(ns, slug, uuid string) ([]backend.PipelineStep, error) {
			return []backend.PipelineStep{
				{UUID: "s1", Name: "Build", State: "SUCCESSFUL", Duration: 42},
				{UUID: "s2", Name: "Test", State: "FAILED", Duration: 17},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := steps.NewCmdSteps(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "p-uuid"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "Build")
	assert.Contains(t, got, "SUCCESSFUL")
	assert.Contains(t, got, "Test")
	assert.Contains(t, got, "FAILED")
}

func TestSteps_PassesPipelineUUIDThrough(t *testing.T) {
	t.Parallel()
	var gotUUID string
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineStepsFn: func(ns, slug, uuid string) ([]backend.PipelineStep, error) {
			gotUUID = uuid
			return nil, nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := steps.NewCmdSteps(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "{abc-123}"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "{abc-123}", gotUUID)
}

func TestSteps_JSON_EmitsArray(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineStepsFn: func(ns, slug, uuid string) ([]backend.PipelineStep, error) {
			return []backend.PipelineStep{{UUID: "s1", Name: "Build", State: "SUCCESSFUL"}}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := steps.NewCmdSteps(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "p-uuid", "--json", "name,state"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, `"name":"Build"`)
	assert.Contains(t, got, `"state":"SUCCESSFUL"`)
}

func TestSteps_APIError_PropagatesError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineStepsFn: func(ns, slug, uuid string) ([]backend.PipelineStep, error) {
			return nil, errors.New("boom")
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := steps.NewCmdSteps(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "p-uuid"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}

func TestSteps_ClientNotPipelineCapable_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &cmdtest.NoPipelineFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := steps.NewCmdSteps(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "p-uuid"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipelines")
}
