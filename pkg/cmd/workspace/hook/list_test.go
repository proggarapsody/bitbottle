package hook_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/workspace/hook"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const cloudConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

func TestNewCmdList_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := hook.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	assert.NotNil(t, cmd.Flag("hostname"))
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("jq"))
}

func TestNewCmdList_AcceptsOptionalWorkspaceArg(t *testing.T) {
	t.Parallel()
	var gotWorkspace string
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := hook.NewCmdList(f, func(opts *hook.ListOptions) error {
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
	cmd := hook.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"ws1", "ws2"})
	require.Error(t, cmd.Execute())
}

func TestList_PrintsHookFields(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspaceWebhooksFn: func(workspace string) ([]backend.Webhook, error) {
			assert.Equal(t, "acme", workspace)
			return []backend.Webhook{
				{ID: "ws-uuid-1", URL: "https://example.com/hook", Active: true, Events: []string{"repo:push", "pullrequest:created"}},
				{ID: "ws-uuid-2", URL: "https://other.example/hook", Active: false, Events: []string{"repo:push"}},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := hook.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"acme"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "ws-uuid-1")
	assert.Contains(t, got, "https://example.com/hook")
	assert.Contains(t, got, "repo:push")
	assert.Contains(t, got, "ws-uuid-2")
}

func TestList_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWorkspaceWebhooksFn: func(workspace string) ([]backend.Webhook, error) {
			return []backend.Webhook{
				{ID: "ws-uuid-1", URL: "https://example.com/hook", Active: true, Events: []string{"repo:push"}},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := hook.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"acme", "--json"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "ws-uuid-1")
	assert.Contains(t, got, "https://example.com/hook")
}

func TestList_NoWorkspaceArg_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, &testhelpers.FakeClient{T: t})

	cmd := hook.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace required")
}

// noWorkspaceWebhookFake wraps backend.Client without satisfying WorkspaceWebhookClient.
type noWorkspaceWebhookFake struct {
	backend.Client
}

func TestList_ServerBackend_ReturnsUnsupportedError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, noWorkspaceWebhookFake{Client: &testhelpers.FakeClient{T: t}})

	cmd := hook.NewCmdList(f, nil)
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
		ListWorkspaceWebhooksFn: func(workspace string) ([]backend.Webhook, error) {
			return []backend.Webhook{
				{ID: "ws-uuid-1", URL: "https://example.com/hook", Active: true, Events: []string{"repo:push"}},
			}, listErr
		},
	}
	f, out, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := hook.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"acme"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, out.String(), "ws-uuid-1")
	assert.Contains(t, errOut.String(), "warning: partial results")
}
