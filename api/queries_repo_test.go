package api_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/proggarapsody/bitbottle/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newClientWithFixture(t *testing.T, pathSuffix, fixturePath string, status int) (*api.Client, *httptest.Server) {
	t.Helper()
	body, err := os.ReadFile(fixturePath)
	require.NoError(t, err)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	client := api.NewClient(srv.Client(), srv.URL, api.AuthConfig{Token: "tok"})
	return client, srv
}

func TestListRepos_ReturnsRepos(t *testing.T) {
	t.Parallel()
	client, _ := newClientWithFixture(t, "/repos", "testdata/repo_list.json", 200)
	repos, err := client.ListRepos(25)
	require.NoError(t, err)
	assert.Len(t, repos, 2)
	assert.Equal(t, "my-service", repos[0].Slug)
}

func TestListRepos_Empty(t *testing.T) {
	t.Parallel()
	client, _ := newClientWithFixture(t, "/repos", "testdata/repo_list_empty.json", 200)
	repos, err := client.ListRepos(25)
	require.NoError(t, err)
	assert.Empty(t, repos)
}

func TestListRepos_RespectsLimit(t *testing.T) {
	t.Parallel()
	client, _ := newClientWithFixture(t, "/repos", "testdata/repo_list.json", 200)
	repos, err := client.ListRepos(1)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(repos), 1)
}

func TestGetRepo_ReturnsRepo(t *testing.T) {
	t.Parallel()
	client, _ := newClientWithFixture(t, "/repos/my-service", "testdata/repo_get.json", 200)
	repo, err := client.GetRepo("MYPROJ", "my-service")
	require.NoError(t, err)
	assert.Equal(t, "my-service", repo.Slug)
	assert.Equal(t, "MYPROJ", repo.Project.Key)
}

func TestGetRepo_NotFound(t *testing.T) {
	t.Parallel()
	client, _ := newClientWithFixture(t, "/repos/missing", "testdata/error_404.json", 404)
	_, err := client.GetRepo("MYPROJ", "missing")
	require.Error(t, err)
	var httpErr *api.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}

func TestCreateRepo_Success(t *testing.T) {
	t.Parallel()
	client, _ := newClientWithFixture(t, "/repos", "testdata/repo_create.json", 201)
	repo, err := client.CreateRepo("MYPROJ", api.CreateRepoInput{Name: "new-repo", ScmID: "git"})
	require.NoError(t, err)
	assert.Equal(t, "new-repo", repo.Slug)
}

func TestCreateRepo_Conflict(t *testing.T) {
	t.Parallel()
	client, _ := newClientWithFixture(t, "/repos", "testdata/error_409.json", 409)
	_, err := client.CreateRepo("MYPROJ", api.CreateRepoInput{Name: "my-service"})
	require.Error(t, err)
	var httpErr *api.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, 409, httpErr.StatusCode)
}

func TestDeleteRepo_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	client := api.NewClient(srv.Client(), srv.URL, api.AuthConfig{Token: "tok"})
	err := client.DeleteRepo("MYPROJ", "my-service")
	require.NoError(t, err)
}

func TestDeleteRepo_NotFound(t *testing.T) {
	t.Parallel()
	client, _ := newClientWithFixture(t, "/repos/missing", "testdata/error_404.json", 404)
	err := client.DeleteRepo("MYPROJ", "missing")
	require.Error(t, err)
}

func TestGetApplicationProperties_ReturnsVersion(t *testing.T) {
	t.Parallel()
	client, _ := newClientWithFixture(t, "/application-properties", "testdata/application_properties.json", 200)
	props, err := client.GetApplicationProperties()
	require.NoError(t, err)
	assert.Equal(t, "8.9.1", props.Version)
}
