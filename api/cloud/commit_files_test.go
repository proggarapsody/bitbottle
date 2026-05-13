package cloud_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/cloud"
)

const cloudCommitFilesJSON = `{
  "values": [
    {"status": "modified", "lines_added": 5, "lines_removed": 2, "new": {"path": "foo.go"}, "old": {"path": "foo.go"}},
    {"status": "added",    "lines_added": 10, "lines_removed": 0, "new": {"path": "bar.go"}},
    {"status": "removed",  "lines_added": 0, "lines_removed": 3, "old": {"path": "baz.go"}}
  ]
}`

func newCloudCommitFilesClient(t *testing.T, handler http.HandlerFunc) *cloud.Client {
	t.Helper()
	return newCloudDeployKeyServer(t, handler)
}

func TestCloudClient_ListCommitFiles_Path(t *testing.T) {
	t.Parallel()
	var gotPath string
	client := newCloudCommitFilesClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	})
	_, err := client.ListCommitFiles("myws", "my-repo", "abc123")
	require.NoError(t, err)
	// The spec "abc123~1..abc123" must appear URL-path-escaped as one segment.
	assert.Equal(t, "/repositories/myws/my-repo/diffstat/abc123~1..abc123", gotPath)
}

func TestCloudClient_ListCommitFiles_Maps(t *testing.T) {
	t.Parallel()
	client := newCloudCommitFilesClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cloudCommitFilesJSON))
	})
	entries, err := client.ListCommitFiles("myws", "my-repo", "abc123")
	require.NoError(t, err)
	require.Len(t, entries, 3)

	assert.Equal(t, "foo.go", entries[0].Path)
	assert.Equal(t, "modified", entries[0].Status)
	assert.Equal(t, 5, entries[0].Additions)
	assert.Equal(t, 2, entries[0].Deletions)

	assert.Equal(t, "bar.go", entries[1].Path)
	assert.Equal(t, "added", entries[1].Status)

	// "removed" maps to "deleted"
	assert.Equal(t, "baz.go", entries[2].Path)
	assert.Equal(t, "deleted", entries[2].Status)
}
