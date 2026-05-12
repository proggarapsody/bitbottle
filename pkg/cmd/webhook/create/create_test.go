package create_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	cmdCreate "github.com/proggarapsody/bitbottle/pkg/cmd/webhook/create"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdCreate_RequiresURLAndEvents(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := cmdCreate.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.Error(t, cmd.Execute())
}

func TestCreate_PassesParsedEventsAndURL(t *testing.T) {
	t.Parallel()
	var got backend.CreateWebhookInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreateWebhookFn: func(ns, slug string, in backend.CreateWebhookInput) (backend.Webhook, error) {
			got = in
			return backend.Webhook{ID: "abc-1", URL: in.URL, Events: in.Events, Active: in.Active}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdCreate.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{
		"myworkspace/my-service",
		"--url", "https://example.com/hook",
		"--events", "repo:push, pullrequest:created ,repo:push",
	})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "https://example.com/hook", got.URL)
	assert.Equal(t, []string{"repo:push", "pullrequest:created"}, got.Events)
	assert.True(t, got.Active)
	assert.Empty(t, got.Secret)
	assert.Contains(t, out.String(), "abc-1")
}

func TestCreate_ActiveFalseFlag(t *testing.T) {
	t.Parallel()
	var got backend.CreateWebhookInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreateWebhookFn: func(ns, slug string, in backend.CreateWebhookInput) (backend.Webhook, error) {
			got = in
			return backend.Webhook{ID: "x", URL: in.URL, Active: in.Active}, nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdCreate.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{
		"myworkspace/my-service",
		"--url", "https://example.com/hook",
		"--events", "repo:push",
		"--active=false",
	})
	require.NoError(t, cmd.Execute())
	assert.False(t, got.Active)
}

func TestCreate_SecretInline_PassesThrough(t *testing.T) {
	t.Parallel()
	var got backend.CreateWebhookInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreateWebhookFn: func(ns, slug string, in backend.CreateWebhookInput) (backend.Webhook, error) {
			got = in
			return backend.Webhook{ID: "x", URL: in.URL}, nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdCreate.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{
		"myworkspace/my-service",
		"--url", "https://example.com/hook",
		"--events", "repo:push",
		"--secret", "redacted-test-value",
	})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "redacted-test-value", got.Secret)
}

func TestCreate_SecretFromFile_TrimsTrailingNewline(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "secret")
	require.NoError(t, os.WriteFile(path, []byte("from-file\n"), 0o600))

	var got backend.CreateWebhookInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreateWebhookFn: func(ns, slug string, in backend.CreateWebhookInput) (backend.Webhook, error) {
			got = in
			return backend.Webhook{ID: "x", URL: in.URL}, nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdCreate.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{
		"myworkspace/my-service",
		"--url", "https://example.com/hook",
		"--events", "repo:push",
		"--secret", "@" + path,
	})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "from-file", got.Secret, "trailing newline must be trimmed when reading from file")
}

func TestCreate_SecretStdin_CapturedOptionsHasDashSentinel(t *testing.T) {
	t.Parallel()
	// Mirror scope-H pattern: assert flag plumbing via runF capture; full stdin
	// roundtrip is exercised via cmd-level integration in manual tests.
	var captured *cmdCreate.Options
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdCreate.NewCmdCreate(f, func(o *cmdCreate.Options) error {
		captured = o
		captured.Stdin = strings.NewReader("piped-secret\n")
		return nil
	})
	cmd.SetArgs([]string{
		"myworkspace/my-service",
		"--url", "https://example.com/hook",
		"--events", "repo:push",
		"--secret", "-",
	})
	require.NoError(t, cmd.Execute())
	require.NotNil(t, captured)
	assert.Equal(t, "-", captured.Secret)
	assert.Equal(t, "https://example.com/hook", captured.URL)
}

func TestCreate_EventsAllWhitespaceRejected(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdCreate.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{
		"myworkspace/my-service",
		"--url", "https://example.com/hook",
		"--events", " , ,  ",
	})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--events")
}
