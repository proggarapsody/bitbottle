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

// reactorFakeClient wraps FakeClient and also satisfies CommentReactor,
// simulating a Server/DC backend where reactions are supported.
type reactorFakeClient struct {
	*testhelpers.FakeClient
	ListCommentReactionsFn  func(ns, slug string, prID, commentID int) ([]backend.CommentReaction, error)
	AddCommentReactionFn    func(ns, slug string, prID, commentID int, emoji string) error
	RemoveCommentReactionFn func(ns, slug string, prID, commentID int, emoji string) error
}

func (r *reactorFakeClient) ListCommentReactions(ns, slug string, prID, commentID int) ([]backend.CommentReaction, error) {
	if r.ListCommentReactionsFn != nil {
		return r.ListCommentReactionsFn(ns, slug, prID, commentID)
	}
	if r.T != nil {
		r.T.Fatalf("unexpected call to reactorFakeClient.ListCommentReactions; set ListCommentReactionsFn in your test")
	}
	return nil, nil
}

func (r *reactorFakeClient) AddCommentReaction(ns, slug string, prID, commentID int, emoji string) error {
	if r.AddCommentReactionFn != nil {
		return r.AddCommentReactionFn(ns, slug, prID, commentID, emoji)
	}
	if r.T != nil {
		r.T.Fatalf("unexpected call to reactorFakeClient.AddCommentReaction; set AddCommentReactionFn in your test")
	}
	return nil
}

func (r *reactorFakeClient) RemoveCommentReaction(ns, slug string, prID, commentID int, emoji string) error {
	if r.RemoveCommentReactionFn != nil {
		return r.RemoveCommentReactionFn(ns, slug, prID, commentID, emoji)
	}
	if r.T != nil {
		r.T.Fatalf("unexpected call to reactorFakeClient.RemoveCommentReaction; set RemoveCommentReactionFn in your test")
	}
	return nil
}

// newReactorHandlers builds a handlers instance backed by a reactorFakeClient
// so that AsCommentReactor succeeds (simulating Server/DC).
func newReactorHandlers(t *testing.T, fake *reactorFakeClient) *handlers {
	t.Helper()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: singleHostConfig})
	factorytest.UseBackend(f, fake)
	return newHandlers(f)
}

// ── listCommentReactions ──────────────────────────────────────────────────────

func TestMCP_ListCommentReactions_ReturnsGroupedReactions(t *testing.T) {
	t.Parallel()
	fake := &reactorFakeClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		ListCommentReactionsFn: func(ns, slug string, prID, commentID int) ([]backend.CommentReaction, error) {
			assert.Equal(t, "MYPROJ", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, 42, prID)
			assert.Equal(t, 7, commentID)
			return []backend.CommentReaction{
				{Emoji: "thumbs_up", Users: []backend.User{{DisplayName: "Alice"}, {DisplayName: "Bob"}}},
				{Emoji: "heart", Users: []backend.User{{DisplayName: "Carol"}}},
			}, nil
		},
	}
	h := newReactorHandlers(t, fake)
	result, err := h.listCommentReactions(context.Background(), makeReq(map[string]any{
		"project":    "myproj",
		"repo":       "my-repo",
		"pr_id":      float64(42),
		"comment_id": float64(7),
	}))
	require.NoError(t, err)
	text := extractText(t, result)
	assert.Contains(t, text, "thumbs_up")
	assert.Contains(t, text, "heart")
	assert.Contains(t, text, "Alice")
	assert.Contains(t, text, "Carol")
}

func TestMCP_ListCommentReactions_ProjectUppercased(t *testing.T) {
	t.Parallel()
	var gotNS string
	fake := &reactorFakeClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		ListCommentReactionsFn: func(ns, slug string, prID, commentID int) ([]backend.CommentReaction, error) {
			gotNS = ns
			return nil, nil
		},
	}
	h := newReactorHandlers(t, fake)
	_, err := h.listCommentReactions(context.Background(), makeReq(map[string]any{
		"project":    "lower",
		"repo":       "my-repo",
		"pr_id":      float64(42),
		"comment_id": float64(7),
	}))
	require.NoError(t, err)
	assert.Equal(t, "LOWER", gotNS)
}

