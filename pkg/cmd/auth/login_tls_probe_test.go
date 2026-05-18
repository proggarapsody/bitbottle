package auth_test

import (
	"bytes"
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"io"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/tlsprobe"
	"github.com/proggarapsody/bitbottle/pkg/cmd/auth"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// TestAuthLogin_TLSProbe_TrustedCA_SilentlyProceeds: when the OS
// already trusts the host's CA, the probe runs but neither prompts
// nor writes skip_tls_verify=true.
func TestAuthLogin_TLSProbe_TrustedCA_SilentlyProceeds(t *testing.T) {
	t.Parallel()
	ios, out, _ := tlsProbeTestIO("new-token\n", false)

	f, _, _ := factorytest.New(t, factorytest.Opts{})
	factorytest.UseBackend(f, fakeBackend(t))
	f.IOStreams = ios
	f.Keyring = testhelpers.NewFakeKeyring()
	probeCalled := false
	f.TLSProber = func(_ context.Context, host string, _ tlsprobe.Options) (*tlsprobe.Result, error) {
		probeCalled = true
		assert.Equal(t, "bb.example.com:443", host)
		return &tlsprobe.Result{TrustedByOS: true}, nil
	}

	cmd := auth.NewCmdAuthLogin(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com", "--username", "alice", "--with-token"})
	require.NoError(t, cmd.Execute())

	assert.True(t, probeCalled, "probe should run on Server/DC login")
	assert.NotContains(t, out.String(), "Trust this certificate", "no prompt when OS trusts CA")

	cfg, _ := f.Config()
	hc, ok := cfg.Get("bb.example.com")
	require.True(t, ok)
	assert.False(t, hc.SkipTLSVerify, "trusted CA must NOT flip skip_tls_verify")
}

// TestAuthLogin_TLSProbe_SelfSigned_UserConfirms_SetsSkipTLS: on a
// self-signed host, the user is shown the cert fingerprint and can
// type `y` to trust it; that persists as skip_tls_verify=true.
func TestAuthLogin_TLSProbe_SelfSigned_UserConfirms_SetsSkipTLS(t *testing.T) {
	t.Parallel()
	cert := mintTestCert(t, "git.example.com")
	fp := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	// stdin = token line, then "y" to confirm the trust prompt. TTY on.
	ios, out, _ := tlsProbeTestIO("new-token\ny\n", true)

	f, _, _ := factorytest.New(t, factorytest.Opts{})
	factorytest.UseBackend(f, fakeBackend(t))
	f.IOStreams = ios
	f.Keyring = testhelpers.NewFakeKeyring()
	f.TLSProber = func(_ context.Context, _ string, _ tlsprobe.Options) (*tlsprobe.Result, error) {
		return &tlsprobe.Result{TrustedByOS: false, LeafCert: cert, FingerprintSHA256: fp}, nil
	}

	cmd := auth.NewCmdAuthLogin(f)
	cmd.SetArgs([]string{"--hostname", "git.example.com", "--username", "alice", "--with-token"})
	require.NoError(t, cmd.Execute())

	o := out.String()
	assert.Contains(t, o, "Trust this certificate", "expected confirmation prompt")
	assert.Contains(t, o, cert.Subject.CommonName, "expected cert subject CN in prompt")
	assert.Contains(t, o, fp[:16], "expected SHA-256 fingerprint in prompt")

	cfg, _ := f.Config()
	hc, ok := cfg.Get("git.example.com")
	require.True(t, ok)
	assert.True(t, hc.SkipTLSVerify, "confirm should write skip_tls_verify=true to hosts.yml")
}

// TestAuthLogin_TLSProbe_SelfSigned_UserDeclines_ReturnsError: on a
// self-signed host, typing anything other than `y` aborts the login
// with an error mentioning the cert and the manual recovery path.
func TestAuthLogin_TLSProbe_SelfSigned_UserDeclines_ReturnsError(t *testing.T) {
	t.Parallel()
	cert := mintTestCert(t, "git.example.com")
	ios, _, _ := tlsProbeTestIO("new-token\nn\n", true)

	f, _, _ := factorytest.New(t, factorytest.Opts{})
	f.IOStreams = ios
	f.Keyring = testhelpers.NewFakeKeyring()
	f.TLSProber = func(_ context.Context, _ string, _ tlsprobe.Options) (*tlsprobe.Result, error) {
		return &tlsprobe.Result{TrustedByOS: false, LeafCert: cert, FingerprintSHA256: "abc"}, nil
	}

	cmd := auth.NewCmdAuthLogin(f)
	cmd.SetArgs([]string{"--hostname", "git.example.com", "--username", "alice", "--with-token"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "certificate not trusted")

	cfg, _ := f.Config()
	_, ok := cfg.Get("git.example.com")
	assert.False(t, ok, "declined login must not persist any host config")
}

// TestAuthLogin_TLSProbe_NonTTY_ReturnsErrorWithoutPrompt: in a
// non-TTY environment (CI, piped) the probe must NOT prompt — it
// must fail so the operator passes --skip-tls-verify explicitly.
func TestAuthLogin_TLSProbe_NonTTY_ReturnsErrorWithoutPrompt(t *testing.T) {
	t.Parallel()
	cert := mintTestCert(t, "git.example.com")
	ios, out, _ := tlsProbeTestIO("new-token\n", false)

	f, _, _ := factorytest.New(t, factorytest.Opts{})
	f.IOStreams = ios
	f.Keyring = testhelpers.NewFakeKeyring()
	f.TLSProber = func(_ context.Context, _ string, _ tlsprobe.Options) (*tlsprobe.Result, error) {
		return &tlsprobe.Result{TrustedByOS: false, LeafCert: cert, FingerprintSHA256: "abc"}, nil
	}

	cmd := auth.NewCmdAuthLogin(f)
	cmd.SetArgs([]string{"--hostname", "git.example.com", "--username", "alice", "--with-token"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--skip-tls-verify")
	assert.NotContains(t, out.String(), "Trust this certificate", "non-TTY must not show prompt")
}

// TestAuthLogin_TLSProbe_SkippedWhenFlagSet: passing
// --skip-tls-verify explicitly bypasses the probe entirely.
func TestAuthLogin_TLSProbe_SkippedWhenFlagSet(t *testing.T) {
	t.Parallel()
	ios, _, _ := tlsProbeTestIO("new-token\n", false)

	f, _, _ := factorytest.New(t, factorytest.Opts{})
	factorytest.UseBackend(f, fakeBackend(t))
	f.IOStreams = ios
	f.Keyring = testhelpers.NewFakeKeyring()
	probeCalled := false
	f.TLSProber = func(_ context.Context, _ string, _ tlsprobe.Options) (*tlsprobe.Result, error) {
		probeCalled = true
		return nil, nil
	}

	cmd := auth.NewCmdAuthLogin(f)
	cmd.SetArgs([]string{"--hostname", "git.example.com", "--username", "alice", "--with-token", "--skip-tls-verify"})
	require.NoError(t, cmd.Execute())
	assert.False(t, probeCalled, "explicit --skip-tls-verify must skip the probe")
}

// TestAuthLogin_TLSProbe_NotInvokedForCloud: Cloud hosts use
// Atlassian-managed certs; the probe is bypassed unconditionally.
func TestAuthLogin_TLSProbe_NotInvokedForCloud(t *testing.T) {
	t.Parallel()
	ios, _, _ := tlsProbeTestIO("new-token\n", false)

	f, _, _ := factorytest.New(t, factorytest.Opts{BackendType: "cloud"})
	factorytest.UseBackend(f, fakeBackend(t))
	f.IOStreams = ios
	f.Keyring = testhelpers.NewFakeKeyring()
	probeCalled := false
	f.TLSProber = func(_ context.Context, _ string, _ tlsprobe.Options) (*tlsprobe.Result, error) {
		probeCalled = true
		return nil, nil
	}

	cmd := auth.NewCmdAuthLogin(f)
	cmd.SetArgs([]string{"--hostname", "bitbucket.org", "--email", "user@example.com", "--with-token"})
	require.NoError(t, cmd.Execute())
	assert.False(t, probeCalled, "Cloud login must not run the TLS probe")
}

// tlsProbeTestIO builds an IOStreams with the given stdin contents,
// returning the stdout/stderr buffers so the caller can assert on
// rendered output. tty toggles IsStdoutTTY.
func tlsProbeTestIO(stdin string, tty bool) (*iostreams.IOStreams, *bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ios := &iostreams.IOStreams{
		In:          io.NopCloser(strings.NewReader(stdin)),
		Out:         out,
		ErrOut:      errOut,
		IsStdoutTTY: func() bool { return tty },
		IsStderrTTY: func() bool { return tty },
	}
	return ios, out, errOut
}

func fakeBackend(t *testing.T) *testhelpers.FakeClient {
	return &testhelpers.FakeClient{
		T:                t,
		GetCurrentUserFn: func() (backend.User, error) { return testhelpers.BackendUserFactory(), nil },
	}
}

// mintTestCert returns a real *x509.Certificate (not parsed from a
// live server) so we can pass it through f.TLSProber without standing
// up an httptest server in every test case.
func mintTestCert(t *testing.T, cn string) *x509.Certificate {
	t.Helper()
	return &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: cn},
		Issuer:       pkix.Name{CommonName: "Self-Signed Test CA"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		Raw:          []byte("dummy-der-bytes"),
	}
}
