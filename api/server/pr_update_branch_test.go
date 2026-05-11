package server_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerClient_UpdatePRBranch_PostsToCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{}`)
	})
	err := client.UpdatePRBranch("MYPROJ", "my-service", 42)
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/projects/MYPROJ/repos/my-service/pull-requests/42/rebase", gotPath)
}

func TestServerClient_UpdatePRBranch_ReturnsErrorOnFailure(t *testing.T) {
	t.Parallel()
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = io.WriteString(w, `{"errors":[{"message":"PR not found"}]}`)
	})
	err := client.UpdatePRBranch("MYPROJ", "my-service", 999)
	require.Error(t, err)
}
