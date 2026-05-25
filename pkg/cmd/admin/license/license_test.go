package license_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/admin/license"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestLicense_TextOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetLicenseFn: func() (backend.AdminLicense, error) {
			return backend.AdminLicense{
				Tier:          "ENTERPRISE",
				Users:         500,
				ServerId:      "srv-abc123",
				ExpiryDate:    "2027-01-01",
				SupportExpiry: "2026-01-01",
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := license.NewCmdLicense(f, nil)
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "ENTERPRISE")
	assert.Contains(t, out.String(), "srv-abc123")
	assert.Contains(t, out.String(), "2027-01-01")
}

func TestLicense_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetLicenseFn: func() (backend.AdminLicense, error) {
			return backend.AdminLicense{
				Tier:     "STARTER",
				Users:    25,
				ServerId: "srv-xyz",
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := license.NewCmdLicense(f, nil)
	cmd.SetArgs([]string{"--json"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), `"Tier":"STARTER"`)
	assert.Contains(t, out.String(), `"Users":25`)
}

func TestLicense_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	type noAdminClient struct{ backend.Client }
	wrapped := noAdminClient{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, wrapped, cmdtest.NewRunner())
	cmd := license.NewCmdLicense(f, nil)
	err := cmd.Execute()
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrUnsupportedOnHost, de.Kind)
}
