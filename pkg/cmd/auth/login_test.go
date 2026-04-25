package auth_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/auth"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
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

func TestAuthLogin_WithToken_StoresCredentials(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return testhelpers.BackendUserFactory(), nil
		},
	}
	ios := iostreams.Test()
	ios.In = io.NopCloser(strings.NewReader("new-token\n"))
	kr := testhelpers.NewFakeKeyring()

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		BackendOverride: fake,
		IOStreams:       ios,
		Keyring:         kr,
	})
	cmd := auth.NewCmdAuthLogin(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com", "--with-token"})
	require.NoError(t, cmd.Execute())

	cfg, err := f.Config()
	require.NoError(t, err)
	hc, ok := cfg.Get("bb.example.com")
	require.True(t, ok, "expected host to be persisted in config")
	assert.Equal(t, "new-token", hc.OAuthToken)
	assert.Equal(t, "alice", hc.User)

	// keyring entry should also be set (best-effort).
	got, err := kr.Get("bitbottle", "alice")
	require.NoError(t, err)
	assert.Equal(t, "new-token", got)
}

func TestAuthLogin_MissingHostname_Errors(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test()
	ios.In = io.NopCloser(strings.NewReader("tok\n"))
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		IOStreams: ios,
	})
	cmd := auth.NewCmdAuthLogin(f)
	cmd.SetArgs([]string{"--with-token"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hostname")
}

func TestAuthLogin_GetCurrentUser_Fails_Errors(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{}, errors.New("401 unauthorized")
		},
	}
	ios := iostreams.Test()
	ios.In = io.NopCloser(strings.NewReader("bad-token\n"))

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		BackendOverride: fake,
		IOStreams:       ios,
	})
	cmd := auth.NewCmdAuthLogin(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com", "--with-token"})
	err := cmd.Execute()
	require.Error(t, err)

	cfg, cerr := f.Config()
	require.NoError(t, cerr)
	assert.Empty(t, cfg.Hosts(), "config must not be saved when validation fails")
}

func TestAuthLogin_KeyringError_IsNonFatal(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return testhelpers.BackendUserFactory(), nil
		},
	}
	ios := iostreams.Test()
	ios.In = io.NopCloser(strings.NewReader("tok\n"))
	kr := testhelpers.NewFakeKeyring()
	kr.SetErr = errors.New("keyring unavailable")

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		BackendOverride: fake,
		IOStreams:       ios,
		Keyring:         kr,
	})
	cmd := auth.NewCmdAuthLogin(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com", "--with-token"})
	err := cmd.Execute()
	require.NoError(t, err, "keyring failures must not fail the login command")

	cfg, cerr := f.Config()
	require.NoError(t, cerr)
	hc, ok := cfg.Get("bb.example.com")
	require.True(t, ok)
	assert.Equal(t, "tok", hc.OAuthToken)
}
