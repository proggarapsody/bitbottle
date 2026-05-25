package server_test

import (
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// ── PutFile ───────────────────────────────────────────────────────────────────

func TestServerClient_PutFile_OK(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath, gotContent, gotBranch, gotMessage string
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path

		// Parse multipart body.
		ct := r.Header.Get("Content-Type")
		mediaType, params, err := mime.ParseMediaType(ct)
		require.NoError(t, err)
		assert.Equal(t, "multipart/form-data", mediaType)

		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
			val, _ := io.ReadAll(part)
			switch part.FormName() {
			case "content":
				gotContent = string(val)
			case "branch":
				gotBranch = string(val)
			case "message":
				gotMessage = string(val)
			}
		}
		w.WriteHeader(http.StatusOK)
	})
	err := c.PutFile("MYPROJ", "my-svc", "README.md", backend.PutFileInput{
		Content: "# Hello",
		Branch:  "main",
		Message: "Update README",
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPut, gotMethod)
	assert.Equal(t, "/projects/MYPROJ/repos/my-svc/browse/README.md", gotPath)
	assert.Equal(t, "# Hello", gotContent)
	assert.Equal(t, "main", gotBranch)
	assert.Equal(t, "Update README", gotMessage)
}

func TestServerClient_PutFile_WithSourceCommit(t *testing.T) {
	t.Parallel()
	var gotSourceCommit string
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		_, params, _ := mime.ParseMediaType(ct)
		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			require.NoError(t, err)
			val, _ := io.ReadAll(part)
			if part.FormName() == "sourceCommitId" {
				gotSourceCommit = string(val)
			}
		}
		w.WriteHeader(http.StatusOK)
	})
	err := c.PutFile("MYPROJ", "my-svc", "config.yaml", backend.PutFileInput{
		Content:      "key: value",
		Branch:       "main",
		Message:      "Update config",
		SourceCommit: "deadbeef1234",
	})
	require.NoError(t, err)
	assert.Equal(t, "deadbeef1234", gotSourceCommit)
}

func TestServerClient_PutFile_NestedPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})
	err := c.PutFile("MYPROJ", "my-svc", "pkg/cmd/main.go", backend.PutFileInput{
		Content: "package main",
		Branch:  "feat/x",
		Message: "Add main",
	})
	require.NoError(t, err)
	assert.Equal(t, "/projects/MYPROJ/repos/my-svc/browse/pkg/cmd/main.go", gotPath)
}

func TestServerClient_PutFile_EmptyPath_ReturnsDomainError(t *testing.T) {
	t.Parallel()
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	err := c.PutFile("MYPROJ", "my-svc", "", backend.PutFileInput{
		Content: "x",
		Branch:  "main",
		Message: "msg",
	})
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrInvalidRequest, de.Kind)
}

func TestServerClient_PutFile_409_ReturnsDomainError(t *testing.T) {
	t.Parallel()
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	})
	err := c.PutFile("MYPROJ", "my-svc", "foo.txt", backend.PutFileInput{
		Content: "x",
		Branch:  "main",
		Message: "msg",
	})
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrConflict, de.Kind)
}

// ── SourceWriter is satisfied by *server.Client ───────────────────────────────

func TestServerClient_ImplementsSourceWriter(t *testing.T) {
	t.Parallel()
	c, _ := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})
	var _ backend.SourceWriter = c
	_ = strings.HasPrefix // keep strings import
}
