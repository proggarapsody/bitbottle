package server_test

import (
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
