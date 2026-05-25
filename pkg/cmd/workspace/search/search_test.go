package search_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/workspace/search"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const cloudConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

func TestNewCmdSearch_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := search.NewCmdSearch(f, nil)
	format.RegisterOutputFlags(cmd)
	assert.NotNil(t, cmd.Flag("query"))
	assert.NotNil(t, cmd.Flag("role"))
	assert.NotNil(t, cmd.Flag("limit"))
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("jq"))
	assert.NotNil(t, cmd.Flag("hostname"))
	assert.Equal(t, "30", cmd.Flag("limit").DefValue)
}

func TestNewCmdSearch_RejectsArgs(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := search.NewCmdSearch(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"unexpected"})
	require.Error(t, cmd.Execute())
}

func TestSearch_PrintsWorkspaceSlugs(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		SearchWorkspacesFn: func(opts backend.WorkspaceSearchOpts) ([]backend.Workspace, error) {
			return []backend.Workspace{
				{Slug: "acme", Name: "Acme Inc", UUID: "u-1", WebURL: "https://bitbucket.org/acme/"},
				{Slug: "beta", Name: "Beta Co", UUID: "u-2", WebURL: "https://bitbucket.org/beta/"},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := search.NewCmdSearch(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "acme")
	assert.Contains(t, got, "beta")
	assert.Contains(t, got, "Acme Inc")
}

func TestSearch_ForwardsQueryAndRole(t *testing.T) {
	t.Parallel()
	var gotOpts backend.WorkspaceSearchOpts
	fake := &testhelpers.FakeClient{
		T: t,
		SearchWorkspacesFn: func(opts backend.WorkspaceSearchOpts) ([]backend.Workspace, error) {
			gotOpts = opts
			return nil, nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := search.NewCmdSearch(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"--query", "myws", "--role", "owner", "--limit", "5"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "myws", gotOpts.Query)
	assert.Equal(t, "owner", gotOpts.Role)
	assert.Equal(t, 5, gotOpts.Limit)
}

func TestSearch_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		SearchWorkspacesFn: func(opts backend.WorkspaceSearchOpts) ([]backend.Workspace, error) {
			return []backend.Workspace{{Slug: "acme", Name: "Acme", UUID: "abc-123"}}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := search.NewCmdSearch(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"--json"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "abc-123")
	assert.Contains(t, got, "acme")
}

// noWorkspaceSearchFake wraps backend.Client without satisfying WorkspaceSearcher,
// simulating a Server backend invocation.
type noWorkspaceSearchFake struct {
	backend.Client
}

func TestSearch_ServerBackend_ReturnsUnsupportedError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, noWorkspaceSearchFake{Client: &testhelpers.FakeClient{T: t}})

	cmd := search.NewCmdSearch(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Cloud only")
}

func TestSearch_PartialResults(t *testing.T) {
	t.Parallel()
	searchErr := errors.New("429 Too Many Requests")
	fake := &testhelpers.FakeClient{
		T: t,
		SearchWorkspacesFn: func(opts backend.WorkspaceSearchOpts) ([]backend.Workspace, error) {
			return []backend.Workspace{
				{Slug: "acme", Name: "Acme Inc"},
			}, searchErr
		},
	}
	f, out, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := search.NewCmdSearch(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, out.String(), "acme")
	assert.Contains(t, errOut.String(), "warning: partial results")
}
