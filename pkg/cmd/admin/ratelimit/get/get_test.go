package get_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/ratelimit/get"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestRateLimitGet_TextOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetRateLimitConfigFn: func() (backend.RateLimitConfig, error) {
			return backend.RateLimitConfig{Enabled: true, RequestsPerHour: 3600, ThrottleWaitMS: 500}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := get.NewCmdGet(f, nil)
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Enabled: true")
	assert.Contains(t, out.String(), "RequestsPerHour: 3600")
	assert.Contains(t, out.String(), "ThrottleWaitMS: 500")
}

func TestRateLimitGet_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetRateLimitConfigFn: func() (backend.RateLimitConfig, error) {
			return backend.RateLimitConfig{Enabled: false, RequestsPerHour: 1000, ThrottleWaitMS: 0}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := get.NewCmdGet(f, nil)
	cmd.SetArgs([]string{"--json"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), `"enabled":false`)
	assert.Contains(t, out.String(), `"requests_per_hour":1000`)
}

func TestRateLimitGet_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	type noAdminClient struct{ backend.Client }
	wrapped := noAdminClient{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, wrapped, cmdtest.NewRunner())
	cmd := get.NewCmdGet(f, nil)
	err := cmd.Execute()
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrUnsupportedOnHost, de.Kind)
}
