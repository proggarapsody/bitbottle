package branch_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/branch"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestBranchCheckout_RequiresOneArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := branch.NewCmdBranchCheckout(f)
	cmd.SetArgs([]string{}) // no branch name
	err := cmd.Execute()
	require.Error(t, err)
}

func TestBranchCheckout_FetchesAndChecksOutExistingBranch(t *testing.T) {
	t.Parallel()
	// Runner responses:
	//   1. git fetch origin feat/x  → success
	//   2. git branch --list feat/x → "feat/x\n" (branch exists locally)
	//   3. git checkout feat/x      → success
	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{},                   // fetch
		testhelpers.RunResponse{Stdout: "feat/x\n"}, // branch --list (exists)
		testhelpers.RunResponse{},                   // checkout
	)
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		GitRunner: runner,
	})
	cmd := branch.NewCmdBranchCheckout(f)
	cmd.SetArgs([]string{"feat/x"})
	require.NoError(t, cmd.Execute())

	runner.AssertCalled(t, "fetch", "origin", "feat/x")
	runner.AssertCalled(t, "branch", "--list", "feat/x")
	runner.AssertCalled(t, "checkout", "feat/x")
}

func TestBranchCheckout_CreatesTrackingBranchWhenNotLocal(t *testing.T) {
	t.Parallel()
	// Runner responses:
	//   1. git fetch origin feat/new → success
	//   2. git branch --list feat/new → "" (branch does NOT exist locally)
	//   3. git checkout -b feat/new --track origin/feat/new → success
	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{},           // fetch
		testhelpers.RunResponse{Stdout: ""}, // branch --list (absent)
		testhelpers.RunResponse{},           // checkout -b ...
	)
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		GitRunner: runner,
	})
	cmd := branch.NewCmdBranchCheckout(f)
	cmd.SetArgs([]string{"feat/new"})
	require.NoError(t, cmd.Execute())

	runner.AssertCalled(t, "fetch", "origin", "feat/new")
	runner.AssertCalled(t, "branch", "--list", "feat/new")
	runner.AssertCalled(t, "checkout", "-b", "feat/new", "--track", "origin/feat/new")
	// Plain checkout must NOT have been called without -b.
	assert.Len(t, filterArgs(runner.Calls, "checkout"), 1, "only one checkout call expected")
}

// filterArgs returns calls whose first argument equals cmd.
func filterArgs(calls []testhelpers.Call, cmd string) []testhelpers.Call {
	var out []testhelpers.Call
	for _, c := range calls {
		if len(c.Args) > 0 && c.Args[0] == cmd {
			out = append(out, c)
		}
	}
	return out
}
