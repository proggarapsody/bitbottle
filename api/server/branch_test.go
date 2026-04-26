package server_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func TestServerClient_CreateBranch_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	var gotMethod string
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"id":"refs/heads/feat/x","displayId":"feat/x","isDefault":false,"latestCommit":"abc123"}`)
	})
	_, err := client.CreateBranch("PROJ", "my-repo", backend.CreateBranchInput{
		Name:    "feat/x",
		StartAt: "main",
	})
	require.NoError(t, err)
	assert.Equal(t, "/projects/PROJ/repos/my-repo/branches", gotPath)
	assert.Equal(t, http.MethodPost, gotMethod)
}

func TestServerClient_CreateBranch_MapsResponse(t *testing.T) {
	t.Parallel()
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"id":"refs/heads/feat/new","displayId":"feat/new","isDefault":false,"latestCommit":"deadbeef123"}`)
	})
	branch, err := client.CreateBranch("PROJ", "my-repo", backend.CreateBranchInput{
		Name:    "feat/new",
		StartAt: "main",
	})
	require.NoError(t, err)
	assert.Equal(t, "feat/new", branch.Name)
	assert.Equal(t, "deadbeef123", branch.LatestHash)
	assert.False(t, branch.IsDefault)
}

func TestServerClient_CreateBranch_PassesStartPoint(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"id":"refs/heads/feat/y","displayId":"feat/y","isDefault":false,"latestCommit":"abc"}`)
	})
	_, err := client.CreateBranch("PROJ", "my-repo", backend.CreateBranchInput{
		Name:    "feat/y",
		StartAt: "main",
	})
	require.NoError(t, err)
	assert.Equal(t, "feat/y", gotBody["name"])
	assert.Equal(t, "main", gotBody["startPoint"])
}
