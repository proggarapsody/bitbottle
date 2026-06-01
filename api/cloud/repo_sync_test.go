package cloud_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

func TestCloudClient_SyncRepo_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"behind":3,"commits_merged":3}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	_, err := client.SyncRepo("myworkspace", "my-fork", "main")
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/repositories/myworkspace/my-fork/merge-upstream", gotPath)
}

func TestCloudClient_SyncRepo_SendsBranchInBody(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"behind":1,"commits_merged":1}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	_, err := client.SyncRepo("myworkspace", "my-fork", "develop")
	require.NoError(t, err)
	assert.Equal(t, "develop", gotBody["branch"])
}

func TestCloudClient_SyncRepo_MapsResult(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"behind":5,"commits_merged":5}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	result, err := client.SyncRepo("ws", "fork-repo", "main")
	require.NoError(t, err)
	assert.Equal(t, 5, result.Behind)
	assert.Equal(t, 5, result.CommitsMerged)
}

func TestCloudClient_SyncRepo_AlreadyUpToDate(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"behind":0,"commits_merged":0}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	result, err := client.SyncRepo("ws", "fork-repo", "main")
	require.NoError(t, err)
	assert.Equal(t, 0, result.Behind)
	assert.Equal(t, 0, result.CommitsMerged)
}

func TestCloudClient_SyncRepo_409ReturnsConflict(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"type":"error","error":{"message":"fork history has diverged"}}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	_, err := client.SyncRepo("ws", "fork-repo", "main")
	require.Error(t, err)
	assert.ErrorIs(t, err, backend.ErrConflict)
}

func TestCloudClient_SyncRepo_404ReturnsNotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"type":"error","error":{"message":"Repository not found"}}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	_, err := client.SyncRepo("ws", "no-such-repo", "main")
	require.Error(t, err)
	assert.ErrorIs(t, err, backend.ErrNotFound)
}

func TestCloudClient_SyncRepo_OmitsBranchWhenEmpty(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"behind":0,"commits_merged":0}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	_, err := client.SyncRepo("ws", "fork-repo", "")
	require.NoError(t, err)
	_, hasBranch := gotBody["branch"]
	assert.False(t, hasBranch, "branch field must be omitted when empty")
}
