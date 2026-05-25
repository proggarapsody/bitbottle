package ipallowlist_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/workspace/ipallowlist"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdAdd_TwoArgs_WorkspaceAndCIDR(t *testing.T) {
	t.Parallel()
	var gotWorkspace, gotCIDR string
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := ipallowlist.NewCmdAdd(f, func(opts *ipallowlist.AddOptions) error {
		gotWorkspace = opts.Workspace
		gotCIDR = opts.CIDR
		return nil
	})
	cmd.SetArgs([]string{"myws", "10.0.0.0/8"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "myws", gotWorkspace)
	assert.Equal(t, "10.0.0.0/8", gotCIDR)
}

func TestNewCmdAdd_OneArg_CIDROnly(t *testing.T) {
	t.Parallel()
	var gotCIDR string
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := ipallowlist.NewCmdAdd(f, func(opts *ipallowlist.AddOptions) error {
		gotCIDR = opts.CIDR
		return nil
	})
	cmd.SetArgs([]string{"192.168.0.0/16"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "192.168.0.0/16", gotCIDR)
}

func TestNewCmdAdd_NoArgs_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := ipallowlist.NewCmdAdd(f, nil)
	cmd.SetArgs([]string{})
	require.Error(t, cmd.Execute())
}

func TestAdd_CallsBackendWithCorrectInput(t *testing.T) {
	t.Parallel()
	var gotWorkspace string
	var gotInput backend.CreateIPAllowlistInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreateIPAllowlistFn: func(workspace string, in backend.CreateIPAllowlistInput) (backend.IPAllowlist, error) {
			gotWorkspace = workspace
			gotInput = in
			return backend.IPAllowlist{
				UUID:        "new-uuid-1234",
				CIDR:        in.CIDR,
				Description: in.Description,
				Enabled:     in.Enabled,
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := ipallowlist.NewCmdAdd(f, nil)
	cmd.SetArgs([]string{"acme", "10.0.0.0/8", "--description", "Corp VPN", "--enabled"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "acme", gotWorkspace)
	assert.Equal(t, "10.0.0.0/8", gotInput.CIDR)
	assert.Equal(t, "Corp VPN", gotInput.Description)
	assert.True(t, gotInput.Enabled)
	assert.Contains(t, out.String(), "Added IP allowlist entry new-uuid-1234")
	assert.Contains(t, out.String(), "10.0.0.0/8")
}

func TestAdd_NoWorkspaceArg_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, &testhelpers.FakeClient{T: t})

	cmd := ipallowlist.NewCmdAdd(f, nil)
	// Only one arg (CIDR), no workspace resolvable from repo
	cmd.SetArgs([]string{"10.0.0.0/8"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace required")
}
