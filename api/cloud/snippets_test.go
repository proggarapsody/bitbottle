package cloud_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

const snippetListJSON = `{
  "values": [
    {
      "id": "Xqjyp1GV",
      "title": "My snippet",
      "is_private": false,
      "owner": {"nickname": "alice"},
      "created_on": "2026-01-10T08:00:00+00:00",
      "files": {"hello.go": {}, "readme.txt": {}},
      "links": {"html": {"href": "https://bitbucket.org/snippets/alice/Xqjyp1GV"}}
    },
    {
      "id": "zAbc123",
      "title": "Secret snippet",
      "is_private": true,
      "owner": {"nickname": "bob"},
      "created_on": "2026-02-01T12:00:00+00:00",
      "files": {"data.json": {}},
      "links": {"html": {"href": "https://bitbucket.org/snippets/bob/zAbc123"}}
    }
  ]
}`

const snippetGetJSON = `{
  "id": "Xqjyp1GV",
  "title": "My snippet",
  "is_private": false,
  "owner": {"nickname": "alice"},
  "created_on": "2026-01-10T08:00:00+00:00",
  "files": {"hello.go": {}},
  "links": {"html": {"href": "https://bitbucket.org/snippets/alice/Xqjyp1GV"}}
}`

func newCloudSnippetServer(t *testing.T, handler http.HandlerFunc) (*cloud.Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return cloud.NewClient(srv.Client(), srv.URL, "tok", ""), srv
}

func TestCloudClient_ListSnippets_PathAndQuery(t *testing.T) {
	t.Parallel()
	var gotPath, gotQuery string
	client, _ := newCloudSnippetServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	})
	_, err := client.ListSnippets("myworkspace", 20)
	require.NoError(t, err)
	assert.Equal(t, "/snippets/myworkspace", gotPath)
	assert.Contains(t, gotQuery, "pagelen=20")
}

func TestCloudClient_ListSnippets_OmitsPagelenWhenZero(t *testing.T) {
	t.Parallel()
	var gotQuery string
	client, _ := newCloudSnippetServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	})
	_, err := client.ListSnippets("myworkspace", 0)
	require.NoError(t, err)
	assert.NotContains(t, gotQuery, "pagelen", "limit=0 must omit pagelen")
}

func TestCloudClient_ListSnippets_Decodes(t *testing.T) {
	t.Parallel()
	client, _ := newCloudSnippetServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(snippetListJSON))
	})
	got, err := client.ListSnippets("acme", 0)
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, "Xqjyp1GV", got[0].ID)
	assert.Equal(t, "My snippet", got[0].Title)
	assert.Equal(t, "alice", got[0].Owner)
	assert.False(t, got[0].IsPrivate)
	assert.Equal(t, "https://bitbucket.org/snippets/alice/Xqjyp1GV", got[0].WebURL)
	assert.Len(t, got[0].Files, 2)

	assert.True(t, got[1].IsPrivate)
	assert.Equal(t, "bob", got[1].Owner)
}

func TestCloudClient_GetSnippet_PathAndDecode(t *testing.T) {
	t.Parallel()
	var gotPath string
	client, _ := newCloudSnippetServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(snippetGetJSON))
	})
	got, err := client.GetSnippet("alice", "Xqjyp1GV")
	require.NoError(t, err)
	assert.Equal(t, "/snippets/alice/Xqjyp1GV", gotPath)
	assert.Equal(t, "Xqjyp1GV", got.ID)
	assert.Equal(t, "My snippet", got.Title)
	assert.Len(t, got.Files, 1)
	assert.Equal(t, "hello.go", got.Files[0].Name)
}

func TestCloudClient_CreateSnippet_BodyAndPath(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	var gotBody []byte
	client, _ := newCloudSnippetServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(snippetGetJSON))
	})

	got, err := client.CreateSnippet("alice", backend.CreateSnippetInput{
		Title:     "My snippet",
		IsPrivate: false,
		Files: []backend.SnippetFile{
			{Name: "hello.go", Content: "package main"},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/snippets/alice", gotPath)
	assert.Equal(t, "Xqjyp1GV", got.ID)

	var sent map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &sent))
	assert.Equal(t, "My snippet", sent["title"])
	assert.Equal(t, false, sent["is_private"])
	files, ok := sent["files"].(map[string]any)
	require.True(t, ok, "files must be a JSON object")
	_, hasFile := files["hello.go"]
	assert.True(t, hasFile, "hello.go must be present in files")
}

func TestCloudClient_CreateSnippet_RejectsEmptyTitle(t *testing.T) {
	t.Parallel()
	client := cloud.NewClient(http.DefaultClient, "https://example.invalid", "tok", "")
	_, err := client.CreateSnippet("alice", backend.CreateSnippetInput{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "title required")
}

func TestCloudClient_DeleteSnippet_PathAndMethod(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	client, _ := newCloudSnippetServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})
	err := client.DeleteSnippet("alice", "Xqjyp1GV")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/snippets/alice/Xqjyp1GV", gotPath)
}
