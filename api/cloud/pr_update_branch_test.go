package cloud_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloudClient_UpdatePRBranch_PostsToCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(202)
		_, _ = io.WriteString(w, `{}`)
	})
	err := client.UpdatePRBranch("myworkspace", "my-service", 42)
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/repositories/myworkspace/my-service/pullrequests/42/update-branch", gotPath)
}

func TestCloudClient_UpdatePRBranch_ReturnsErrorOnFailure(t *testing.T) {
	t.Parallel()
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		_, _ = io.WriteString(w, `{"type":"error","error":{"message":"PR not found"}}`)
	})
	err := client.UpdatePRBranch("myworkspace", "my-service", 999)
	require.Error(t, err)
}
