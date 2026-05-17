package list_test

// json_golden_test.go — golden-file tests for --json field stability.
//
// To regenerate golden files after an intentional field rename:
//
//	BITBOTTLE_UPDATE_GOLDEN=1 go test ./pkg/cmd/webhook/list/...

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	cmdList "github.com/proggarapsody/bitbottle/pkg/cmd/webhook/list"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestWebhookList_JSONGolden(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListWebhooksFn: func(ns, slug string) ([]backend.Webhook, error) {
			return []backend.Webhook{
				{
					ID:     "webhook-uuid-1",
					URL:    "https://example.com/hook",
					Active: true,
					Events: []string{"repo:push", "pullrequest:created"},
				},
			}, nil
		},
	}

	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := cmdList.NewCmdList(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--json"})
	require.NoError(t, cmd.Execute())

	testhelpers.AssertGolden(t, "json/webhook-list", out.String())
}
