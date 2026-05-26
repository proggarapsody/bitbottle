package mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestDryRunMergePR_CanMerge(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		DryRunMergePRFn: func(ns, slug string, prID int, strategy string) (backend.MergeDryRunResult, error) {
			assert.Equal(t, "myproj", ns)
			assert.Equal(t, "my-repo", slug)
			assert.Equal(t, 7, prID)
			assert.Equal(t, "squash", strategy)
			return backend.MergeDryRunResult{CanMerge: true, Message: "OK"}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.dryRunMergePR(context.Background(), makeReq(map[string]any{
		"project":  "myproj",
		"slug":     "my-repo",
		"pr_id":    float64(7),
		"strategy": "squash",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "can_merge", "true")
}

func TestDryRunMergePR_CannotMerge_ReturnsResult(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		DryRunMergePRFn: func(ns, slug string, prID int, strategy string) (backend.MergeDryRunResult, error) {
			return backend.MergeDryRunResult{
				CanMerge: false,
				Vetoes: []backend.MergeVeto{
					{SummaryMessage: "Approvals required", DetailMessage: "2 more needed"},
				},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.dryRunMergePR(context.Background(), makeReq(map[string]any{
		"project": "myproj",
		"slug":    "my-repo",
		"pr_id":   float64(7),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "vetoes", "Approvals required")
}

func TestDryRunMergePR_MissingProject(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.dryRunMergePR(context.Background(), makeReq(map[string]any{
		"slug":  "my-repo",
		"pr_id": float64(7),
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "project")
}

func TestDryRunMergePR_MissingPRID(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleHostConfig, &testhelpers.FakeClient{T: t})
	result, err := h.dryRunMergePR(context.Background(), makeReq(map[string]any{
		"project": "myproj",
		"slug":    "my-repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "pr_id")
}

func TestDryRunMergePR_APIError_ReturnsErrorResult(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		DryRunMergePRFn: func(ns, slug string, prID int, strategy string) (backend.MergeDryRunResult, error) {
			return backend.MergeDryRunResult{}, errors.New("timeout")
		},
	}
	h := newHandlersWithFake(t, singleHostConfig, fake)
	result, err := h.dryRunMergePR(context.Background(), makeReq(map[string]any{
		"project": "myproj",
		"slug":    "my-repo",
		"pr_id":   float64(7),
	}))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}
