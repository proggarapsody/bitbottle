package hook_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/workspace/hook"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdCreate_RequiredFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig})
	cmd := hook.NewCmdCreate(f, nil)
	// --url and --events are required; executing without them must fail.
	cmd.SetArgs([]string{"acme"})
	require.Error(t, cmd.Execute())
}

func TestCreate_CallsBackendWithCorrectInput(t *testing.T) {
	t.Parallel()
	var gotWorkspace string
	var gotInput backend.CreateWebhookInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreateWorkspaceWebhookFn: func(workspace string, in backend.CreateWebhookInput) (backend.Webhook, error) {
			gotWorkspace = workspace
			gotInput = in
			return backend.Webhook{ID: "new-uuid-1", URL: in.URL, Active: in.Active, Events: in.Events}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := hook.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{"acme", "--url", "https://example.com/hook", "--events", "repo:push,pullrequest:created", "--active"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "acme", gotWorkspace)
	assert.Equal(t, "https://example.com/hook", gotInput.URL)
	assert.True(t, gotInput.Active)
	assert.Contains(t, gotInput.Events, "repo:push")
	assert.Contains(t, gotInput.Events, "pullrequest:created")
	assert.Contains(t, out.String(), "new-uuid-1")
	assert.Contains(t, out.String(), "Created workspace webhook")
}

func TestCreate_RepeatableEventsFlag(t *testing.T) {
	t.Parallel()
	var gotInput backend.CreateWebhookInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreateWorkspaceWebhookFn: func(workspace string, in backend.CreateWebhookInput) (backend.Webhook, error) {
			gotInput = in
			return backend.Webhook{ID: "uuid-x", URL: in.URL, Active: in.Active, Events: in.Events}, nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := hook.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{"acme", "--url", "https://example.com/hook", "--events", "repo:push", "--events", "pullrequest:created"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, gotInput.Events, "repo:push")
	assert.Contains(t, gotInput.Events, "pullrequest:created")
}

func TestCreate_NoWorkspaceArg_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, &testhelpers.FakeClient{T: t})

	cmd := hook.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{"--url", "https://example.com/hook", "--events", "repo:push"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace required")
}
