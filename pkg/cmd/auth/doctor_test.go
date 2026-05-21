package auth_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/auth"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// doctorConfig is a single-host config for doctor tests.
const doctorConfig = "bb.example.com:\n  user: alice\n  git_protocol: ssh\n"

func TestNewCmdDoctor_HasHostnameFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := auth.NewCmdDoctor(f)
	assert.NotNil(t, cmd.Flag("hostname"))
}

func TestDoctor_NoHosts_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := auth.NewCmdDoctor(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestDoctor_UnknownHostFlag_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: doctorConfig})
	cmd := auth.NewCmdDoctor(f)
	cmd.SetArgs([]string{"--hostname", "unknown.example.com"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not logged into")
}

func TestDoctor_NoTokenInKeyring_ReportsFailure(t *testing.T) {
	t.Parallel()

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: doctorConfig})
	// Keyring is empty (FakeKeyring with no entries).

	// Wire a backend stub that returns ErrAuth (host is up, no valid token).
	fakeClient := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{}, &backend.DomainError{
				Kind:    backend.ErrAuth,
				Message: "authentication required",
			}
		},
	}
	factorytest.UseBackend(f, fakeClient)

	cmd := auth.NewCmdDoctor(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com"})

	// The command exits with error (token not stored → check fails).
	err := cmd.Execute()
	require.Error(t, err)

	got := out.String()
	assert.Contains(t, got, "bb.example.com")
	assert.Contains(t, got, "token stored")
	assert.Contains(t, got, "✗")
	// The raw config token value must not appear (authConfig uses "tok").
	// We check a longer substring to avoid false positives from words like "token".
	assert.NotContains(t, got, " tok\n")
	assert.NotContains(t, got, "oauth_token")
}

func TestDoctor_TokenInKeyring_ReportsFormat_BBDC(t *testing.T) {
	t.Parallel()

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: doctorConfig})

	// Pre-load the fake keyring with a BBDC- token.
	kr := testhelpers.NewFakeKeyring()
	require.NoError(t, kr.Set("bitbottle", "alice", "BBDC-secrettoken"))
	f.Keyring = kr

	// Wire a backend that returns ErrAuth so reachability passes but auth fails
	// (to keep the test focused on token format output, not auth outcome).
	fakeClient := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{}, &backend.DomainError{
				Kind:    backend.ErrAuth,
				Message: "authentication required",
			}
		},
	}
	factorytest.UseBackend(f, fakeClient)

	cmd := auth.NewCmdDoctor(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com"})

	_ = cmd.Execute()

	got := out.String()
	assert.Contains(t, got, "token stored")
	assert.Contains(t, got, "Server app-password")
	// Token value must never appear.
	assert.NotContains(t, got, "BBDC-secrettoken")
	assert.NotContains(t, got, "secrettoken")
}

func TestDoctor_TokenInKeyring_ReportsFormat_ATATT(t *testing.T) {
	t.Parallel()

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: doctorConfig})

	kr := testhelpers.NewFakeKeyring()
	require.NoError(t, kr.Set("bitbottle", "alice", "ATATTmytoken"))
	f.Keyring = kr

	fakeClient := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{}, &backend.DomainError{
				Kind:    backend.ErrAuth,
				Message: "authentication required",
			}
		},
	}
	factorytest.UseBackend(f, fakeClient)

	cmd := auth.NewCmdDoctor(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com"})

	_ = cmd.Execute()

	got := out.String()
	assert.Contains(t, got, "Cloud OAuth")
	assert.NotContains(t, got, "ATATTmytoken")
}

func TestDoctor_TokenInKeyring_ReportsFormat_Unknown(t *testing.T) {
	t.Parallel()

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: doctorConfig})

	kr := testhelpers.NewFakeKeyring()
	require.NoError(t, kr.Set("bitbottle", "alice", "someothertoken"))
	f.Keyring = kr

	fakeClient := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{}, &backend.DomainError{
				Kind:    backend.ErrAuth,
				Message: "authentication required",
			}
		},
	}
	factorytest.UseBackend(f, fakeClient)

	cmd := auth.NewCmdDoctor(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com"})

	_ = cmd.Execute()

	got := out.String()
	assert.Contains(t, got, "unknown")
	assert.NotContains(t, got, "someothertoken")
}

