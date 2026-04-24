package factory_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// TestBackend_Integration_CloudEndToEnd_ListRepos verifies that a bitbucket.org
// hostname is dispatched to the cloud adapter and that a cloud-shaped JSON
// response is decoded into backend.Repository domain objects with the
// namespace extracted from full_name.
func TestBackend_Integration_CloudEndToEnd_ListRepos(t *testing.T) {
	t.Parallel()

	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"pagelen":10,"values":[{"full_name":"ws/my-svc","slug":"my-svc","name":"my-svc","scm":"git","links":{"html":{"href":"https://bitbucket.org/ws/my-svc"}}}],"page":1,"size":1}`)
	}))
	t.Cleanup(srv.Close)

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		HTTPClient: srv.Client(),
		BaseURL:    func(hostname string) string { return srv.URL },
	})

	client, err := f.Backend("bitbucket.org")
	require.NoError(t, err)

	repos, err := client.ListRepos(10)
	require.NoError(t, err)
	require.Len(t, repos, 1)

	// Cloud path shape: /repositories
	assert.Contains(t, gotPath, "/repositories")
	// full_name "ws/my-svc" must be split into ns + slug
	assert.Equal(t, "ws", repos[0].Namespace)
	assert.Equal(t, "my-svc", repos[0].Slug)
	assert.Equal(t, "git", repos[0].SCM)
}

// TestBackend_Integration_ServerEndToEnd_ListRepos verifies that a custom
// hostname is dispatched to the server adapter and decodes the Data-Center
// paged envelope into backend.Repository.
func TestBackend_Integration_ServerEndToEnd_ListRepos(t *testing.T) {
	t.Parallel()

	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"values":[{"slug":"my-svc","name":"My Service","project":{"key":"PROJ"},"scmId":"git","links":{"self":[{"href":"https://git.example.com/projects/PROJ/repos/my-svc"}]}}],"isLastPage":true}`)
	}))
	t.Cleanup(srv.Close)

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		HTTPClient: srv.Client(),
		BaseURL:    func(hostname string) string { return srv.URL },
	})

	client, err := f.Backend("git.example.com")
	require.NoError(t, err)

	repos, err := client.ListRepos(10)
	require.NoError(t, err)
	require.Len(t, repos, 1)

	// Server path shape: /repos (not /repositories)
	assert.Contains(t, gotPath, "/repos")
	assert.NotContains(t, gotPath, "/repositories")
	assert.Equal(t, "my-svc", repos[0].Slug)
}

// TestBackend_Integration_CloudErrorContract_Translates4xx verifies that a
// Bitbucket Cloud error envelope (`{"type":"error","error":{"message":"..."}}`)
// is decoded into *backend.HTTPError with the correct StatusCode and Message.
func TestBackend_Integration_CloudErrorContract_Translates4xx(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, `{"type":"error","error":{"message":"forbidden by cloud"}}`)
	}))
	t.Cleanup(srv.Close)

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		HTTPClient: srv.Client(),
		BaseURL:    func(hostname string) string { return srv.URL },
	})

	client, err := f.Backend("bitbucket.org")
	require.NoError(t, err)

	_, err = client.ListRepos(10)
	require.Error(t, err)

	var httpErr *backend.HTTPError
	require.True(t, errors.As(err, &httpErr), "cloud adapter must translate 4xx into *backend.HTTPError")
	assert.Equal(t, http.StatusForbidden, httpErr.StatusCode)
	assert.Equal(t, "forbidden by cloud", httpErr.Message)
}

// TestBackend_Integration_ServerErrorContract_Translates4xx verifies that a
// Bitbucket Server/DC error envelope (`{"errors":[{"message":"..."}]}`) is
// decoded into *backend.HTTPError with the correct StatusCode and Message.
func TestBackend_Integration_ServerErrorContract_Translates4xx(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"errors":[{"message":"unauthorized by server"}]}`)
	}))
	t.Cleanup(srv.Close)

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		HTTPClient: srv.Client(),
		BaseURL:    func(hostname string) string { return srv.URL },
	})

	client, err := f.Backend("git.example.com")
	require.NoError(t, err)

	_, err = client.ListRepos(10)
	require.Error(t, err)

	var httpErr *backend.HTTPError
	require.True(t, errors.As(err, &httpErr), "server adapter must translate 4xx into *backend.HTTPError")
	assert.Equal(t, http.StatusUnauthorized, httpErr.StatusCode)
	assert.Equal(t, "unauthorized by server", httpErr.Message)
}

// TestBackend_Integration_ConfigDrivenDispatch_BackendTypeCloudForcesCloud
// verifies that HostConfig.BackendType="cloud" on a non-bitbucket.org host
// causes the factory to dispatch to the cloud adapter. This test uses
// InitialConfig (hosts.yml on disk) as the source of truth rather than the
// TestFactoryOpts.BackendType override, confirming end-to-end config-driven
// routing.
func TestBackend_Integration_ConfigDrivenDispatch_BackendTypeCloudForcesCloud(t *testing.T) {
	t.Parallel()

	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"pagelen":10,"values":[],"page":1,"size":0}`)
	}))
	t.Cleanup(srv.Close)

	// Non-bitbucket.org host configured with backend_type: cloud.
	hostsYML := "git.example.com:\n  oauth_token: tok\n  git_protocol: ssh\n  backend_type: cloud\n"

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: hostsYML,
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
	})

	client, err := f.Backend("git.example.com")
	require.NoError(t, err)

	_, err = client.ListRepos(10)
	require.NoError(t, err)

	// Cloud dispatch proven by path containing /repositories (not /repos).
	assert.Contains(t, gotPath, "/repositories")
}

// TestBackend_Integration_ConfigDrivenDispatch_BackendTypeServerForcesServer
// verifies that HostConfig.BackendType="server" forces server dispatch even
// for bitbucket.org (the default cloud hostname).
func TestBackend_Integration_ConfigDrivenDispatch_BackendTypeServerForcesServer(t *testing.T) {
	t.Parallel()

	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"values":[],"isLastPage":true}`)
	}))
	t.Cleanup(srv.Close)

	// bitbucket.org overridden to server.
	hostsYML := "bitbucket.org:\n  oauth_token: tok\n  git_protocol: ssh\n  backend_type: server\n"

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: hostsYML,
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
	})

	client, err := f.Backend("bitbucket.org")
	require.NoError(t, err)

	_, err = client.ListRepos(10)
	require.NoError(t, err)

	// Server dispatch proven by path ending /repos.
	assert.Contains(t, gotPath, "/repos")
	assert.NotContains(t, gotPath, "/repositories")
}
