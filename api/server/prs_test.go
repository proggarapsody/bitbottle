package server_test

import (
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func TestServerClient_ListPRs_MapsAuthorSlug(t *testing.T) {
	t.Parallel()
	client := fixtureClient(t, "testdata/pr_list.json", 200)
	prs, err := client.ListPRs("MYPROJ", "my-service", "OPEN", 25)
	require.NoError(t, err)
	require.Len(t, prs, 1)
	assert.Equal(t, "alice", prs[0].Author.Slug)
}

func TestServerClient_ListPRs_MapsFromBranch(t *testing.T) {
	t.Parallel()
	client := fixtureClient(t, "testdata/pr_list.json", 200)
	prs, err := client.ListPRs("MYPROJ", "my-service", "OPEN", 25)
	require.NoError(t, err)
	require.Len(t, prs, 1)
	assert.Equal(t, "fix/login", prs[0].FromBranch)
}

func TestServerClient_GetPR_MapsAllFields(t *testing.T) {
	t.Parallel()
	client := fixtureClient(t, "testdata/pr_get.json", 200)
	pr, err := client.GetPR("MYPROJ", "my-service", 42)
	require.NoError(t, err)
	assert.Equal(t, 42, pr.ID)
	assert.Equal(t, "Fix login bug", pr.Title)
	assert.Equal(t, "Fixes auth", pr.Description)
	assert.Equal(t, "OPEN", pr.State)
	assert.Equal(t, "alice", pr.Author.Slug)
	assert.Equal(t, "Alice", pr.Author.DisplayName)
	assert.Equal(t, "fix/login", pr.FromBranch)
	assert.Equal(t, "main", pr.ToBranch)
	assert.Contains(t, pr.WebURL, "pull-requests/42")
}

func TestServerClient_CreatePR_AddsRefsHeadsPrefix(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		body, _ := os.ReadFile("testdata/pr_get.json")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(body)
	})
	_, err := client.CreatePR("MYPROJ", "my-service", backend.CreatePRInput{
		Title:      "Test PR",
		FromBranch: "feat/x",
		ToBranch:   "main",
	})
	require.NoError(t, err)
	assert.Contains(t, string(gotBody), `"id":"refs/heads/feat/x"`)
}

func TestServerClient_CreatePR_NoDoublePrefix(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		body, _ := os.ReadFile("testdata/pr_get.json")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(body)
	})
	_, err := client.CreatePR("MYPROJ", "my-service", backend.CreatePRInput{
		Title:      "Test PR",
		FromBranch: "refs/heads/feat/x",
		ToBranch:   "main",
	})
	require.NoError(t, err)
	bodyStr := string(gotBody)
	// Should appear exactly once, not refs/heads/refs/heads/
	assert.Contains(t, bodyStr, `"id":"refs/heads/feat/x"`)
	assert.NotContains(t, bodyStr, "refs/heads/refs/heads/")
}

func TestServerClient_MergePR_SendsStrategy(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		body, _ := os.ReadFile("testdata/pr_get.json")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
	_, err := client.MergePR("MYPROJ", "my-service", 42, backend.MergePRInput{Strategy: "squash"})
	require.NoError(t, err)
	assert.Contains(t, string(gotBody), `"strategy":"squash"`)
}

func TestServerClient_MergePR_EmptyStrategyOmitted(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		body, _ := os.ReadFile("testdata/pr_get.json")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
	_, err := client.MergePR("MYPROJ", "my-service", 42, backend.MergePRInput{})
	require.NoError(t, err)
	assert.NotContains(t, string(gotBody), `"strategy"`)
}

func TestServerClient_ApprovePR_PostsToApproveEndpoint(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{}`)
	})
	err := client.ApprovePR("MYPROJ", "my-service", 42)
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod, "ApprovePR must POST, not PUT")
	assert.Equal(t, "/projects/MYPROJ/repos/my-service/pull-requests/42/approve", gotPath)
}

func TestServerClient_DeleteBranch_SendsDeleteWithBody(t *testing.T) {
	t.Parallel()
	var gotMethod string
	var gotBody []byte
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusNoContent)
	})
	err := client.DeleteBranch("MYPROJ", "my-service", "feat/old")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Contains(t, string(gotBody), `"name"`)
}

func TestServerClient_GetCurrentUser_MapsSlug(t *testing.T) {
	t.Parallel()
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"slug":"alice","displayName":"Alice"}`)
	})
	user, err := client.GetCurrentUser()
	require.NoError(t, err)
	assert.Equal(t, "alice", user.Slug)
}
