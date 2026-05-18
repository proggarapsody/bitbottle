package factory_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

// TestFactory_DebugHTTP_LogsExchanges verifies that f.DebugHTTP=true wraps
// the transport in a debug roundtripper that writes → METHOD URL and ← STATUS
// lines to IOStreams.ErrOut for each HTTP exchange.
func TestFactory_DebugHTTP_LogsExchanges(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	configDir := t.TempDir()
	hostsYML := "bb.example.com:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"
	require.NoError(t, os.MkdirAll(filepath.Join(configDir, "bitbottle"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "bitbottle", "hosts.yml"), []byte(hostsYML), 0o600))
	t.Setenv("XDG_CONFIG_HOME", configDir)

	var errBuf strings.Builder
	ios := iostreams.Test()
	ios.ErrOut = &errBuf

	f := factory.New()
	f.IOStreams = ios
	f.DebugHTTP = true

	hc, err := f.HTTPClient("bb.example.com")
	require.NoError(t, err)

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/ping", nil)
	resp, err := hc.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	out := errBuf.String()
	assert.Contains(t, out, "→ GET")
	assert.Contains(t, out, "/ping")
	assert.Contains(t, out, "← 204")
}

// TestFactory_DebugHTTP_Off_NoOutput verifies that debug logging is silent
// when f.DebugHTTP is false (the default).
func TestFactory_DebugHTTP_Off_NoOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	configDir := t.TempDir()
	hostsYML := "bb.example.com:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"
	require.NoError(t, os.MkdirAll(filepath.Join(configDir, "bitbottle"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "bitbottle", "hosts.yml"), []byte(hostsYML), 0o600))
	t.Setenv("XDG_CONFIG_HOME", configDir)

	var errBuf strings.Builder
	ios := iostreams.Test()
	ios.ErrOut = &errBuf

	f := factory.New()
	f.IOStreams = ios
	// f.DebugHTTP intentionally left false

	hc, err := f.HTTPClient("bb.example.com")
	require.NoError(t, err)

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/", nil)
	resp, err := hc.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Empty(t, errBuf.String(), "expected no debug output when DebugHTTP is false")
}
