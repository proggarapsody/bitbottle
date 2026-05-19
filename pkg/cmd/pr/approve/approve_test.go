package approve_test

import (
	"errors"
	"testing"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr/approve"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		T: t,
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
		T: t,
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
