package approve_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr/approve"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func openPRFn(id int) func(ns, slug string, id int) (backend.PullRequest, error) {
	return func(_, _ string, _ int) (backend.PullRequest, error) {
		return backend.PullRequest{ID: id, State: "OPEN"}, nil
	}
}

func TestNewCmdApprove_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := approve.NewCmdApprove(f)
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
		T:       t,
		GetPRFn: openPRFn(42),
		ApprovePRFn: func(ns, slug string, id int) error {
			calledNS = ns
			calledSlug = slug
			calledID = id
			return nil
		},
	}
	f, out, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := approve.NewCmdApprove(f)
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
		T:       t,
		GetPRFn: openPRFn(42),
		ApprovePRFn: func(ns, slug string, id int) error {
			return apiErr
		},
	}
	f, _, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := approve.NewCmdApprove(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}

func TestPRApprove_DeclinedPR_GuardBlocksMutation(t *testing.T) {
	t.Parallel()

	approveCalled := false
	fake := &testhelpers.FakeClient{
		T: t,
		GetPRFn: func(_, _ string, _ int) (backend.PullRequest, error) {
			return backend.PullRequest{ID: 42, State: "DECLINED"}, nil
		},
		ApprovePRFn: func(_, _ string, _ int) error {
			approveCalled = true
			return nil
		},
	}
	f, _, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := approve.NewCmdApprove(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	require.True(t, errors.Is(err, backend.ErrConflict), "want ErrConflict, got %v", err)
	assert.False(t, approveCalled, "ApprovePR must not be called for a DECLINED PR")
}
