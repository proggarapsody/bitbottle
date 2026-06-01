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

// ---- ListCommitComments ----

func TestCloudClient_ListCommitComments_Single(t *testing.T) {
	t.Parallel()
	const body = `{"values":[{"id":1,"content":{"raw":"Looks good"},"user":{"account_id":"abc","display_name":"Alice","nickname":"alice"},"created_on":"2024-01-01T10:00:00Z","updated_on":"2024-01-01T10:00:00Z"}]}`
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	cmts, err := client.ListCommitComments("myws", "my-repo", "abc123", 0)
	require.NoError(t, err)

	assert.Equal(t, "/repositories/myws/my-repo/commit/abc123/comments", gotPath)
	require.Len(t, cmts, 1)
	assert.Equal(t, 1, cmts[0].ID)
	assert.Equal(t, "Looks good", cmts[0].Body)
	assert.Equal(t, "alice", cmts[0].Author.Slug)
	assert.Equal(t, "Alice", cmts[0].Author.DisplayName)
	assert.False(t, cmts[0].CreatedAt.IsZero())
}

func TestCloudClient_ListCommitComments_Pagination(t *testing.T) {
	t.Parallel()
	page1 := `{"values":[{"id":1,"content":{"raw":"first"},"user":{"nickname":"alice"},"created_on":"2024-01-01T10:00:00Z"}],"next":"NEXTURL"}`
	page2 := `{"values":[{"id":2,"content":{"raw":"second"},"user":{"nickname":"bob"},"created_on":"2024-01-01T11:00:00Z"}]}`
	callCount := 0
	var srv *httptest.Server
	srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			// Return page1 with a next pointer to the same server
			p1 := `{"values":[{"id":1,"content":{"raw":"first"},"user":{"nickname":"alice"},"created_on":"2024-01-01T10:00:00Z"}],"next":"` + srv.URL + `/repositories/myws/my-repo/commit/abc123/comments?page=2"}`
			_, _ = w.Write([]byte(p1))
		} else {
			_, _ = w.Write([]byte(page2))
		}
	}))
	t.Cleanup(srv.Close)
	_ = page1
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	cmts, err := client.ListCommitComments("myws", "my-repo", "abc123", 0)
	require.NoError(t, err)
	assert.Len(t, cmts, 2)
	assert.Equal(t, 2, callCount)
}

func TestCloudClient_ListCommitComments_NicknameFallback(t *testing.T) {
	t.Parallel()
	const body = `{"values":[{"id":5,"content":{"raw":"hi"},"user":{"account_id":"xyz","display_name":"Bot","nickname":""},"created_on":"2024-01-01T10:00:00Z"}]}`
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	cmts, err := client.ListCommitComments("myws", "my-repo", "abc123", 0)
	require.NoError(t, err)
	require.Len(t, cmts, 1)
	// When nickname is empty, fall back to account_id
	assert.Equal(t, "xyz", cmts[0].Author.Slug)
}

// ---- AddCommitComment ----

func TestCloudClient_AddCommitComment(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":99,"content":{"raw":"hello"},"user":{"nickname":"alice","display_name":"Alice"},"created_on":"2024-01-01T10:00:00Z","updated_on":"2024-01-01T10:00:00Z"}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	got, err := client.AddCommitComment("myws", "my-repo", "abc123", backend.AddCommitCommentInput{Body: "hello"})
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/repositories/myws/my-repo/commit/abc123/comments", gotPath)
	content, _ := gotBody["content"].(map[string]any)
	assert.Equal(t, "hello", content["raw"])
	assert.Equal(t, 99, got.ID)
	assert.Equal(t, "hello", got.Body)
	assert.Equal(t, "alice", got.Author.Slug)
}

// ---- EditCommitComment ----

func TestCloudClient_EditCommitComment(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":99,"content":{"raw":"updated"},"user":{"nickname":"alice"},"created_on":"2024-01-01T10:00:00Z","updated_on":"2024-01-01T11:00:00Z"}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	got, err := client.EditCommitComment("myws", "my-repo", "abc123", 99, "updated")
	require.NoError(t, err)

	assert.Equal(t, http.MethodPut, gotMethod)
	assert.Equal(t, "/repositories/myws/my-repo/commit/abc123/comments/99", gotPath)
	content, _ := gotBody["content"].(map[string]any)
	assert.Equal(t, "updated", content["raw"])
	assert.Equal(t, 99, got.ID)
	assert.Equal(t, "updated", got.Body)
}

// ---- DeleteCommitComment ----

func TestCloudClient_DeleteCommitComment(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	require.NoError(t, client.DeleteCommitComment("myws", "my-repo", "abc123", 99))

	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/repositories/myws/my-repo/commit/abc123/comments/99", gotPath)
}
