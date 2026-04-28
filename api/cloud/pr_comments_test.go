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
