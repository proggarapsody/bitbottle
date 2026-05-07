package list_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/workspace/list"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// cloudConfig is a single-host Cloud config — workspaces only exist on Cloud.
const cloudConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

func TestNewCmdList_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := list.NewCmdList(f, nil)
	assert.NotNil(t, cmd.Flag("limit"))
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("jq"))
	assert.NotNil(t, cmd.Flag("hostname"))
	assert.Equal(t, "30", cmd.Flag("limit").DefValue)
}

func TestNewCmdList_RejectsArgs(t *testing.T) {
	// Workspace listing has no positional args. Catching this in cobra's
	// arg validator stops typos like `bitbottle workspace list myworkspace`
	// from silently being ignored.
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := list.NewCmdList(f, nil)
	cmd.SetArgs([]string{"unexpected"})
	require.Error(t, cmd.Execute())
}

func TestList_PrintsWorkspaceSlugs(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspacesFn: func(limit int) ([]backend.Workspace, error) {
			return []backend.Workspace{
				{Slug: "acme", Name: "Acme Inc", UUID: "u-1", WebURL: "https://bitbucket.org/acme/"},
				{Slug: "beta", Name: "Beta Co", UUID: "u-2", WebURL: "https://bitbucket.org/beta/"},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := list.NewCmdList(f, nil)
	cmd.SetArgs([]string{})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "acme")
	assert.Contains(t, got, "beta")
	assert.Contains(t, got, "Acme Inc")
}

func TestList_ForwardsLimit(t *testing.T) {
	t.Parallel()
	var gotLimit int
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspacesFn: func(limit int) ([]backend.Workspace, error) {
			gotLimit = limit
			return nil, nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := list.NewCmdList(f, nil)
	cmd.SetArgs([]string{"--limit", "5"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, 5, gotLimit)
}

func TestList_JSONOutput_OmitsTableButIncludesUUID(t *testing.T) {
	// JSON path must include UUID (JSONOnly field) and must not be coloured
	// — those are properties the format printer enforces; this test guards
	// the wiring in this command.
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspacesFn: func(limit int) ([]backend.Workspace, error) {
			return []backend.Workspace{{Slug: "acme", Name: "Acme", UUID: "abc-123"}}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := list.NewCmdList(f, nil)
	cmd.SetArgs([]string{"--json", "uuid,slug"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "abc-123")
	assert.Contains(t, got, "acme")
}

// noWorkspaceFake wraps backend.Client without satisfying WorkspaceClient,
// simulating a Server backend invocation. The interface embedding (not the
// concrete struct) prevents method promotion.
type noWorkspaceFake struct {
	backend.Client
}

func TestList_ServerBackend_ReturnsUnsupportedError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, noWorkspaceFake{Client: &testhelpers.FakeClient{T: t}})

	cmd := list.NewCmdList(f, nil)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Cloud only")
}
