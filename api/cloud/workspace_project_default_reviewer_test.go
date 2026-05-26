package cloud_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/cloud"
)

const wsProjectDefaultReviewersJSON = `{
  "values": [
    {
      "account_id": "abc123",
      "display_name": "Alice",
      "nickname": "alice",
      "links": {"avatar": {"href": "https://bitbucket.org/account/alice/avatar"}}
    },
    {
      "account_id": "def456",
      "display_name": "Bob",
      "nickname": "bob",
      "links": {"avatar": {"href": "https://bitbucket.org/account/bob/avatar"}}
    }
  ]
}`

func TestCloudClient_ListProjectDefaultReviewers(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(wsProjectDefaultReviewersJSON))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	got, err := client.ListProjectDefaultReviewers("myws", "PROJ", 0)
	require.NoError(t, err)
	assert.Equal(t, "/workspaces/myws/projects/PROJ/default-reviewers", gotPath)
	require.Len(t, got, 2)
	assert.Equal(t, "abc123", got[0].AccountID)
	assert.Equal(t, "Alice", got[0].DisplayName)
	assert.Equal(t, "alice", got[0].Nickname)
	assert.Equal(t, "https://bitbucket.org/account/alice/avatar", got[0].AvatarURL)
	assert.Equal(t, "def456", got[1].AccountID)
}

func TestCloudClient_ListProjectDefaultReviewers_Empty(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	got, err := client.ListProjectDefaultReviewers("myws", "PROJ", 0)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestCloudClient_AddProjectDefaultReviewer(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	var gotBody []byte
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody, _ = json.Marshal(nil)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"account_id":"abc123","display_name":"Alice","nickname":"alice","links":{"avatar":{"href":""}}}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	err := client.AddProjectDefaultReviewer("myws", "PROJ", "abc123")
	require.NoError(t, err)
	assert.Equal(t, "/workspaces/myws/projects/PROJ/default-reviewers/abc123", gotPath)
	assert.Equal(t, http.MethodPut, gotMethod)
	_ = gotBody
}

func TestCloudClient_RemoveProjectDefaultReviewer(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	err := client.RemoveProjectDefaultReviewer("myws", "PROJ", "abc123")
	require.NoError(t, err)
	assert.Equal(t, "/workspaces/myws/projects/PROJ/default-reviewers/abc123", gotPath)
	assert.Equal(t, http.MethodDelete, gotMethod)
}

func TestCloudClient_ListProjectDefaultReviewers_PathEscape(t *testing.T) {
	t.Parallel()
	var gotEscapedPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotEscapedPath = r.URL.EscapedPath()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	_, err := client.ListProjectDefaultReviewers("my ws", "MY PROJ", 0)
	require.NoError(t, err)
	assert.Equal(t, "/workspaces/my%20ws/projects/MY%20PROJ/default-reviewers", gotEscapedPath)
}
