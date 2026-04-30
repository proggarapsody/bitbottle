package cloud_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

func cloudFixtureClient(t *testing.T, fixturePath string, status int) (*cloud.Client, *httptest.Server) {
	t.Helper()
	body, err := os.ReadFile(fixturePath)
	require.NoError(t, err)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return cloud.NewClient(srv.Client(), srv.URL, "tok", ""), srv
}

func TestCloudClient_GetRepo_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		body, _ := os.ReadFile("testdata/repo_get.json")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.GetRepo("myworkspace", "my-service")
	require.NoError(t, err)
	assert.Equal(t, "/repositories/myworkspace/my-service", gotPath)
}

func TestCloudClient_ListRepos_MapsFullName(t *testing.T) {
	t.Parallel()
	// Cloud ListRepos fetches /repositories/{workspace} — test via GetRepo instead
	// for the full_name split logic.
	client, _ := cloudFixtureClient(t, "testdata/repo_get.json", 200)
	repo, err := client.GetRepo("myworkspace", "my-service")
	require.NoError(t, err)
	assert.Equal(t, "myworkspace", repo.Namespace)
	assert.Equal(t, "my-service", repo.Slug)
}

func TestCloudClient_ListRepos_MapsWebURL(t *testing.T) {
	t.Parallel()
	client, _ := cloudFixtureClient(t, "testdata/repo_get.json", 200)
	repo, err := client.GetRepo("myworkspace", "my-service")
	require.NoError(t, err)
	assert.Equal(t, "https://bitbucket.org/myworkspace/my-service", repo.WebURL)
}

func TestCloudClient_ListRepos_Empty(t *testing.T) {
	t.Parallel()
	client, _ := cloudFixtureClient(t, "testdata/repo_list_empty.json", 200)
	repos, err := client.ListRepos(10)
	require.NoError(t, err)
	assert.Empty(t, repos)
}

func TestCloudClient_ListRepos_UsesWorkspacePath(t *testing.T) {
	t.Parallel()
	var gotRepoPaths []string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/user" {
			_, _ = w.Write([]byte(`{"nickname":"testws","account_id":"abc","display_name":"Test"}`))
			return
		}
		gotRepoPaths = append(gotRepoPaths, r.URL.Path)
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.ListRepos(10)
	require.NoError(t, err)
	require.Len(t, gotRepoPaths, 1)
	assert.Equal(t, "/repositories/testws", gotRepoPaths[0])
}

func TestCloudClient_CreateRepo_SendsScmNotScmId(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		body, _ := os.ReadFile("testdata/repo_get.json")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(body)
	})
	_, err := client.CreateRepo("myworkspace", backend.CreateRepoInput{Name: "new-repo", SCM: "git"})
	require.NoError(t, err)
	bodyStr := string(gotBody)
	assert.Contains(t, bodyStr, `"scm":"git"`)
	assert.NotContains(t, bodyStr, `"scmId"`)
}

func TestCloudClient_CreateRepo_SendsIsPrivate(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		body, _ := os.ReadFile("testdata/repo_get.json")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(body)
	})
	_, err := client.CreateRepo("myworkspace", backend.CreateRepoInput{Name: "new-repo", Public: false})
	require.NoError(t, err)
	assert.Contains(t, string(gotBody), `"is_private":true`)
}

func TestCloudClient_DeleteRepo_204(t *testing.T) {
	t.Parallel()
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	err := client.DeleteRepo("myworkspace", "my-service")
	require.NoError(t, err)
}
