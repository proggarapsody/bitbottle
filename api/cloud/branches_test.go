package cloud_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/cloud"
)

func TestCloudClient_ListBranches_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.ListBranches("myworkspace", "my-service", 10)
	require.NoError(t, err)
	assert.Equal(t, "/repositories/myworkspace/my-service/refs/branches", gotPath)
}

func TestCloudClient_ListBranches_MapsName(t *testing.T) {
	t.Parallel()
	client, _ := cloudFixtureClient(t, "testdata/branch_list.json", 200)
	branches, err := client.ListBranches("myworkspace", "my-service", 25)
	require.NoError(t, err)
	require.Len(t, branches, 2)
	assert.Equal(t, "main", branches[0].Name)
	assert.Equal(t, "feature/login", branches[1].Name)
}

func TestCloudClient_ListBranches_MapsLatestHash(t *testing.T) {
	t.Parallel()
	client, _ := cloudFixtureClient(t, "testdata/branch_list.json", 200)
	branches, err := client.ListBranches("myworkspace", "my-service", 25)
	require.NoError(t, err)
	require.Len(t, branches, 2)
	assert.Equal(t, "abc1234def567890", branches[0].LatestHash)
}

func TestCloudClient_ListBranches_IsDefaultFalse(t *testing.T) {
	t.Parallel()
	// Cloud branch list doesn't carry isDefault; all should be false
	client, _ := cloudFixtureClient(t, "testdata/branch_list.json", 200)
	branches, err := client.ListBranches("myworkspace", "my-service", 25)
	require.NoError(t, err)
	for _, b := range branches {
		assert.False(t, b.IsDefault)
	}
}

func TestCloudClient_ListBranches_Empty(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	branches, err := client.ListBranches("myworkspace", "my-service", 10)
	require.NoError(t, err)
	assert.Empty(t, branches)
}
