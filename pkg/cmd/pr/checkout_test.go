package pr_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdPRCheckout_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRCheckout(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestPRCheckout_FetchesAndChecksOutBranch(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetPRFn: func(ns, slug string, id int) (backend.PullRequest, error) {
			return testhelpers.BackendPRFactory(
				testhelpers.BackendPRWithID(42),
				testhelpers.BackendPRWithFromBranch("feat/x"),
			), nil
		},
	}
	// newPRRunner with two extra empty responses: git fetch + git checkout.
	runner := newPRRunner(testhelpers.RunResponse{}, testhelpers.RunResponse{})
	f, _, _ := newPRFactory(t, fake, runner)
	cmd := pr.NewCmdPRCheckout(f)
	cmd.SetArgs([]string{"42"})
	require.NoError(t, cmd.Execute())

	runner.AssertCalled(t, "fetch", "origin", "feat/x")
	runner.AssertCalled(t, "checkout", "feat/x")
}

func TestPRCheckout_GitError_PropagatesError(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetPRFn: func(ns, slug string, id int) (backend.PullRequest, error) {
			return testhelpers.BackendPRFactory(
				testhelpers.BackendPRWithID(42),
				testhelpers.BackendPRWithFromBranch("feat/x"),
			), nil
		},
	}
	// remote get-url ok; git fetch fails.
	runner := newPRRunner(testhelpers.RunResponse{Err: errors.New("fetch failed")})
	f, _, _ := newPRFactory(t, fake, runner)
	cmd := pr.NewCmdPRCheckout(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fetch failed")
}
