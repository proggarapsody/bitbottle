package cloud_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/cloud"
)

func newCloudDiffServer(t *testing.T, handler http.HandlerFunc) *cloud.Client {
	t.Helper()
	return newCloudDeployKeyServer(t, handler) // reuse the TLS server helper
}

func TestCloudClient_GetDiff(t *testing.T) {
	t.Parallel()
	const diffText = `diff --git a/foo.go b/foo.go
--- a/foo.go
+++ b/foo.go
@@ -1 +1 @@
-old
+new
`
	var gotPath string
	client := newCloudDiffServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(diffText))
	})
	text, err := client.GetDiff("myws", "my-repo", "main", "feature")
	require.NoError(t, err)
	assert.Equal(t, "/repositories/myws/my-repo/diff/main..feature", gotPath)
	assert.Equal(t, diffText, text)
}

func TestCloudClient_GetDiffStat(t *testing.T) {
	t.Parallel()
	const body = `{
  "values": [
    {"status":"modified","lines_added":10,"lines_removed":2,"new":{"path":"api/foo.go"}},
    {"status":"added","lines_added":5,"lines_removed":0,"new":{"path":"api/bar.go"}},
    {"status":"removed","lines_added":0,"lines_removed":2,"old":{"path":"api/old.go"}}
  ]
}`
	var gotPath string
	client := newCloudDiffServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	})
	stat, err := client.GetDiffStat("myws", "my-repo", "main", "feature")
	require.NoError(t, err)
	assert.Equal(t, "/repositories/myws/my-repo/diffstat/main..feature", gotPath)
	assert.Equal(t, 3, stat.FilesChanged)
	assert.Equal(t, 15, stat.Additions)
	assert.Equal(t, 4, stat.Deletions)
	require.Len(t, stat.Files, 3)
	assert.Equal(t, "api/foo.go", stat.Files[0].Path)
	assert.Equal(t, "modified", stat.Files[0].Status)
	assert.Equal(t, 10, stat.Files[0].Additions)
	assert.Equal(t, 2, stat.Files[0].Deletions)
	assert.Equal(t, "api/bar.go", stat.Files[1].Path)
	assert.Equal(t, "added", stat.Files[1].Status)
	// "removed" status should map to "deleted"
	assert.Equal(t, "api/old.go", stat.Files[2].Path)
	assert.Equal(t, "deleted", stat.Files[2].Status)
}