func TestMCP_ListCommentReactions_MissingProject(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.listCommentReactions(context.Background(), makeReq(map[string]any{
		"repo":       "my-repo",
		"pr_id":      float64(42),
		"comment_id": float64(7),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestMCP_ListCommentReactions_MissingPRID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.listCommentReactions(context.Background(), makeReq(map[string]any{
		"project":    "myproj",
		"repo":       "my-repo",
		"comment_id": float64(7),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "pr_id")
}

func TestMCP_ListCommentReactions_MissingCommentID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.listCommentReactions(context.Background(), makeReq(map[string]any{
		"project": "myproj",
		"repo":    "my-repo",
		"pr_id":   float64(42),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "comment_id")
}

func TestMCP_ListCommentReactions_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	// Plain FakeClient does NOT implement CommentReactor → AsCommentReactor returns ErrUnsupportedOnHost.
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listCommentReactions(context.Background(), makeReq(map[string]any{
		"project":    "myproj",
		"repo":       "my-repo",
		"pr_id":      float64(42),
		"comment_id": float64(7),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "host.unsupported")
}

func TestMCP_ListCommentReactions_BackendError(t *testing.T) {
	t.Parallel()
	fake := &reactorFakeClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		ListCommentReactionsFn: func(ns, slug string, prID, commentID int) ([]backend.CommentReaction, error) {
			return nil, errors.New("reactions unavailable")
		},
	}
	h := newReactorHandlers(t, fake)
	result, err := h.listCommentReactions(context.Background(), makeReq(map[string]any{
		"project":    "myproj",
		"repo":       "my-repo",
		"pr_id":      float64(42),
		"comment_id": float64(7),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "reactions unavailable")
}

// ── addCommentReaction ────────────────────────────────────────────────────────

func TestMCP_AddCommentReaction_NormalisesAndCallsAPI(t *testing.T) {
	t.Parallel()
	var gotEmoji string
	var gotNS, gotSlug string
	var gotPRID, gotCommentID int
	fake := &reactorFakeClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		AddCommentReactionFn: func(ns, slug string, prID, commentID int, emoji string) error {
			gotNS = ns
			gotSlug = slug
			gotPRID = prID
			gotCommentID = commentID
			gotEmoji = emoji
			return nil
		},
	}
	h := newReactorHandlers(t, fake)
	result, err := h.addCommentReaction(context.Background(), makeReq(map[string]any{
		"project":    "myproj",
		"repo":       "my-repo",
		"pr_id":      float64(42),
		"comment_id": float64(7),
		"emoji":      ":thumbsup:",
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError, "expected success result")
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-repo", gotSlug)
	assert.Equal(t, 42, gotPRID)
	assert.Equal(t, 7, gotCommentID)
	assert.Equal(t, "thumbs_up", gotEmoji)
}

func TestMCP_AddCommentReaction_NormalisesVariousEmojiForms(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input    string
		expected string
	}{
		{":thumbsup:", "thumbs_up"},
		{":thumbsdown:", "thumbs_down"},
		{":heart:", "heart"},
		{":smile:", "laugh"},
		{":tada:", "hooray"},
		{":confused:", "confused"},
		{"thumbs_up", "thumbs_up"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			var gotEmoji string
			fake := &reactorFakeClient{
				FakeClient: &testhelpers.FakeClient{T: t},
				AddCommentReactionFn: func(ns, slug string, prID, commentID int, emoji string) error {
					gotEmoji = emoji
					return nil
				},
			}
			h := newReactorHandlers(t, fake)
			_, err := h.addCommentReaction(context.Background(), makeReq(map[string]any{
				"project":    "myproj",
				"repo":       "my-repo",
				"pr_id":      float64(42),
				"comment_id": float64(7),
				"emoji":      tc.input,
			}))
			require.NoError(t, err)
			assert.Equal(t, tc.expected, gotEmoji)
		})
	}
}

func TestMCP_AddCommentReaction_MissingEmoji(t *testing.T) {
	t.Parallel()
	// Missing emoji is validated before resolving the backend, so any client works.
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.addCommentReaction(context.Background(), makeReq(map[string]any{
		"project":    "myproj",
		"repo":       "my-repo",
		"pr_id":      float64(42),
		"comment_id": float64(7),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "emoji")
}

func TestMCP_AddCommentReaction_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	// Plain FakeClient does NOT implement CommentReactor.
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.addCommentReaction(context.Background(), makeReq(map[string]any{
		"project":    "myproj",
		"repo":       "my-repo",
		"pr_id":      float64(42),
		"comment_id": float64(7),
		"emoji":      "thumbs_up",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "host.unsupported")
}

func TestMCP_AddCommentReaction_BackendError(t *testing.T) {
	t.Parallel()
	fake := &reactorFakeClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		AddCommentReactionFn: func(ns, slug string, prID, commentID int, emoji string) error {
			return errors.New("reaction failed")
		},
	}
	h := newReactorHandlers(t, fake)
	result, err := h.addCommentReaction(context.Background(), makeReq(map[string]any{
		"project":    "myproj",
		"repo":       "my-repo",
		"pr_id":      float64(42),
		"comment_id": float64(7),
		"emoji":      "thumbs_up",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "reaction failed")
}

// ── removeCommentReaction ─────────────────────────────────────────────────────

func TestMCP_RemoveCommentReaction_NormalisesAndCallsAPI(t *testing.T) {
	t.Parallel()
	var gotEmoji string
	var gotNS, gotSlug string
	var gotPRID, gotCommentID int
	fake := &reactorFakeClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		RemoveCommentReactionFn: func(ns, slug string, prID, commentID int, emoji string) error {
			gotNS = ns
			gotSlug = slug
			gotPRID = prID
			gotCommentID = commentID
			gotEmoji = emoji
			return nil
		},
	}
	h := newReactorHandlers(t, fake)
	result, err := h.removeCommentReaction(context.Background(), makeReq(map[string]any{
		"project":    "myproj",
		"repo":       "my-repo",
		"pr_id":      float64(42),
		"comment_id": float64(7),
		"emoji":      ":thumbsdown:",
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError, "expected success result")
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-repo", gotSlug)
	assert.Equal(t, 42, gotPRID)
	assert.Equal(t, 7, gotCommentID)
	assert.Equal(t, "thumbs_down", gotEmoji)
}

func TestMCP_RemoveCommentReaction_MissingEmoji(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.removeCommentReaction(context.Background(), makeReq(map[string]any{
		"project":    "myproj",
		"repo":       "my-repo",
		"pr_id":      float64(42),
		"comment_id": float64(7),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "emoji")
}

func TestMCP_RemoveCommentReaction_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	// Plain FakeClient does NOT implement CommentReactor.
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.removeCommentReaction(context.Background(), makeReq(map[string]any{
		"project":    "myproj",
		"repo":       "my-repo",
		"pr_id":      float64(42),
		"comment_id": float64(7),
		"emoji":      "thumbs_up",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "host.unsupported")
}

func TestMCP_RemoveCommentReaction_BackendError(t *testing.T) {
	t.Parallel()
	fake := &reactorFakeClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		RemoveCommentReactionFn: func(ns, slug string, prID, commentID int, emoji string) error {
			return errors.New("delete failed")
		},
	}
	h := newReactorHandlers(t, fake)
	result, err := h.removeCommentReaction(context.Background(), makeReq(map[string]any{
		"project":    "myproj",
		"repo":       "my-repo",
		"pr_id":      float64(42),
		"comment_id": float64(7),
		"emoji":      "thumbs_up",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "delete failed")
}

// ── listPRComments with include_reactions ─────────────────────────────────────

func TestMCP_ListPRComments_WithReactions_FetchesAndIncludesReactions(t *testing.T) {
	t.Parallel()
	fake := &reactorFakeClient{
		FakeClient: &testhelpers.FakeClient{T: t,
			ListPRCommentsFn: func(ns, slug string, id int) ([]backend.PRComment, error) {
				return []backend.PRComment{
					{ID: 10, Text: "looks good"},
					{ID: 11, Text: "needs work"},
				}, nil
			},
		},
		ListCommentReactionsFn: func(ns, slug string, prID, commentID int) ([]backend.CommentReaction, error) {
			if commentID == 10 {
				return []backend.CommentReaction{
					{Emoji: "thumbs_up", Users: []backend.User{{DisplayName: "Alice"}}},
				}, nil
			}
			return nil, nil
		},
	}
	h := newReactorHandlers(t, fake)
	result, err := h.listPRComments(context.Background(), makeReq(map[string]any{
		"project":           "myproj",
		"slug":              "my-repo",
		"id":                float64(42),
		"include_reactions": true,
	}))
	require.NoError(t, err)
	text := extractText(t, result)
	assert.Contains(t, text, "thumbs_up")
	assert.Contains(t, text, "Alice")
	assert.Contains(t, text, "looks good")
	assert.Contains(t, text, "needs work")
}

func TestMCP_ListPRComments_WithoutReactions_DoesNotFetchReactions(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		ListPRCommentsFn: func(ns, slug string, id int) ([]backend.PRComment, error) {
			return []backend.PRComment{
				{ID: 10, Text: "looks good"},
			}, nil
		},
		// ListCommentReactionsFn deliberately NOT set; FakeClient.T.Fatalf would fire.
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listPRComments(context.Background(), makeReq(map[string]any{
		"project": "myproj",
		"slug":    "my-repo",
		"id":      float64(42),
		// include_reactions omitted → false
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "looks good", "")
}
