package list_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	cmdList "github.com/proggarapsody/bitbottle/pkg/cmd/webhook/list"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdList_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := cmdList.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{})
	require.Error(t, cmd.Execute())
}

func TestList_PrintsHooksOnTTY(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWebhooksFn: func(ns, slug string) ([]backend.Webhook, error) {
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "my-service", slug)
			return []backend.Webhook{
				{ID: "abc-1", URL: "https://example.com/h1", Active: true, Events: []string{"repo:push", "pullrequest:created"}},
				{ID: "abc-2", URL: "https://example.com/h2", Active: false, Events: []string{"repo:push"}},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdList.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "abc-1")
	assert.Contains(t, got, "https://example.com/h1")
	assert.Contains(t, got, "repo:push")
	assert.Contains(t, got, "abc-2")
}

func TestList_JSON(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListWebhooksFn: func(ns, slug string) ([]backend.Webhook, error) {
			return []backend.Webhook{
				{ID: "abc-1", URL: "https://example.com/h", Active: true, Events: []string{"repo:push"}},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdList.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--json"})
	require.NoError(t, cmd.Execute())
	var rows []map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &rows))
	require.Len(t, rows, 1)
	assert.Equal(t, "abc-1", rows[0]["id"])
	assert.Equal(t, "https://example.com/h", rows[0]["url"])
	assert.Equal(t, true, rows[0]["active"])
	events, ok := rows[0]["events"].([]any)
	require.True(t, ok)
	require.Len(t, events, 1)
	assert.Equal(t, "repo:push", events[0])
}

func TestList_PartialResults(t *testing.T) {
	t.Parallel()
	listErr := errors.New("429 Too Many Requests")
	fake := &testhelpers.FakeClient{
		T: t,
		ListWebhooksFn: func(ns, slug string) ([]backend.Webhook, error) {
			return []backend.Webhook{
				{ID: "partial-hook", URL: "https://example.com/partial", Active: true, Events: []string{"repo:push"}},
			}, listErr
		},
	}
	f, out, errOut := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdList.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, out.String(), "partial-hook")
	assert.Contains(t, errOut.String(), "warning: partial results")
}
