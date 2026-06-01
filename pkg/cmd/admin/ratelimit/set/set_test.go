package set_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/ratelimit/set"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// stubCurrent returns a fake GetRateLimitConfigFn that responds with a fixed config.
func stubCurrent(cfg backend.RateLimitConfig) func() (backend.RateLimitConfig, error) {
	return func() (backend.RateLimitConfig, error) { return cfg, nil }
}

func TestRateLimitSet_Enable_Succeeds(t *testing.T) {
	t.Parallel()
	var gotIn backend.RateLimitConfig
	fake := &testhelpers.FakeClient{
		T:                    t,
		GetRateLimitConfigFn: stubCurrent(backend.RateLimitConfig{Enabled: false, RequestsPerHour: 1000, ThrottleWaitMS: 100}),
		SetRateLimitConfigFn: func(in backend.RateLimitConfig) error {
			gotIn = in
			return nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := set.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"--enabled"})
	require.NoError(t, cmd.Execute())
	assert.True(t, gotIn.Enabled)
	assert.Equal(t, 1000, gotIn.RequestsPerHour) // unchanged
	assert.Contains(t, out.String(), "updated")
}

func TestRateLimitSet_RequestsPerHour_Succeeds(t *testing.T) {
	t.Parallel()
	var gotIn backend.RateLimitConfig
	fake := &testhelpers.FakeClient{
		T:                    t,
		GetRateLimitConfigFn: stubCurrent(backend.RateLimitConfig{Enabled: true, RequestsPerHour: 500, ThrottleWaitMS: 0}),
		SetRateLimitConfigFn: func(in backend.RateLimitConfig) error {
			gotIn = in
			return nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := set.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"--requests-per-hour", "7200"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, 7200, gotIn.RequestsPerHour)
	assert.True(t, gotIn.Enabled) // unchanged from current
}

func TestRateLimitSet_ThrottleWaitMS_Succeeds(t *testing.T) {
	t.Parallel()
	var gotIn backend.RateLimitConfig
	fake := &testhelpers.FakeClient{
		T:                    t,
		GetRateLimitConfigFn: stubCurrent(backend.RateLimitConfig{Enabled: true, RequestsPerHour: 3600, ThrottleWaitMS: 100}),
		SetRateLimitConfigFn: func(in backend.RateLimitConfig) error {
			gotIn = in
			return nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := set.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"--throttle-wait-ms", "250"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, 250, gotIn.ThrottleWaitMS)
}

func TestRateLimitSet_NoFlags_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := set.NewCmdSet(f, nil)
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one of")
}

func TestRateLimitSet_PermissionError_PrintsHint(t *testing.T) {
	t.Parallel()
	permErr := &backend.DomainError{Kind: backend.ErrPermission, Message: "forbidden"}
	fake := &testhelpers.FakeClient{
		T:                    t,
		GetRateLimitConfigFn: stubCurrent(backend.RateLimitConfig{}),
		SetRateLimitConfigFn: func(in backend.RateLimitConfig) error { return permErr },
	}
	f, _, errOut := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := set.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"--enabled"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, errOut.String(), "SYS_ADMIN")
}

func TestRateLimitSet_RequestsPerHour_Zero_Succeeds(t *testing.T) {
	t.Parallel()
	var gotIn backend.RateLimitConfig
	fake := &testhelpers.FakeClient{
		T:                    t,
		GetRateLimitConfigFn: stubCurrent(backend.RateLimitConfig{Enabled: true, RequestsPerHour: 500, ThrottleWaitMS: 200}),
		SetRateLimitConfigFn: func(in backend.RateLimitConfig) error {
			gotIn = in
			return nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := set.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"--requests-per-hour", "0"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, 0, gotIn.RequestsPerHour)
	assert.Equal(t, 200, gotIn.ThrottleWaitMS) // unchanged
}

func TestRateLimitSet_ThrottleWaitMS_Zero_Succeeds(t *testing.T) {
	t.Parallel()
	var gotIn backend.RateLimitConfig
	fake := &testhelpers.FakeClient{
		T:                    t,
		GetRateLimitConfigFn: stubCurrent(backend.RateLimitConfig{Enabled: true, RequestsPerHour: 3600, ThrottleWaitMS: 100}),
		SetRateLimitConfigFn: func(in backend.RateLimitConfig) error {
			gotIn = in
			return nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := set.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"--throttle-wait-ms", "0"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, 0, gotIn.ThrottleWaitMS)
	assert.Equal(t, 3600, gotIn.RequestsPerHour) // unchanged
}

func TestRateLimitSet_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	type noAdminClient struct{ backend.Client }
	wrapped := noAdminClient{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, wrapped, cmdtest.NewRunner())
	cmd := set.NewCmdSet(f, nil)
	cmd.SetArgs([]string{"--enabled"})
	err := cmd.Execute()
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrUnsupportedOnHost, de.Kind)
}
