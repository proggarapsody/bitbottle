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

const listPRActivitiesJSON = `{"values":[{"action":"COMMENTED","comment":{"id":1,"text":"LGTM!","author":{"slug":"alice","displayName":"Alice"},"createdDate":1714000000000}},{"action":"OPENED"},{"action":"COMMENTED","comment":{"id":2,"text":"please add tests","author":{"slug":"bob","displayName":"Bob"},"createdDate":1714000100000}}],"isLastPage":true,"size":3}`

func TestServerClient_ListPRComments(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(listPRActivitiesJSON))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	cmts, err := client.ListPRComments("MYPROJ", "my-svc", 42)
	require.NoError(t, err)

	assert.Equal(t, "/rest/api/1.0/projects/MYPROJ/repos/my-svc/pull-requests/42/activities", gotPath)
	require.Len(t, cmts, 2)
	assert.Equal(t, 1, cmts[0].ID)
	assert.Equal(t, "LGTM!", cmts[0].Text)
	assert.Equal(t, "alice", cmts[0].Author.Slug)
	assert.Equal(t, 2, cmts[1].ID)
}

func TestServerClient_ListPRComments_InlineAnchor(t *testing.T) {
	t.Parallel()
	const body = `{"values":[{"action":"COMMENTED","commentAnchor":{"path":"src/foo.go","line":42,"lineType":"ADDED","fileType":"TO","fromHash":"aaa","toHash":"bbb"},"comment":{"id":11,"text":"nit","author":{"slug":"alice","displayName":"Alice"},"createdDate":1714000000000,"updatedDate":1714003600000}}],"isLastPage":true,"size":1}`
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	cmts, err := client.ListPRComments("MYPROJ", "my-svc", 42)
	require.NoError(t, err)
	require.Len(t, cmts, 1)
	require.NotNil(t, cmts[0].Inline)
	assert.Equal(t, "src/foo.go", cmts[0].Inline.Path)
	assert.Equal(t, "new", cmts[0].Inline.Side)
	assert.Equal(t, 42, cmts[0].Inline.Line)
	assert.False(t, cmts[0].UpdatedAt.IsZero())
}

func TestServerClient_ListPRComments_FlattensNestedReplies(t *testing.T) {
	t.Parallel()
	// Top-level comment 11 has reply 12, which has nested reply 13.
	const body = `{"values":[{"action":"COMMENTED","comment":{"id":11,"text":"top","author":{"slug":"alice"},"createdDate":1714000000000,"comments":[{"id":12,"text":"reply","author":{"slug":"bob"},"createdDate":1714000100000,"comments":[{"id":13,"text":"reply-to-reply","author":{"slug":"alice"},"createdDate":1714000200000}]}]}}],"isLastPage":true,"size":1}`
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	cmts, err := client.ListPRComments("MYPROJ", "my-svc", 42)
	require.NoError(t, err)
	require.Len(t, cmts, 3)
	assert.Equal(t, 11, cmts[0].ID)
	assert.Equal(t, 0, cmts[0].ParentID)
	assert.Equal(t, 12, cmts[1].ID)
	assert.Equal(t, 11, cmts[1].ParentID)
	assert.Equal(t, 13, cmts[2].ID)
	assert.Equal(t, 12, cmts[2].ParentID)
}

func TestServerClient_ListPRComments_InlineAnchorOnTopOnlyNotReplies(t *testing.T) {
	t.Parallel()
	const body = `{"values":[{"action":"COMMENTED","commentAnchor":{"path":"src/foo.go","line":7,"fileType":"FROM"},"comment":{"id":21,"text":"q","author":{"slug":"alice"},"createdDate":1714000000000,"comments":[{"id":22,"text":"a","author":{"slug":"bob"},"createdDate":1714000100000}]}}],"isLastPage":true,"size":1}`
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	cmts, err := client.ListPRComments("MYPROJ", "my-svc", 42)
	require.NoError(t, err)
	require.Len(t, cmts, 2)
	require.NotNil(t, cmts[0].Inline)
	assert.Equal(t, "old", cmts[0].Inline.Side) // FROM → old
	assert.Equal(t, 7, cmts[0].Inline.Line)
	assert.Nil(t, cmts[1].Inline, "replies do not carry the inline anchor")
	assert.Equal(t, 21, cmts[1].ParentID)
}

func TestServerClient_AddPRComment(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":99,"text":"hello","author":{"slug":"alice","displayName":"Alice"},"createdDate":1714000200000}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	got, err := client.AddPRComment("MYPROJ", "my-svc", 42, backend.AddPRCommentInput{Text: "hello"})
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/rest/api/1.0/projects/MYPROJ/repos/my-svc/pull-requests/42/comments", gotPath)
	assert.Equal(t, "hello", gotBody["text"])
	assert.Equal(t, 99, got.ID)
	assert.Equal(t, "hello", got.Text)
}
