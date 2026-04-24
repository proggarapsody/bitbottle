package repo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
)

func TestNewCmdRepoDelete_HasConfirmFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := repo.NewCmdRepoDelete(f)
	assert.NotNil(t, cmd.Flag("confirm"))
}

func TestNewCmdRepoDelete_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := repo.NewCmdRepoDelete(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestNewCmdRepoDelete_NotImplemented(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := repo.NewCmdRepoDelete(f)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--confirm"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}
