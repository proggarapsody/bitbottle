package server_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
)

func TestServerClient_GetFileContent_RawBytes(t *testing.T) {
	t.Parallel()
	var gotPath, gotQuery string
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("package main\n"))
	})
	body, err := client.GetFileContent("KEY", "my-svc", "main", "main.go")
	require.NoError(t, err)
	assert.Equal(t, "/projects/KEY/repos/my-svc/raw/main.go", gotPath)
	assert.Contains(t, gotQuery, "at=main")
	assert.Equal(t, "package main\n", string(body))
}

func TestServerClient_GetFileContent_PreservesBinary(t *testing.T) {
	t.Parallel()
	binary := []byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10}
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		_, _ = w.Write(binary)
	})
	body, err := client.GetFileContent("KEY", "my-svc", "main", "logo.jpg")
	require.NoError(t, err)
	assert.Equal(t, binary, body)
}

func TestServerClient_GetFileContent_EscapesPathSegments(t *testing.T) {
	t.Parallel()
	var gotURI string
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotURI = r.RequestURI
		_, _ = w.Write([]byte("ok"))
	})
	_, err := client.GetFileContent("KEY", "my-svc", "feature/x", "docs/User Guide.md")
	require.NoError(t, err)
	assert.Contains(t, gotURI, "User%20Guide.md")
	assert.Contains(t, gotURI, "at=feature%2Fx")
}

func TestServerClient_GetFileContent_404Maps(t *testing.T) {
	t.Parallel()
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"path not found"}]}`))
	})
	_, err := client.GetFileContent("KEY", "my-svc", "main", "missing.go")
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrNotFound, de.Kind)
}

func TestServerClient_GetFileContent_EmptyPath(t *testing.T) {
	t.Parallel()
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("transport should not be called for empty path")
	})
	_, err := client.GetFileContent("KEY", "my-svc", "main", "")
	require.Error(t, err)
}

func TestServerClient_ListTree_Root(t *testing.T) {
	t.Parallel()
	var gotPath, gotQuery string
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "children": {
    "values": [
      {"path": {"toString": "README.md"}, "type": "FILE", "size": 1234, "contentId": "abc"},
      {"path": {"toString": "cmd"}, "type": "DIRECTORY"},
      {"path": {"toString": "vendor/foo"}, "type": "SUBMODULE", "latestCommit": "feed"}
    ],
    "isLastPage": true
  }
}`))
	})
	entries, err := client.ListTree("KEY", "my-svc", "main", "")
	require.NoError(t, err)
	assert.Equal(t, "/projects/KEY/repos/my-svc/browse", gotPath)
	assert.Contains(t, gotQuery, "at=main")
	require.Len(t, entries, 3)
	assert.Equal(t, backend.TreeEntry{Path: "README.md", Type: "file", Size: 1234, Hash: "abc"}, entries[0])
	assert.Equal(t, backend.TreeEntry{Path: "cmd", Type: "dir"}, entries[1])
	// Submodules normalise to "dir"; latestCommit fills Hash when contentId
	// is absent.
	assert.Equal(t, "dir", entries[2].Type)
	assert.Equal(t, "vendor/foo", entries[2].Path)
	assert.Equal(t, "feed", entries[2].Hash)
}

func TestServerClient_ListTree_NestedPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		// Server reports each child by basename in path.toString — adapter
		// must prepend the parent so callers see full repo-relative paths.
		_, _ = w.Write([]byte(`{
  "children": {
    "values": [
      {"path": {"toString": "main.go"}, "type": "FILE", "size": 100}
    ],
    "isLastPage": true
  }
}`))
	})
	entries, err := client.ListTree("KEY", "my-svc", "main", "cmd/foo")
	require.NoError(t, err)
	assert.Equal(t, "/projects/KEY/repos/my-svc/browse/cmd/foo", gotPath)
	require.Len(t, entries, 1)
	assert.Equal(t, "cmd/foo/main.go", entries[0].Path)
}

func TestServerClient_ListTree_404Maps(t *testing.T) {
	t.Parallel()
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"path not found"}]}`))
	})
	_, err := client.ListTree("KEY", "my-svc", "main", "missing")
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrNotFound, de.Kind)
}
