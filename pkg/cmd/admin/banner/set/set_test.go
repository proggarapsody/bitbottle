package set_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/banner/set"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestBannerSet_CallsBackendWithArgs(t *testing.T) {
	t.Parallel()
	var gotCfg backend.BannerConfig
	fake := &testhelpers.FakeClient{
		T: t,
		SetBannerFn: func(in backend.BannerConfig) error {
			gotCfg = in
			return nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := set.NewCmdBannerSet(f, nil)
	cmd.SetArgs([]string{"Scheduled maintenance", "--audience", "authenticated"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "Scheduled maintenance", gotCfg.Message)
	assert.Equal(t, "AUTHENTICATED", gotCfg.Audience)
	assert.True(t, gotCfg.Enabled)
	assert.Contains(t, out.String(), "updated")
}

func TestBannerSet_DefaultAudience(t *testing.T) {
	t.Parallel()
	var gotCfg backend.BannerConfig
	fake := &testhelpers.FakeClient{
		T: t,
		SetBannerFn: func(in backend.BannerConfig) error {
			gotCfg = in
			return nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := set.NewCmdBannerSet(f, nil)
	cmd.SetArgs([]string{"Hello everyone"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "ALL", gotCfg.Audience)
}

func TestBannerSet_InvalidAudience(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := set.NewCmdBannerSet(f, nil)
	cmd.SetArgs([]string{"msg", "--audience", "INVALID"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "audience")
}

func TestBannerSet_RequiresMessage(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := set.NewCmdBannerSet(f, nil)
	err := cmd.Execute()
	require.Error(t, err)
}

func TestBannerSet_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	type noAdminClient struct{ backend.Client }
	wrapped := noAdminClient{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, wrapped, cmdtest.NewRunner())
	cmd := set.NewCmdBannerSet(f, nil)
	cmd.SetArgs([]string{"test message"})
	err := cmd.Execute()
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrUnsupportedOnHost, de.Kind)
}
