package cloud_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/cloud"
)

const listCommitsJSON = `{"values":[{"hash":"abc1234def456abc1234def456abc1234def456ab","message":"Fix null pointer in auth\n\nLonger body here","author":{"raw":"Alice <alice@example.com>","user":{"account_id":"123","display_name":"Alice"}},"date":"2026-04-24T10:00:00Z","links":{"html":{"href":"https://bitbucket.org/myworkspace/my-service/commits/abc1234def456abc1234def456abc1234def456ab"}}}]}`

const getCommitJSON = `{"hash":"abc1234def456abc1234def456abc1234def456ab","message":"Fix null pointer in auth","author":{"raw":"Alice <alice@example.com>","user":{"account_id":"123","display_name":"Alice"}},"date":"2026-04-24T10:00:00Z","links":{"html":{"href":"https://bitbucket.org/myworkspace/my-service/commits/abc1234def456abc1234def456abc1234def456ab"}}}`

func TestCloudClient_ListCommits_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath, gotBranch, gotPagelen string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBranch = r.URL.Query().Get("branch")
		gotPagelen = r.URL.Query().Get("pagelen")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(listCommitsJSON))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	commits, err := client.ListCommits("myworkspace", "my-service", "main", 10)
	require.NoError(t, err)

	assert.Equal(t, "/repositories/myworkspace/my-service/commits", gotPath)
	assert.Equal(t, "main", gotBranch)
	assert.Equal(t, "10", gotPagelen)

	require.Len(t, commits, 1)
	assert.Equal(t, "abc1234def456abc1234def456abc1234def456ab", commits[0].Hash)
	assert.Equal(t, "Fix null pointer in auth", commits[0].Message)
	assert.Equal(t, "Alice", commits[0].Author.Slug)
	assert.Contains(t, commits[0].WebURL, "commits/")
}

func TestCloudClient_GetCommit_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(getCommitJSON))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	const hash = "abc1234def456abc1234def456abc1234def456ab"
	commit, err := client.GetCommit("myworkspace", "my-service", hash)
	require.NoError(t, err)

	assert.Equal(t, "/repositories/myworkspace/my-service/commit/abc1234def456abc1234def456abc1234def456ab", gotPath)
	assert.Equal(t, hash, commit.Hash)
	assert.Equal(t, "Fix null pointer in auth", commit.Message)
	assert.Equal(t, "Alice", commit.Author.Slug)
	assert.Contains(t, commit.WebURL, "commits/")
}
