package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/server"
)

func TestServerClient_CherryPickCommit(t *testing.T) {
	t.Parallel()
	var seenPath string
	var seenBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&seenBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "newabcd1234",
			"message": "Fix bug\n\ncherry-picked from abc123",
			"author": {"name": "alice", "emailAddress": "alice@example.com"},
			"authorTimestamp": 1700000000000
		}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "alice")

	got, err := client.CherryPickCommit("MYPROJ", "myrepo", backend.CherryPickInput{
		SourceHash:   "abc123",
		TargetBranch: "main",
		Message:      "Fix bug",
	})
	require.NoError(t, err)

	assert.Equal(t, "/rest/branch-utils/1.0/projects/MYPROJ/repos/myrepo/cherry-pick", seenPath)
	assert.Equal(t, "abc123", seenBody["sourceCommit"])
	assert.Equal(t, map[string]any{"id": "refs/heads/main"}, seenBody["targetRef"])
	assert.Equal(t, "Fix bug", seenBody["message"])

	assert.Equal(t, "newabcd1234", got.Hash)
	assert.Equal(t, "Fix bug", got.Message) // subject only
	assert.NotEmpty(t, got.WebURL)
}

func TestServerClient_CherryPickCommit_NoMessage(t *testing.T) {
	t.Parallel()
	var seenBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&seenBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id": "xyz789", "message": "original message", "author": {"name": "bob"}, "authorTimestamp": 0}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "bob")

	_, err := client.CherryPickCommit("PROJ", "repo", backend.CherryPickInput{
		SourceHash:   "deadbeef",
		TargetBranch: "develop",
	})
	require.NoError(t, err)
	// message field must be omitted when empty
	_, hasMsg := seenBody["message"]
	assert.False(t, hasMsg, "message should not be sent when empty")
}

func TestServerClient_CherryPickCommit_Error(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"errors":[{"message":"merge conflict"}]}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "")

	_, err := client.CherryPickCommit("P", "r", backend.CherryPickInput{
		SourceHash: "abc", TargetBranch: "main",
	})
	require.Error(t, err)
}
