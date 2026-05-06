package logs_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/logs"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdLogs_RequiresThreeArgs(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := logs.NewCmdLogs(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "p-uuid"}) // missing step UUID
	require.Error(t, cmd.Execute())
}

func TestLogs_StreamsBytesToStdout(t *testing.T) {
	t.Parallel()
	const payload = "step-1: starting\nstep-1: line two\nstep-1: done\n"
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelineStepLogFn: func(ns, slug, p, s string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(payload)), nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := logs.NewCmdLogs(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "p-uuid", "s-uuid"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, payload, out.String())
}

func TestLogs_PassesUUIDsToBackend(t *testing.T) {
	t.Parallel()
	var gotPipeline, gotStep string
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelineStepLogFn: func(ns, slug, p, s string) (io.ReadCloser, error) {
			gotPipeline, gotStep = p, s
			return io.NopCloser(strings.NewReader("")), nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := logs.NewCmdLogs(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "{p-uuid}", "{s-uuid}"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "{p-uuid}", gotPipeline)
	assert.Equal(t, "{s-uuid}", gotStep)
}

func TestLogs_APIError_PropagatesError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelineStepLogFn: func(ns, slug, p, s string) (io.ReadCloser, error) {
			return nil, errors.New("not found")
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := logs.NewCmdLogs(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "p", "s"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestLogs_ClientNotPipelineCapable_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &cmdtest.NoPipelineFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := logs.NewCmdLogs(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "p", "s"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipelines")
}
