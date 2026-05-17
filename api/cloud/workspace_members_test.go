package cloud_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/cloud"
)

const wsMemberPageJSON = `{
  "values": [
    {
      "user": {"account_id":"alice-id","nickname":"alice","display_name":"Alice Smith"},
      "workspace":{"slug":"acme"}
    },
    {
      "user": {"account_id":"bob-id","nickname":"bob","display_name":"Bob Jones"},
      "workspace":{"slug":"acme"}
    }
  ]
}`

func TestCloudClient_ListWorkspaceMembers_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.ListWorkspaceMembers("acme", 0)
	require.NoError(t, err)
	assert.Equal(t, "/workspaces/acme/members", gotPath)
}

func TestCloudClient_ListWorkspaceMembers_IncludesPagelenWhenLimited(t *testing.T) {
	t.Parallel()
	var gotQuery string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.ListWorkspaceMembers("acme", 50)
	require.NoError(t, err)
	assert.Equal(t, "pagelen=50", gotQuery)
}

func TestCloudClient_ListWorkspaceMembers_DecodesValues(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(wsMemberPageJSON))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	got, err := client.ListWorkspaceMembers("acme", 0)
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, "alice", got[0].User.Slug)
	assert.Equal(t, "Alice Smith", got[0].User.DisplayName)
	assert.Equal(t, "acme", got[0].Workspace)
	assert.Equal(t, "bob", got[1].User.Slug)
}

func TestCloudClient_ListWorkspaceMembers_RejectsEmptyWorkspace(t *testing.T) {
	t.Parallel()
	// No HTTP server: an empty workspace must short-circuit before the call.
	client := cloud.NewClient(http.DefaultClient, "https://example.invalid", "tok", "")
	_, err := client.ListWorkspaceMembers("", 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workspace required")
}

func TestCloudClient_ListWorkspaceMembers_PropagatesAPIError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"type":"error","error":{"message":"Unauthorized"}}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	_, err := client.ListWorkspaceMembers("acme", 0)
	require.Error(t, err)
}
