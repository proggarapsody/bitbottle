package cloud_test

import (
	"encoding/json"
	"io"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func TestCloudClient_CreateBranch_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	var gotMethod string
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"name":"feat/x","target":{"hash":"abc123def456abc123def456abc123def456abc123"}}`)
	})
	_, err := client.CreateBranch("myworkspace", "my-service", backend.CreateBranchInput{
		Name:    "feat/x",
		StartAt: "abc123def456abc123def456abc123def456abc123",
	})
	require.NoError(t, err)
	assert.Equal(t, "/repositories/myworkspace/my-service/refs/branches", gotPath)
	assert.Equal(t, http.MethodPost, gotMethod)
}

func TestCloudClient_CreateBranch_MapsResponse(t *testing.T) {
	t.Parallel()
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"name":"feat/abc","target":{"hash":"deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"}}`)
	})
	branch, err := client.CreateBranch("ws", "repo", backend.CreateBranchInput{
		Name:    "feat/abc",
		StartAt: "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
	})
	require.NoError(t, err)
	assert.Equal(t, "feat/abc", branch.Name)
	assert.Equal(t, "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef", branch.LatestHash)
	assert.False(t, branch.IsDefault)
}

func TestCloudClient_CreateBranch_ResolvesHashWhenBranchName(t *testing.T) {
	t.Parallel()
	// First request: GET to resolve the branch → returns commit hash.
	// Second request: POST to create the branch.
	var requestCount int32
	var gotPostBody map[string]any
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&requestCount, 1)
		w.Header().Set("Content-Type", "application/json")
		switch n {
		case 1:
			// GET /repositories/ws/repo/refs/branches/main
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Contains(t, r.URL.Path, "main")
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, `{"name":"main","target":{"hash":"1111111111111111111111111111111111111111"}}`)
		case 2:
			// POST to create branch
			assert.Equal(t, http.MethodPost, r.Method)
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &gotPostBody)
			w.WriteHeader(http.StatusCreated)
			_, _ = io.WriteString(w, `{"name":"feat/new","target":{"hash":"1111111111111111111111111111111111111111"}}`)
		}
	})
	branch, err := client.CreateBranch("ws", "repo", backend.CreateBranchInput{
		Name:    "feat/new",
		StartAt: "main", // branch name, not a hash → must be resolved
	})
	require.NoError(t, err)
	assert.Equal(t, "feat/new", branch.Name)
	assert.Equal(t, int32(2), atomic.LoadInt32(&requestCount), "should make 2 requests: GET to resolve + POST to create")
	// The POST body should contain the resolved hash.
	target, ok := gotPostBody["target"].(map[string]any)
	require.True(t, ok, "expected target object in POST body")
	assert.Equal(t, "1111111111111111111111111111111111111111", target["hash"])
}

func TestCloudClient_CreateBranch_SkipsResolveWhenAlreadyHash(t *testing.T) {
	t.Parallel()
	// When StartAt is already a 40-char hex string, only one request (POST) should be made.
	var requestCount int32
	hash := "abcdef1234567890abcdef1234567890abcdef12"
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"name":"feat/y","target":{"hash":"`+hash+`"}}`)
	})
	_, err := client.CreateBranch("ws", "repo", backend.CreateBranchInput{
		Name:    "feat/y",
		StartAt: hash,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&requestCount), "should make only 1 request when StartAt is already a hash")
}
