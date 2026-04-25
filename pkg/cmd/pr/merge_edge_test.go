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

// TestPRMerge_DeleteBranchFailure_WrapsError verifies that when MergePR
// succeeds but the subsequent DeleteBranch call fails, the command:
//   - returns a non-nil error
//   - the error message wraps "merge succeeded but failed to delete branch"
//   - the underlying delete error is preserved via %w wrapping (errors.Is)
func TestPRMerge_DeleteBranchFailure_WrapsError(t *testing.T) {
	t.Parallel()

	deleteErr := errors.New("403 forbidden")
	fake := &testhelpers.FakeClient{
		T: t,
		MergePRFn: func(ns, slug string, id int, in backend.MergePRInput) (backend.PullRequest, error) {
			return testhelpers.BackendPRFactory(
				testhelpers.BackendPRWithState("MERGED"),
				testhelpers.BackendPRWithFromBranch("feat/x"),
			), nil
		},
		DeleteBranchFn: func(ns, slug, branch string) error {
			return deleteErr
		},
	}

	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRMerge(f)
	cmd.SetArgs([]string{"42", "--delete-branch"})
	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "merge succeeded but failed to delete branch",
		"error must distinguish partial success — merge done, branch left behind")
	assert.True(t, errors.Is(err, deleteErr),
		"wrapped error must preserve the underlying DeleteBranch error via %%w")
}
