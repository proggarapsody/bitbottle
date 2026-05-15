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

// suggestionFakeClient wraps FakeClient and also satisfies SuggestionApplier,
// simulating a Server/DC backend where suggestions are supported.
type suggestionFakeClient struct {
	*testhelpers.FakeClient
	ApplySuggestionFn      func(ns, slug string, prID, commentID, suggestionID int) (backend.SuggestionApplyResult, error)
	GetSuggestionPreviewFn func(ns, slug string, prID, commentID int) (string, error)
}

func (s *suggestionFakeClient) ApplySuggestion(ns, slug string, prID, commentID, suggestionID int) (backend.SuggestionApplyResult, error) {
	if s.ApplySuggestionFn != nil {
		return s.ApplySuggestionFn(ns, slug, prID, commentID, suggestionID)
	}
	s.T.Fatalf("unexpected call to suggestionFakeClient.ApplySuggestion; set ApplySuggestionFn")
	return backend.SuggestionApplyResult{}, nil
}

func (s *suggestionFakeClient) GetSuggestionPreview(ns, slug string, prID, commentID int) (string, error) {
	if s.GetSuggestionPreviewFn != nil {
		return s.GetSuggestionPreviewFn(ns, slug, prID, commentID)
	}
	s.T.Fatalf("unexpected call to suggestionFakeClient.GetSuggestionPreview; set GetSuggestionPreviewFn")
	return "", nil
}

// newSuggestionHandlers builds a handlers instance backed by a suggestionFakeClient.
func newSuggestionHandlers(t *testing.T, fake *suggestionFakeClient) *handlers {
	t.Helper()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: singleHostConfig})
	factorytest.UseBackend(f, fake)
	return newHandlers(f)
}

// ── prSuggestionApply ─────────────────────────────────────────────────────────

func TestMCP_PRSuggestionApply_OK(t *testing.T) {
	t.Parallel()
	var gotPRID, gotCommentID, gotSuggestionID int
	fake := &suggestionFakeClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		ApplySuggestionFn: func(ns, slug string, prID, commentID, suggestionID int) (backend.SuggestionApplyResult, error) {
			gotPRID = prID
			gotCommentID = commentID
			gotSuggestionID = suggestionID
			return backend.SuggestionApplyResult{CommitHash: "abc123", CommitMessage: "Applied"}, nil
		},
	}
	h := newSuggestionHandlers(t, fake)
	result, err := h.prSuggestionApply(context.Background(), makeReq(map[string]any{
		"project":       "myproj",
		"slug":          "my-repo",
		"pr_id":         float64(42),
		"comment_id":    float64(7),
		"suggestion_id": float64(1),
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assertJSONContains(t, result, "abc123", "")
	assert.Equal(t, 42, gotPRID)
	assert.Equal(t, 7, gotCommentID)
	assert.Equal(t, 1, gotSuggestionID)
}

func TestMCP_PRSuggestionApply_Preview(t *testing.T) {
	t.Parallel()
	fake := &suggestionFakeClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		GetSuggestionPreviewFn: func(ns, slug string, prID, commentID int) (string, error) {
			return "suggestion body text", nil
		},
	}
	h := newSuggestionHandlers(t, fake)
	result, err := h.prSuggestionApply(context.Background(), makeReq(map[string]any{
		"project":       "myproj",
		"slug":          "my-repo",
		"pr_id":         float64(42),
		"comment_id":    float64(7),
		"suggestion_id": float64(1),
		"preview":       true,
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "suggestion body text", "")
}

func TestMCP_PRSuggestionApply_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	// FakeClient does NOT implement SuggestionApplier → Cloud-like backend.
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.prSuggestionApply(context.Background(), makeReq(map[string]any{
		"project":       "myproj",
		"slug":          "my-repo",
		"pr_id":         float64(42),
		"comment_id":    float64(7),
		"suggestion_id": float64(1),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "host.unsupported")
}

func TestMCP_PRSuggestionApply_MissingProject(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, nil)
	result, err := h.prSuggestionApply(context.Background(), makeReq(map[string]any{
		"slug":          "my-repo",
		"pr_id":         float64(42),
		"comment_id":    float64(7),
		"suggestion_id": float64(1),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestMCP_PRSuggestionApply_BackendError(t *testing.T) {
	t.Parallel()
	fake := &suggestionFakeClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		ApplySuggestionFn: func(ns, slug string, prID, commentID, suggestionID int) (backend.SuggestionApplyResult, error) {
			return backend.SuggestionApplyResult{}, errors.New("server error")
		},
	}
	h := newSuggestionHandlers(t, fake)
	result, err := h.prSuggestionApply(context.Background(), makeReq(map[string]any{
		"project":       "myproj",
		"slug":          "my-repo",
		"pr_id":         float64(42),
		"comment_id":    float64(7),
		"suggestion_id": float64(1),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "server error")
}
