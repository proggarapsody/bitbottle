package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/auth"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// TestAuthLogin_NonTTY_NoToken_Errors verifies that in a non-TTY context
// (e.g. piped stdin) without --with-token and without any stored token,
// login returns an actionable error rather than silently calling the API
// with an empty credential.
func TestAuthLogin_NonTTY_NoToken_Errors(t *testing.T) {
	t.Parallel()

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		// TestFactory defaults: IsStdoutTTY = false, no stored config.
	})
	cmd := auth.NewCmdAuthLogin(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--with-token",
		"error must guide the user to --with-token when running non-interactively")
}

// TestAuthLogin_NonTTY_StoredToken_Revalidates verifies that when a token is
// already stored in the config file, login without --with-token uses the
// stored token to re-validate the credentials (useful for refreshing the
// saved user slug after a server migration, etc.).
func TestAuthLogin_NonTTY_StoredToken_Revalidates(t *testing.T) {
	t.Parallel()

	called := false
	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			called = true
			return testhelpers.BackendUserFactory(), nil
		},
	}

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		BackendOverride: fake,
		// Seed the config with an existing token so the non-TTY path
		// can pick it up without --with-token.
		InitialConfig: "bb.example.com:\n  oauth_token: existing-token\n  git_protocol: ssh\n",
	})
	cmd := auth.NewCmdAuthLogin(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com"})
	require.NoError(t, cmd.Execute())

	assert.True(t, called, "GetCurrentUser must run to validate the stored token")

	cfg, err := f.Config()
	require.NoError(t, err)
	hc, ok := cfg.Get("bb.example.com")
	require.True(t, ok, "host must remain persisted")
	assert.Equal(t, "alice", hc.User)
	assert.Equal(t, "existing-token", hc.OAuthToken)
}
