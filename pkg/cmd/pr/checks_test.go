package pr_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdPRChecks_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := pr.NewCmdPRChecks(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestNewCmdPRChecks_HasWatchFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := pr.NewCmdPRChecks(f)
	assert.NotNil(t, cmd.Flag("watch"))
}

func TestNewCmdPRChecks_HasIntervalFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := pr.NewCmdPRChecks(f)
	assert.NotNil(t, cmd.Flag("interval"))
	assert.Equal(t, "10", cmd.Flag("interval").DefValue)
}

func TestPRChecks_ListsStatuses(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetPRFn: func(ns, slug string, id int) (backend.PullRequest, error) {
			return backend.PullRequest{
				ID:             42,
				Title:          "Fix something",
				HeadCommitHash: "abc1234",
			}, nil
		},
		ListCommitStatusesFn: func(ns, slug, hash string) ([]backend.CommitStatus, error) {
			return []backend.CommitStatus{
				{Key: "ci/build", State: "SUCCESSFUL", Name: "CI Build", URL: "https://ci.example.com/1"},
			}, nil
		},
	}

	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRChecks(f)
	cmd.SetArgs([]string{"42"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "ci/build")
	assert.Contains(t, got, "SUCCESSFUL")
}

func TestPRChecks_ErrorWhenNoHeadCommit(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetPRFn: func(ns, slug string, id int) (backend.PullRequest, error) {
			return backend.PullRequest{ID: 42, Title: "PR", HeadCommitHash: ""}, nil
		},
	}

	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRChecks(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "head commit hash unavailable")
}
