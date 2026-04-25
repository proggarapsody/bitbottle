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

func TestCloudClient_GetPR_PathHasNoPullRequestsHyphen(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		body, _ := os.ReadFile("testdata/pr_get.json")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.GetPR("myworkspace", "my-service", 42)
	require.NoError(t, err)
	assert.Contains(t, gotPath, "/pullrequests/42")
	assert.NotContains(t, gotPath, "/pull-requests/")
}

func TestCloudClient_ListPRs_MapsSourceBranch(t *testing.T) {
	t.Parallel()
	client, _ := cloudFixtureClient(t, "testdata/pr_list.json", 200)
	prs, err := client.ListPRs("myworkspace", "my-service", "OPEN", 25)
	require.NoError(t, err)
	require.Len(t, prs, 1)
	assert.Equal(t, "fix/login", prs[0].FromBranch)
}

func TestCloudClient_ListPRs_MapsAuthorDisplayName(t *testing.T) {
	t.Parallel()
	client, _ := cloudFixtureClient(t, "testdata/pr_list.json", 200)
	prs, err := client.ListPRs("myworkspace", "my-service", "OPEN", 25)
	require.NoError(t, err)
	require.Len(t, prs, 1)
	assert.Equal(t, "Alice", prs[0].Author.DisplayName)
}

func TestCloudClient_ListPRs_MapsAuthorAccountId(t *testing.T) {
	t.Parallel()
	client, _ := cloudFixtureClient(t, "testdata/pr_list.json", 200)
	prs, err := client.ListPRs("myworkspace", "my-service", "OPEN", 25)
	require.NoError(t, err)
	require.Len(t, prs, 1)
	assert.Equal(t, "alice-uuid", prs[0].Author.Slug)
}

func TestCloudClient_ApprovePR_SendsPOST(t *testing.T) {
	t.Parallel()
	var gotMethod string
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{}`)
	})
	err := client.ApprovePR("myworkspace", "my-service", 42)
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
}

func TestCloudClient_DeleteBranch_PathAndMethod(t *testing.T) {
	t.Parallel()
	var gotRawPath, gotMethod string
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		// Use RequestURI which preserves percent-encoding.
		gotRawPath = r.RequestURI
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	})
	err := client.DeleteBranch("myworkspace", "my-service", "feat/my")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Contains(t, gotRawPath, "/refs/branches/feat%2Fmy")
}

func TestCloudClient_GetCurrentUser_PathIsSlashUser(t *testing.T) {
	t.Parallel()
	var gotPath string
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"account_id":"u123","display_name":"User"}`)
	})
	_, err := client.GetCurrentUser()
	require.NoError(t, err)
	assert.Equal(t, "/user", gotPath)
}

func TestCloudClient_CreatePR_DraftFieldSent(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		body, _ := os.ReadFile("testdata/pr_get.json")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(body)
	})
	_, err := client.CreatePR("myworkspace", "my-service", backend.CreatePRInput{
		Title:      "Test PR",
		Draft:      true,
		FromBranch: "feat/x",
		ToBranch:   "main",
	})
	require.NoError(t, err)
	assert.Contains(t, string(gotBody), `"draft":true`)
}

func TestCloudClient_MergePR_MapsSquashStrategy(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		body, _ := os.ReadFile("testdata/pr_get.json")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
	_, err := client.MergePR("myworkspace", "my-service", 42, backend.MergePRInput{Strategy: "squash"})
	require.NoError(t, err)
	assert.Contains(t, string(gotBody), `"merge_strategy":"squash"`)
}

func TestCloudClient_MergePR_EmptyStrategyOmitted(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		body, _ := os.ReadFile("testdata/pr_get.json")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
	_, err := client.MergePR("myworkspace", "my-service", 42, backend.MergePRInput{})
	require.NoError(t, err)
	assert.NotContains(t, string(gotBody), `"merge_strategy"`)
}

func TestCloudClient_CloudNotServerCapabilities(t *testing.T) {
	t.Parallel()
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	// cloud.Client must NOT implement backend.ServerCapabilities
	var iface any = client
	_, ok := iface.(interface {
		GetApplicationProperties() (backend.AppProperties, error)
	})
	assert.False(t, ok, "cloud.Client must not implement ServerCapabilities")
}
