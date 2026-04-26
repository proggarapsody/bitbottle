package branch_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/branch"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestBranchCreate_RequiresStartAt(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   branchConfig,
		BackendOverride: fake,
	})
	cmd := branch.NewCmdBranchCreate(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "feat/x"}) // missing --start-at
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "start-at")
}

func TestBranchCreate_PrintsConfirmation(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		CreateBranchFn: func(ns, slug string, in backend.CreateBranchInput) (backend.Branch, error) {
			return backend.Branch{Name: in.Name, LatestHash: "abc123"}, nil
		},
	}
	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   branchConfig,
		BackendOverride: fake,
	})
	cmd := branch.NewCmdBranchCreate(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "feat/x", "--start-at", "main"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "feat/x")
}

func TestBranchCreate_PassesNameToAPI(t *testing.T) {
	t.Parallel()
	var gotName string
	fake := &testhelpers.FakeClient{
		T: t,
		CreateBranchFn: func(ns, slug string, in backend.CreateBranchInput) (backend.Branch, error) {
			gotName = in.Name
			return backend.Branch{Name: in.Name}, nil
		},
	}
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   branchConfig,
		BackendOverride: fake,
	})
	cmd := branch.NewCmdBranchCreate(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "feat/my-feature", "--start-at", "main"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "feat/my-feature", gotName)
}

func TestBranchCreate_PassesStartAtToAPI(t *testing.T) {
	t.Parallel()
	var gotStartAt string
	fake := &testhelpers.FakeClient{
		T: t,
		CreateBranchFn: func(ns, slug string, in backend.CreateBranchInput) (backend.Branch, error) {
			gotStartAt = in.StartAt
			return backend.Branch{Name: in.Name}, nil
		},
	}
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   branchConfig,
		BackendOverride: fake,
	})
	cmd := branch.NewCmdBranchCreate(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "feat/x", "--start-at", "develop"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "develop", gotStartAt)
}

func TestBranchCreate_APIError_PropagatesError(t *testing.T) {
	t.Parallel()
	apiErr := errors.New("branch already exists")
	fake := &testhelpers.FakeClient{
		T: t,
		CreateBranchFn: func(ns, slug string, in backend.CreateBranchInput) (backend.Branch, error) {
			return backend.Branch{}, apiErr
		},
	}
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   branchConfig,
		BackendOverride: fake,
	})
	cmd := branch.NewCmdBranchCreate(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "feat/x", "--start-at", "main"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "branch already exists")
}
