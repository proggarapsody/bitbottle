package cloud_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/cloud"
)

const listStatusesJSON = `{"values":[{"key":"build-123","state":"SUCCESSFUL","name":"CI","description":"All good","url":"https://ci.example.com/123"},{"key":"build-124","state":"FAILED","name":"Lint","description":"4 errors","url":"https://ci.example.com/124"}]}`

func TestCloudClient_ListCommitStatuses(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(listStatusesJSON))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	const hash = "abc1234def567890"
	statuses, err := client.ListCommitStatuses("myws", "my-svc", hash)
	require.NoError(t, err)

	assert.Equal(t, "/repositories/myws/my-svc/commit/abc1234def567890/statuses", gotPath)
	require.Len(t, statuses, 2)
	assert.Equal(t, "build-123", statuses[0].Key)
	assert.Equal(t, "SUCCESSFUL", statuses[0].State)
	assert.Equal(t, "FAILED", statuses[1].State)
}

func TestCloudClient_ListCommitStatuses_FollowsNextPage(t *testing.T) {
	t.Parallel()
	page2JSON := `{"values":[{"key":"build-200","state":"SUCCESSFUL","name":"E2E","description":"ok","url":"https://ci.example.com/200"}]}`

	var callCount int
	var srvURL string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/page2" {
			_, _ = w.Write([]byte(page2JSON))
		} else {
			page1 := fmt.Sprintf(`{"values":[{"key":"build-100","state":"FAILED","name":"CI","description":"err","url":"https://ci.example.com/100"}],"next":"%s/page2"}`, srvURL)
			_, _ = w.Write([]byte(page1))
		}
	}))
	srvURL = srv.URL
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	statuses, err := client.ListCommitStatuses("ws", "repo", "deadbeef")
	require.NoError(t, err)
	assert.Equal(t, 2, callCount)
	require.Len(t, statuses, 2)
	assert.Equal(t, "build-100", statuses[0].Key)
	assert.Equal(t, "build-200", statuses[1].Key)
}
