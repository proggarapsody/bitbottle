package cloud_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
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

// TestCloudClient_ListBranches_RespectsTotalLimit verifies that --limit caps
// the total branches returned across paginated Cloud responses. PRD #47.
func TestCloudClient_ListBranches_RespectsTotalLimit(t *testing.T) {
	t.Parallel()
	var count int32
	var srv *httptest.Server
	srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		c := atomic.AddInt32(&count, 1)
		if c == 1 {
			next := srv.URL + "/repositories/ws/repo/refs/branches?page=2"
			body := fmt.Sprintf(`{"pagelen":2,"values":[`+
				`{"name":"a","target":{"hash":"1"}},`+
				`{"name":"b","target":{"hash":"2"}}],"next":"%s"}`, next)
			_, _ = io.WriteString(w, body)
			return
		}
		_, _ = io.WriteString(w, `{"pagelen":2,"values":[`+
			`{"name":"c","target":{"hash":"3"}},`+
			`{"name":"d","target":{"hash":"4"}}],"next":""}`)
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	got, err := client.ListBranches("ws", "repo", 2)
	require.NoError(t, err)
	assert.Len(t, got, 2)
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
