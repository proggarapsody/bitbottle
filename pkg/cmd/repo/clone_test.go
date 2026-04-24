package repo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
)

func TestNewCmdRepoClone_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := repo.NewCmdRepoClone(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestNewCmdRepoClone_NotImplemented(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := repo.NewCmdRepoClone(f)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}

func TestNewCmdRepoClone_AcceptsTargetDir(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := repo.NewCmdRepoClone(f)
	cmd.SetArgs([]string{"MYPROJ/my-service", "/tmp/dest"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}
