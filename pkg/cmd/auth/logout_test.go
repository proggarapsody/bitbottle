package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/auth"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func TestNewCmdAuthLogout_HasHostnameFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := auth.NewCmdAuthLogout(f)
	assert.NotNil(t, cmd.Flag("hostname"))
}

func TestNewCmdAuthLogout_NotImplemented(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := auth.NewCmdAuthLogout(f)
	cmd.SetArgs([]string{"--hostname", "bitbucket.example.com"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}
