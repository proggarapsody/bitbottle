package auth_test

import (
	"bytes"
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
	// Cloud uses --email; Server/DC uses --username — they are separate flags.
	assert.NotNil(t, cmd.Flag("email"), "--email flag must exist for Cloud API token auth")
	assert.NotNil(t, cmd.Flag("username"), "--username flag must exist for Server/DC auth")
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
	cmd.SetArgs([]string{"--hostname", "bb.example.com", "--username", "alice", "--with-token"})
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

func TestAuthLogin_Cloud_WithEmail_StoresAuthUser(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return testhelpers.BackendUserFactory(), nil
		},
	}
	ios := iostreams.Test()
	ios.In = io.NopCloser(strings.NewReader("my-api-token\n"))
	kr := testhelpers.NewFakeKeyring()

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		BackendOverride: fake,
		IOStreams:       ios,
		Keyring:         kr,
	})
	cmd := auth.NewCmdAuthLogin(f)
	cmd.SetArgs([]string{
		"--hostname", "bitbucket.org",
		"--email", "alice@example.com",
		"--with-token",
	})
	require.NoError(t, cmd.Execute())

	cfg, err := f.Config()
	require.NoError(t, err)
	hc, ok := cfg.Get("bitbucket.org")
	require.True(t, ok, "host must be persisted")
	assert.Equal(t, "alice@example.com", hc.AuthUser,
		"email must be stored as AuthUser for Cloud Basic auth")
	assert.Equal(t, "my-api-token", hc.OAuthToken)
}

func TestAuthLogin_Cloud_MissingEmail_Errors(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test() // non-TTY so no interactive prompt
	ios.In = io.NopCloser(strings.NewReader("my-api-token\n"))

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		IOStreams: ios,
	})
	cmd := auth.NewCmdAuthLogin(f)
	cmd.SetArgs([]string{"--hostname", "bitbucket.org", "--with-token"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--email",
		"error must point Cloud users to the --email flag")
}

func TestAuthLogin_Cloud_UsernameFlag_Errors(t *testing.T) {
	t.Parallel()

	ios := iostreams.Test()
	ios.In = io.NopCloser(strings.NewReader("my-api-token\n"))

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		IOStreams: ios,
	})
	cmd := auth.NewCmdAuthLogin(f)
	// --username is for Server/DC; on Cloud it must error and guide user to --email
	cmd.SetArgs([]string{
		"--hostname", "bitbucket.org",
		"--username", "alice",
		"--with-token",
	})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--email",
		"error must tell Cloud users to use --email instead of --username")
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

// ttyInput builds a multi-line stdin string for interactive TTY tests.
// Each element is one line (newline appended automatically).
func ttyInput(lines ...string) io.Reader {
	return strings.NewReader(strings.Join(lines, "\n") + "\n")
}

func TestAuthLogin_TTY_Cloud_PromptsForEmail(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return testhelpers.BackendUserFactory(), nil
		},
	}
	ios := iostreams.TestTTY()
	// Sequence: email → choice "2" (paste) → token
	ios.In = io.NopCloser(ttyInput("alice@example.com", "2", "my-token"))

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		BackendOverride: fake,
		IOStreams:       ios,
	})
	cmd := auth.NewCmdAuthLogin(f)
	cmd.SetArgs([]string{"--hostname", "bitbucket.org"}) // no --email
	require.NoError(t, cmd.Execute())

	outBuf := ios.Out.(*bytes.Buffer)
	assert.Contains(t, outBuf.String(), "Atlassian account email for bitbucket.org",
		"must prompt for email on Cloud when --email not provided")

	cfg, err := f.Config()
	require.NoError(t, err)
	hc, ok := cfg.Get("bitbucket.org")
	require.True(t, ok)
	assert.Equal(t, "alice@example.com", hc.AuthUser)
}

