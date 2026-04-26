package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/server"
)

const commitListJSON = `{"values":[{"id":"abc1234def456abc1234def456abc1234def456ab","message":"Fix null pointer\n\nBody text","author":{"name":"Alice","emailAddress":"alice@example.com"},"authorTimestamp":1714118400000}],"isLastPage":true}`

func TestServerClient_ListCommits_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	var gotQuery string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(commitListJSON))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "")

	commits, err := client.ListCommits("PROJ", "my-repo", "main", 10)
	require.NoError(t, err)

	assert.Equal(t, "/projects/PROJ/repos/my-repo/commits", gotPath)
	_ = gotQuery // checked via individual query param assertions below

	// Re-run with a request inspector to check query params individually.
	var capturedR *http.Request
	srv2 := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedR = r
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(commitListJSON))
	}))
	t.Cleanup(srv2.Close)
	client2 := server.NewClient(srv2.Client(), srv2.URL, "tok", "")
	_, err = client2.ListCommits("PROJ", "my-repo", "main", 10)
	require.NoError(t, err)
	assert.Equal(t, "main", capturedR.URL.Query().Get("until"))
	assert.Equal(t, "10", capturedR.URL.Query().Get("limit"))

	// Assert field mapping from first response.
	require.Len(t, commits, 1)
	assert.Equal(t, "abc1234def456abc1234def456abc1234def456ab", commits[0].Hash)
	assert.Equal(t, "Fix null pointer", commits[0].Message)
	assert.Equal(t, "Alice", commits[0].Author.Slug)
	assert.Contains(t, commits[0].WebURL, "/projects/PROJ/repos/my-repo/commits/")
}

func TestServerClient_GetCommit_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"abc1234def456abc1234def456abc1234def456ab","message":"Fix null pointer","author":{"name":"Alice","emailAddress":"alice@example.com"},"authorTimestamp":1714118400000}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "")

	_, err := client.GetCommit("PROJ", "my-repo", "abc1234def456abc1234def456abc1234def456ab")
	require.NoError(t, err)
	assert.Equal(t, "/projects/PROJ/repos/my-repo/commits/abc1234def456abc1234def456abc1234def456ab", gotPath)
}
