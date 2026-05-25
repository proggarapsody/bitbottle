package cloud_test

import (
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

// ── PutFile ───────────────────────────────────────────────────────────────────

func TestCloudClient_PutFile_OK(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	fields := map[string]string{}

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path

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
			fields[part.FormName()] = string(val)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	err := client.PutFile("myws", "my-svc", "README.md", backend.PutFileInput{
		Content: "# Hello",
		Branch:  "main",
		Message: "Update README",
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/repositories/myws/my-svc/src", gotPath)
	assert.Equal(t, "# Hello", fields["README.md"])
	assert.Equal(t, "main", fields["branch"])
	assert.Equal(t, "Update README", fields["message"])
}

func TestCloudClient_PutFile_WithSourceCommit(t *testing.T) {
	t.Parallel()
	fields := map[string]string{}
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			fields[part.FormName()] = string(val)
		}
		w.WriteHeader(http.StatusCreated)
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	err := client.PutFile("myws", "my-svc", "config.yaml", backend.PutFileInput{
		Content:      "key: val",
		Branch:       "main",
		Message:      "Update config",
		SourceCommit: "abc123",
	})
	require.NoError(t, err)
	assert.Equal(t, "abc123", fields["parents"])
}

func TestCloudClient_PutFile_EmptyPath_ReturnsDomainError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	err := client.PutFile("myws", "my-svc", "", backend.PutFileInput{
		Content: "x",
		Branch:  "main",
		Message: "msg",
	})
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrInvalidRequest, de.Kind)
}

// ── SourceWriter is satisfied by *cloud.Client ───────────────────────────────

func TestCloudClient_ImplementsSourceWriter(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)
	c := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	var _ backend.SourceWriter = c
}
