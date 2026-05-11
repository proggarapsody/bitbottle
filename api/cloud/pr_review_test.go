package cloud_test

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// reviewRecorder captures the path of each request the SubmitReview flow
// makes against the httptest server, so tests can assert on the call
// sequence (body comment → inline comments → action) without mocking each
// dependency individually. Concurrency-safe: SubmitReview is sequential, but
// the test server can serve from multiple goroutines under -race.
type reviewRecorder struct {
	mu    sync.Mutex
	paths []string
}

func (r *reviewRecorder) record(path string) {
	r.mu.Lock()
	r.paths = append(r.paths, path)
	r.mu.Unlock()
}

func (r *reviewRecorder) snapshot() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, len(r.paths))
	copy(out, r.paths)
	return out
}

// commentResponse returns a minimal but valid Cloud PR comment payload so
// that AddPRComment's response-parsing path is exercised, not bypassed.
func commentResponse() string {
	return `{"id":1,"content":{"raw":"x"},"user":{"display_name":"a","account_id":"a"},"created_on":"2024-01-01T00:00:00Z","updated_on":"2024-01-01T00:00:00Z"}`
}

func TestCloudClient_SubmitReview_ApprovePath_PostsBodyThenApprove(t *testing.T) {
	t.Parallel()
	rec := &reviewRecorder{}
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		rec.record(r.Method + " " + r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, commentResponse())
	})
	err := client.SubmitReview("ws", "svc", 7, backend.SubmitReviewInput{
		Action: "approve",
		Body:   "looks good",
	})
	require.NoError(t, err)
	got := rec.snapshot()
	require.Len(t, got, 2, "expect comment then approve")
	assert.Equal(t, "POST /repositories/ws/svc/pullrequests/7/comments", got[0])
	assert.Equal(t, "POST /repositories/ws/svc/pullrequests/7/approve", got[1])
}

func TestCloudClient_SubmitReview_RequestChangesPath_PostsAction(t *testing.T) {
	t.Parallel()
	rec := &reviewRecorder{}
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		rec.record(r.Method + " " + r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, commentResponse())
	})
	err := client.SubmitReview("ws", "svc", 7, backend.SubmitReviewInput{
		Action: "request_changes",
	})
	require.NoError(t, err)
	got := rec.snapshot()
	require.Len(t, got, 1, "no body, no inline → only the action POST")
	assert.Equal(t, "POST /repositories/ws/svc/pullrequests/7/request-changes", got[0])
}

func TestCloudClient_SubmitReview_CommentOnly_NoActionCall(t *testing.T) {
	t.Parallel()
	rec := &reviewRecorder{}
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		rec.record(r.Method + " " + r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, commentResponse())
	})
	err := client.SubmitReview("ws", "svc", 7, backend.SubmitReviewInput{
		Action: "comment",
		Body:   "fyi",
	})
	require.NoError(t, err)
	got := rec.snapshot()
	require.Len(t, got, 1, "comment-only → just the body POST")
	assert.Equal(t, "POST /repositories/ws/svc/pullrequests/7/comments", got[0])
}

func TestCloudClient_SubmitReview_PostsInlineCommentsBeforeAction(t *testing.T) {
	t.Parallel()
	rec := &reviewRecorder{}
	var bodies []string
	var bodyMu sync.Mutex
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		rec.record(r.Method + " " + r.URL.Path)
		if strings.HasSuffix(r.URL.Path, "/comments") {
			b, _ := io.ReadAll(r.Body)
			bodyMu.Lock()
			bodies = append(bodies, string(b))
			bodyMu.Unlock()
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, commentResponse())
	})
	err := client.SubmitReview("ws", "svc", 7, backend.SubmitReviewInput{
		Action: "approve",
		Inline: []backend.SubmitReviewInline{
			{Path: "pkg/a.go", Line: 12, Side: "new", Body: "fix"},
			{Path: "pkg/b.go", Line: 5, Side: "old", Body: "rm"},
		},
	})
	require.NoError(t, err)
	got := rec.snapshot()
	require.Len(t, got, 3, "two inline comments + approve")
	assert.Equal(t, "POST /repositories/ws/svc/pullrequests/7/comments", got[0])
	assert.Equal(t, "POST /repositories/ws/svc/pullrequests/7/comments", got[1])
	assert.Equal(t, "POST /repositories/ws/svc/pullrequests/7/approve", got[2])
	require.Len(t, bodies, 2)
	assert.Contains(t, bodies[0], `"path":"pkg/a.go"`)
	assert.Contains(t, bodies[0], `"to":12`)
	assert.Contains(t, bodies[1], `"path":"pkg/b.go"`)
	assert.Contains(t, bodies[1], `"from":5`)
}

func TestCloudClient_SubmitReview_UnknownAction_Errors(t *testing.T) {
	t.Parallel()
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, commentResponse())
	})
	err := client.SubmitReview("ws", "svc", 7, backend.SubmitReviewInput{Action: "bogus"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bogus")
}
