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

const listPRCommentsJSON = `{"values":[{"id":1,"content":{"raw":"LGTM!"},"user":{"account_id":"123","display_name":"Alice","nickname":"alice"},"created_on":"2026-04-24T10:00:00Z"},{"id":2,"content":{"raw":"Please add tests"},"user":{"account_id":"456","display_name":"Bob","nickname":"bob"},"created_on":"2026-04-24T11:00:00Z"}]}`

func TestCloudClient_ListPRComments(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(listPRCommentsJSON))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	cmts, err := client.ListPRComments("myws", "my-svc", 42)
	require.NoError(t, err)

	assert.Equal(t, "/repositories/myws/my-svc/pullrequests/42/comments", gotPath)
	require.Len(t, cmts, 2)
	assert.Equal(t, 1, cmts[0].ID)
	assert.Equal(t, "LGTM!", cmts[0].Text)
	assert.Equal(t, "alice", cmts[0].Author.Slug)
	assert.Equal(t, "Alice", cmts[0].Author.DisplayName)
	assert.False(t, cmts[0].CreatedAt.IsZero())
}

func TestCloudClient_ListPRComments_InlineSingleLineNewSide(t *testing.T) {
	t.Parallel()
	const body = `{"values":[{"id":7,"content":{"raw":"nit"},"user":{"nickname":"alice","display_name":"Alice"},"created_on":"2026-04-24T10:00:00Z","inline":{"path":"main.go","to":42}}]}`
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	cmts, err := client.ListPRComments("myws", "my-svc", 42)
	require.NoError(t, err)
	require.Len(t, cmts, 1)
	require.NotNil(t, cmts[0].Inline)
	assert.Equal(t, "main.go", cmts[0].Inline.Path)
	assert.Equal(t, "new", cmts[0].Inline.Side)
	assert.Equal(t, 42, cmts[0].Inline.Line)
	assert.Equal(t, 0, cmts[0].Inline.StartLine)
}

func TestCloudClient_ListPRComments_InlineOldSideMultiLine(t *testing.T) {
	t.Parallel()
	const body = `{"values":[{"id":8,"content":{"raw":"range nit"},"user":{"nickname":"alice"},"created_on":"2026-04-24T10:00:00Z","inline":{"path":"main.go","from":10,"start_from":7}}]}`
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	cmts, err := client.ListPRComments("myws", "my-svc", 42)
	require.NoError(t, err)
	require.Len(t, cmts, 1)
	require.NotNil(t, cmts[0].Inline)
	assert.Equal(t, "old", cmts[0].Inline.Side)
	assert.Equal(t, 10, cmts[0].Inline.Line)
	assert.Equal(t, 7, cmts[0].Inline.StartLine)
}

func TestCloudClient_ListPRComments_ParentResolutionUpdatedAt(t *testing.T) {
	t.Parallel()
	const body = `{"values":[{"id":9,"content":{"raw":"reply"},"user":{"nickname":"bob"},"created_on":"2026-04-24T10:00:00Z","updated_on":"2026-04-24T11:30:00Z","parent":{"id":7},"resolution":{"type":"resolved"}}]}`
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	cmts, err := client.ListPRComments("myws", "my-svc", 42)
	require.NoError(t, err)
	require.Len(t, cmts, 1)
	assert.Equal(t, 7, cmts[0].ParentID)
	assert.True(t, cmts[0].Resolved)
	assert.False(t, cmts[0].UpdatedAt.IsZero())
	assert.Nil(t, cmts[0].Inline)
}

func TestCloudClient_AddPRComment(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":99,"content":{"raw":"hello"},"user":{"display_name":"Alice","nickname":"alice"},"created_on":"2026-04-24T12:00:00Z"}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	got, err := client.AddPRComment("myws", "my-svc", 42, backend.AddPRCommentInput{Text: "hello"})
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/repositories/myws/my-svc/pullrequests/42/comments", gotPath)
	content, _ := gotBody["content"].(map[string]any)
	assert.Equal(t, "hello", content["raw"])
	assert.Equal(t, 99, got.ID)
	assert.Equal(t, "hello", got.Text)
}
