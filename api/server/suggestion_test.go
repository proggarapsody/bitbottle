package server_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/server"
)

func TestServerClient_ApplySuggestion_OK(t *testing.T) {
	t.Parallel()
	var paths []string
	var applyBody map[string]any
	callCount := 0
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		callCount++
		if callCount == 1 {
			// First call: GET PR to fetch version.
			_, _ = w.Write([]byte(`{"id":42,"version":3,"title":"my pr","state":"OPEN","fromRef":{"displayId":"feat/x"},"toRef":{"displayId":"main"}}`))
		} else {
			// Second call: POST apply.
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &applyBody)
			_, _ = w.Write([]byte(`{"commitHash":"abc123","commitMessage":"Applied suggestion"}`))
		}
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	result, err := client.ApplySuggestion("MYPROJ", "my-svc", 42, 7, 1)
	require.NoError(t, err)
	require.Len(t, paths, 2)
	assert.Equal(t, "/rest/api/1.0/projects/MYPROJ/repos/my-svc/pull-requests/42", paths[0])
	assert.Equal(t, "/rest/api/1.0/projects/MYPROJ/repos/my-svc/pull-requests/42/comments/7/suggestions/1/apply", paths[1])
	assert.Equal(t, float64(3), applyBody["version"])
	assert.Equal(t, "abc123", result.CommitHash)
	assert.Equal(t, "Applied suggestion", result.CommitMessage)
}

func TestServerClient_ApplySuggestion_Error(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"not found"}]}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	_, err := client.ApplySuggestion("MYPROJ", "my-svc", 42, 7, 1)
	require.Error(t, err)
}

func TestServerClient_GetSuggestionPreview_OK(t *testing.T) {
	t.Parallel()
	var gotPath string
	// Bitbucket Server comment body with a suggestion block.
	commentJSON := "{\"id\":7,\"text\":\"Here is a suggestion:\\n```suggestion\\nfoo := bar\\n```\",\"version\":2}"
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(commentJSON))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	preview, err := client.GetSuggestionPreview("MYPROJ", "my-svc", 42, 7)
	require.NoError(t, err)
	assert.Equal(t, "/rest/api/1.0/projects/MYPROJ/repos/my-svc/pull-requests/42/comments/7", gotPath)
	assert.Contains(t, preview, "foo := bar")
}

func TestServerClient_GetSuggestionPreview_Error(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"message":"comment not found"}]}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	_, err := client.GetSuggestionPreview("MYPROJ", "my-svc", 42, 7)
	require.Error(t, err)
}
