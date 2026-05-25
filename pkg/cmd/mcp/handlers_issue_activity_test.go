package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// noIssueActivityClient embeds backend.Client without satisfying
// IssueActivityClient, so AsIssueActivityClient type-assertion fails.
type noIssueActivityClient struct {
	backend.Client
}

func TestListIssueActivity_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListIssueActivityFn: func(ns, slug string, issueID int, limit int) ([]backend.IssueChange, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "myrepo", slug)
			assert.Equal(t, 7, issueID)
			return []backend.IssueChange{
				{
					ID:        1,
					Kind:      "status",
					OldVal:    "new",
					NewVal:    "open",
					CreatedOn: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
					User:      backend.User{Slug: "alice", DisplayName: "Alice"},
				},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listIssueActivity(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"repo_slug": "myrepo",
		"issue_id":  float64(7),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "status", "")
	assertJSONContains(t, result, "alice", "")
}

func TestListIssueActivity_UnsupportedOnServer(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: singleCloudConfig})
	factorytest.UseBackend(f, noIssueActivityClient{Client: &testhelpers.FakeClient{T: t}})
	h := newHandlers(f)
	result, err := h.listIssueActivity(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"repo_slug": "myrepo",
		"issue_id":  float64(7),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "host.unsupported")
}

func TestListIssueActivity_MissingIssueID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listIssueActivity(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
		"repo_slug": "myrepo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "issue_id")
}
