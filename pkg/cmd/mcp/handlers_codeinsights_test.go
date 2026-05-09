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

func fakeCIHandlers(t *testing.T, fake *testhelpers.FakeClient) *handlers {
	t.Helper()
	return newHandlersWithFake(t, singleHostConfig, fake)
}

// ── ListReports ───────────────────────────────────────────────────────────────

func TestMCP_ListCodeInsightsReports_OK(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		ListReportsFn: func(project, slug, hash string) ([]backend.CodeInsightsReport, error) {
			assert.Equal(t, "MYPROJ", project)
			assert.Equal(t, "myrepo", slug)
			assert.Equal(t, "abc123", hash)
			return []backend.CodeInsightsReport{{Key: "tool-1", Title: "Tool 1", Result: "PASS"}}, nil
		},
	}
	h := fakeCIHandlers(t, fake)
	result, err := h.listCodeInsightsReports(context.Background(), makeReq(map[string]any{
		"project": "MYPROJ", "slug": "myrepo", "hash": "abc123",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "tool-1", "PASS")
}

// ── GetReport ────────────────────────────────────────────────────────────────

func TestMCP_GetCodeInsightsReport_OK(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		GetReportFn: func(project, slug, hash, key string) (backend.CodeInsightsReport, error) {
			return backend.CodeInsightsReport{Key: key, Title: "T", Result: "FAIL"}, nil
		},
	}
	h := fakeCIHandlers(t, fake)
	result, err := h.getCodeInsightsReport(context.Background(), makeReq(map[string]any{
		"project": "P", "slug": "s", "hash": "abc", "key": "k1",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "k1", "FAIL")
}

// ── SetReport ────────────────────────────────────────────────────────────────

func TestMCP_SetCodeInsightsReport_OK(t *testing.T) {
	t.Parallel()
	var gotIn backend.CodeInsightsReportInput
	fake := &testhelpers.FakeClient{T: t,
		SetReportFn: func(project, slug, hash, key string, in backend.CodeInsightsReportInput) (backend.CodeInsightsReport, error) {
			gotIn = in
			return backend.CodeInsightsReport{Key: key, Title: in.Title, Result: in.Result}, nil
		},
	}
	h := fakeCIHandlers(t, fake)
	result, err := h.setCodeInsightsReport(context.Background(), makeReq(map[string]any{
		"project": "P", "slug": "s", "hash": "abc", "key": "k1",
		"title": "My Report", "result": "pass",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "My Report", "")
	assert.Equal(t, "PASS", gotIn.Result)
}

// ── DeleteReport ─────────────────────────────────────────────────────────────

func TestMCP_DeleteCodeInsightsReport_OK(t *testing.T) {
	t.Parallel()
	var called bool
	fake := &testhelpers.FakeClient{T: t,
		DeleteReportFn: func(project, slug, hash, key string) error {
			called = true
			return nil
		},
	}
	h := fakeCIHandlers(t, fake)
	result, err := h.deleteCodeInsightsReport(context.Background(), makeReq(map[string]any{
		"project": "P", "slug": "s", "hash": "abc", "key": "k1",
	}))
	require.NoError(t, err)
	assert.True(t, called)
	assertJSONContains(t, result, "deleted", "")
}

// ── ListAnnotations ───────────────────────────────────────────────────────────

func TestMCP_ListCodeInsightsAnnotations_OK(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		ListAnnotationsFn: func(project, slug, hash, key string) ([]backend.CodeInsightsAnnotation, error) {
			return []backend.CodeInsightsAnnotation{
				{Path: "main.go", Line: 10, Severity: "HIGH", Message: "null ptr"},
			}, nil
		},
	}
	h := fakeCIHandlers(t, fake)
	result, err := h.listCodeInsightsAnnotations(context.Background(), makeReq(map[string]any{
		"project": "P", "slug": "s", "hash": "abc", "key": "k1",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "main.go", "HIGH")
}

// ── AddAnnotations ────────────────────────────────────────────────────────────

func TestMCP_AddCodeInsightsAnnotations_OK(t *testing.T) {
	t.Parallel()
	var gotAnns []backend.CodeInsightsAnnotationInput
	fake := &testhelpers.FakeClient{T: t,
		AddAnnotationsFn: func(project, slug, hash, key string, in []backend.CodeInsightsAnnotationInput) error {
			gotAnns = in
			return nil
		},
	}
	h := fakeCIHandlers(t, fake)
	annsJSON, _ := json.Marshal([]backend.CodeInsightsAnnotation{
		{Path: "a.go", Line: 5, Message: "issue", Severity: "LOW"},
	})
	result, err := h.addCodeInsightsAnnotations(context.Background(), makeReq(map[string]any{
		"project": "P", "slug": "s", "hash": "abc", "key": "k1",
		"annotations_json": string(annsJSON),
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "added", "")
	require.Len(t, gotAnns, 1)
	assert.Equal(t, "a.go", gotAnns[0].Path)
}

func TestMCP_AddCodeInsightsAnnotations_InvalidJSON(t *testing.T) {
	t.Parallel()
	h := fakeCIHandlers(t, &testhelpers.FakeClient{T: t})
	result, err := h.addCodeInsightsAnnotations(context.Background(), makeReq(map[string]any{
		"project": "P", "slug": "s", "hash": "abc", "key": "k1",
		"annotations_json": "not-json",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "invalid annotations_json")
}

// ── DeleteAnnotations ─────────────────────────────────────────────────────────

func TestMCP_DeleteCodeInsightsAnnotations_OK(t *testing.T) {
	t.Parallel()
	var called bool
	fake := &testhelpers.FakeClient{T: t,
		DeleteAnnotationsFn: func(project, slug, hash, key string) error {
			called = true
			return nil
		},
	}
	h := fakeCIHandlers(t, fake)
	result, err := h.deleteCodeInsightsAnnotations(context.Background(), makeReq(map[string]any{
		"project": "P", "slug": "s", "hash": "abc", "key": "k1",
	}))
	require.NoError(t, err)
	assert.True(t, called)
	assertJSONContains(t, result, "deleted", "")
}

// ── SetMergeCheck ─────────────────────────────────────────────────────────────

func TestMCP_SetMergeCheck_OK(t *testing.T) {
	t.Parallel()
	var gotIn backend.MergeCheckInput
	fake := &testhelpers.FakeClient{T: t,
		SetMergeCheckFn: func(project, slug, key string, in backend.MergeCheckInput) error {
			gotIn = in
			return nil
		},
	}
	h := fakeCIHandlers(t, fake)
	result, err := h.setMergeCheck(context.Background(), makeReq(map[string]any{
		"project": "P", "slug": "s", "key": "check-1",
		"report_key": "my-tool", "must_pass": true, "min_severity": "medium",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "ok", "")
	assert.Equal(t, "my-tool", gotIn.ReportKey)
	assert.True(t, gotIn.MustPass)
	assert.Equal(t, "MEDIUM", gotIn.MinSeverity)
}

// ── GetMergeCheck ─────────────────────────────────────────────────────────────

func TestMCP_GetMergeCheck_OK(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t,
		GetMergeCheckFn: func(project, slug, key string) (backend.MergeCheck, error) {
			return backend.MergeCheck{Key: key, ReportKey: "r1", MustPass: true}, nil
		},
	}
	h := fakeCIHandlers(t, fake)
	result, err := h.getMergeCheck(context.Background(), makeReq(map[string]any{
		"project": "P", "slug": "s", "key": "check-1",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "check-1", "r1")
}

// ── DeleteMergeCheck ─────────────────────────────────────────────────────────

func TestMCP_DeleteMergeCheck_OK(t *testing.T) {
	t.Parallel()
	var called bool
	fake := &testhelpers.FakeClient{T: t,
		DeleteMergeCheckFn: func(project, slug, key string) error {
			called = true
			return nil
		},
	}
	h := fakeCIHandlers(t, fake)
	result, err := h.deleteMergeCheck(context.Background(), makeReq(map[string]any{
		"project": "P", "slug": "s", "key": "check-1",
	}))
	require.NoError(t, err)
	assert.True(t, called)
	assertJSONContains(t, result, "deleted", "")
}
