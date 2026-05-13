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

// commitReactorFakeClient wraps FakeClient and also satisfies CommitCommentReactor.
type commitReactorFakeClient struct {
	*testhelpers.FakeClient
	ListCommitCommentReactionsFn  func(ns, slug, hash string, commentID int) ([]backend.CommentReaction, error)
	AddCommitCommentReactionFn    func(ns, slug, hash string, commentID int, emoji string) error
	RemoveCommitCommentReactionFn func(ns, slug, hash string, commentID int, emoji string) error
}

func (r *commitReactorFakeClient) ListCommitCommentReactions(ns, slug, hash string, commentID int) ([]backend.CommentReaction, error) {
	if r.ListCommitCommentReactionsFn != nil {
		return r.ListCommitCommentReactionsFn(ns, slug, hash, commentID)
	}
	if r.T != nil {
		r.T.Fatalf("unexpected call to commitReactorFakeClient.ListCommitCommentReactions")
	}
	return nil, nil
}

func (r *commitReactorFakeClient) AddCommitCommentReaction(ns, slug, hash string, commentID int, emoji string) error {
	if r.AddCommitCommentReactionFn != nil {
		return r.AddCommitCommentReactionFn(ns, slug, hash, commentID, emoji)
	}
	if r.T != nil {
		r.T.Fatalf("unexpected call to commitReactorFakeClient.AddCommitCommentReaction")
	}
	return nil
}

func (r *commitReactorFakeClient) RemoveCommitCommentReaction(ns, slug, hash string, commentID int, emoji string) error {
	if r.RemoveCommitCommentReactionFn != nil {
		return r.RemoveCommitCommentReactionFn(ns, slug, hash, commentID, emoji)
	}
	if r.T != nil {
		r.T.Fatalf("unexpected call to commitReactorFakeClient.RemoveCommitCommentReaction")
	}
	return nil
}

func newCommitReactorHandlers(t *testing.T, fake *commitReactorFakeClient) *handlers {
	t.Helper()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: singleHostConfig})
	factorytest.UseBackend(f, fake)
	return newHandlers(f)
}

// ── listCommitCommentReactions ────────────────────────────────────────────────

func TestMCP_ListCommitCommentReactions_ReturnsGroupedReactions(t *testing.T) {
	t.Parallel()
	fake := &commitReactorFakeClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		ListCommitCommentReactionsFn: func(ns, slug, hash string, commentID int) ([]backend.CommentReaction, error) {
			assert.Equal(t, "MYPROJ", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, "abc123", hash)
			assert.Equal(t, 7, commentID)
			return []backend.CommentReaction{
				{Emoji: "thumbs_up", Users: []backend.User{{DisplayName: "Alice"}, {DisplayName: "Bob"}}},
				{Emoji: "heart", Users: []backend.User{{DisplayName: "Carol"}}},
			}, nil
		},
	}
	h := newCommitReactorHandlers(t, fake)
	result, err := h.listCommitCommentReactions(context.Background(), makeReq(map[string]any{
		"project":    "myproj",
		"slug":       "my-repo",
		"hash":       "abc123",
		"comment_id": float64(7),
	}))
	require.NoError(t, err)
	text := extractText(t, result)
	assert.Contains(t, text, "thumbs_up")
	assert.Contains(t, text, "heart")
	assert.Contains(t, text, "Alice")
	assert.Contains(t, text, "Carol")
}

func TestMCP_ListCommitCommentReactions_ProjectUppercased(t *testing.T) {
	t.Parallel()
	var gotNS string
	fake := &commitReactorFakeClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		ListCommitCommentReactionsFn: func(ns, slug, hash string, commentID int) ([]backend.CommentReaction, error) {
			gotNS = ns
			return nil, nil
		},
	}
	h := newCommitReactorHandlers(t, fake)
	_, err := h.listCommitCommentReactions(context.Background(), makeReq(map[string]any{
		"project":    "lower",
		"slug":       "my-repo",
		"hash":       "abc123",
		"comment_id": float64(7),
	}))
	require.NoError(t, err)
	assert.Equal(t, "LOWER", gotNS)
}

