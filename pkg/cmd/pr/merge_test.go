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

func TestNewCmdPRMerge_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRMerge(f)
	assert.NotNil(t, cmd.Flag("merge"))
	assert.NotNil(t, cmd.Flag("squash"))
	assert.NotNil(t, cmd.Flag("delete-branch"))
}

func TestNewCmdPRMerge_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRMerge(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestPRMerge_MergeStrategy_CallsAPI(t *testing.T) {
	t.Parallel()

	var captured backend.MergePRInput
	fake := &testhelpers.FakeClient{
		T: t,
		MergePRFn: func(ns, slug string, id int, in backend.MergePRInput) (backend.PullRequest, error) {
			captured = in
			return testhelpers.BackendPRFactory(testhelpers.BackendPRWithState("MERGED")), nil
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRMerge(f)
	cmd.SetArgs([]string{"42", "--merge"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "merge-commit", captured.Strategy)
}

func TestPRMerge_SquashStrategy_CallsAPI(t *testing.T) {
	t.Parallel()

	var captured backend.MergePRInput
	fake := &testhelpers.FakeClient{
		T: t,
		MergePRFn: func(ns, slug string, id int, in backend.MergePRInput) (backend.PullRequest, error) {
			captured = in
			return testhelpers.BackendPRFactory(testhelpers.BackendPRWithState("MERGED")), nil
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRMerge(f)
	cmd.SetArgs([]string{"42", "--squash"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "squash", captured.Strategy)
}

func TestPRMerge_BothMergeAndSquash_Errors(t *testing.T) {
	t.Parallel()

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: prConfig,
		GitRunner:     newPRRunner(),
	})
	cmd := pr.NewCmdPRMerge(f)
	cmd.SetArgs([]string{"42", "--merge", "--squash"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--merge")
	assert.Contains(t, err.Error(), "--squash")
}

func TestPRMerge_DeleteBranch_CallsDeleteBranchAfterMerge(t *testing.T) {
	t.Parallel()

	mergeCalled := false
	deleteCalled := false
	var capturedBranch string

	fake := &testhelpers.FakeClient{
		T: t,
		MergePRFn: func(ns, slug string, id int, in backend.MergePRInput) (backend.PullRequest, error) {
			mergeCalled = true
			return testhelpers.BackendPRFactory(
				testhelpers.BackendPRWithState("MERGED"),
				testhelpers.BackendPRWithFromBranch("feat/x"),
			), nil
		},
		DeleteBranchFn: func(ns, slug, branch string) error {
			deleteCalled = true
			capturedBranch = branch
			return nil
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRMerge(f)
	cmd.SetArgs([]string{"42", "--delete-branch"})
	require.NoError(t, cmd.Execute())

	assert.True(t, mergeCalled, "MergePR must be called")
	assert.True(t, deleteCalled, "DeleteBranch must be called when --delete-branch is set")
	assert.Equal(t, "feat/x", capturedBranch)
}

func TestPRMerge_APIError_PropagatesError(t *testing.T) {
	t.Parallel()

	apiErr := errors.New("500 internal error")
	fake := &testhelpers.FakeClient{
		T: t,
		MergePRFn: func(ns, slug string, id int, in backend.MergePRInput) (backend.PullRequest, error) {
			return backend.PullRequest{}, apiErr
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRMerge(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}
