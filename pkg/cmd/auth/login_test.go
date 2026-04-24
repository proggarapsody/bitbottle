package auth_test

import (
	"testing"

	"github.com/aleksey/bitbottle/pkg/cmd/auth"
	"github.com/aleksey/bitbottle/pkg/cmd/factory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdAuthLogin_HasRequiredFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := auth.NewCmdAuthLogin(f)
	assert.NotNil(t, cmd.Flag("hostname"))
	assert.NotNil(t, cmd.Flag("git-protocol"))
	assert.NotNil(t, cmd.Flag("skip-tls-verify"))
	assert.NotNil(t, cmd.Flag("with-token"))
}

func TestNewCmdAuthLogin_GitProtocolDefault(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := auth.NewCmdAuthLogin(f)
	flag := cmd.Flag("git-protocol")
	require.NotNil(t, flag)
	assert.Equal(t, "ssh", flag.DefValue)
}

func TestNewCmdAuthLogin_NotImplemented(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := auth.NewCmdAuthLogin(f)
	cmd.SetArgs([]string{"--hostname", "bitbucket.example.com"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}

func TestNewCmdAuthLogin_WithToken_NotImplemented(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := auth.NewCmdAuthLogin(f)
	cmd.SetArgs([]string{"--hostname", "bitbucket.example.com", "--with-token"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}
