package pr_test

import (
	"testing"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdPRCreate_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRCreate(f)
	assert.NotNil(t, cmd.Flag("title"))
	assert.NotNil(t, cmd.Flag("body"))
	assert.NotNil(t, cmd.Flag("base"))
	assert.NotNil(t, cmd.Flag("draft"))
}

func TestNewCmdPRCreate_NotImplemented(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRCreate(f)
	cmd.SetArgs([]string{"--title", "My PR", "--base", "main"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}
