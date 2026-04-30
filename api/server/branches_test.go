package server_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/server"
)

func TestServerClient_ListBranches_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"values":[],"isLastPage":true}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.ListBranches("PROJ", "my-repo", 10)
	require.NoError(t, err)
	assert.Equal(t, "/projects/PROJ/repos/my-repo/branches", gotPath)
}

func TestServerClient_ListBranches_MapsDisplayID(t *testing.T) {
	t.Parallel()
	client := fixtureClient(t, "testdata/branch_list.json", 200)
	branches, err := client.ListBranches("PROJ", "my-repo", 25)
	require.NoError(t, err)
	require.Len(t, branches, 2)
	assert.Equal(t, "main", branches[0].Name)
	assert.Equal(t, "feature/login", branches[1].Name)
}

func TestServerClient_ListBranches_MarksDefault(t *testing.T) {
	t.Parallel()
	client := fixtureClient(t, "testdata/branch_list.json", 200)
	branches, err := client.ListBranches("PROJ", "my-repo", 25)
	require.NoError(t, err)
	require.Len(t, branches, 2)
	assert.True(t, branches[0].IsDefault)
	assert.False(t, branches[1].IsDefault)
}

func TestServerClient_ListBranches_MapsLatestHash(t *testing.T) {
	t.Parallel()
	client := fixtureClient(t, "testdata/branch_list.json", 200)
	branches, err := client.ListBranches("PROJ", "my-repo", 25)
	require.NoError(t, err)
	require.Len(t, branches, 2)
	assert.Equal(t, "abc1234def5", branches[0].LatestHash)
}

// TestServerClient_ListBranches_RespectsTotalLimit asserts that --limit
// caps the total number of branches returned across all pages, not the
// per-page hint. See PRD #47, audit concern 1.
func TestServerClient_ListBranches_RespectsTotalLimit(t *testing.T) {
	t.Parallel()
	page1 := `{"size":2,"isLastPage":false,"nextPageStart":2,"start":0,"values":[` +
		`{"id":"refs/heads/a","displayId":"a","isDefault":false,"latestCommit":"1"},` +
		`{"id":"refs/heads/b","displayId":"b","isDefault":false,"latestCommit":"2"}]}`
	page2 := `{"size":2,"isLastPage":true,"start":2,"values":[` +
		`{"id":"refs/heads/c","displayId":"c","isDefault":false,"latestCommit":"3"},` +
		`{"id":"refs/heads/d","displayId":"d","isDefault":false,"latestCommit":"4"}]}`
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("start") == "2" {
			_, _ = io.WriteString(w, page2)
			return
		}
		_, _ = io.WriteString(w, page1)
	})

	got, err := client.ListBranches("PROJ", "my-repo", 2)
	require.NoError(t, err)
	assert.Len(t, got, 2)
}

func TestServerClient_ListBranches_Empty(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[],"isLastPage":true}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "")
	branches, err := client.ListBranches("PROJ", "my-repo", 10)
	require.NoError(t, err)
	assert.Empty(t, branches)
}
