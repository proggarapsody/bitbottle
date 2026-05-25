package ipallowlist_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/workspace/ipallowlist"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const cloudConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

func TestNewCmdList_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := ipallowlist.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	assert.NotNil(t, cmd.Flag("hostname"))
	assert.NotNil(t, cmd.Flag("limit"))
	assert.NotNil(t, cmd.Flag("json"))
}

func TestNewCmdList_AcceptsOptionalWorkspaceArg(t *testing.T) {
	t.Parallel()
	var gotWorkspace string
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := ipallowlist.NewCmdList(f, func(opts *ipallowlist.ListOptions) error {
		gotWorkspace = opts.Workspace
		return nil
	})
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myws"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "myws", gotWorkspace)
}

func TestNewCmdList_RejectsTooManyArgs(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := ipallowlist.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"ws1", "ws2"})
	require.Error(t, cmd.Execute())
}

func TestList_PrintsEntryFields(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListIPAllowlistsFn: func(workspace string) ([]backend.IPAllowlist, error) {
			assert.Equal(t, "acme", workspace)
			return []backend.IPAllowlist{
				{UUID: "{aaaa-1111}", CIDR: "10.0.0.0/8", Description: "Corporate VPN", Enabled: true},
				{UUID: "{bbbb-2222}", CIDR: "192.168.1.0/24", Description: "Office", Enabled: false},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := ipallowlist.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"acme"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "10.0.0.0/8")
	assert.Contains(t, got, "Corporate VPN")
	assert.Contains(t, got, "192.168.1.0/24")
	assert.Contains(t, got, "Office")
}

func TestList_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListIPAllowlistsFn: func(workspace string) ([]backend.IPAllowlist, error) {
			return []backend.IPAllowlist{
				{UUID: "{aaaa-1111}", CIDR: "10.0.0.0/8", Description: "Corp VPN", Enabled: true},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := ipallowlist.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"acme", "--json"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "10.0.0.0/8")
	assert.Contains(t, got, "Corp VPN")
}

func TestList_NoWorkspaceArg_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, &testhelpers.FakeClient{T: t})

	cmd := ipallowlist.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace required")
}

// noIPAllowlistFake wraps backend.Client without satisfying IPAllowlistClient.
type noIPAllowlistFake struct {
	backend.Client
}

func TestList_ServerBackend_ReturnsUnsupportedError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, noIPAllowlistFake{Client: &testhelpers.FakeClient{T: t}})

	cmd := ipallowlist.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"acme"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Cloud only")
}

func TestList_PartialResults(t *testing.T) {
	t.Parallel()
	listErr := errors.New("429 Too Many Requests")
	fake := &testhelpers.FakeClient{
		T: t,
		ListIPAllowlistsFn: func(workspace string) ([]backend.IPAllowlist, error) {
			return []backend.IPAllowlist{
				{UUID: "{aaaa-1111}", CIDR: "10.0.0.0/8", Description: "Corp VPN", Enabled: true},
			}, listErr
		},
	}
	f, out, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := ipallowlist.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"acme"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, out.String(), "10.0.0.0/8")
	assert.Contains(t, errOut.String(), "warning: partial results")
}
