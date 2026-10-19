package auth_test

import (
	"testing"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/auth"
)

func TestAuthToken_PrintsToken(t *testing.T) {
	t.Parallel()

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: authConfig})
	cmd := auth.NewCmdAuthToken(f)
	cmd.SetArgs([]string{})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "tok\n", out.String())
}

func TestAuthToken_PrintsKeyringToken(t *testing.T) {
	t.Parallel()

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "bb.example.com:\n  user: alice\n  git_protocol: ssh\n"})
	require.NoError(t, f.Keyring.Set("bitbottle", "alice", "keyring-secret"))
	cmd := auth.NewCmdAuthToken(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com"})

	require.NoError(t, cmd.Execute())
	assert.Equal(t, "keyring-secret\n", out.String())
}

func TestAuthToken_BBTokenOverridesStoredToken(t *testing.T) {
	t.Setenv("BB_TOKEN", "environment-secret")

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: authConfig})
	cmd := auth.NewCmdAuthToken(f)
	cmd.SetArgs([]string{})

	require.NoError(t, cmd.Execute())
	assert.Equal(t, "environment-secret\n", out.String())
}

func TestAuthToken_NoToken_ReturnsError(t *testing.T) {
	t.Parallel()

	const noTokenConfig = "bb.example.com:\n  oauth_token: \"\"\n  user: alice\n  git_protocol: ssh\n"

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: noTokenConfig})
	cmd := auth.NewCmdAuthToken(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no token")
	assert.NotContains(t, err.Error(), "secret")
	assert.Empty(t, out.String())
}

func TestAuthToken_UnknownHost_ReturnsErrorWithoutOutput(t *testing.T) {
	t.Parallel()

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: authConfig})
	cmd := auth.NewCmdAuthToken(f)
	cmd.SetArgs([]string{"--hostname", "unknown.example.com"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not logged into unknown.example.com")
	assert.Empty(t, out.String())
}