func TestDoctor_Reachable_ReportsSuccess(t *testing.T) {
	t.Parallel()

	cfg := "bb.example.com:\n  user: alice\n  git_protocol: ssh\n  backend_type: server\n"
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cfg})

	kr := testhelpers.NewFakeKeyring()
	require.NoError(t, kr.Set("bitbottle", "alice", "BBDC-goodtoken"))
	f.Keyring = kr

	// Backend returns ErrAuth → host is reachable but auth fails.
	fakeClient := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{}, &backend.DomainError{
				Kind:    backend.ErrAuth,
				Message: "authentication required",
			}
		},
	}
	factorytest.UseBackend(f, fakeClient)

	cmd := auth.NewCmdDoctor(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com"})
	_ = cmd.Execute()

	got := out.String()
	// keyring backend line always present
	assert.Contains(t, got, "keyring backend")
	// token format should appear since we put a BBDC- token
	assert.Contains(t, got, "Server app-password")
	// Token raw value must not be in output under any circumstance
	assert.NotContains(t, got, "BBDC-goodtoken")
	assert.NotContains(t, got, "goodtoken")
}

func TestDoctor_AllPassed_NoErrorLine(t *testing.T) {
	t.Parallel()

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: doctorConfig})

	kr := testhelpers.NewFakeKeyring()
	require.NoError(t, kr.Set("bitbottle", "alice", "BBDC-goodtoken"))
	f.Keyring = kr

	// Backend returns success → reachable: yes, authenticated: yes.
	fakeClient := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{Slug: "alice"}, nil
		},
	}
	factorytest.UseBackend(f, fakeClient)

	cmd := auth.NewCmdDoctor(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com"})
	err := cmd.Execute()

	// All checks pass → no error.
	assert.NoError(t, err)

	got := out.String()
	// Token raw value must never appear in output.
	assert.NotContains(t, got, "BBDC-goodtoken")
	assert.NotContains(t, got, "goodtoken")
	// keyring backend line always present.
	assert.Contains(t, got, "keyring backend")
	// Both checks pass.
	assert.Contains(t, got, "reachable: yes")
	assert.Contains(t, got, "authenticated: yes")
}

func TestDoctor_ErrTransport_ReachableNo(t *testing.T) {
	t.Parallel()

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: doctorConfig})

	kr := testhelpers.NewFakeKeyring()
	require.NoError(t, kr.Set("bitbottle", "alice", "BBDC-goodtoken"))
	f.Keyring = kr

	// Backend returns ErrTransport → host is unreachable.
	fakeClient := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{}, &backend.DomainError{
				Kind:    backend.ErrTransport,
				Message: "connection refused",
			}
		},
	}
	factorytest.UseBackend(f, fakeClient)

	cmd := auth.NewCmdDoctor(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com"})
	err := cmd.Execute()

	require.Error(t, err)

	got := out.String()
	assert.Contains(t, got, "reachable")
	// The reachable check should report failure.
	assert.Contains(t, got, "✗")
	assert.NotContains(t, got, "BBDC-goodtoken")
}

func TestDoctor_ErrAuth_AuthenticatedNo(t *testing.T) {
	t.Parallel()

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: doctorConfig})

	kr := testhelpers.NewFakeKeyring()
	require.NoError(t, kr.Set("bitbottle", "alice", "BBDC-badtoken"))
	f.Keyring = kr

	// Backend returns ErrAuth → host reachable, token invalid.
	fakeClient := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{}, &backend.DomainError{
				Kind:    backend.ErrAuth,
				Message: "HTTP 401: authentication required",
			}
		},
	}
	factorytest.UseBackend(f, fakeClient)

	cmd := auth.NewCmdDoctor(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com"})
	err := cmd.Execute()

	require.Error(t, err)

	got := out.String()
	assert.Contains(t, got, "reachable: yes")
	assert.Contains(t, got, "authenticated")
	assert.Contains(t, got, "auth error")
	assert.NotContains(t, got, "BBDC-badtoken")
	assert.NotContains(t, got, "badtoken")
}

func TestDoctor_PlainError_ReachableUnknown(t *testing.T) {
	t.Parallel()

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: doctorConfig})

	kr := testhelpers.NewFakeKeyring()
	require.NoError(t, kr.Set("bitbottle", "alice", "BBDC-goodtoken"))
	f.Keyring = kr

	// Backend returns a plain (non-domain) error.
	plainErr := errors.New("unexpected nil pointer")
	fakeClient := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{}, plainErr
		},
	}
	factorytest.UseBackend(f, fakeClient)

	cmd := auth.NewCmdDoctor(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com"})
	err := cmd.Execute()

	require.Error(t, err)

	got := out.String()
	assert.Contains(t, got, "reachable")
	assert.Contains(t, got, "unknown")
	assert.NotContains(t, got, "BBDC-goodtoken")
}
