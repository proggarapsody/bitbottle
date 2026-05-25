package activity_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/run"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/issue/activity"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const cloudConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

func runner() *testhelpers.FakeRunner {
	return testhelpers.NewFakeRunner(testhelpers.RunResponse{
		Stdout: "https://bitbucket.org/acme/repo.git\n",
	})
}

func newFactory(t *testing.T, fake backend.Client) (*factory.Factory, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	f, out, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)
	r := runner()
	f.GitRunner = func() run.Runner { return r }
	return f, out, errOut
}

// noActivityFake wraps backend.Client without satisfying IssueActivityClient.
type noActivityFake struct {
	backend.Client
}

func TestActivity_PrintsChanges(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListIssueActivityFn: func(ns, slug string, issueID int, limit int) ([]backend.IssueChange, error) {
			assert.Equal(t, "acme", ns)
			assert.Equal(t, "repo", slug)
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
	f, out, _ := newFactory(t, fake)
	cmd := activity.NewCmdIssueActivity(f)
	cmd.SetArgs([]string{"7", "acme/repo"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "status")
	assert.Contains(t, got, "new")
	assert.Contains(t, got, "open")
	assert.Contains(t, got, "Alice")
}

func TestActivity_InvalidID(t *testing.T) {
	t.Parallel()
	f, _, _ := newFactory(t, &testhelpers.FakeClient{T: t})
	cmd := activity.NewCmdIssueActivity(f)
	cmd.SetArgs([]string{"not-a-number", "acme/repo"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid issue ID")
}

func TestActivity_UnsupportedOnServer(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, noActivityFake{Client: &testhelpers.FakeClient{T: t}})
	r := runner()
	f.GitRunner = func() run.Runner { return r }
	cmd := activity.NewCmdIssueActivity(f)
	cmd.SetArgs([]string{"7", "acme/repo"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Cloud only")
}
