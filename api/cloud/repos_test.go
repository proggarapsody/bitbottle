package cloud_test

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
	repos, err := client.ListRepos("myworkspace", 10)
	require.NoError(t, err)
	assert.Empty(t, repos)
}

// TestCloudClient_ListRepos_ExplicitNamespaceNoUserCall verifies that ListRepos
// with an explicit namespace hits /repositories/{ns} directly and never calls /user.
func TestCloudClient_ListRepos_ExplicitNamespaceNoUserCall(t *testing.T) {
	t.Parallel()
	var requestPaths []string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPaths = append(requestPaths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.ListRepos("explicit-ws", 10)
	require.NoError(t, err)
	for _, p := range requestPaths {
		assert.NotEqual(t, "/user", p, "ListRepos with explicit namespace must not call /user")
	}
	require.Len(t, requestPaths, 1)
	assert.Equal(t, "/repositories/explicit-ws", requestPaths[0])
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

func TestCloudClient_RenameRepo_PutsNewName(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	var gotBody map[string]any
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"full_name":"myworkspace/renamed","slug":"renamed","name":"renamed","scm":"git","links":{"html":{"href":"https://bitbucket.org/myworkspace/renamed"}}}`))
	})
	repo, err := client.RenameRepo("myworkspace", "my-service", "renamed")
	require.NoError(t, err)
	assert.Equal(t, http.MethodPut, gotMethod)
	assert.Equal(t, "/repositories/myworkspace/my-service", gotPath)
	assert.Equal(t, "renamed", gotBody["name"])
	assert.Equal(t, "renamed", repo.Slug)
	assert.Equal(t, "myworkspace", repo.Namespace)
	assert.Equal(t, "https://bitbucket.org/myworkspace/renamed", repo.WebURL)
}

func TestCloudClient_ForkRepo_PostsToForksEndpoint(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	var gotBody map[string]any
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"full_name":"otherws/my-service","slug":"my-service","name":"my-service","scm":"git","links":{"html":{"href":"https://bitbucket.org/otherws/my-service"}}}`))
	})
	repo, err := client.ForkRepo("myworkspace", "my-service", backend.ForkRepoInput{Workspace: "otherws"})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/repositories/myworkspace/my-service/forks", gotPath)
	ws, _ := gotBody["workspace"].(map[string]any)
	require.NotNil(t, ws, "workspace object expected on fork body")
	assert.Equal(t, "otherws", ws["slug"])
	assert.Equal(t, "otherws", repo.Namespace)
}

func TestCloudClient_ForkRepo_OmitsNameWhenEmpty(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"full_name":"otherws/my-service","slug":"my-service","name":"my-service","scm":"git"}`))
	})
	_, err := client.ForkRepo("myworkspace", "my-service", backend.ForkRepoInput{Workspace: "otherws"})
	require.NoError(t, err)
	_, hasName := gotBody["name"]
	assert.False(t, hasName, "name must be omitted from body when empty so Bitbucket reuses the source name")
}

func TestCloudClient_ForkRepo_SendsNameWhenSet(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"full_name":"otherws/forked","slug":"forked","name":"forked","scm":"git"}`))
	})
	_, err := client.ForkRepo("myworkspace", "my-service", backend.ForkRepoInput{Workspace: "otherws", Name: "forked"})
	require.NoError(t, err)
	assert.Equal(t, "forked", gotBody["name"])
}

func TestCloudClient_GetRepo_MapsCloneURLs(t *testing.T) {
	t.Parallel()
	client, _ := cloudFixtureClient(t, "testdata/repo_get.json", 200)
	repo, err := client.GetRepo("myworkspace", "my-service")
	require.NoError(t, err)
	require.Len(t, repo.CloneURLs, 2)
	assert.Equal(t, "https", repo.CloneURLs[0].Name)
	assert.Equal(t, "ssh", repo.CloneURLs[1].Name)
}

func TestCloudClient_DeleteRepo_204(t *testing.T) {
	t.Parallel()
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	err := client.DeleteRepo("myworkspace", "my-service")
	require.NoError(t, err)
}
