package cloud_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

const searchCommitsCloudJSON = `{"values":[` +
	`{"hash":"abc1234def456abc1234def456abc1234def456ab","message":"feat: add search endpoint","author":{"raw":"Alice <alice@example.com>","user":{"account_id":"123","display_name":"Alice"}},"date":"2026-04-24T10:00:00Z","links":{"html":{"href":"https://bitbucket.org/ws/repo/commits/abc1234def456abc1234def456abc1234def456ab"}}},` +
	`{"hash":"bbb2234def456bbb2234def456bbb2234def456bb","message":"fix: null pointer","author":{"raw":"Bob <bob@example.com>","user":{"account_id":"456","display_name":"Bob"}},"date":"2026-04-23T10:00:00Z","links":{"html":{"href":"https://bitbucket.org/ws/repo/commits/bbb2234def456bbb2234def456bbb2234def456bb"}}}` +
	`]}`

func TestCloudClient_SearchCommits_NoFilter(t *testing.T) {
	t.Parallel()
	var gotPath, gotQ string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQ = r.URL.Query().Get("q")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(searchCommitsCloudJSON))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	commits, err := client.SearchCommits("myws", "my-repo", backend.CommitSearchOpts{Limit: 10})
	require.NoError(t, err)

	assert.Equal(t, "/repositories/myws/my-repo/commits", gotPath)
	assert.Empty(t, gotQ)
	require.Len(t, commits, 2)
}

func TestCloudClient_SearchCommits_MessageFilter(t *testing.T) {
	t.Parallel()
	var gotQ string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQ = r.URL.Query().Get("q")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(searchCommitsCloudJSON))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	commits, err := client.SearchCommits("myws", "my-repo", backend.CommitSearchOpts{
		Query: "search",
		Limit: 10,
	})
	require.NoError(t, err)
	assert.Contains(t, gotQ, `message~"search"`)
	require.Len(t, commits, 2) // API returns both; message filter is in q param
}

func TestCloudClient_SearchCommits_AuthorFilterClientSide(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(searchCommitsCloudJSON))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	commits, err := client.SearchCommits("myws", "my-repo", backend.CommitSearchOpts{
		Author: "alice",
		Limit:  10,
	})
	require.NoError(t, err)
	require.Len(t, commits, 1)
	assert.Equal(t, "abc1234def456abc1234def456abc1234def456ab", commits[0].Hash)
}

func TestCloudClient_SearchCommits_SinceUntilInQuery(t *testing.T) {
	t.Parallel()
	var gotQ string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQ = r.URL.Query().Get("q")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	_, err := client.SearchCommits("myws", "my-repo", backend.CommitSearchOpts{
		Since: "2026-01-01",
		Until: "2026-06-01",
		Limit: 10,
	})
	require.NoError(t, err)
	assert.Contains(t, gotQ, `date>"2026-01-01"`)
	assert.Contains(t, gotQ, `date<"2026-06-01"`)
	assert.Contains(t, gotQ, " AND ")
}

func TestCloudClient_SearchCommits_LimitCap(t *testing.T) {
	t.Parallel()
	var gotPagelen string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPagelen = r.URL.Query().Get("pagelen")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	_, err := client.SearchCommits("myws", "my-repo", backend.CommitSearchOpts{Limit: 200})
	require.NoError(t, err)
	// pagelen must be capped at 100
	assert.Equal(t, "100", gotPagelen)
}
