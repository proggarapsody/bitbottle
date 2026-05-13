package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/server"
)

const prCommitListJSON = `{"values":[{"id":"abc1234def456abc1234def456abc1234def456ab","message":"Fix null pointer\n\nBody text","author":{"name":"Alice","emailAddress":"alice@example.com"},"authorTimestamp":1714118400000}],"isLastPage":true}`

func TestServerClient_ListPRCommits_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(prCommitListJSON))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "")

	commits, err := client.ListPRCommits("PROJ", "my-repo", 42)
	require.NoError(t, err)

	assert.Equal(t, "/projects/PROJ/repos/my-repo/pull-requests/42/commits", gotPath)

	require.Len(t, commits, 1)
	assert.Equal(t, "abc1234def456abc1234def456abc1234def456ab", commits[0].Hash)
	assert.Equal(t, "Fix null pointer", commits[0].Message)
	assert.Equal(t, "Alice", commits[0].Author.Slug)
	assert.Contains(t, commits[0].WebURL, "/projects/PROJ/repos/my-repo/commits/")
}
