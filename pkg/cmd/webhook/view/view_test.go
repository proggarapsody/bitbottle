package view_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	cmdView "github.com/proggarapsody/bitbottle/pkg/cmd/webhook/view"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdView_RequiresTwoArgs(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := cmdView.NewCmdView(f, nil)
	cmd.SetArgs([]string{"only-repo"})
	require.Error(t, cmd.Execute())
}

func TestView_PrintsHook(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetWebhookFn: func(ns, slug, id string) (backend.Webhook, error) {
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "my-service", slug)
			assert.Equal(t, "abc-1", id)
			return backend.Webhook{
				ID:     "abc-1",
				URL:    "https://example.com/h",
				Active: true,
				Events: []string{"repo:push"},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdView.NewCmdView(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "abc-1"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "abc-1")
	assert.Contains(t, got, "https://example.com/h")
	assert.Contains(t, got, "repo:push")
}
