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

func TestNewCmdGet_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := config.NewCmdGet(f, nil)
	assert.NotNil(t, cmd.Flag("hostname"))
}

func TestPipelineConfigGet_PrintsEnabled(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelinesConfigFn: func(ws, slug string) (backend.PipelineConfig, error) {
			assert.Equal(t, "myworkspace", ws)
			assert.Equal(t, "my-service", slug)
			return backend.PipelineConfig{Enabled: true}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := config.NewCmdGet(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Pipelines enabled: true")
}

func TestPipelineConfigGet_PrintsDisabled(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelinesConfigFn: func(ws, slug string) (backend.PipelineConfig, error) {
			return backend.PipelineConfig{Enabled: false}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := config.NewCmdGet(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Pipelines enabled: false")
}

func TestPipelineConfigGet_ClientNotCapable_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &noPipelineConfigFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := config.NewCmdGet(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipeline config")
}

func TestPipelineConfigGet_JSON(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelinesConfigFn: func(ws, slug string) (backend.PipelineConfig, error) {
			return backend.PipelineConfig{Enabled: true}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := config.NewCmdGet(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "--json"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), `"enabled"`)
	assert.Contains(t, out.String(), "true")
}

// noPipelineConfigFake wraps backend.Client without implementing
// backend.PipelineConfigClient — simulates a Bitbucket Server backend.
type noPipelineConfigFake struct {
	backend.Client
}
