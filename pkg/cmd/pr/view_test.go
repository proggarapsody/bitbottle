package pr_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
)

func TestNewCmdPRView_HasWebFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRView(f)
	assert.NotNil(t, cmd.Flag("web"))
}

func TestNewCmdPRView_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRView(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestNewCmdPRView_NotImplemented(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRView(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}
