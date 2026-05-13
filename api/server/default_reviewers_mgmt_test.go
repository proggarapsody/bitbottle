package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/server"
)

func newServerDefaultReviewerMgmtClient(t *testing.T, handler http.HandlerFunc) *server.Client {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return server.NewClient(srv.Client(), srv.URL, "tok", "")
}

const serverDefaultReviewerListJSON = `{
  "values": [
    {"slug": "jsmith", "displayName": "John Smith", "emailAddress": "j@co.com"},
    {"slug": "alice",  "displayName": "Alice",      "emailAddress": "alice@co.com"}
  ],
  "isLastPage": true
}`

func TestServerClient_ListDefaultReviewers_Path(t *testing.T) {
	t.Parallel()
	var gotPath string
	client := newServerDefaultReviewerMgmtClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[],"isLastPage":true}`))
	})
	_, err := client.ListDefaultReviewers("PROJ", "my-repo")
	require.NoError(t, err)
	assert.Equal(t, "/rest/default-reviewers/1.0/projects/PROJ/repos/my-repo/reviewers", gotPath)
}

func TestServerClient_ListDefaultReviewers_Maps(t *testing.T) {
	t.Parallel()
	client := newServerDefaultReviewerMgmtClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(serverDefaultReviewerListJSON))
	})
	reviewers, err := client.ListDefaultReviewers("PROJ", "my-repo")
	require.NoError(t, err)
	require.Len(t, reviewers, 2)
	assert.Equal(t, "jsmith", reviewers[0].UserSlug)
	assert.Equal(t, "John Smith", reviewers[0].DisplayName)
	assert.Equal(t, "j@co.com", reviewers[0].EmailAddress)
	assert.Equal(t, "alice", reviewers[1].UserSlug)
}

func TestServerClient_AddDefaultReviewer(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	client := newServerDefaultReviewerMgmtClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})
	err := client.AddDefaultReviewer("PROJ", "my-repo", "jsmith")
	require.NoError(t, err)
	assert.Equal(t, http.MethodPut, gotMethod)
	assert.Equal(t, "/rest/default-reviewers/1.0/projects/PROJ/repos/my-repo/reviewers/jsmith", gotPath)
}

func TestServerClient_RemoveDefaultReviewer(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	client := newServerDefaultReviewerMgmtClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})
	err := client.RemoveDefaultReviewer("PROJ", "my-repo", "jsmith")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/rest/default-reviewers/1.0/projects/PROJ/repos/my-repo/reviewers/jsmith", gotPath)
}
