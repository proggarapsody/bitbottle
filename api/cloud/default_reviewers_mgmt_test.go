package cloud_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/cloud"
)

func newCloudDefaultReviewerServer(t *testing.T, handler http.HandlerFunc) *cloud.Client {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return cloud.NewClient(srv.Client(), srv.URL, "tok", "")
}

const cloudEffectiveDefaultReviewersJSON = `{
  "values": [
    {"user": {"account_id": "uuid-alice", "display_name": "Alice Smith", "nickname": "alice"}},
    {"user": {"account_id": "uuid-bob",   "display_name": "Bob Jones",   "nickname": "bob"}}
  ]
}`

func TestCloudClient_ListDefaultReviewers(t *testing.T) {
	t.Parallel()
	client := newCloudDefaultReviewerServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repositories/myws/my-repo/effective-default-reviewers", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cloudEffectiveDefaultReviewersJSON))
	})
	reviewers, err := client.ListDefaultReviewers("myws", "my-repo")
	require.NoError(t, err)
	require.Len(t, reviewers, 2)
	assert.Equal(t, "alice", reviewers[0].UserSlug)
	assert.Equal(t, "Alice Smith", reviewers[0].DisplayName)
	assert.Equal(t, "bob", reviewers[1].UserSlug)
	assert.Equal(t, "Bob Jones", reviewers[1].DisplayName)
}

func TestCloudClient_AddDefaultReviewer(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	client := newCloudDefaultReviewerServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})
	err := client.AddDefaultReviewer("myws", "my-repo", "uuid-alice")
	require.NoError(t, err)
	assert.Equal(t, http.MethodPut, gotMethod)
	assert.Equal(t, "/repositories/myws/my-repo/default-reviewers/uuid-alice", gotPath)
}

func TestCloudClient_RemoveDefaultReviewer(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	client := newCloudDefaultReviewerServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})
	err := client.RemoveDefaultReviewer("myws", "my-repo", "uuid-alice")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/repositories/myws/my-repo/default-reviewers/uuid-alice", gotPath)
}
