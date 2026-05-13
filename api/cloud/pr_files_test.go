package cloud_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/cloud"
)

const listPRFilesJSON = `{"values":[{"status":"added","lines_added":10,"lines_removed":0,"new":{"path":"foo.go"}},{"status":"modified","lines_added":3,"lines_removed":2,"new":{"path":"bar.go"},"old":{"path":"bar.go"}},{"status":"removed","lines_added":0,"lines_removed":5,"old":{"path":"baz.go"}},{"status":"renamed","lines_added":1,"lines_removed":1,"new":{"path":"newname.go"},"old":{"path":"oldname.go"}}]}`

func TestCloudClient_ListPRFiles_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(listPRFilesJSON))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	files, err := client.ListPRFiles("myworkspace", "my-service", 42)
	require.NoError(t, err)

	assert.Equal(t, "/repositories/myworkspace/my-service/pullrequests/42/diffstat", gotPath)
	require.Len(t, files, 4)
	assert.Equal(t, "foo.go", files[0].Path)
	assert.Equal(t, "added", files[0].Status)
	assert.Equal(t, 10, files[0].Additions)
	assert.Equal(t, 0, files[0].Deletions)

	assert.Equal(t, "bar.go", files[1].Path)
	assert.Equal(t, "modified", files[1].Status)

	// "removed" maps to "deleted"
	assert.Equal(t, "baz.go", files[2].Path)
	assert.Equal(t, "deleted", files[2].Status)

	assert.Equal(t, "newname.go", files[3].Path)
	assert.Equal(t, "renamed", files[3].Status)
}
