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

func TestCloudClient_AddPRComment_InlineNewSide(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":7,"content":{"raw":"nit"},"user":{"nickname":"alice"},"created_on":"2026-04-24T12:00:00Z","inline":{"path":"main.go","to":42}}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	_, err := client.AddPRComment("myws", "my-svc", 42, backend.AddPRCommentInput{
		Text:   "nit",
		Inline: &backend.PRCommentInline{Path: "main.go", Side: "new", Line: 42},
	})
	require.NoError(t, err)

	inline, ok := gotBody["inline"].(map[string]any)
	require.True(t, ok, "expected inline object in request body, got %#v", gotBody)
	assert.Equal(t, "main.go", inline["path"])
	assert.EqualValues(t, 42, inline["to"])
	_, hasFrom := inline["from"]
	assert.False(t, hasFrom, "old-side fields should be absent on a new-side comment")
}

func TestCloudClient_AddPRComment_InlineOldSideMultiLine(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":8,"content":{"raw":"range"},"user":{"nickname":"alice"},"created_on":"2026-04-24T12:00:00Z"}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	_, err := client.AddPRComment("myws", "my-svc", 42, backend.AddPRCommentInput{
		Text:   "range",
		Inline: &backend.PRCommentInline{Path: "main.go", Side: "old", Line: 20, StartLine: 15},
	})
	require.NoError(t, err)

	inline := gotBody["inline"].(map[string]any)
	assert.Equal(t, "main.go", inline["path"])
	assert.EqualValues(t, 20, inline["from"])
	assert.EqualValues(t, 15, inline["start_from"])
	_, hasTo := inline["to"]
	assert.False(t, hasTo)
}

func TestCloudClient_EditPRComment(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":99,"content":{"raw":"updated body"},"user":{"nickname":"alice"},"created_on":"2026-04-24T12:00:00Z","updated_on":"2026-04-24T13:00:00Z"}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	got, err := client.EditPRComment("myws", "my-svc", 42, 99, "updated body")
	require.NoError(t, err)

	assert.Equal(t, http.MethodPut, gotMethod)
	assert.Equal(t, "/repositories/myws/my-svc/pullrequests/42/comments/99", gotPath)
	content := gotBody["content"].(map[string]any)
	assert.Equal(t, "updated body", content["raw"])
	assert.Equal(t, "updated body", got.Text)
}

func TestCloudClient_DeletePRComment(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	require.NoError(t, client.DeletePRComment("myws", "my-svc", 42, 99))

	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/repositories/myws/my-svc/pullrequests/42/comments/99", gotPath)
}

func TestCloudClient_ResolvePRComment(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":99,"content":{"raw":"x"},"user":{"nickname":"alice"},"created_on":"2026-04-24T12:00:00Z","resolution":{"type":"resolved"}}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	require.NoError(t, client.ResolvePRComment("myws", "my-svc", 42, 99))

	// Cloud requires a write — verb is PUT or PATCH; assert one of those plus the resolved body.
	assert.Contains(t, []string{http.MethodPut, http.MethodPatch, http.MethodPost}, gotMethod)
	assert.Equal(t, "/repositories/myws/my-svc/pullrequests/42/comments/99", gotPath)
	resolution, ok := gotBody["resolution"].(map[string]any)
	require.True(t, ok, "expected resolution object in body, got %#v", gotBody)
	assert.Equal(t, "resolved", resolution["type"])
}

func TestCloudClient_AddPRComment_Reply(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":11,"content":{"raw":"agreed"},"user":{"nickname":"bob"},"created_on":"2026-04-24T12:00:00Z","parent":{"id":7}}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	parent := 7
	_, err := client.AddPRComment("myws", "my-svc", 42, backend.AddPRCommentInput{
		Text:   "agreed",
		Parent: &parent,
	})
	require.NoError(t, err)

	parentObj := gotBody["parent"].(map[string]any)
	assert.EqualValues(t, 7, parentObj["id"])
	_, hasInline := gotBody["inline"]
	assert.False(t, hasInline)
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

// TestCloudClient_AddPRComment_IgnoresSeverity verifies that Cloud silently
// ignores Severity (Cloud has no task concept) and creates a normal comment.
func TestCloudClient_AddPRComment_IgnoresSeverity(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":300,"content":{"raw":"task body"},"user":{"display_name":"Alice","nickname":"alice"},"created_on":"2026-04-24T12:00:00Z"}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	_, err := client.AddPRComment("myws", "my-svc", 42, backend.AddPRCommentInput{
		Text:     "task body",
		Severity: "BLOCKER",
	})
	require.NoError(t, err)
	// Cloud wire body has no "severity" key.
	_, hasSeverity := gotBody["severity"]
	assert.False(t, hasSeverity, "Cloud must not forward Severity to the API")
}

// TestCloudClient_DoesNotImplementCommentReactor verifies that the Cloud client
// does NOT satisfy CommentReactor, so AsCommentReactor returns
// ErrUnsupportedOnHost instead of succeeding silently.
func TestCloudClient_DoesNotImplementCommentReactor(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("no HTTP call expected; got %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	_, err := backend.AsCommentReactor(client, "bitbucket.org")
	require.Error(t, err)

	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.ErrorIs(t, de, backend.ErrUnsupportedOnHost)
	assert.Equal(t, backend.CodeHostUnsupported, de.Code)
}

// TestCloudClient_DoesNotImplementPRCommentStateSetter verifies that the Cloud
// client does NOT satisfy PRCommentStateSetter, so AsPRCommentStateSetter
// returns ErrUnsupportedOnHost instead of succeeding silently.
func TestCloudClient_DoesNotImplementPRCommentStateSetter(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("no HTTP call expected; got %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	_, err := backend.AsPRCommentStateSetter(client, "bitbucket.org")
	require.Error(t, err)

	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.ErrorIs(t, de, backend.ErrUnsupportedOnHost)
	assert.Equal(t, backend.CodeHostUnsupported, de.Code)
}
