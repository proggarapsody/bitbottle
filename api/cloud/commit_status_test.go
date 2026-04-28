package cloud_test

import (
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
