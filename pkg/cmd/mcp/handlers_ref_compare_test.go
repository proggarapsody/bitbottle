package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestCompareRefs_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		CompareRefsFn: func(ns, slug, base, head string, limit int) (backend.RefComparison, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, "main", base)
			assert.Equal(t, "feature", head)
			return backend.RefComparison{
				Base:     "main",
				Head:     "feature",
				AheadBy:  3,
				BehindBy: 1,
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.compareRefs(context.Background(), makeReq(map[string]any{
		"project": "myws",
		"slug":    "my-repo",
		"base":    "main",
		"head":    "feature",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "ahead_by", "3")
}

// MCP-04: the legacy {repo} shape still works for one release but the result
// carries a deprecation note nudging the client toward {project, slug}.
func TestCompareRefs_LegacyRepoShape_StillWorksWithDeprecation(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		CompareRefsFn: func(ns, slug, base, head string, limit int) (backend.RefComparison, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			return backend.RefComparison{Base: "main", Head: "feature", AheadBy: 3}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.compareRefs(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"base": "main",
		"head": "feature",
	}))
	require.NoError(t, err)
	require.False(t, result.IsError)
	require.Len(t, result.Content, 2, "legacy shape prepends a deprecation note")
	assert.Contains(t, extractTextAt(t, result, 0), "DEPRECATION")
	assert.Contains(t, extractTextAt(t, result, 1), "ahead_by")
}

func TestCompareRefs_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.compareRefs(context.Background(), makeReq(map[string]any{
		"base": "main",
		"head": "feature",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestCompareRefs_MissingBase(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.compareRefs(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"head": "feature",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "base")
}
