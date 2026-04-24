package pr_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
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

func TestNewCmdPRMerge_NotImplemented(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRMerge(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}

func TestNewCmdPRMerge_SquashFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRMerge(f)
	cmd.SetArgs([]string{"42", "--squash"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}
