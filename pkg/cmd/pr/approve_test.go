package pr_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
)

func TestNewCmdPRApprove_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRApprove(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestNewCmdPRApprove_NotImplemented(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRApprove(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}
