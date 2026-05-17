package member_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/workspace/member"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// cloudConfig is a single-host Cloud config — workspace members are Cloud only.
const cloudConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

func TestNewCmdList_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := member.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	assert.NotNil(t, cmd.Flag("limit"))
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("jq"))
	assert.NotNil(t, cmd.Flag("hostname"))
	assert.Equal(t, "50", cmd.Flag("limit").DefValue)
}

func TestNewCmdList_AcceptsOptionalWorkspaceArg(t *testing.T) {
	t.Parallel()
	var gotWorkspace string
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := member.NewCmdList(f, func(opts *member.Options) error {
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
	cmd := member.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"ws1", "ws2"})
	require.Error(t, cmd.Execute())
}

func TestList_PrintsMemberSlugsAndNames(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspaceMembersFn: func(workspace string, limit int) ([]backend.WorkspaceMember, error) {
			assert.Equal(t, "acme", workspace)
			return []backend.WorkspaceMember{
				{User: backend.User{Slug: "alice", DisplayName: "Alice Smith"}, Workspace: "acme"},
				{User: backend.User{Slug: "bob", DisplayName: "Bob Jones"}, Workspace: "acme"},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := member.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"acme"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "alice")
	assert.Contains(t, got, "Alice Smith")
	assert.Contains(t, got, "bob")
}

func TestList_ForwardsLimit(t *testing.T) {
	t.Parallel()
	var gotLimit int
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspaceMembersFn: func(workspace string, limit int) ([]backend.WorkspaceMember, error) {
			gotLimit = limit
			return nil, nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := member.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"acme", "--limit", "10"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, 10, gotLimit)
}

func TestList_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspaceMembersFn: func(workspace string, limit int) ([]backend.WorkspaceMember, error) {
			return []backend.WorkspaceMember{
				{User: backend.User{Slug: "alice", DisplayName: "Alice Smith"}, Workspace: "acme"},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := member.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"acme", "--json"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "alice")
	assert.Contains(t, got, "Alice Smith")
}

func TestList_NoWorkspaceArg_ReturnsError(t *testing.T) {
	// When there's no positional arg and we're not inside a git checkout
	// that can supply the workspace, listRun must return a helpful error.
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, &testhelpers.FakeClient{T: t})

	cmd := member.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace required")
}

// noWorkspaceMemberFake wraps backend.Client without satisfying
// WorkspaceMemberClient, simulating a Server backend invocation.
type noWorkspaceMemberFake struct {
	backend.Client
}

func TestList_ServerBackend_ReturnsUnsupportedError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, noWorkspaceMemberFake{Client: &testhelpers.FakeClient{T: t}})

	cmd := member.NewCmdList(f, nil)
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
		ListWorkspaceMembersFn: func(workspace string, limit int) ([]backend.WorkspaceMember, error) {
			return []backend.WorkspaceMember{
				{User: backend.User{Slug: "alice", DisplayName: "Alice Smith"}, Workspace: "acme"},
			}, listErr
		},
	}
	f, out, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := member.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"acme"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, out.String(), "alice")
	assert.Contains(t, errOut.String(), "warning: partial results")
}
