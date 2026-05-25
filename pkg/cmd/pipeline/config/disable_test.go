package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/config"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdDisable_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := config.NewCmdDisable(f, nil)
	assert.NotNil(t, cmd.Flag("hostname"))
}

func TestPipelineConfigDisable_PrintsSuccess(t *testing.T) {
	t.Parallel()
	var calledWith backend.PipelineConfig
	fake := &testhelpers.FakeClient{
		T: t,
		UpdatePipelinesConfigFn: func(ws, slug string, in backend.PipelineConfig) (backend.PipelineConfig, error) {
			assert.Equal(t, "myworkspace", ws)
			assert.Equal(t, "my-service", slug)
			calledWith = in
			return backend.PipelineConfig{Enabled: false}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := config.NewCmdDisable(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())
	assert.False(t, calledWith.Enabled)
	assert.Contains(t, out.String(), "Pipelines disabled for myworkspace/my-service.")
}

func TestPipelineConfigDisable_ClientNotCapable_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &noPipelineConfigFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := config.NewCmdDisable(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipeline config")
}
