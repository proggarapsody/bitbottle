package pr_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdPRUpdateBranch_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := pr.NewCmdPRUpdateBranch(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestPRUpdateBranch_CallsAPI(t *testing.T) {
	t.Parallel()

	var calledNS, calledSlug string
	var calledID int
	fake := &testhelpers.FakeClient{
		T: t,
		UpdatePRBranchFn: func(ns, slug string, prID int) error {
			calledNS = ns
			calledSlug = slug
			calledID = prID
			return nil
		},
	}

	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRUpdateBranch(f)
	cmd.SetArgs([]string{"42"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "MYPROJ", calledNS)
	assert.Equal(t, "my-service", calledSlug)
	assert.Equal(t, 42, calledID)
	assert.Contains(t, out.String(), "branch updated")
}

func TestPRUpdateBranch_APIError_PropagatesError(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		UpdatePRBranchFn: func(ns, slug string, prID int) error {
			return errors.New("rebase conflict")
		},
	}

	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRUpdateBranch(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rebase conflict")
}
