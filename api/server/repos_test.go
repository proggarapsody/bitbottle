package server_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/server"
)

// fixtureClient returns a server.Client backed by a TLS server that always
// responds with the contents of fixturePath at the given status code.
func fixtureClient(t *testing.T, fixturePath string, status int) *server.Client {
	t.Helper()
	body, err := os.ReadFile(fixturePath)
	require.NoError(t, err)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return server.NewClient(srv.Client(), srv.URL, "tok", "")
}

func TestServerClient_ListRepos_MapsNamespace(t *testing.T) {
	t.Parallel()
	client := fixtureClient(t, "testdata/repo_list.json", 200)
	repos, err := client.ListRepos("", 25)
	require.NoError(t, err)
	require.Len(t, repos, 1)
	assert.Equal(t, "MYPROJ", repos[0].Namespace)
}

func TestServerClient_ListRepos_MapsSCM(t *testing.T) {
	t.Parallel()
	client := fixtureClient(t, "testdata/repo_list.json", 200)
	repos, err := client.ListRepos("", 25)
	require.NoError(t, err)
	require.Len(t, repos, 1)
	assert.Equal(t, "git", repos[0].SCM)
}

func TestServerClient_ListRepos_MapsWebURL(t *testing.T) {
	t.Parallel()
	client := fixtureClient(t, "testdata/repo_list.json", 200)
	repos, err := client.ListRepos("", 25)
	require.NoError(t, err)
	require.Len(t, repos, 1)
	assert.Contains(t, repos[0].WebURL, "browse")
}

func TestServerClient_ListRepos_Empty(t *testing.T) {
	t.Parallel()
	client := fixtureClient(t, "testdata/repo_list_empty.json", 200)
	repos, err := client.ListRepos("", 25)
	require.NoError(t, err)
	assert.Empty(t, repos)
}

func TestServerClient_GetRepo_MapsFields(t *testing.T) {
	t.Parallel()
	client := fixtureClient(t, "testdata/repo_get.json", 200)
	repo, err := client.GetRepo("MYPROJ", "my-service")
	require.NoError(t, err)
	assert.Equal(t, "my-service", repo.Slug)
	assert.Equal(t, "MYPROJ", repo.Namespace)
	assert.Equal(t, "git", repo.SCM)
	assert.Contains(t, repo.WebURL, "browse")
}

func TestServerClient_GetRepo_NotFound(t *testing.T) {
	t.Parallel()
	client := fixtureClient(t, "testdata/error_404.json", 404)
	_, err := client.GetRepo("MYPROJ", "missing")
	require.Error(t, err)
	var httpErr *backend.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}

func TestServerClient_GetRepo_MapsCloneURLs(t *testing.T) {
	t.Parallel()
	client := fixtureClient(t, "testdata/repo_get.json", 200)
	repo, err := client.GetRepo("MYPROJ", "my-service")
	require.NoError(t, err)
	require.Len(t, repo.CloneURLs, 2)
	assert.Equal(t, "http", repo.CloneURLs[0].Name)
	assert.Equal(t, "ssh", repo.CloneURLs[1].Name)
}

func TestServerClient_CreateRepo_SendsScmId(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		body, _ := os.ReadFile("testdata/repo_create.json")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(body)
	})
	_, err := client.CreateRepo("MYPROJ", backend.CreateRepoInput{Name: "new-repo", SCM: "git"})
	require.NoError(t, err)
	assert.Contains(t, string(gotBody), `"scmId":"git"`)
}

func TestServerClient_RenameRepo_PutsNewName(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	var gotBody map[string]any
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"slug":"renamed","name":"renamed","scmId":"git","project":{"key":"MYPROJ"},"links":{"self":[{"href":"https://srv/projects/MYPROJ/repos/renamed/browse"}]}}`))
	})
	repo, err := client.RenameRepo("MYPROJ", "my-service", "renamed")
	require.NoError(t, err)
	assert.Equal(t, http.MethodPut, gotMethod)
	assert.Equal(t, "/projects/MYPROJ/repos/my-service", gotPath)
	assert.Equal(t, "renamed", gotBody["name"])
	assert.Equal(t, "renamed", repo.Slug)
	assert.Equal(t, "MYPROJ", repo.Namespace)
}

func TestServerClient_DeleteRepo_204(t *testing.T) {
	t.Parallel()
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	err := client.DeleteRepo("MYPROJ", "my-service")
	require.NoError(t, err)
}
