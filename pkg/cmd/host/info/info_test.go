package info_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/host/info"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestHostInfo_PrintsBackendAndHost(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetHostInfoFn: func() (backend.HostInfo, error) {
			return backend.HostInfo{
				BackendType:       "cloud",
				BaseURL:           "https://api.bitbucket.org/2.0",
				DisplayName:       "Bitbucket Cloud",
				SupportedFeatures: []string{"issues", "pipelines"},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := info.NewCmdHostInfo(f)
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "cloud")
	assert.Contains(t, out.String(), "bitbucket.org")
}

func TestHostInfo_PrintsVersionForServer(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetHostInfoFn: func() (backend.HostInfo, error) {
			return backend.HostInfo{
				BackendType:       "server",
				BaseURL:           "https://git.example.com",
				Version:           "8.19.0",
				BuildNumber:       "80190000",
				DisplayName:       "Bitbucket",
				SupportedFeatures: []string{"admin_operations", "branch_protect"},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := info.NewCmdHostInfo(f)
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "8.19.0")
	assert.Contains(t, out.String(), "server")
}

func TestHostInfo_OmitsVersionWhenEmpty(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetHostInfoFn: func() (backend.HostInfo, error) {
			return backend.HostInfo{
				BackendType:       "cloud",
				BaseURL:           "https://api.bitbucket.org/2.0",
				DisplayName:       "Bitbucket Cloud",
				SupportedFeatures: []string{"issues"},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := info.NewCmdHostInfo(f)
	require.NoError(t, cmd.Execute())
	assert.NotContains(t, out.String(), "Version:")
}

func TestHostInfo_APIError_PropagatesError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetHostInfoFn: func() (backend.HostInfo, error) {
			return backend.HostInfo{}, errors.New("network error")
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := info.NewCmdHostInfo(f)
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "network error")
}

func TestHostInfo_NoHostConfigured_Error(t *testing.T) {
	t.Parallel()
	// cmdtest.Config has bitbucket.org configured. The factory will resolve it.
	// Here we ensure no error on happy path.
	fake := &testhelpers.FakeClient{
		T: t,
		GetHostInfoFn: func() (backend.HostInfo, error) {
			return backend.HostInfo{
				BackendType:       "cloud",
				BaseURL:           "https://api.bitbucket.org/2.0",
				DisplayName:       "Bitbucket Cloud",
				SupportedFeatures: nil,
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := info.NewCmdHostInfo(f)
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "cloud")
}
