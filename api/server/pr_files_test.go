package server_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/server"
)

const serverPRFilesJSON = `{
  "values": [
    {"type": "ADD",    "path": {"toString": "foo.go"}},
    {"type": "MODIFY", "path": {"toString": "bar.go"}},
    {"type": "DELETE", "path": {"toString": "baz.go"}},
    {"type": "RENAME", "path": {"toString": "qux.go"}},
    {"type": "COPY",   "path": {"toString": "quux.go"}}
  ],
  "isLastPage": true
}`

func newServerPRFilesClient(t *testing.T, handler http.HandlerFunc) *server.Client {
	t.Helper()
	return newServerDeployKeyClient(t, handler)
}

func TestServerClient_ListPRFiles_Path(t *testing.T) {
	t.Parallel()
	var gotPath string
	client := newServerPRFilesClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[],"isLastPage":true}`))
	})
	_, err := client.ListPRFiles("PROJ", "my-repo", 42)
	require.NoError(t, err)
	assert.Equal(t, "/projects/PROJ/repos/my-repo/pull-requests/42/changes", gotPath)
}

func TestServerClient_ListPRFiles_Maps(t *testing.T) {
	t.Parallel()
	client := newServerPRFilesClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(serverPRFilesJSON))
	})
	entries, err := client.ListPRFiles("PROJ", "my-repo", 42)
	require.NoError(t, err)
	require.Len(t, entries, 5)

	assert.Equal(t, "foo.go", entries[0].Path)
	assert.Equal(t, "added", entries[0].Status)

	assert.Equal(t, "bar.go", entries[1].Path)
	assert.Equal(t, "modified", entries[1].Status)

	assert.Equal(t, "baz.go", entries[2].Path)
	assert.Equal(t, "deleted", entries[2].Status)

	assert.Equal(t, "qux.go", entries[3].Path)
	assert.Equal(t, "renamed", entries[3].Status)

	// COPY maps to "added"
	assert.Equal(t, "quux.go", entries[4].Path)
	assert.Equal(t, "added", entries[4].Status)
}
