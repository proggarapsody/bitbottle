package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/auth"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func TestNewCmdAuthStatus_HasHostnameFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := auth.NewCmdAuthStatus(f)
	assert.NotNil(t, cmd.Flag("hostname"))
}

func TestNewCmdAuthStatus_NotImplemented(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := auth.NewCmdAuthStatus(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}
