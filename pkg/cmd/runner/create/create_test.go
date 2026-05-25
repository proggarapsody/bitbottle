package create_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/runner/create"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const cloudConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

func TestNewCmdCreate_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := create.NewCmdCreate(f, nil)
	assert.NotNil(t, cmd.Flag("hostname"))
	assert.NotNil(t, cmd.Flag("name"))
	assert.NotNil(t, cmd.Flag("label"))
	assert.NotNil(t, cmd.Flag("platform"))
}

func TestRunnerCreate_CreatesRunner(t *testing.T) {
	t.Parallel()
	var gotWorkspace string
	var gotIn backend.CreateRunnerInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreateRunnerFn: func(workspace string, in backend.CreateRunnerInput) (backend.Runner, error) {
			gotWorkspace = workspace
			gotIn = in
			return backend.Runner{UUID: "new-uuid", Name: in.Name, State: "OFFLINE"}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := create.NewCmdCreate(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace", "--name", "my-runner", "--label", "self.hosted", "--platform", "linux_amd64"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "myworkspace", gotWorkspace)
	assert.Equal(t, "my-runner", gotIn.Name)
	assert.Equal(t, []string{"self.hosted"}, gotIn.Labels)
	assert.Equal(t, backend.RunnerPlatform{Operating: "LINUX", Arch: "AMD64"}, gotIn.Platform)
	assert.Contains(t, out.String(), "new-uuid")
	assert.Contains(t, out.String(), "my-runner")
}

func TestRunnerCreate_InvalidPlatform_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, &testhelpers.FakeClient{T: t})

	cmd := create.NewCmdCreate(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace", "--name", "r", "--platform", "invalid_platform"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --platform")
}

func TestRunnerCreate_NoWorkspace_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, &testhelpers.FakeClient{T: t})

	cmd := create.NewCmdCreate(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"--name", "r"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace required")
}

func TestRunnerCreate_ClientNotCapable_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &noRunnerFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := create.NewCmdCreate(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace", "--name", "r"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipeline runners")
}

func TestRunnerCreate_PlatformVariants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		platform string
		wantOS   string
		wantArch string
	}{
		{"linux_amd64", "LINUX", "AMD64"},
		{"linux_arm64", "LINUX", "ARM64"},
		{"windows_amd64", "WINDOWS", "AMD64"},
		{"macos_arm64", "MACOS", "ARM64"},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.platform, func(t *testing.T) {
			t.Parallel()
			var gotIn backend.CreateRunnerInput
			fake := &testhelpers.FakeClient{
				T: t,
				CreateRunnerFn: func(workspace string, in backend.CreateRunnerInput) (backend.Runner, error) {
					gotIn = in
					return backend.Runner{UUID: "u", Name: in.Name}, nil
				},
			}
			f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
			factorytest.UseBackend(f, fake)

			cmd := create.NewCmdCreate(f, nil)
			format.RegisterOutputFlags(cmd)
			cmd.SetArgs([]string{"myws", "--name", "r", "--platform", tc.platform})
			require.NoError(t, cmd.Execute())
			assert.Equal(t, tc.wantOS, gotIn.Platform.Operating)
			assert.Equal(t, tc.wantArch, gotIn.Platform.Arch)
		})
	}
}

// noRunnerFake wraps backend.Client without implementing backend.RunnerClient.
type noRunnerFake struct {
	backend.Client
}
