package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/auth"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdAuthLogout_HasHostnameFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := auth.NewCmdAuthLogout(f)
	assert.NotNil(t, cmd.Flag("hostname"))
}

func TestAuthLogout_RemovesCredentials(t *testing.T) {
	t.Parallel()

	kr := testhelpers.NewFakeKeyring()
	require.NoError(t, kr.Set("bitbottle", "alice", "tok"))

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: authConfig,
		Keyring:       kr,
	})
	cmd := auth.NewCmdAuthLogout(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com"})
	require.NoError(t, cmd.Execute())

	cfg, err := f.Config()
	require.NoError(t, err)
	_, ok := cfg.Get("bb.example.com")
	assert.False(t, ok, "host should be removed from config after logout")

	_, kerr := kr.Get("bitbottle", "alice")
	assert.Error(t, kerr, "keyring entry should be removed")
}

func TestAuthLogout_UnknownHost_Errors(t *testing.T) {
	t.Parallel()

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: authConfig,
	})
	cmd := auth.NewCmdAuthLogout(f)
	cmd.SetArgs([]string{"--hostname", "unknown.example.com"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown.example.com")
}
