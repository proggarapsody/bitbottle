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

// ── helpers ───────────────────────────────────────────────────────────────────

func fakeCloudCIHandlers(t *testing.T, fake *testhelpers.FakeClient) *handlers {
	t.Helper()
	return newHandlersWithFake(t, singleHostConfig, fake)
}

// ── ListCloudCIReports ────────────────────────────────────────────────────────

func TestMCP_ListCloudCIReports_OK(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		ListCodeInsightsReportsFn: func(project, slug, hash string) ([]backend.CodeInsightsReport, error) {
			assert.Equal(t, "myworkspace", project)
			assert.Equal(t, "myrepo", slug)
			assert.Equal(t, "abc123", hash)
			return []backend.CodeInsightsReport{{Key: "scan-1", Title: "Security Scan", Result: "PASSED"}}, nil
		},
	}
	h := fakeCloudCIHandlers(t, fake)
	result, err := h.listCloudCIReports(context.Background(), makeReq(map[string]any{
		"project": "myworkspace", "slug": "myrepo", "hash": "abc123",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "scan-1", "PASSED")
}

func TestMCP_ListCloudCIReports_MissingHash(t *testing.T) {
	t.Parallel()
	h := fakeCloudCIHandlers(t, &testhelpers.FakeClient{T: t})
	result, err := h.listCloudCIReports(context.Background(), makeReq(map[string]any{
		"project": "ws", "slug": "repo",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "hash")
}

// ── GetCloudCIReport ──────────────────────────────────────────────────────────

func TestMCP_GetCloudCIReport_OK(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		GetCodeInsightsReportFn: func(project, slug, hash, key string) (backend.CodeInsightsReport, error) {
			return backend.CodeInsightsReport{Key: key, Title: "T", Result: "FAILED"}, nil
		},
	}
	h := fakeCloudCIHandlers(t, fake)
	result, err := h.getCloudCIReport(context.Background(), makeReq(map[string]any{
		"project": "ws", "slug": "repo", "hash": "abc", "key": "k1",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "k1", "FAILED")
}

// ── PutCloudCIReport ──────────────────────────────────────────────────────────

func TestMCP_PutCloudCIReport_OK(t *testing.T) {
	t.Parallel()
	var gotIn backend.CodeInsightsReportInput
	fake := &testhelpers.FakeClient{T: t,
		PutCodeInsightsReportFn: func(project, slug, hash, key string, in backend.CodeInsightsReportInput) (backend.CodeInsightsReport, error) {
			gotIn = in
			return backend.CodeInsightsReport{Key: key, Title: in.Title, Result: in.Result}, nil
		},
	}
	h := fakeCloudCIHandlers(t, fake)
	result, err := h.putCloudCIReport(context.Background(), makeReq(map[string]any{
		"project": "ws", "slug": "repo", "hash": "abc", "key": "k1",
		"title": "My Report", "result": "passed",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "My Report", "")
	assert.Equal(t, "PASSED", gotIn.Result)
}

// ── DeleteCloudCIReport ───────────────────────────────────────────────────────

func TestMCP_DeleteCloudCIReport_OK(t *testing.T) {
	t.Parallel()
	var called bool
	fake := &testhelpers.FakeClient{T: t,
		DeleteCodeInsightsReportFn: func(project, slug, hash, key string) error {
			called = true
			return nil
		},
	}
	h := fakeCloudCIHandlers(t, fake)
	result, err := h.deleteCloudCIReport(context.Background(), makeReq(map[string]any{
		"project": "ws", "slug": "repo", "hash": "abc", "key": "k1",
	}))
	require.NoError(t, err)
	assert.True(t, called)
	assertJSONContains(t, result, "deleted", "")
}

// ── ListCloudCIAnnotations ────────────────────────────────────────────────────

func TestMCP_ListCloudCIAnnotations_OK(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		ListCodeInsightsAnnotationsFn: func(project, slug, hash, key string) ([]backend.CodeInsightsAnnotation, error) {
			return []backend.CodeInsightsAnnotation{
				{Path: "main.go", Line: 10, Severity: "HIGH", Message: "null ptr"},
			}, nil
		},
	}
	h := fakeCloudCIHandlers(t, fake)
	result, err := h.listCloudCIAnnotations(context.Background(), makeReq(map[string]any{
		"project": "ws", "slug": "repo", "hash": "abc", "key": "k1",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "main.go", "HIGH")
}

// ── PutCloudCIAnnotations ─────────────────────────────────────────────────────

func TestMCP_PutCloudCIAnnotations_OK(t *testing.T) {
	t.Parallel()
	var gotAnns []backend.CodeInsightsAnnotationInput
	fake := &testhelpers.FakeClient{T: t,
		PutCodeInsightsAnnotationsFn: func(project, slug, hash, key string, in []backend.CodeInsightsAnnotationInput) error {
			gotAnns = in
			return nil
		},
	}
	h := fakeCloudCIHandlers(t, fake)
	annsJSON, _ := json.Marshal([]backend.CodeInsightsAnnotation{
		{Path: "a.go", Line: 5, Message: "issue", Severity: "LOW"},
	})
	result, err := h.putCloudCIAnnotations(context.Background(), makeReq(map[string]any{
		"project": "ws", "slug": "repo", "hash": "abc", "key": "k1",
		"annotations_json": string(annsJSON),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "added", "")
	require.Len(t, gotAnns, 1)
	assert.Equal(t, "a.go", gotAnns[0].Path)
}

func TestMCP_PutCloudCIAnnotations_InvalidJSON(t *testing.T) {
	t.Parallel()
	h := fakeCloudCIHandlers(t, &testhelpers.FakeClient{T: t})
	result, err := h.putCloudCIAnnotations(context.Background(), makeReq(map[string]any{
		"project": "ws", "slug": "repo", "hash": "abc", "key": "k1",
		"annotations_json": "not-json",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "invalid annotations_json")
}
