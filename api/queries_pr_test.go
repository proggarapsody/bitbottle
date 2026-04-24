package api_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aleksey/bitbottle/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func prClient(t *testing.T, fixturePath string, status int) *api.Client {
	t.Helper()
	var body []byte
	if fixturePath != "" {
		var err error
		body, err = os.ReadFile(fixturePath)
		require.NoError(t, err)
	}
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if body != nil {
			_, _ = w.Write(body)
		}
	}))
	t.Cleanup(srv.Close)
	return api.NewClient(srv.Client(), srv.URL, api.AuthConfig{Token: "tok"})
}

func TestListPRs_Open(t *testing.T) {
	t.Parallel()
	var gotState string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotState = r.URL.Query().Get("state")
		body, _ := os.ReadFile("testdata/pr_list.json")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	client := api.NewClient(srv.Client(), srv.URL, api.AuthConfig{Token: "tok"})
	prs, err := client.ListPRs("MYPROJ", "my-service", "OPEN", 25)
	require.NoError(t, err)
	assert.Equal(t, "OPEN", gotState)
	assert.Len(t, prs, 2)
}

func TestListPRs_Merged(t *testing.T) {
	t.Parallel()
	client := prClient(t, "testdata/pr_list_merged.json", 200)
	prs, err := client.ListPRs("MYPROJ", "my-service", "MERGED", 25)
	require.NoError(t, err)
	assert.Equal(t, "MERGED", prs[0].State)
}

func TestListPRs_Empty(t *testing.T) {
	t.Parallel()
	client := prClient(t, "testdata/repo_list_empty.json", 200)
	prs, err := client.ListPRs("MYPROJ", "my-service", "OPEN", 25)
	require.NoError(t, err)
	assert.Empty(t, prs)
}

func TestGetPR_ReturnsDetails(t *testing.T) {
	t.Parallel()
	client := prClient(t, "testdata/pr_get.json", 200)
	pr, err := client.GetPR("MYPROJ", "my-service", 42)
	require.NoError(t, err)
	assert.Equal(t, 42, pr.ID)
	assert.Equal(t, "Fix login bug", pr.Title)
	assert.Equal(t, "alice", pr.Author.User.Slug)
}

func TestGetPR_NotFound(t *testing.T) {
	t.Parallel()
	client := prClient(t, "testdata/error_404.json", 404)
	_, err := client.GetPR("MYPROJ", "my-service", 999)
	require.Error(t, err)
	var httpErr *api.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, 404, httpErr.StatusCode)
}

func TestCreatePR_Success(t *testing.T) {
	t.Parallel()
	client := prClient(t, "testdata/pr_create.json", 201)
	pr, err := client.CreatePR("MYPROJ", "my-service", api.CreatePRInput{
		Title:   "New feature",
		FromRef: api.PRRef{ID: "refs/heads/feat/new", DisplayID: "feat/new"},
		ToRef:   api.PRRef{ID: "refs/heads/main", DisplayID: "main"},
	})
	require.NoError(t, err)
	assert.Equal(t, "New feature", pr.Title)
}

func TestCreatePR_Conflict(t *testing.T) {
	t.Parallel()
	client := prClient(t, "testdata/error_409.json", 409)
	_, err := client.CreatePR("MYPROJ", "my-service", api.CreatePRInput{Title: "dup"})
	require.Error(t, err)
	var httpErr *api.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, 409, httpErr.StatusCode)
}

func newMergePRClient(t *testing.T) (*api.Client, *[]byte) {
	t.Helper()
	var gotBody []byte
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		body, _ := os.ReadFile("testdata/pr_get.json")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return api.NewClient(srv.Client(), srv.URL, api.AuthConfig{Token: "tok"}), &gotBody
}

func TestMergePR_MergeStrategy(t *testing.T) {
	t.Parallel()
	client, gotBody := newMergePRClient(t)
	_, err := client.MergePR("MYPROJ", "my-service", 42, api.MergePRInput{Strategy: "merge-commit"})
	require.NoError(t, err)
	assert.Contains(t, string(*gotBody), "merge-commit")
}

func TestMergePR_SquashStrategy(t *testing.T) {
	t.Parallel()
	client, gotBody := newMergePRClient(t)
	_, err := client.MergePR("MYPROJ", "my-service", 42, api.MergePRInput{Strategy: "squash"})
	require.NoError(t, err)
	assert.Contains(t, string(*gotBody), "squash")
}

func TestMergePR_Conflict(t *testing.T) {
	t.Parallel()
	client := prClient(t, "testdata/error_409.json", 409)
	_, err := client.MergePR("MYPROJ", "my-service", 42, api.MergePRInput{})
	require.Error(t, err)
	var httpErr *api.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, 409, httpErr.StatusCode)
}

func TestApprovePR_Success(t *testing.T) {
	t.Parallel()
	var gotMethod string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"user":{"slug":"alice"},"role":"REVIEWER","approved":true}`)
	}))
	t.Cleanup(srv.Close)
	client := api.NewClient(srv.Client(), srv.URL, api.AuthConfig{Token: "tok"})
	err := client.ApprovePR("MYPROJ", "my-service", 42)
	require.NoError(t, err)
	assert.Equal(t, "PUT", gotMethod)
}

func TestGetPRDiff_ReturnsDiff(t *testing.T) {
	t.Parallel()
	diffBody, err := os.ReadFile("testdata/pr_diff.txt")
	require.NoError(t, err)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write(diffBody)
	}))
	t.Cleanup(srv.Close)
	client := api.NewClient(srv.Client(), srv.URL, api.AuthConfig{Token: "tok"})
	diff, err := client.GetPRDiff("MYPROJ", "my-service", 42)
	require.NoError(t, err)
	assert.Contains(t, diff, "diff --git")
}

func TestGetCurrentUser_ReturnsUser(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"slug":"alice","displayName":"Alice"}`)
	}))
	t.Cleanup(srv.Close)
	client := api.NewClient(srv.Client(), srv.URL, api.AuthConfig{Token: "tok"})
	user, err := client.GetCurrentUser()
	require.NoError(t, err)
	assert.Equal(t, "alice", user.Slug)
}
