package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/server"
)

const searchCommitsServerJSON = `{"values":[` +
	`{"id":"abc1234def456abc1234def456abc1234def456ab","message":"feat: add search endpoint","author":{"name":"alice","emailAddress":"alice@example.com"},"authorTimestamp":1714118400000},` +
	`{"id":"bbb2234def456bbb2234def456bbb2234def456bb","message":"fix: null pointer","author":{"name":"bob","emailAddress":"bob@example.com"},"authorTimestamp":1714032000000}` +
	`],"isLastPage":true}`

func TestServerClient_SearchCommits_Path(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(searchCommitsServerJSON))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "")

	commits, err := client.SearchCommits("PROJ", "my-repo", backend.CommitSearchOpts{Limit: 10})
	require.NoError(t, err)
	assert.Equal(t, "/projects/PROJ/repos/my-repo/commits", gotPath)
	require.Len(t, commits, 2)
}

func TestServerClient_SearchCommits_AuthorParam(t *testing.T) {
	t.Parallel()
	var gotAuthor string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuthor = r.URL.Query().Get("author")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(searchCommitsServerJSON))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "")

	_, err := client.SearchCommits("PROJ", "my-repo", backend.CommitSearchOpts{
		Author: "alice",
		Limit:  10,
	})
	require.NoError(t, err)
	assert.Equal(t, "alice", gotAuthor)
}

func TestServerClient_SearchCommits_MessageFilterClientSide(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(searchCommitsServerJSON))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "")

	commits, err := client.SearchCommits("PROJ", "my-repo", backend.CommitSearchOpts{
		Query: "search",
		Limit: 10,
	})
	require.NoError(t, err)
	// Only "feat: add search endpoint" contains "search"
	require.Len(t, commits, 1)
	assert.Contains(t, commits[0].Message, "search")
}

func TestServerClient_SearchCommits_LimitApplied(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(searchCommitsServerJSON))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "")

	commits, err := client.SearchCommits("PROJ", "my-repo", backend.CommitSearchOpts{Limit: 1})
	require.NoError(t, err)
	require.Len(t, commits, 1)
}

func TestServerClient_SearchCommits_SinceUntilParams(t *testing.T) {
	t.Parallel()
	var gotSince, gotUntil string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSince = r.URL.Query().Get("since")
		gotUntil = r.URL.Query().Get("until")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[],"isLastPage":true}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "")

	_, err := client.SearchCommits("PROJ", "my-repo", backend.CommitSearchOpts{
		Since: "2026-01-01",
		Until: "2026-06-01",
		Limit: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, "2026-01-01", gotSince)
	assert.Equal(t, "2026-06-01", gotUntil)
}
