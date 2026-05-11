package server_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/server"
)

// ---- ListCommitComments ----

const listCommitCommentsJSON = `{"values":[{"id":1,"text":"LGTM!","author":{"slug":"alice","displayName":"Alice"},"createdDate":1704067200000,"updatedDate":0,"version":0},{"id":2,"text":"Minor nit","author":{"slug":"bob","displayName":"Bob"},"createdDate":1704070800000,"updatedDate":0,"version":0}],"isLastPage":true,"size":2}`

func TestServerClient_ListCommitComments(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(listCommitCommentsJSON))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	cmts, err := client.ListCommitComments("MYPROJ", "my-repo", "deadbeef", 0)
	require.NoError(t, err)

	assert.Equal(t, "/rest/api/1.0/projects/MYPROJ/repos/my-repo/commits/deadbeef/comments", gotPath)
	require.Len(t, cmts, 2)
	assert.Equal(t, 1, cmts[0].ID)
	assert.Equal(t, "LGTM!", cmts[0].Body)
	assert.Equal(t, "alice", cmts[0].Author.Slug)
	assert.Equal(t, "Alice", cmts[0].Author.DisplayName)
	assert.False(t, cmts[0].CreatedAt.IsZero())
	assert.Equal(t, 2, cmts[1].ID)
}

func TestServerClient_ListCommitComments_Pagination(t *testing.T) {
	t.Parallel()
	nextStart := 1
	callCount := 0
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			start := nextStart
			page := struct {
				Values        []map[string]any `json:"values"`
				IsLastPage    bool             `json:"isLastPage"`
				NextPageStart *int             `json:"nextPageStart"`
			}{
				Values:        []map[string]any{{"id": 1, "text": "first", "author": map[string]string{"slug": "alice", "displayName": "Alice"}, "createdDate": 1704067200000, "version": 0}},
				IsLastPage:    false,
				NextPageStart: &start,
			}
			b, _ := json.Marshal(page)
			_, _ = w.Write(b)
		} else {
			_, _ = w.Write([]byte(`{"values":[{"id":2,"text":"second","author":{"slug":"bob","displayName":"Bob"},"createdDate":1704070800000,"version":0}],"isLastPage":true,"size":1}`))
		}
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	cmts, err := client.ListCommitComments("MYPROJ", "my-repo", "deadbeef", 0)
	require.NoError(t, err)
	assert.Len(t, cmts, 2)
	assert.Equal(t, 2, callCount)
}

// ---- AddCommitComment ----

func TestServerClient_AddCommitComment(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":42,"text":"hello","author":{"slug":"alice","displayName":"Alice"},"createdDate":1704067200000,"version":0}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	got, err := client.AddCommitComment("MYPROJ", "my-repo", "deadbeef", backend.AddCommitCommentInput{Body: "hello"})
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/rest/api/1.0/projects/MYPROJ/repos/my-repo/commits/deadbeef/comments", gotPath)
	assert.Equal(t, "hello", gotBody["text"])
	assert.Equal(t, 42, got.ID)
	assert.Equal(t, "hello", got.Body)
}

// ---- EditCommitComment ----

func TestServerClient_EditCommitComment(t *testing.T) {
	t.Parallel()
	var paths []string
	var methods []string
	var lastBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.RequestURI())
		methods = append(methods, r.Method)
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			// fetchCommitCommentVersion call
			_, _ = w.Write([]byte(`{"id":99,"version":3}`))
		} else {
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &lastBody)
			_, _ = w.Write([]byte(`{"id":99,"text":"updated","author":{"slug":"alice","displayName":"Alice"},"createdDate":1704067200000,"version":4}`))
		}
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	got, err := client.EditCommitComment("MYPROJ", "my-repo", "deadbeef", 99, "updated")
	require.NoError(t, err)

	// First call: GET to fetch version
	assert.Equal(t, http.MethodGet, methods[0])
	assert.Equal(t, "/rest/api/1.0/projects/MYPROJ/repos/my-repo/commits/deadbeef/comments/99", paths[0])
	// Second call: PUT with version
	assert.Equal(t, http.MethodPut, methods[1])
	assert.Contains(t, paths[1], "version=3")
	assert.Equal(t, "updated", lastBody["text"])
	assert.EqualValues(t, 3, lastBody["version"])
	assert.Equal(t, 99, got.ID)
	assert.Equal(t, "updated", got.Body)
}

// ---- DeleteCommitComment ----

func TestServerClient_DeleteCommitComment(t *testing.T) {
	t.Parallel()
	var paths []string
	var methods []string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.RequestURI())
		methods = append(methods, r.Method)
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"id":99,"version":2}`))
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	require.NoError(t, client.DeleteCommitComment("MYPROJ", "my-repo", "deadbeef", 99))

	assert.Equal(t, http.MethodGet, methods[0])
	assert.Equal(t, http.MethodDelete, methods[1])
	assert.Contains(t, paths[1], "version=2")
}
