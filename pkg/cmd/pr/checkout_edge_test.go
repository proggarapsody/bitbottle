package pr_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// TestPRCheckout_CheckoutFailure_PropagatesError verifies that when `git
// checkout BRANCH` fails (e.g. because the branch already exists locally with
// uncommitted changes, or any other reason git refuses to switch) the error
// surfaces unmodified to the caller. The fetch succeeds but the checkout
// returns a non-nil error.
func TestPRCheckout_CheckoutFailure_PropagatesError(t *testing.T) {
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

	checkoutErr := errors.New("error: Your local changes would be overwritten by checkout")
	// Runner ordering for resolvePRTarget → checkout flow:
	//   1. remote get-url origin (newPRRunner default)
	//   2. fetch origin feat/x   (succeeds)
	//   3. checkout feat/x       (fails — simulates "branch already local with conflicts")
	runner := newPRRunner(
		testhelpers.RunResponse{},
		testhelpers.RunResponse{Err: checkoutErr},
	)
	f, _, _ := newPRFactory(t, fake, runner)

	cmd := pr.NewCmdPRCheckout(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "local changes would be overwritten",
		"checkout failure must propagate the underlying git error")
	runner.AssertCalled(t, "checkout", "feat/x")
}