func TestMCP_ListCommitCommentReactions_MissingProject(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.listCommitCommentReactions(context.Background(), makeReq(map[string]any{
		"slug":       "my-repo",
		"hash":       "abc123",
		"comment_id": float64(7),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestMCP_ListCommitCommentReactions_MissingHash(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.listCommitCommentReactions(context.Background(), makeReq(map[string]any{
		"project":    "myproj",
		"slug":       "my-repo",
		"comment_id": float64(7),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "hash")
}

func TestMCP_ListCommitCommentReactions_MissingCommentID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.listCommitCommentReactions(context.Background(), makeReq(map[string]any{
		"project": "myproj",
		"slug":    "my-repo",
		"hash":    "abc123",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "comment_id")
}

func TestMCP_ListCommitCommentReactions_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listCommitCommentReactions(context.Background(), makeReq(map[string]any{
		"project":    "myproj",
		"slug":       "my-repo",
		"hash":       "abc123",
		"comment_id": float64(7),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "host.unsupported")
}

func TestMCP_ListCommitCommentReactions_BackendError(t *testing.T) {
	t.Parallel()
	fake := &commitReactorFakeClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		ListCommitCommentReactionsFn: func(ns, slug, hash string, commentID int) ([]backend.CommentReaction, error) {
			return nil, errors.New("reactions unavailable")
		},
	}
	h := newCommitReactorHandlers(t, fake)
	result, err := h.listCommitCommentReactions(context.Background(), makeReq(map[string]any{
		"project":    "myproj",
		"slug":       "my-repo",
		"hash":       "abc123",
		"comment_id": float64(7),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "reactions unavailable")
}

// ── addCommitCommentReaction ──────────────────────────────────────────────────

func TestMCP_AddCommitCommentReaction_NormalisesAndCallsAPI(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug, gotHash, gotEmoji string
	var gotCommentID int
	fake := &commitReactorFakeClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		AddCommitCommentReactionFn: func(ns, slug, hash string, commentID int, emoji string) error {
			gotNS = ns
			gotSlug = slug
			gotHash = hash
			gotCommentID = commentID
			gotEmoji = emoji
			return nil
		},
	}
	h := newCommitReactorHandlers(t, fake)
	result, err := h.addCommitCommentReaction(context.Background(), makeReq(map[string]any{
		"project":    "myproj",
		"slug":       "my-repo",
		"hash":       "abc123",
		"comment_id": float64(7),
		"emoji":      ":thumbsup:",
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError, "expected success result")
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-repo", gotSlug)
	assert.Equal(t, "abc123", gotHash)
	assert.Equal(t, 7, gotCommentID)
	assert.Equal(t, "thumbs_up", gotEmoji)
}

func TestMCP_AddCommitCommentReaction_MissingEmoji(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.addCommitCommentReaction(context.Background(), makeReq(map[string]any{
		"project":    "myproj",
		"slug":       "my-repo",
		"hash":       "abc123",
		"comment_id": float64(7),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "emoji")
}

func TestMCP_AddCommitCommentReaction_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.addCommitCommentReaction(context.Background(), makeReq(map[string]any{
		"project":    "myproj",
		"slug":       "my-repo",
		"hash":       "abc123",
		"comment_id": float64(7),
		"emoji":      "thumbs_up",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "host.unsupported")
}

func TestMCP_AddCommitCommentReaction_BackendError(t *testing.T) {
	t.Parallel()
	fake := &commitReactorFakeClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		AddCommitCommentReactionFn: func(ns, slug, hash string, commentID int, emoji string) error {
			return errors.New("reaction failed")
		},
	}
	h := newCommitReactorHandlers(t, fake)
	result, err := h.addCommitCommentReaction(context.Background(), makeReq(map[string]any{
		"project":    "myproj",
		"slug":       "my-repo",
		"hash":       "abc123",
		"comment_id": float64(7),
		"emoji":      "thumbs_up",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "reaction failed")
}

// ── removeCommitCommentReaction ───────────────────────────────────────────────

func TestMCP_RemoveCommitCommentReaction_NormalisesAndCallsAPI(t *testing.T) {
	t.Parallel()
	var gotEmoji string
	fake := &commitReactorFakeClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		RemoveCommitCommentReactionFn: func(ns, slug, hash string, commentID int, emoji string) error {
			gotEmoji = emoji
			return nil
		},
	}
	h := newCommitReactorHandlers(t, fake)
	result, err := h.removeCommitCommentReaction(context.Background(), makeReq(map[string]any{
		"project":    "myproj",
		"slug":       "my-repo",
		"hash":       "abc123",
		"comment_id": float64(7),
		"emoji":      ":thumbsdown:",
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError, "expected success result")
	assert.Equal(t, "thumbs_down", gotEmoji)
}

func TestMCP_RemoveCommitCommentReaction_MissingEmoji(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.removeCommitCommentReaction(context.Background(), makeReq(map[string]any{
		"project":    "myproj",
		"slug":       "my-repo",
		"hash":       "abc123",
		"comment_id": float64(7),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "emoji")
}

func TestMCP_RemoveCommitCommentReaction_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.removeCommitCommentReaction(context.Background(), makeReq(map[string]any{
		"project":    "myproj",
		"slug":       "my-repo",
		"hash":       "abc123",
		"comment_id": float64(7),
		"emoji":      "heart",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "host.unsupported")
}

func TestMCP_RemoveCommitCommentReaction_BackendError(t *testing.T) {
	t.Parallel()
	fake := &commitReactorFakeClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		RemoveCommitCommentReactionFn: func(ns, slug, hash string, commentID int, emoji string) error {
			return errors.New("delete failed")
		},
	}
	h := newCommitReactorHandlers(t, fake)
	result, err := h.removeCommitCommentReaction(context.Background(), makeReq(map[string]any{
		"project":    "myproj",
		"slug":       "my-repo",
		"hash":       "abc123",
		"comment_id": float64(7),
		"emoji":      "heart",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "delete failed")
}

// ── listCommitComments with include_reactions ─────────────────────────────────

func TestMCP_ListCommitComments_WithReactions_FetchesAndIncludesReactions(t *testing.T) {
	t.Parallel()
	fake := &commitReactorFakeClient{
		FakeClient: &testhelpers.FakeClient{T: t,
			ListCommitCommentsFn: func(ns, slug, hash string, limit int) ([]backend.CommitComment, error) {
				return []backend.CommitComment{
					{ID: 10, Body: "looks good"},
					{ID: 11, Body: "needs work"},
				}, nil
			},
		},
		ListCommitCommentReactionsFn: func(ns, slug, hash string, commentID int) ([]backend.CommentReaction, error) {
			if commentID == 10 {
				return []backend.CommentReaction{
					{Emoji: "thumbs_up", Users: []backend.User{{DisplayName: "Alice"}}},
				}, nil
			}
			return nil, nil
		},
	}
	h := newCommitReactorHandlers(t, fake)
	result, err := h.listCommitComments(context.Background(), makeReq(map[string]any{
		"project":           "myproj",
		"slug":              "my-repo",
		"hash":              "abc123",
		"include_reactions": true,
	}))
	require.NoError(t, err)
	text := extractText(t, result)
	assert.Contains(t, text, "thumbs_up")
	assert.Contains(t, text, "Alice")
	assert.Contains(t, text, "looks good")
	assert.Contains(t, text, "needs work")
}

func TestMCP_ListCommitComments_WithoutReactions_DoesNotFetchReactions(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		ListCommitCommentsFn: func(ns, slug, hash string, limit int) ([]backend.CommitComment, error) {
			return []backend.CommitComment{
				{ID: 10, Body: "looks good"},
			}, nil
		},
		// ListCommitCommentReactionsFn deliberately NOT set.
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.listCommitComments(context.Background(), makeReq(map[string]any{
		"project": "myproj",
		"slug":    "my-repo",
		"hash":    "abc123",
		// include_reactions omitted → false
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "looks good", "")
}
