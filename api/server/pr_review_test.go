package server_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/server"
)

// reviewRecorder mirrors the Cloud test helper: thread-safe append-only log
// of "METHOD /path" strings the SubmitReview flow produces. Lets each test
// assert the call sequence is exactly comments-then-action.
type reviewRecorder struct {
	mu    sync.Mutex
	paths []string
}

func (r *reviewRecorder) record(s string) {
	r.mu.Lock()
	r.paths = append(r.paths, s)
	r.mu.Unlock()
}

func (r *reviewRecorder) snapshot() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, len(r.paths))
	copy(out, r.paths)
	return out
}

const serverCommentResponse = `{"id":1,"text":"x","author":{"slug":"a","displayName":"A"},"createdDate":1714000000000}`

func newReviewServerClient(t *testing.T, handler http.HandlerFunc) *server.Client {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return server.NewClient(srv.Client(), srv.URL+"/rest/api/1.0", "tok", "alice")
}

func TestServerClient_SubmitReview_ApprovePath_PostsBodyThenApprove(t *testing.T) {
	t.Parallel()
	rec := &reviewRecorder{}
	client := newReviewServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		rec.record(r.Method + " " + r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, serverCommentResponse)
	})
	err := client.SubmitReview("MYPROJ", "my-svc", 7, backend.SubmitReviewInput{
		Action: "approve",
		Body:   "looks good",
	})
	require.NoError(t, err)
	got := rec.snapshot()
	require.Len(t, got, 2)
	assert.Equal(t, "POST /rest/api/1.0/projects/MYPROJ/repos/my-svc/pull-requests/7/comments", got[0])
	assert.Equal(t, "POST /rest/api/1.0/projects/MYPROJ/repos/my-svc/pull-requests/7/approve", got[1])
}

func TestServerClient_SubmitReview_CommentOnly_NoActionCall(t *testing.T) {
	t.Parallel()
	rec := &reviewRecorder{}
	client := newReviewServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		rec.record(r.Method + " " + r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, serverCommentResponse)
	})
	err := client.SubmitReview("MYPROJ", "my-svc", 7, backend.SubmitReviewInput{
		Action: "comment",
		Body:   "fyi",
	})
	require.NoError(t, err)
	got := rec.snapshot()
	require.Len(t, got, 1)
	assert.Equal(t, "POST /rest/api/1.0/projects/MYPROJ/repos/my-svc/pull-requests/7/comments", got[0])
}

func TestServerClient_SubmitReview_RequestChanges_ReturnsUnsupportedOnHost(t *testing.T) {
	t.Parallel()
	// Server should refuse before posting anything when the action itself
	// is unsupported. We still respond OK to comments so a regression in
	// ordering surfaces as a test failure rather than a network error.
	client := newReviewServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, serverCommentResponse)
	})
	err := client.SubmitReview("git.example.com", "my-svc", 7, backend.SubmitReviewInput{
		Action: "request_changes",
	})
	require.Error(t, err)
	require.True(t, errors.Is(err, backend.ErrUnsupportedOnHost), "must be ErrUnsupportedOnHost: %v", err)
	var de *backend.DomainError
	require.True(t, errors.As(err, &de))
	assert.Equal(t, backend.CodeHostUnsupported, de.Code)
	assert.Equal(t, "pr-review-request-changes", de.Feature)
}

func TestServerClient_SubmitReview_PostsInlineCommentsBeforeAction(t *testing.T) {
	t.Parallel()
	// Server inline comments require a follow-up GET .../diff/{path} per file
	// to look up fromHash/toHash for the anchor. Stub both shapes from one
	// handler so the flow runs end-to-end.
	rec := &reviewRecorder{}
	client := newReviewServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		rec.record(r.Method + " " + r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/diff/"):
			_, _ = io.WriteString(w, `{"fromHash":"aaa","toHash":"bbb","diffs":[]}`)
		default:
			_, _ = io.WriteString(w, serverCommentResponse)
		}
	})
	err := client.SubmitReview("MYPROJ", "my-svc", 7, backend.SubmitReviewInput{
		Action: "approve",
		Inline: []backend.SubmitReviewInline{
			{Path: "pkg/a.go", Line: 12, Side: "new", Body: "fix"},
		},
	})
	require.NoError(t, err)
	got := rec.snapshot()
	// Expected: GET diff/{path}, POST comments, POST approve.
	require.Len(t, got, 3, "got: %v", got)
	assert.Equal(t, "GET /rest/api/1.0/projects/MYPROJ/repos/my-svc/pull-requests/7/diff/pkg/a.go", got[0])
	assert.Equal(t, "POST /rest/api/1.0/projects/MYPROJ/repos/my-svc/pull-requests/7/comments", got[1])
	assert.Equal(t, "POST /rest/api/1.0/projects/MYPROJ/repos/my-svc/pull-requests/7/approve", got[2])
}

func TestServerClient_SubmitReview_UnknownAction_Errors(t *testing.T) {
	t.Parallel()
	client := newReviewServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, serverCommentResponse)
	})
	err := client.SubmitReview("MYPROJ", "my-svc", 7, backend.SubmitReviewInput{Action: "bogus"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bogus")
}
