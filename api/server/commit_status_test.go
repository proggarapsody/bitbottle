package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/server"
)

const listServerStatusesJSON = `{"values":[{"key":"jenkins-build","state":"SUCCESSFUL","name":"Build #42","description":"All passed","url":"https://jenkins.example.com/42"},{"key":"sonar","state":"INPROGRESS","name":"Sonar","description":"running","url":"https://sonar.example.com/p"}],"isLastPage":true,"size":2}`

func TestServerClient_ListCommitStatuses(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(listServerStatusesJSON))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	const hash = "abc1234def567890"
	statuses, err := client.ListCommitStatuses("MYPROJ", "my-svc", hash)
	require.NoError(t, err)

	// Build statuses live on the build-status API root, not /rest/api/1.0.
	assert.Equal(t, "/rest/build-status/1.0/commits/abc1234def567890", gotPath)
	require.Len(t, statuses, 2)
	assert.Equal(t, "jenkins-build", statuses[0].Key)
	assert.Equal(t, "SUCCESSFUL", statuses[0].State)
	assert.Equal(t, "INPROGRESS", statuses[1].State)
}
