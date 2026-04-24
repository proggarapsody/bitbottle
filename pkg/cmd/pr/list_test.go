package pr_test

import (
	"testing"

	"github.com/aleksey/bitbottle/pkg/cmd/factory"
	"github.com/aleksey/bitbottle/pkg/cmd/pr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdPRList_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRList(f)
	assert.NotNil(t, cmd.Flag("state"))
	assert.NotNil(t, cmd.Flag("limit"))
	assert.NotNil(t, cmd.Flag("json"))
}

func TestNewCmdPRList_StateDefault(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRList(f)
	assert.Equal(t, "open", cmd.Flag("state").DefValue)
}

func TestNewCmdPRList_LimitDefault(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRList(f)
	assert.Equal(t, "30", cmd.Flag("limit").DefValue)
}

func TestNewCmdPRList_NoRemoteReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRList(f)
	err := cmd.Execute()
	require.Error(t, err)
	// No git remote and no PROJECT/REPO arg — must error.
	assert.NotNil(t, err)
}

func TestNewCmdPRList_AcceptsMaxOneArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRList(f)
	cmd.SetArgs([]string{"PROJ/repo", "extra"})
	err := cmd.Execute()
	require.Error(t, err)
}
