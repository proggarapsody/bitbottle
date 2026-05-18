package factory_test

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// TestFactory_SkipTLSOverride_OverridesHostConfigFalse verifies that
// setting f.SkipTLSOverride forces InsecureSkipVerify even when
// hostCfg.SkipTLSVerify is false. This is the override that the global
// -k / --skip-tls-verify root flag relies on. Without this, the error
// hint that promises "-k" cannot recover the user from a self-signed
// CA failure on any subcommand other than `auth login`.
//
// The test exercises factory.New() (the production wiring) rather than
// the factorytest helper, because the bug only existed in the
// production HTTPClient closure.
func TestFactory_SkipTLSOverride_OverridesHostConfigFalse(t *testing.T) {
	// Self-signed httptest TLS server — handshake fails with
	// "x509: certificate signed by unknown authority" against any
	// client that doesn't either trust srv.Certificate() or skip
	// verification.
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	configDir := t.TempDir()
	// hostCfg.SkipTLSVerify is explicitly false here — the test asserts
	// the override path, not the stored config.
	hostsYML := "bb.example.com:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "hosts.yml"), []byte(hostsYML), 0o600))
	t.Setenv("XDG_CONFIG_HOME", configDir)
	// factory.New() reads $XDG_CONFIG_HOME/bitbottle/hosts.yml.
	require.NoError(t, os.MkdirAll(filepath.Join(configDir, "bitbottle"), 0o755))
	require.NoError(t, os.Rename(filepath.Join(configDir, "hosts.yml"), filepath.Join(configDir, "bitbottle", "hosts.yml")))

	f := factory.New()

	ctx := context.Background()

	t.Run("without override → handshake fails", func(t *testing.T) {
		hc, err := f.HTTPClient("bb.example.com")
		require.NoError(t, err)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/", nil)
		resp, err := hc.Do(req)
		if resp != nil {
			_ = resp.Body.Close()
		}
		require.Error(t, err, "expected TLS verification to fail against self-signed server")
		assert.Contains(t, err.Error(), "x509")
	})

	t.Run("with override → handshake succeeds", func(t *testing.T) {
		f.SkipTLSOverride = true
		hc, err := f.HTTPClient("bb.example.com")
		require.NoError(t, err)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/", nil)
		resp, err := hc.Do(req)
		require.NoError(t, err, "expected SkipTLSOverride to bypass cert verification")
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// helper: cast the http.Client transport to inspect TLS config, used to
// catch regressions where InsecureSkipVerify silently flips back to
// false because of a Clone()/mutation bug.
func tlsConfig(hc factory.HTTPClient) *tls.Config {
	c, ok := hc.(*http.Client)
	if !ok {
		return nil
	}
	type unwrappable interface{ Unwrap() http.RoundTripper }
	rt := c.Transport
	for {
		u, ok := rt.(unwrappable)
		if !ok {
			break
		}
		rt = u.Unwrap()
	}
	tr, ok := rt.(*http.Transport)
	if !ok {
		return nil
	}
	return tr.TLSClientConfig
}

var _ = tlsConfig // referenced for future regression tests; keep exported helper available
