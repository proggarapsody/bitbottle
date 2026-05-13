package cloud_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/cloud"
)

const listPRCommitsJSON = `{"values":[{"hash":"abc1234def456abc1234def456abc1234def456ab","message":"Fix null pointer in auth\n\nLonger body here","author":{"raw":"Alice <alice@example.com>","user":{"account_id":"123","display_name":"Alice"}},"date":"2026-04-24T10:00:00Z","links":{"html":{"href":"https://bitbucket.org/myworkspace/my-service/commits/abc1234def456abc1234def456abc1234def456ab"}}}]}`

func TestCloudClient_ListPRCommits_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(listPRCommitsJSON))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	commits, err := client.ListPRCommits("myworkspace", "my-service", 42)
	require.NoError(t, err)

	assert.Equal(t, "/repositories/myworkspace/my-service/pullrequests/42/commits", gotPath)
	require.Len(t, commits, 1)
	assert.Equal(t, "abc1234def456abc1234def456abc1234def456ab", commits[0].Hash)
	assert.Equal(t, "Fix null pointer in auth", commits[0].Message)
	assert.Equal(t, "Alice", commits[0].Author.Slug)
	assert.Contains(t, commits[0].WebURL, "commits/")
}
