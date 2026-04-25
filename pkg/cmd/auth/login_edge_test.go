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

// TestAuthLogin_WithoutToken_StillCallsGetCurrentUser verifies the
// non-`--with-token` path: the command does NOT read stdin for a token,
// but it MUST still validate credentials by calling GetCurrentUser.
// The persisted OAuthToken is the empty string in this branch (any token
// would have been picked up from a future interactive prompt), and the
// resolved user slug is what gets recorded.
func TestAuthLogin_WithoutToken_StillCallsGetCurrentUser(t *testing.T) {
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
	})
	cmd := auth.NewCmdAuthLogin(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com"})
	require.NoError(t, cmd.Execute())

	assert.True(t, called,
		"GetCurrentUser must run even without --with-token so credentials are validated")

	cfg, err := f.Config()
	require.NoError(t, err)
	hc, ok := cfg.Get("bb.example.com")
	require.True(t, ok, "host must be persisted")
	assert.Equal(t, "alice", hc.User)
	assert.Equal(t, "", hc.OAuthToken,
		"without --with-token no token is read; OAuthToken stays empty")
}
