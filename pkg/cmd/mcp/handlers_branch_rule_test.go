package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestListBranchRules_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListBranchRulesFn: func(ns, slug string) ([]backend.BranchRule, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			return []backend.BranchRule{
				{ID: 1, Kind: "push", Pattern: "main"},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.listBranchRules(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "push", "main")
}

func TestListBranchRules_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.listBranchRules(context.Background(), makeReq(nil))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}

func TestAddBranchRule_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		AddBranchRuleFn: func(ns, slug string, input backend.BranchRuleInput) (backend.BranchRule, error) {
			return backend.BranchRule{ID: 42, Kind: input.Kind, Pattern: input.Pattern, Value: input.Value}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.addBranchRule(context.Background(), makeReq(map[string]any{
		"repo":    "myws/my-repo",
		"kind":    "push",
		"pattern": "main",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "42", "push")
}

func TestAddBranchRule_MissingKind(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.addBranchRule(context.Background(), makeReq(map[string]any{
		"repo":    "myws/my-repo",
		"pattern": "main",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "kind")
}

func TestAddBranchRule_MissingPattern(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.addBranchRule(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"kind": "push",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "pattern")
}

func TestDeleteBranchRule_Success(t *testing.T) {
	t.Parallel()
	deleted := false
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteBranchRuleFn: func(ns, slug string, id int) error {
			assert.Equal(t, 7, id)
			deleted = true
			return nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.deleteBranchRule(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
		"id":   float64(7),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "deleted", "")
	assert.True(t, deleted)
}

func TestDeleteBranchRule_MissingID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.deleteBranchRule(context.Background(), makeReq(map[string]any{
		"repo": "myws/my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "id")
}

func TestUpdateBranchRule_Success(t *testing.T) {
	t.Parallel()
	var gotInput backend.UpdateBranchRuleInput
	fake := &testhelpers.FakeClient{
		T: t,
		UpdateBranchRuleFn: func(ns, slug string, id int, in backend.UpdateBranchRuleInput) (backend.BranchRule, error) {
			assert.Equal(t, "myws", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, 5, id)
			gotInput = in
			return backend.BranchRule{ID: 5, Kind: "push", Pattern: "release/*"}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.updateBranchRule(context.Background(), makeReq(map[string]any{
		"repo":    "myws/my-repo",
		"id":      float64(5),
		"pattern": "release/*",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "5", "release/*")
	require.NotNil(t, gotInput.Pattern)
	assert.Equal(t, "release/*", *gotInput.Pattern)
}

func TestUpdateBranchRule_MissingID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.updateBranchRule(context.Background(), makeReq(map[string]any{
		"repo":    "myws/my-repo",
		"pattern": "main",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "id")
}

func TestUpdateBranchRule_MissingRepo(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.updateBranchRule(context.Background(), makeReq(map[string]any{
		"id":      float64(5),
		"pattern": "main",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "repo")
}
