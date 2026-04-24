package repo_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
)

const repoListConfig = "bb.example.com:\n  oauth_token: tok\n  git_protocol: ssh\n"

// TestRepoList_Integration_PrintsRepoSlugs verifies that `repo list` fetches
// repositories from the API and writes their slugs to stdout.
func TestRepoList_Integration_PrintsRepoSlugs(t *testing.T) {
	t.Parallel()

	fixture, err := os.ReadFile("../../../api/testdata/repo_list.json")
	require.NoError(t, err)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	t.Cleanup(srv.Close)

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: repoListConfig,
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
	})

	cmd := repo.NewCmdRepoList(f)
	require.NoError(t, cmd.Execute())

	output := out.String()
	assert.Contains(t, output, "my-service")
	assert.Contains(t, output, "another-repo")
}

// TestRepoList_Integration_RespectsLimit verifies that --limit is forwarded
// as the `limit` query parameter.
func TestRepoList_Integration_RespectsLimit(t *testing.T) {
	t.Parallel()

	var gotQuery string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"values":[],"isLastPage":true}`)
	}))
	t.Cleanup(srv.Close)

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: repoListConfig,
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
	})

	cmd := repo.NewCmdRepoList(f)
	cmd.SetArgs([]string{"--limit", "5"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, gotQuery, "limit=5")
}

// TestRepoList_Integration_EmptyResultPrintsNothing verifies that an empty
// API response exits cleanly with no stdout output.
func TestRepoList_Integration_EmptyResultPrintsNothing(t *testing.T) {
	t.Parallel()

	fixture, err := os.ReadFile("../../../api/testdata/repo_list_empty.json")
	require.NoError(t, err)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	t.Cleanup(srv.Close)

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: repoListConfig,
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
	})

	cmd := repo.NewCmdRepoList(f)
	require.NoError(t, cmd.Execute())

	assert.Empty(t, out.String())
}

// TestRepoList_Integration_APIErrorSurfaced verifies that an API 401 causes
// the command to return an error, not silently succeed.
func TestRepoList_Integration_APIErrorSurfaced(t *testing.T) {
	t.Parallel()

	fixture, err := os.ReadFile("../../../api/testdata/error_401.json")
	require.NoError(t, err)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write(fixture)
	}))
	t.Cleanup(srv.Close)

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: repoListConfig,
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
	})

	cmd := repo.NewCmdRepoList(f)
	err = cmd.Execute()
	require.Error(t, err)
}

// TestRepoList_Integration_NoConfigError verifies that the command returns an
// error when no hosts are configured and --hostname is not provided.
func TestRepoList_Integration_NoConfigError(t *testing.T) {
	t.Parallel()

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})

	cmd := repo.NewCmdRepoList(f)
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not authenticated")
}

// TestRepoList_Integration_ExplicitHostname verifies that --hostname bypasses
// config lookup and uses the provided hostname, even when a conflicting host is
// configured.
func TestRepoList_Integration_ExplicitHostname(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"values":[],"isLastPage":true}`)
	}))
	t.Cleanup(srv.Close)

	// Seed a different host in config; --hostname should override it.
	var gotHost string
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: "other.example.com:\n  oauth_token: other-tok\n  git_protocol: ssh\n",
		HTTPClient:    srv.Client(),
		BaseURL: func(hostname string) string {
			gotHost = hostname
			return srv.URL
		},
	})

	cmd := repo.NewCmdRepoList(f)
	cmd.SetArgs([]string{"--hostname", "bb.example.com"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "bb.example.com", gotHost)
}

// TestRepoList_Integration_ServerError verifies that a 5xx response causes the
// command to return an error containing the HTTP status code.
func TestRepoList_Integration_ServerError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"errors":[{"message":"internal server error"}]}`)
	}))
	t.Cleanup(srv.Close)

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: repoListConfig,
		HTTPClient:    srv.Client(),
		BaseURL:       func(hostname string) string { return srv.URL },
	})

	cmd := repo.NewCmdRepoList(f)
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}
