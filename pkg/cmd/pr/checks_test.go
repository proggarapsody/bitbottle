package pr_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
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

func TestPRChecks_JSONOutput(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetPRFn: func(ns, slug string, id int) (backend.PullRequest, error) {
			return backend.PullRequest{ID: 42, HeadCommitHash: "abc1234"}, nil
		},
		ListCommitStatusesFn: func(ns, slug, hash string) ([]backend.CommitStatus, error) {
			assert.Equal(t, "abc1234", hash)
			return []backend.CommitStatus{
				{Key: "ci/build", State: "SUCCESSFUL", Name: "CI Build", Description: "passed", URL: "https://ci.example.com/1"},
				{Key: "ci/lint", State: "FAILED", Name: "Lint", URL: "https://ci.example.com/2"},
			}, nil
		},
	}

	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRChecks(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "--json"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.True(t, strings.HasPrefix(got, "["), "expected JSON array, got: %s", got)
	assert.Contains(t, got, `"ci/build"`)
	assert.Contains(t, got, `"SUCCESSFUL"`)
	assert.Contains(t, got, `"description"`)
	assert.Contains(t, got, `"passed"`)
}

func TestPRChecks_JSONAndWatchMutuallyExclusive(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRChecks(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "--json", "--watch"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}
