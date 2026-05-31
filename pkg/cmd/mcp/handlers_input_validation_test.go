package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// This file is the MCP-INPUT-VALIDATION acceptance matrix: one focused test
// per gap (MCP-06 … MCP-14) asserting the structured {code, field} envelope
// the fix introduces. assertArgEnvelope decodes the error body and checks the
// dotted code + field so the contract is verified, not just a substring.

func decodeArgEnvelope(t *testing.T, text string) struct {
	Code    string `json:"code"`
	Field   string `json:"field"`
	Got     string `json:"got"`
	Message string `json:"message"`
} {
	t.Helper()
	var env struct {
		Code    string `json:"code"`
		Field   string `json:"field"`
		Got     string `json:"got"`
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal([]byte(text), &env), "error body should be a JSON envelope: %s", text)
	return env
}

// MCP-06: a string passed for a numeric id reports a type error, not "missing".
func TestInputValidation_MCP06_WrongTypeID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.getPR(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ", "slug": "my-repo", "id": "not-a-number",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	env := decodeArgEnvelope(t, extractText(t, result))
	assert.Equal(t, "arg.invalid_type", env.Code)
	assert.Equal(t, "id", env.Field)
	assert.Equal(t, "string", env.Got)
	assert.Contains(t, env.Message, "id must be integer")
}

// MCP-07: an explicit id:0 is accepted as present (not falsely "missing").
// With Min(1) it is now an out-of-range error rather than a missing error —
// the key point is it is NOT reported as missing, and 0 flows through as a
// recognised value.
func TestInputValidation_MCP07_ZeroIDIsPresent(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.getPR(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ", "slug": "my-repo", "id": float64(0),
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	env := decodeArgEnvelope(t, extractText(t, result))
	assert.Equal(t, "arg.out_of_range", env.Code, "id:0 should be present-but-out-of-range, not missing")
	assert.Equal(t, "id", env.Field)
	assert.Equal(t, "0", env.Got)
}

// MCP-07 corollary: a legal zero-value field (limit on a list tool) is read
// from key presence, not >0. Passing limit:0 reaches the helper as 0 rather
// than silently defaulting — proving presence is by-key, not by-zero-value.
// Here listBranches clamps via validateRange(1,100), so 0 surfaces an explicit
// range error instead of being swallowed.
func TestInputValidation_MCP07_LimitZeroValidated(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listBranches(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ", "slug": "my-repo", "limit": float64(0),
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, extractText(t, result), "limit")
}

// MCP-08: a negative id is rejected client-side, never forwarded to the API.
func TestInputValidation_MCP08_NegativeID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.getPR(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ", "slug": "my-repo", "id": float64(-5),
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	env := decodeArgEnvelope(t, extractText(t, result))
	assert.Equal(t, "arg.out_of_range", env.Code)
	assert.Equal(t, "id", env.Field)
	assert.Equal(t, "-5", env.Got)
}

// MCP-09: merge_pr strategy enum no longer lists "" — invalid value message
// names only real strategies, and empty/omitted strategy is accepted.
func TestInputValidation_MCP09_StrategyEnumNoEmpty(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.mergePR(context.Background(), makeReq(map[string]any{
		"project": "PROJ", "slug": "repo", "id": float64(1), "strategy": "bogus",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	env := decodeArgEnvelope(t, extractText(t, result))
	assert.Equal(t, "arg.invalid_value", env.Code)
	assert.Equal(t, "strategy", env.Field)
	assert.Contains(t, env.Message, "must be one of merge, squash, rebase")
	assert.NotContains(t, env.Message, "one of ,")
}

func TestInputValidation_MCP09_EmptyStrategyAccepted(t *testing.T) {
	t.Parallel()
	var gotStrategy string
	fake := &testhelpers.FakeClient{
		MergePRFn: func(ns, slug string, id int, in backend.MergePRInput) (backend.PullRequest, error) {
			gotStrategy = in.Strategy
			return backend.PullRequest{ID: id, State: "MERGED"}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.mergePR(context.Background(), makeReq(map[string]any{
		"project": "PROJ", "slug": "repo", "id": float64(1), "strategy": "",
	}))
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Equal(t, "", gotStrategy, "empty strategy means use the server default")
}

// MCP-10: inline_line without inline_path is now caught symmetrically.
func TestInputValidation_MCP10_InlineLineWithoutPath(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.addPRComment(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ", "slug": "my-repo", "id": float64(1),
		"body": "hi", "inline_line": float64(10),
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	env := decodeArgEnvelope(t, extractText(t, result))
	assert.Equal(t, "inline_path", env.Field)
}

func TestInputValidation_MCP10_InlinePathWithoutLine(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.addPRComment(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ", "slug": "my-repo", "id": float64(1),
		"body": "hi", "inline_path": "main.go",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	env := decodeArgEnvelope(t, extractText(t, result))
	assert.Equal(t, "inline_line", env.Field)
}

// MCP-11: add_commit_comment rejects non-hex / too-short hashes client-side.
func TestInputValidation_MCP11_BadHash(t *testing.T) {
	t.Parallel()
	for _, hash := range []string{"a", "NOT_HEX"} {
		h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
		result, err := h.addCommitComment(context.Background(), makeReq(map[string]any{
			"project": "MYPROJ", "slug": "my-repo", "hash": hash, "body": "x",
		}))
		require.NoError(t, err)
		assert.True(t, result.IsError, "hash=%q", hash)
		env := decodeArgEnvelope(t, extractText(t, result))
		assert.Equal(t, "arg.invalid_value", env.Code, "hash=%q", hash)
		assert.Equal(t, "hash", env.Field, "hash=%q", hash)
	}
}

// MCP-12: create_branch rejects Git-invalid ref names.
func TestInputValidation_MCP12_BadBranchName(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"/", "/leading", "trailing/", "a//b", "feat..x"} {
		h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
		result, err := h.createBranch(context.Background(), makeReq(map[string]any{
			"project": "MYPROJ", "slug": "my-repo", "name": name, "start_at": "main",
		}))
		require.NoError(t, err)
		assert.True(t, result.IsError, "name=%q", name)
		env := decodeArgEnvelope(t, extractText(t, result))
		assert.Equal(t, "arg.invalid_value", env.Code, "name=%q", name)
		assert.Equal(t, "name", env.Field, "name=%q", name)
	}
}

// MCP-13: update_pr with neither title nor body is a clean client-side reject.
func TestInputValidation_MCP13_UpdatePRNoOp(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.updatePR(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ", "slug": "my-repo", "id": float64(7),
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	env := decodeArgEnvelope(t, extractText(t, result))
	assert.Equal(t, "arg.missing", env.Code)
	assert.Contains(t, env.Message, "nothing to update")
}

// MCP-14: compare_refs.repo rejects 3-segment input the same way it rejects
// 1-segment input.
func TestInputValidation_MCP14_ThreeSegmentRepoRejected(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.compareRefs(context.Background(), makeReq(map[string]any{
		"repo": "bitbucket.org/proj/repo", "base": "main", "head": "feature",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, extractText(t, result), "WORKSPACE/REPO")
}

func TestInputValidation_MCP14_OneSegmentRepoRejected(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.compareRefs(context.Background(), makeReq(map[string]any{
		"repo": "single", "base": "main", "head": "feature",
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, extractText(t, result), "WORKSPACE/REPO")
}
