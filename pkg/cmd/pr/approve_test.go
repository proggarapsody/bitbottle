package pr_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// TestNewCmdPRApprove_RequiresArg and TestPRApprove_MissingArg_Errors were
// identical (both verified cobra's ExactArgs(1) rejection); only one is kept.

func TestNewCmdPRApprove_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRApprove(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "arg")
}

func TestPRApprove_CallsAPI(t *testing.T) {
	t.Parallel()

	var calledID int
	var calledNS, calledSlug string
	fake := &testhelpers.FakeClient{
		T: t,
		ApprovePRFn: func(ns, slug string, id int) error {
			calledNS = ns
			calledSlug = slug
			calledID = id
			return nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRApprove(f)
	cmd.SetArgs([]string{"42"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "MYPROJ", calledNS)
	assert.Equal(t, "my-service", calledSlug)
	assert.Equal(t, 42, calledID)
	assert.Contains(t, out.String(), "42")
}

func TestPRApprove_APIError_PropagatesError(t *testing.T) {
	t.Parallel()

	apiErr := errors.New("403 forbidden")
	fake := &testhelpers.FakeClient{
		T: t,
		ApprovePRFn: func(ns, slug string, id int) error {
			return apiErr
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRApprove(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}
