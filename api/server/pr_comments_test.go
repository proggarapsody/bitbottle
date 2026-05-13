package server_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

// inlineDiffStub returns a Server diff endpoint stub that yields the given
// (fromHash,toHash) for any path query. The response shape mirrors the
// JSON-flavoured /diff endpoint (fields: fromHash, toHash, diffs[]).
func inlineDiffStub(fromHash, toHash string) string {
	return `{"fromHash":"` + fromHash + `","toHash":"` + toHash + `","diffs":[]}`
}

func TestServerClient_AddPRComment_InlineLooksUpHashesAndPostsAnchor(t *testing.T) {
	t.Parallel()
	var gotCommentBody map[string]any
	var gotDiffPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/diff/"):
			gotDiffPath = r.URL.Path
			_, _ = w.Write([]byte(inlineDiffStub("aaa111", "bbb222")))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/comments"):
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &gotCommentBody)
			_, _ = w.Write([]byte(`{"id":501,"text":"nit","author":{"slug":"alice"},"createdDate":1714000200000}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	_, err := client.AddPRComment("MYPROJ", "my-svc", 42, backend.AddPRCommentInput{
		Text:   "nit",
		Inline: &backend.PRCommentInline{Path: "src/foo.go", Side: "new", Line: 88},
	})
	require.NoError(t, err)

	assert.Equal(t, "/rest/api/1.0/projects/MYPROJ/repos/my-svc/pull-requests/42/diff/src/foo.go", gotDiffPath)

	anchor, ok := gotCommentBody["anchor"].(map[string]any)
	require.True(t, ok, "expected anchor object in body, got %#v", gotCommentBody)
	assert.Equal(t, "src/foo.go", anchor["path"])
	assert.EqualValues(t, 88, anchor["line"])
	assert.Equal(t, "TO", anchor["fileType"]) // new-side
	assert.Equal(t, "aaa111", anchor["fromHash"])
	assert.Equal(t, "bbb222", anchor["toHash"])
}

func TestServerClient_AddPRComment_InlineMultiLineUnsupported(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("HTTP must not be called when input is invalid; got %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	_, err := client.AddPRComment("MYPROJ", "my-svc", 42, backend.AddPRCommentInput{
		Text:   "range",
		Inline: &backend.PRCommentInline{Path: "src/foo.go", Side: "new", Line: 20, StartLine: 15},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "multi-line")
}

func TestServerClient_AddPRComment_Reply(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":600,"text":"reply","author":{"slug":"bob"},"createdDate":1714000300000}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	parent := 7
	_, err := client.AddPRComment("MYPROJ", "my-svc", 42, backend.AddPRCommentInput{
		Text:   "reply",
		Parent: &parent,
	})
	require.NoError(t, err)
	parentObj, ok := gotBody["parent"].(map[string]any)
	require.True(t, ok, "expected parent object, got %#v", gotBody)
	assert.EqualValues(t, 7, parentObj["id"])
	_, hasAnchor := gotBody["anchor"]
	assert.False(t, hasAnchor, "reply must not carry an inline anchor")
}

func TestServerClient_EditPRComment_FetchesVersionThenPuts(t *testing.T) {
	t.Parallel()
	var gotPaths []string
	var gotPutQuery string
	var gotPutBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"id":99,"version":3,"text":"old"}`))
		case http.MethodPut:
			gotPutQuery = r.URL.RawQuery
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &gotPutBody)
			_, _ = w.Write([]byte(`{"id":99,"version":4,"text":"new","author":{"slug":"alice"},"createdDate":1714000400000,"updatedDate":1714000500000}`))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	got, err := client.EditPRComment("MYPROJ", "my-svc", 42, 99, "new")
	require.NoError(t, err)

	require.Len(t, gotPaths, 2, "expected GET-then-PUT, got %v", gotPaths)
	assert.Equal(t, "GET /rest/api/1.0/projects/MYPROJ/repos/my-svc/pull-requests/42/comments/99", gotPaths[0])
	assert.Equal(t, "PUT /rest/api/1.0/projects/MYPROJ/repos/my-svc/pull-requests/42/comments/99", gotPaths[1])
	assert.Equal(t, "version=3", gotPutQuery)
	assert.Equal(t, "new", gotPutBody["text"])
	assert.EqualValues(t, 3, gotPutBody["version"])
	assert.Equal(t, "new", got.Text)
}

