package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/auth"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestAuthMigrate_MovesTokenToKeyring(t *testing.T) {
	t.Parallel()

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: authConfig})
	kr := testhelpers.NewFakeKeyring()
	f.Keyring = kr

	cmd := auth.NewCmdAuthMigrate(f)
	cmd.SetArgs([]string{})
	require.NoError(t, cmd.Execute())

	// Output should indicate migration.
	assert.Contains(t, out.String(), "token migrated to keyring")

	// Token must be in the keyring under the host key.
	stored, err := kr.Get("bitbottle", "bb.example.com")
	require.NoError(t, err)
	assert.Equal(t, "tok", stored)

	// Token must have been zeroed in the in-memory config (and thus stripped
	// from any subsequent Save).
	cfg, err := f.Config()
	require.NoError(t, err)
	hc, ok := cfg.Get("bb.example.com")
	require.True(t, ok)
	assert.Empty(t, hc.OAuthToken, "token must be zeroed from config after migration")
}

func TestAuthMigrate_NoToken_Skips(t *testing.T) {
	t.Parallel()

	const noTokenConfig = "bb.example.com:\n  user: alice\n  git_protocol: ssh\n"
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: noTokenConfig})

	cmd := auth.NewCmdAuthMigrate(f)
	cmd.SetArgs([]string{})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "no config-file token found")
}

func TestAuthMigrate_HostnameFlag_TargetsSingleHost(t *testing.T) {
	t.Parallel()

	const twoHostConfig = "bb.example.com:\n  oauth_token: tok1\n  user: alice\n  git_protocol: ssh\n" +
		"other.example.com:\n  oauth_token: tok2\n  user: bob\n  git_protocol: ssh\n"
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: twoHostConfig})
	kr := testhelpers.NewFakeKeyring()
	f.Keyring = kr

	cmd := auth.NewCmdAuthMigrate(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "bb.example.com: token migrated to keyring")

	// Only the targeted host's token was moved.
	stored, err := kr.Get("bitbottle", "bb.example.com")
	require.NoError(t, err)
	assert.Equal(t, "tok1", stored)

	// other.example.com was NOT touched.
	_, err = kr.Get("bitbottle", "other.example.com")
	require.Error(t, err, "other host must not have been migrated")
}

func TestAuthMigrate_UnknownHostname_ReturnsError(t *testing.T) {
	t.Parallel()

	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: authConfig})
	cmd := auth.NewCmdAuthMigrate(f)
	cmd.SetArgs([]string{"--hostname", "unknown.example.com"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not logged into")
}
