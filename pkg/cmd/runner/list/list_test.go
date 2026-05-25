package list_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/runner/list"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const cloudConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

func TestNewCmdList_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	assert.NotNil(t, cmd.Flag("hostname"))
	assert.NotNil(t, cmd.Flag("json"))
}

func TestRunnerList_PrintsRunners(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListRunnersFn: func(workspace string) ([]backend.Runner, error) {
			assert.Equal(t, "myworkspace", workspace)
			return []backend.Runner{
				{
					UUID:     "runner-1",
					Name:     "my-runner",
					State:    "ONLINE",
					Platform: backend.RunnerPlatform{Operating: "LINUX", Arch: "AMD64"},
					Labels:   []string{"self.hosted", "linux"},
				},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "runner-1")
	assert.Contains(t, got, "my-runner")
	assert.Contains(t, got, "ONLINE")
}

func TestRunnerList_EmptyList(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListRunnersFn: func(workspace string) ([]backend.Runner, error) {
			return []backend.Runner{}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "No runners found")
}

func TestRunnerList_NoWorkspace_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, &testhelpers.FakeClient{T: t})

	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace required")
}

func TestRunnerList_ClientNotCapable_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &noRunnerFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipeline runners")
}

func TestRunnerList_PartialResults(t *testing.T) {
	t.Parallel()
	listErr := errors.New("429 Too Many Requests")
	fake := &testhelpers.FakeClient{
		T: t,
		ListRunnersFn: func(workspace string) ([]backend.Runner, error) {
			return []backend.Runner{
				{UUID: "partial-runner", Name: "r1", State: "ONLINE"},
			}, listErr
		},
	}
	f, out, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := list.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, out.String(), "partial-runner")
	assert.Contains(t, errOut.String(), "warning: partial results")
}

// noRunnerFake wraps backend.Client without implementing backend.RunnerClient.
type noRunnerFake struct {
	backend.Client
}
