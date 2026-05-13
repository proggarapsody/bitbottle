package mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// serverFakeClient wraps FakeClient and also satisfies PRCommentStateSetter,
// simulating a Server/DC backend where tasks are supported.
type serverFakeClient struct {
	*testhelpers.FakeClient
	SetPRCommentStateFn func(ns, slug string, id, commentID int, state string) error
}

func (s *serverFakeClient) SetPRCommentState(ns, slug string, id, commentID int, state string) error {
	if s.SetPRCommentStateFn != nil {
		return s.SetPRCommentStateFn(ns, slug, id, commentID, state)
	}
	s.T.Fatalf("unexpected call to serverFakeClient.SetPRCommentState; set SetPRCommentStateFn in your test")
	return nil
}

// newServerHandlers builds a handlers instance backed by a serverFakeClient
// so that AsPRCommentStateSetter succeeds (simulating Server/DC).
func newServerHandlers(t *testing.T, fake *serverFakeClient) *handlers {
	t.Helper()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: singleHostConfig})
	factorytest.UseBackend(f, fake)
	return newHandlers(f)
}

// ── listPRTasks ───────────────────────────────────────────────────────────────

func TestMCP_ListPRTasks_ServerReturnsBlockerTasks(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		ListPRCommentsFn: func(ns, slug string, id int) ([]backend.PRComment, error) {
			return []backend.PRComment{
				{ID: 1, Text: "fix the null check", Severity: "BLOCKER", State: "OPEN"},
				{ID: 2, Text: "regular comment", Severity: ""},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listPRTasks(context.Background(), makeReq(map[string]any{
		"project": "myproj",
		"slug":    "my-repo",
		"pr_id":   float64(42),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "fix the null check", "")
	// regular comment must not appear when tasks are present
	text := extractText(t, result)
	assert.NotContains(t, text, "regular comment")
}

func TestMCP_ListPRTasks_FilterResolved(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		ListPRCommentsFn: func(ns, slug string, id int) ([]backend.PRComment, error) {
			return []backend.PRComment{
				{ID: 1, Text: "open task", Severity: "BLOCKER", State: "OPEN"},
				{ID: 2, Text: "done task", Severity: "BLOCKER", State: "RESOLVED"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listPRTasks(context.Background(), makeReq(map[string]any{
		"project": "myproj",
		"slug":    "my-repo",
		"pr_id":   float64(42),
		"state":   "resolved",
	}))
	require.NoError(t, err)
	text := extractText(t, result)
	assert.Contains(t, text, "done task")
	assert.NotContains(t, text, "open task")
}

func TestMCP_ListPRTasks_MissingProject(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.listPRTasks(context.Background(), makeReq(map[string]any{
		"slug":  "my-repo",
		"pr_id": float64(42),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestMCP_ListPRTasks_MissingPRID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.listPRTasks(context.Background(), makeReq(map[string]any{
		"project": "myproj",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "pr_id")
}

func TestMCP_ListPRTasks_BackendError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		ListPRCommentsFn: func(ns, slug string, id int) ([]backend.PRComment, error) {
			return nil, errors.New("upstream failure")
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listPRTasks(context.Background(), makeReq(map[string]any{
		"project": "myproj",
		"slug":    "my-repo",
		"pr_id":   float64(42),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "upstream failure")
}

// ── createPRTask ──────────────────────────────────────────────────────────────

func TestMCP_CreatePRTask_OK(t *testing.T) {
	t.Parallel()
	var gotIn backend.AddPRCommentInput
	fake := &testhelpers.FakeClient{T: t,
		AddPRCommentFn: func(ns, slug string, id int, in backend.AddPRCommentInput) (backend.PRComment, error) {
			gotIn = in
			return backend.PRComment{ID: 99, Text: in.Text, Severity: "BLOCKER"}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.createPRTask(context.Background(), makeReq(map[string]any{
		"project": "myproj",
		"slug":    "my-repo",
		"pr_id":   float64(42),
		"body":    "fix the null check",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "99", "")
	assertJSONContains(t, result, "BLOCKER", "")
	assert.Equal(t, "BLOCKER", gotIn.Severity)
	assert.Equal(t, "fix the null check", gotIn.Text)
}

func TestMCP_CreatePRTask_WithParent(t *testing.T) {
	t.Parallel()
	var gotIn backend.AddPRCommentInput
	fake := &testhelpers.FakeClient{T: t,
		AddPRCommentFn: func(ns, slug string, id int, in backend.AddPRCommentInput) (backend.PRComment, error) {
			gotIn = in
			return backend.PRComment{ID: 100, Text: in.Text}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.createPRTask(context.Background(), makeReq(map[string]any{
		"project":           "myproj",
		"slug":              "my-repo",
		"pr_id":             float64(42),
		"body":              "reply task",
		"parent_comment_id": float64(55),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "100", "")
	require.NotNil(t, gotIn.Parent)
	assert.Equal(t, 55, *gotIn.Parent)
}

func TestMCP_CreatePRTask_MissingBody(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.createPRTask(context.Background(), makeReq(map[string]any{
		"project": "myproj",
		"slug":    "my-repo",
		"pr_id":   float64(42),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "body")
}

func TestMCP_CreatePRTask_BackendError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		AddPRCommentFn: func(ns, slug string, id int, in backend.AddPRCommentInput) (backend.PRComment, error) {
			return backend.PRComment{}, errors.New("server error")
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.createPRTask(context.Background(), makeReq(map[string]any{
		"project": "myproj",
		"slug":    "my-repo",
		"pr_id":   float64(42),
		"body":    "task text",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "server error")
}

// ── resolvePRTask / reopenPRTask ──────────────────────────────────────────────

func TestMCP_ResolvePRTask_OK(t *testing.T) {
	t.Parallel()
	var gotState string
	fake := &serverFakeClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		SetPRCommentStateFn: func(ns, slug string, id, commentID int, state string) error {
			gotState = state
			return nil
		},
	}
	h := newServerHandlers(t, fake)
	result, err := h.resolvePRTask(context.Background(), makeReq(map[string]any{
		"project": "myproj",
		"slug":    "my-repo",
		"pr_id":   float64(42),
		"task_id": float64(7),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "RESOLVED", "")
	assert.Equal(t, "RESOLVED", gotState)
}

func TestMCP_ReopenPRTask_OK(t *testing.T) {
	t.Parallel()
	var gotState string
	fake := &serverFakeClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		SetPRCommentStateFn: func(ns, slug string, id, commentID int, state string) error {
			gotState = state
			return nil
		},
	}
	h := newServerHandlers(t, fake)
	result, err := h.reopenPRTask(context.Background(), makeReq(map[string]any{
		"project": "myproj",
		"slug":    "my-repo",
		"pr_id":   float64(42),
		"task_id": float64(7),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "OPEN", "")
	assert.Equal(t, "OPEN", gotState)
}

func TestMCP_ResolvePRTask_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	// FakeClient does NOT implement PRCommentStateSetter → Cloud-like backend.
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.resolvePRTask(context.Background(), makeReq(map[string]any{
		"project": "myproj",
		"slug":    "my-repo",
		"pr_id":   float64(42),
		"task_id": float64(7),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "host.unsupported")
}

func TestMCP_ResolvePRTask_MissingTaskID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.resolvePRTask(context.Background(), makeReq(map[string]any{
		"project": "myproj",
		"slug":    "my-repo",
		"pr_id":   float64(42),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "task_id")
}

func TestMCP_ResolvePRTask_BackendError(t *testing.T) {
	t.Parallel()
	fake := &serverFakeClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		SetPRCommentStateFn: func(ns, slug string, id, commentID int, state string) error {
			return errors.New("task not found")
		},
	}
	h := newServerHandlers(t, fake)
	result, err := h.resolvePRTask(context.Background(), makeReq(map[string]any{
		"project": "myproj",
		"slug":    "my-repo",
		"pr_id":   float64(42),
		"task_id": float64(7),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "task not found")
}
