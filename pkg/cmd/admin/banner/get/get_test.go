package get_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/banner/get"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestBannerGet_TextOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetBannerFn: func() (backend.BannerConfig, error) {
			return backend.BannerConfig{
				Message:  "Maintenance Friday 22:00 UTC",
				Audience: "ALL",
				Enabled:  true,
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := get.NewCmdBannerGet(f, nil)
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Maintenance Friday 22:00 UTC")
	assert.Contains(t, out.String(), "ALL")
	assert.Contains(t, out.String(), "true")
}

func TestBannerGet_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetBannerFn: func() (backend.BannerConfig, error) {
			return backend.BannerConfig{
				Message:  "Upgrade tonight",
				Audience: "AUTHENTICATED",
				Enabled:  false,
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := get.NewCmdBannerGet(f, nil)
	cmd.SetArgs([]string{"--json"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), `"message":"Upgrade tonight"`)
	assert.Contains(t, out.String(), `"audience":"AUTHENTICATED"`)
	assert.Contains(t, out.String(), `"enabled":false`)
}

func TestBannerGet_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	type noAdminClient struct{ backend.Client }
	wrapped := noAdminClient{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, wrapped, cmdtest.NewRunner())
	cmd := get.NewCmdBannerGet(f, nil)
	err := cmd.Execute()
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrUnsupportedOnHost, de.Kind)
}
