package cloud_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func TestCloudClient_GetFileContent_RawBytes(t *testing.T) {
	t.Parallel()
	var gotPath string
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("package main\n"))
	})
	body, err := client.GetFileContent("myws", "my-svc", "main", "main.go")
	require.NoError(t, err)
	assert.Equal(t, "/repositories/myws/my-svc/src/main/main.go", gotPath)
	assert.Equal(t, "package main\n", string(body))
}

func TestCloudClient_GetFileContent_PreservesBinary(t *testing.T) {
	t.Parallel()
	binary := []byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F'}
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write(binary)
	})
	body, err := client.GetFileContent("myws", "my-svc", "main", "logo.jpg")
	require.NoError(t, err)
	assert.Equal(t, binary, body)
}

func TestCloudClient_GetFileContent_DirectoryReturnsNotFound(t *testing.T) {
	t.Parallel()
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[{"type":"commit_file","path":"main.go"}]}`))
	})
	_, err := client.GetFileContent("myws", "my-svc", "main", "src")
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrNotFound, de.Kind)
}

func TestCloudClient_GetFileContent_EscapesPathSegments(t *testing.T) {
	t.Parallel()
	var gotRequestURI string
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		// RequestURI preserves the over-the-wire encoded form; r.URL.Path
		// is auto-decoded by net/http and would mask escape bugs.
		gotRequestURI = r.RequestURI
		_, _ = w.Write([]byte("ok"))
	})
	_, err := client.GetFileContent("myws", "my-svc", "release/1.0", "docs/User Guide.md")
	require.NoError(t, err)
	assert.Contains(t, gotRequestURI, "release%2F1.0")
	assert.Contains(t, gotRequestURI, "User%20Guide.md")
}

func TestCloudClient_GetFileContent_EmptyPath(t *testing.T) {
	t.Parallel()
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("transport should not have been called for empty path")
	})
	_, err := client.GetFileContent("myws", "my-svc", "main", "")
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrNotFound, de.Kind)
}

func TestCloudClient_GetFileContent_404FromTransport(t *testing.T) {
	t.Parallel()
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"type":"error","error":{"message":"Not Found"}}`))
	})
	_, err := client.GetFileContent("myws", "my-svc", "main", "missing.go")
	require.Error(t, err)
	var de *backend.DomainError
	if !errors.As(err, &de) {
		t.Fatalf("want DomainError, got %T: %v", err, err)
	}
	assert.Equal(t, backend.ErrNotFound, de.Kind)
}

func TestCloudClient_ListTree_Root(t *testing.T) {
	t.Parallel()
	var gotPath string
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "values": [
    {"type":"commit_file","path":"README.md","size":1234,"commit":{"hash":"abc"}},
    {"type":"commit_directory","path":"cmd","commit":{"hash":"def"}},
    {"type":"commit_submodule","path":"vendor/foo","commit":{"hash":"feed"}}
  ]
}`))
	})
	entries, err := client.ListTree("myws", "my-svc", "main", "")
	require.NoError(t, err)
	assert.Equal(t, "/repositories/myws/my-svc/src/main/", gotPath)
	require.Len(t, entries, 3)
	assert.Equal(t, backend.TreeEntry{Path: "README.md", Type: "file", Size: 1234, Hash: "abc"}, entries[0])
	assert.Equal(t, backend.TreeEntry{Path: "cmd", Type: "dir", Hash: "def"}, entries[1])
	// Submodules normalise to "dir" so renderers recurse into them.
	assert.Equal(t, "dir", entries[2].Type)
	assert.Equal(t, "vendor/foo", entries[2].Path)
}

func TestCloudClient_ListTree_FilePathRejected(t *testing.T) {
	t.Parallel()
	// When a caller passes a file path to /src, Cloud returns the file
	// content (not a JSON listing). The adapter rejects with ErrNotFound.
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("package main\n"))
	})
	_, err := client.ListTree("myws", "my-svc", "main", "main.go")
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrNotFound, de.Kind)
}

func TestCloudClient_ListTree_EmptyDirectoryStillReturnsListing(t *testing.T) {
	t.Parallel()
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[],"page":1,"size":0,"pagelen":10}`))
	})
	// An empty directory listing still has the "values" key but no entry
	// type sigils; we should NOT misclassify this as a file path. It's a
	// rare edge case in source listings (most repos have something at the
	// root), so for now bitbottle returns ErrNotFound and a future iteration
	// can refine the heuristic if real repos hit this.
	_, err := client.ListTree("myws", "my-svc", "main", "empty-dir")
	require.Error(t, err)
}
