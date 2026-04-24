package factory_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/aleksey/bitbottle/api"
	"github.com/aleksey/bitbottle/pkg/cmd/factory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFactory_Integration_ConfigLoadsFromDiskAndWiresClient verifies that:
// 1. hosts.yml written to disk is accessible via Config().Get(), and
// 2. The factory-built client makes real HTTP requests (the test-token is wired,
//    not the pipeline-token — NewTestFactory always uses "test-token" for isolation).
func TestFactory_Integration_ConfigLoadsFromDiskAndWiresClient(t *testing.T) {
	t.Parallel()

	var gotAuth string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"slug":"alice","displayName":"Alice"}`)
	}))
	t.Cleanup(srv.Close)

	// Write a real hosts.yml that the factory will read.
	configDir := t.TempDir()
	hostsYML := "bb.example.com:\n  oauth_token: pipeline-token\n  git_protocol: ssh\n"
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "hosts.yml"), []byte(hostsYML), 0o600))

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		ConfigDir: configDir,
		HTTPClient: srv.Client(),
		BaseURL: func(hostname string) string { return srv.URL },
	})

	client, err := f.HttpClient("bb.example.com")
	require.NoError(t, err)

	// The test factory wires the config but uses test-token. To test the real
	// token flow we need to read the config and verify it's accessible.
	cfg, err := f.Config()
	require.NoError(t, err)
	hc, ok := cfg.Get("bb.example.com")
	require.True(t, ok)
	assert.Equal(t, "pipeline-token", hc.OAuthToken)

	// The test factory's HttpClient uses "test-token" (see testing.go:118).
	// Verify the client is wired correctly by making a real request.
	_, err = client.GetCurrentUser()
	require.NoError(t, err)
	assert.Equal(t, "Bearer test-token", gotAuth)
}

// TestFactory_Integration_MissingConfigNotAnError verifies that HttpClient
// succeeds even when hosts.yml does not exist (unauthenticated scenario).
func TestFactory_Integration_MissingConfigNotAnError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"slug":"anon","displayName":"Anon"}`)
	}))
	t.Cleanup(srv.Close)

	// Config dir is empty — no hosts.yml.
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		HTTPClient: srv.Client(),
		BaseURL:    func(hostname string) string { return srv.URL },
	})

	// HttpClient must not error when config file is absent.
	client, err := f.HttpClient("bb.example.com")
	require.NoError(t, err)
	assert.NotNil(t, client)
}

// TestFactory_Integration_MultipleClosureCallsShareConfig verifies that calling
// Config() and HttpClient() multiple times does not cause double-load races.
func TestFactory_Integration_MultipleClosureCallsShareConfig(t *testing.T) {
	t.Parallel()

	configDir := t.TempDir()
	hostsYML := "bb.example.com:\n  oauth_token: tok\n  git_protocol: ssh\n"
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "hosts.yml"), []byte(hostsYML), 0o600))

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{ConfigDir: configDir})

	// Call Config twice — should not error.
	cfg1, err := f.Config()
	require.NoError(t, err)
	cfg2, err := f.Config()
	require.NoError(t, err)

	// Both calls return consistent data.
	h1, _ := cfg1.Get("bb.example.com")
	h2, _ := cfg2.Get("bb.example.com")
	assert.Equal(t, h1.OAuthToken, h2.OAuthToken)
}

// TestFactory_Integration_BaseURLNoDoubleSlash verifies that the factory's
// BaseURL function produces a URL that, combined with the path in NewClient,
// never creates double-slash paths.
func TestFactory_Integration_BaseURLNoDoubleSlash(t *testing.T) {
	t.Parallel()

	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"values":[],"isLastPage":true}`)
	}))
	t.Cleanup(srv.Close)

	// BaseURL returns URL with trailing slash — NewClient must strip it.
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		HTTPClient: srv.Client(),
		BaseURL:    func(hostname string) string { return srv.URL + "/" },
	})

	client, err := f.HttpClient("bb.example.com")
	require.NoError(t, err)

	_, err = client.ListRepos(10)
	require.NoError(t, err)
	assert.NotContains(t, gotPath, "//", "double slash in request path from factory-built client")
}

// TestFactory_Integration_HTTPClientUsesProvidedHTTPClient verifies that the
// api.HTTPClient injected via TestFactoryOpts is actually used for requests
// (not the default transport). This ensures test isolation is guaranteed.
func TestFactory_Integration_HTTPClientUsesProvidedHTTPClient(t *testing.T) {
	t.Parallel()

	callCount := 0
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"slug":"u","displayName":"U"}`)
	}))
	t.Cleanup(srv.Close)

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		HTTPClient: srv.Client(),
		BaseURL:    func(hostname string) string { return srv.URL },
	})

	client, err := f.HttpClient("bb.example.com")
	require.NoError(t, err)

	_, err = client.GetCurrentUser()
	require.NoError(t, err)
	assert.Equal(t, 1, callCount, "request should have hit the injected test server")
}

// TestFactory_Integration_noopHTTPClientPreventsRealNetwork verifies that the
// default noopHTTPClient in NewTestFactory returns 404 and prevents accidental
// real network calls when no HTTPClient is provided.
func TestFactory_Integration_noopHTTPClientPreventsRealNetwork(t *testing.T) {
	t.Parallel()

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	client, err := f.HttpClient("bb.example.com")
	require.NoError(t, err)

	// The noop client returns 404 for everything.
	_, err = client.GetCurrentUser()
	require.Error(t, err)
	var httpErr *api.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}