func TestServerClient_DeletePRComment_FetchesVersionThenDeletes(t *testing.T) {
	t.Parallel()
	var gotPaths []string
	var gotDeleteQuery string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"id":99,"version":5,"text":"x"}`))
		case http.MethodDelete:
			gotDeleteQuery = r.URL.RawQuery
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	require.NoError(t, client.DeletePRComment("MYPROJ", "my-svc", 42, 99))
	require.Len(t, gotPaths, 2)
	assert.Equal(t, "GET /rest/api/1.0/projects/MYPROJ/repos/my-svc/pull-requests/42/comments/99", gotPaths[0])
	assert.Equal(t, "DELETE /rest/api/1.0/projects/MYPROJ/repos/my-svc/pull-requests/42/comments/99", gotPaths[1])
	assert.Equal(t, "version=5", gotDeleteQuery)
}

// ── Task (BLOCKER comment) tests ─────────────────────────────────────────────

func TestServerClient_ListPRComments_SeverityAndStatePopulated(t *testing.T) {
	t.Parallel()
	const body = `{"values":[{"action":"COMMENTED","comment":{"id":10,"text":"fix this","severity":"BLOCKER","state":"OPEN","version":2,"author":{"slug":"alice","displayName":"Alice"},"createdDate":1714000000000}}],"isLastPage":true,"size":1}`
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	cmts, err := client.ListPRComments("MYPROJ", "my-svc", 42)
	require.NoError(t, err)
	require.Len(t, cmts, 1)
	assert.Equal(t, "BLOCKER", cmts[0].Severity)
	assert.Equal(t, "OPEN", cmts[0].State)
	assert.Equal(t, 2, cmts[0].Version)
}

func TestServerClient_AddPRComment_SeverityBlockerIncludedInBody(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":200,"text":"must fix","severity":"BLOCKER","state":"OPEN","version":0,"author":{"slug":"alice"},"createdDate":1714000000000}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	c, err := client.AddPRComment("MYPROJ", "my-svc", 42, backend.AddPRCommentInput{
		Text:     "must fix",
		Severity: "BLOCKER",
	})
	require.NoError(t, err)
	assert.Equal(t, "must fix", gotBody["text"])
	assert.Equal(t, "BLOCKER", gotBody["severity"])
	assert.Equal(t, "BLOCKER", c.Severity)
}

func TestServerClient_AddPRComment_NoSeverityOmitsField(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":201,"text":"hi","author":{"slug":"alice"},"createdDate":1714000000000}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	_, err := client.AddPRComment("MYPROJ", "my-svc", 42, backend.AddPRCommentInput{Text: "hi"})
	require.NoError(t, err)
	_, hasSeverity := gotBody["severity"]
	assert.False(t, hasSeverity, "severity should be omitted when empty")
}

func TestServerClient_SetPRCommentState_GetsThenPuts(t *testing.T) {
	t.Parallel()
	var gotPaths []string
	var gotPutBody map[string]any
	var gotPutQuery string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPaths = append(gotPaths, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"id":77,"version":4,"text":"task","severity":"BLOCKER","state":"OPEN"}`))
		case http.MethodPut:
			gotPutQuery = r.URL.RawQuery
			b, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(b, &gotPutBody)
			_, _ = w.Write([]byte(`{"id":77,"version":5,"text":"task","severity":"BLOCKER","state":"RESOLVED"}`))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")

	err := client.SetPRCommentState("MYPROJ", "my-svc", 42, 77, "RESOLVED")
	require.NoError(t, err)

	require.Len(t, gotPaths, 2, "expected GET-then-PUT, got %v", gotPaths)
	assert.Equal(t, "GET /rest/api/1.0/projects/MYPROJ/repos/my-svc/pull-requests/42/comments/77", gotPaths[0])
	assert.Equal(t, "PUT /rest/api/1.0/projects/MYPROJ/repos/my-svc/pull-requests/42/comments/77", gotPaths[1])
	assert.Equal(t, "version=4", gotPutQuery)
	assert.Equal(t, "RESOLVED", gotPutBody["state"])
	assert.EqualValues(t, 4, gotPutBody["version"])
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
