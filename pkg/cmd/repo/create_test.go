package repo_test

import (
	"testing"

	"github.com/aleksey/bitbottle/pkg/cmd/factory"
	"github.com/aleksey/bitbottle/pkg/cmd/repo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdRepoCreate_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := repo.NewCmdRepoCreate(f)
	assert.NotNil(t, cmd.Flag("project"))
	assert.NotNil(t, cmd.Flag("description"))
	assert.NotNil(t, cmd.Flag("private"))
}

func TestNewCmdRepoCreate_NotImplemented(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := repo.NewCmdRepoCreate(f)
	cmd.SetArgs([]string{"new-repo", "--project", "MYPROJ"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}
