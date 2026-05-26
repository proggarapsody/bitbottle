package cloud_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/cloud"
)

const snippetCommentListJSON = `{
  "values": [
    {
      "id": 17,
      "content": {"raw": "Nice snippet!"},
      "created_on": "2024-01-01T12:00:00Z",
      "updated_on": "2024-01-01T12:00:00Z",
      "author": {"display_name": "Alice"},
      "links": {"html": {"href": "https://bitbucket.org/snippets/alice/Xqjyp1GV/_/diff#comment-17"}}
    },
    {
      "id": 18,
      "content": {"raw": "Thanks!"},
      "created_on": "2024-01-02T10:00:00Z",
      "updated_on": "2024-01-02T10:00:00Z",
      "author": {"display_name": "Bob"},
      "links": {"html": {"href": "https://bitbucket.org/snippets/alice/Xqjyp1GV/_/diff#comment-18"}}
    }
  ]
}`

const snippetCommentAddJSON = `{
  "id": 42,
  "content": {"raw": "Hello world"},
  "created_on": "2024-03-01T08:00:00Z",
  "updated_on": "2024-03-01T08:00:00Z",
  "author": {"display_name": "Charlie"},
  "links": {"html": {"href": "https://bitbucket.org/snippets/alice/Xqjyp1GV/_/diff#comment-42"}}
}`

func newCloudSnippetCommentServer(t *testing.T, handler http.HandlerFunc) *cloud.Client {
	t.Helper()
	// reuse the existing helper from snippets_test.go (same package).
	client, _ := newCloudSnippetServer(t, handler)
	return client
}

func TestCloudClient_ListSnippetComments_PathAndQuery(t *testing.T) {
	t.Parallel()
	var gotPath, gotQuery string
	client := newCloudSnippetCommentServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	})
	_, err := client.ListSnippetComments("alice", "Xqjyp1GV", 25)
	require.NoError(t, err)
	assert.Equal(t, "/snippets/alice/Xqjyp1GV/comments", gotPath)
	assert.Contains(t, gotQuery, "pagelen=25")
}

func TestCloudClient_ListSnippetComments_Decodes(t *testing.T) {
	t.Parallel()
	client := newCloudSnippetCommentServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(snippetCommentListJSON))
	})
	got, err := client.ListSnippetComments("alice", "Xqjyp1GV", 0)
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, 17, got[0].ID)
	assert.Equal(t, "Nice snippet!", got[0].Body)
	assert.Equal(t, "Alice", got[0].Author)
	assert.Equal(t, "2024-01-01T12:00:00Z", got[0].CreatedOn)
	assert.Contains(t, got[0].WebURL, "comment-17")

	assert.Equal(t, 18, got[1].ID)
	assert.Equal(t, "Bob", got[1].Author)
}

func TestCloudClient_AddSnippetComment_BodyAndPath(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	var gotBody []byte
	client := newCloudSnippetCommentServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(snippetCommentAddJSON))
	})

	got, err := client.AddSnippetComment("alice", "Xqjyp1GV", "Hello world")
	require.NoError(t, err)
	assert.Equal(t, "POST", gotMethod)
	assert.Equal(t, "/snippets/alice/Xqjyp1GV/comments", gotPath)
	assert.Equal(t, 42, got.ID)
	assert.Equal(t, "Hello world", got.Body)

	var sent map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &sent))
	content, ok := sent["content"].(map[string]any)
	require.True(t, ok, "content must be a JSON object")
	assert.Equal(t, "Hello world", content["raw"])
}

func TestCloudClient_DeleteSnippetComment_PathAndMethod(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	client := newCloudSnippetCommentServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})
	err := client.DeleteSnippetComment("alice", "Xqjyp1GV", 17)
	require.NoError(t, err)
	assert.Equal(t, "DELETE", gotMethod)
	assert.Equal(t, "/snippets/alice/Xqjyp1GV/comments/17", gotPath)
}