func TestAuthLogin_TTY_OpensBrowser_Server(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return testhelpers.BackendUserFactory(), nil
		},
	}
	ios := iostreams.TestTTY()
	// Sequence: choice "1" → Enter (press-enter gate) → token
	ios.In = io.NopCloser(ttyInput("1", "", "my-token"))
	browser := &testhelpers.FakeBrowserLauncher{}
	kr := testhelpers.NewFakeKeyring()

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		BackendOverride: fake,
		IOStreams:       ios,
		Keyring:         kr,
		Browser:         browser,
	})
	cmd := auth.NewCmdAuthLogin(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com", "--username", "alice"})
	require.NoError(t, cmd.Execute())

	require.Len(t, browser.URLs, 1, "browser must be opened exactly once")
	assert.Contains(t, browser.URLs[0], "bb.example.com")
	assert.Contains(t, browser.URLs[0], "access-tokens")
}

func TestAuthLogin_TTY_OpensBrowser_Cloud(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return testhelpers.BackendUserFactory(), nil
		},
	}
	ios := iostreams.TestTTY()
	// Sequence: choice "1" → Enter (press-enter gate) → token
	// --email is provided via flag so no email prompt is shown.
	ios.In = io.NopCloser(ttyInput("1", "", "my-token"))
	browser := &testhelpers.FakeBrowserLauncher{}
	kr := testhelpers.NewFakeKeyring()

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		BackendOverride: fake,
		IOStreams:       ios,
		Keyring:         kr,
		Browser:         browser,
	})
	cmd := auth.NewCmdAuthLogin(f)
	cmd.SetArgs([]string{"--hostname", "bitbucket.org", "--email", "alice@example.com"})
	require.NoError(t, cmd.Execute())

	require.Len(t, browser.URLs, 1, "browser must be opened exactly once")
	assert.Contains(t, browser.URLs[0], "bitbucket.org")
	assert.Contains(t, browser.URLs[0], "app-passwords")
}

func TestAuthLogin_TTY_ChoiceTwo_NoBrowser(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return testhelpers.BackendUserFactory(), nil
		},
	}
	ios := iostreams.TestTTY()
	// Sequence: choice "2" (paste directly) → token
	ios.In = io.NopCloser(ttyInput("2", "my-token"))
	browser := &testhelpers.FakeBrowserLauncher{}

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		BackendOverride: fake,
		IOStreams:       ios,
		Browser:         browser,
	})
	cmd := auth.NewCmdAuthLogin(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com", "--username", "alice"})
	require.NoError(t, cmd.Execute())

	assert.Empty(t, browser.URLs, "browser must NOT be opened when choice is 2")
}

func TestAuthLogin_TTY_PromptsForUsername(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return testhelpers.BackendUserFactory(), nil
		},
	}
	ios := iostreams.TestTTY()
	// Sequence: username → choice "2" → token
	ios.In = io.NopCloser(ttyInput("alice", "2", "my-token"))

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		BackendOverride: fake,
		IOStreams:       ios,
	})
	cmd := auth.NewCmdAuthLogin(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com"}) // no --username
	require.NoError(t, cmd.Execute())

	outBuf := ios.Out.(*bytes.Buffer)
	assert.Contains(t, outBuf.String(), "Bitbucket username for bb.example.com")
}

func TestAuthLogin_TTY_BrowserError_IsNonFatal(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return testhelpers.BackendUserFactory(), nil
		},
	}
	ios := iostreams.TestTTY()
	// Sequence: choice "1" → Enter → token
	ios.In = io.NopCloser(ttyInput("1", "", "my-token"))
	browser := &testhelpers.FakeBrowserLauncher{Err: errors.New("no display")}

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		BackendOverride: fake,
		IOStreams:       ios,
		Browser:         browser,
	})
	cmd := auth.NewCmdAuthLogin(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com", "--username", "alice"})
	require.NoError(t, cmd.Execute(), "browser failure must not abort login")

	// When browser fails the fallback URL is printed so the user can open it manually.
	outBuf := ios.Out.(*bytes.Buffer)
	assert.Contains(t, outBuf.String(), "access-tokens")
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
	cmd.SetArgs([]string{"--hostname", "bb.example.com", "--username", "alice", "--with-token"})
	err := cmd.Execute()
	require.NoError(t, err, "keyring failures must not fail the login command")

	cfg, cerr := f.Config()
	require.NoError(t, cerr)
	hc, ok := cfg.Get("bb.example.com")
	require.True(t, ok)
	assert.Equal(t, "tok", hc.OAuthToken)
}
