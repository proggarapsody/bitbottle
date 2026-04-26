package branch_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/branch"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdBranchDelete_RequiresTwoArgs(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := branch.NewCmdBranchDelete(f)
	cmd.SetArgs([]string{"myworkspace/my-service"}) // missing branch name
	err := cmd.Execute()
	require.Error(t, err)
}

func TestBranchDelete_PrintsConfirmation(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T:              t,
		DeleteBranchFn: func(ns, slug, b string) error { return nil },
	}

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   branchConfig,
		BackendOverride: fake,
	})
	cmd := branch.NewCmdBranchDelete(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "feature/login"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "feature/login")
}

func TestBranchDelete_PassesBranchNameToAPI(t *testing.T) {
	t.Parallel()

	var gotBranch string
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteBranchFn: func(ns, slug, b string) error {
			gotBranch = b
			return nil
		},
	}

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   branchConfig,
		BackendOverride: fake,
	})
	cmd := branch.NewCmdBranchDelete(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "feature/login"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "feature/login", gotBranch)
}

func TestBranchDelete_APIError_PropagatesError(t *testing.T) {
	t.Parallel()

	apiErr := errors.New("branch not found")
	fake := &testhelpers.FakeClient{
		T:              t,
		DeleteBranchFn: func(ns, slug, b string) error { return apiErr },
	}

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   branchConfig,
		BackendOverride: fake,
	})
	cmd := branch.NewCmdBranchDelete(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "feature/login"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "branch not found")
}
