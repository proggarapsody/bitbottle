package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/server"
)

// ── fixtures ──────────────────────────────────────────────────────────────────

const reportsJSON = `{
  "values": [
    {
      "key": "my-tool",
      "title": "My Tool",
      "result": "PASS",
      "reportType": "TESTING",
      "details": "all good",
      "reporter": "ci-bot",
      "link": "https://ci.example.com/build/1",
      "logoUrl": "https://example.com/logo.png",
      "data": [{"title":"Coverage","type":"PERCENTAGE","value":87.5}],
      "createdDate": 1700000000000,
      "updatedDate": 1700000001000
    }
  ],
  "size": 1,
  "isLastPage": true,
  "start": 0
}`

const singleReportJSON = `{
  "key": "my-tool",
  "title": "My Tool",
  "result": "PASS",
  "reportType": "TESTING",
  "details": "all good",
  "reporter": "ci-bot",
  "link": "https://ci.example.com/build/1",
  "logoUrl": "https://example.com/logo.png",
  "data": [],
  "createdDate": 1700000000000,
  "updatedDate": 1700000001000
}`

const annotationsJSON = `{
  "annotations": [
    {
      "externalId": "ann-1",
      "path": "src/main.go",
      "line": 42,
      "message": "potential null dereference",
      "severity": "HIGH",
      "type": "BUG",
      "link": "https://linter.example.com/rules/NPE"
    }
  ]
}`

const mergeCheckJSON = `{
  "key": "check-1",
  "reportKey": "my-tool",
  "mustPass": true,
  "minSeverity": "MEDIUM"
}`

// ── helpers ───────────────────────────────────────────────────────────────────

func newInsightsClient(t *testing.T, handler http.Handler) *server.Client {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return server.NewClient(srv.Client(), srv.URL, "tok", "alice")
}

// ── ListReports ───────────────────────────────────────────────────────────────

func TestServerClient_ListReports(t *testing.T) {
	t.Parallel()
	var seenPath string
	client := newInsightsClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(reportsJSON))
	}))

	got, err := client.ListReports("PROJ", "repo", "abc123")
	require.NoError(t, err)
	require.Len(t, got, 1)

	assert.Equal(t, "/rest/insights/1.0/projects/PROJ/repos/repo/commits/abc123/reports", seenPath)
	r := got[0]
	assert.Equal(t, "my-tool", r.Key)
	assert.Equal(t, "My Tool", r.Title)
	assert.Equal(t, "PASS", r.Result)
	assert.Equal(t, "TESTING", r.ReportType)
	assert.Equal(t, "all good", r.Details)
	assert.Equal(t, "ci-bot", r.Reporter)
	require.Len(t, r.Data, 1)
	assert.Equal(t, "Coverage", r.Data[0].Title)
	assert.Equal(t, "PERCENTAGE", r.Data[0].Type)
}

// ── GetReport ────────────────────────────────────────────────────────────────

func TestServerClient_GetReport(t *testing.T) {
	t.Parallel()
	var seenPath string
	client := newInsightsClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(singleReportJSON))
	}))

	got, err := client.GetReport("PROJ", "repo", "abc123", "my-tool")
	require.NoError(t, err)
	assert.Equal(t, "/rest/insights/1.0/projects/PROJ/repos/repo/commits/abc123/reports/my-tool", seenPath)
	assert.Equal(t, "my-tool", got.Key)
	assert.Equal(t, "PASS", got.Result)
}

// ── SetReport ────────────────────────────────────────────────────────────────

func TestServerClient_SetReport(t *testing.T) {
	t.Parallel()
	var seenPath, seenMethod string
	var seenBody map[string]any
	client := newInsightsClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		seenMethod = r.Method
		_ = json.NewDecoder(r.Body).Decode(&seenBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(singleReportJSON))
	}))

	in := backend.CodeInsightsReportInput{
		Title:  "My Tool",
		Result: "PASS",
	}
	got, err := client.SetReport("PROJ", "repo", "abc123", "my-tool", in)
	require.NoError(t, err)
	assert.Equal(t, "PUT", seenMethod)
	assert.Equal(t, "/rest/insights/1.0/projects/PROJ/repos/repo/commits/abc123/reports/my-tool", seenPath)
	assert.Equal(t, "My Tool", seenBody["title"])
	assert.Equal(t, "PASS", seenBody["result"])
	assert.Equal(t, "my-tool", got.Key)
}

// ── DeleteReport ─────────────────────────────────────────────────────────────

func TestServerClient_DeleteReport(t *testing.T) {
	t.Parallel()
	var seenPath, seenMethod string
	client := newInsightsClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		seenMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))

	err := client.DeleteReport("PROJ", "repo", "abc123", "my-tool")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", seenMethod)
	assert.Equal(t, "/rest/insights/1.0/projects/PROJ/repos/repo/commits/abc123/reports/my-tool", seenPath)
}

// ── ListAnnotations ───────────────────────────────────────────────────────────

func TestServerClient_ListAnnotations(t *testing.T) {
	t.Parallel()
	var seenPath string
	client := newInsightsClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(annotationsJSON))
	}))

	got, err := client.ListAnnotations("PROJ", "repo", "abc123", "my-tool")
	require.NoError(t, err)
	assert.Equal(t, "/rest/insights/1.0/projects/PROJ/repos/repo/commits/abc123/reports/my-tool/annotations", seenPath)
	require.Len(t, got, 1)
	a := got[0]
	assert.Equal(t, "ann-1", a.ExternalID)
	assert.Equal(t, "src/main.go", a.Path)
	assert.Equal(t, 42, a.Line)
	assert.Equal(t, "HIGH", a.Severity)
	assert.Equal(t, "BUG", a.Type)
}

// ── AddAnnotations ────────────────────────────────────────────────────────────

func TestServerClient_AddAnnotations(t *testing.T) {
	t.Parallel()
	var seenPath, seenMethod string
	var seenBody map[string]any
	client := newInsightsClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		seenMethod = r.Method
		_ = json.NewDecoder(r.Body).Decode(&seenBody)
		w.WriteHeader(http.StatusNoContent)
	}))

	in := []backend.CodeInsightsAnnotationInput{
		{Path: "src/main.go", Line: 10, Message: "issue", Severity: "LOW"},
	}
	err := client.AddAnnotations("PROJ", "repo", "abc123", "my-tool", in)
	require.NoError(t, err)
	assert.Equal(t, "POST", seenMethod)
	assert.Equal(t, "/rest/insights/1.0/projects/PROJ/repos/repo/commits/abc123/reports/my-tool/annotations", seenPath)
	anns, ok := seenBody["annotations"].([]any)
	require.True(t, ok)
	require.Len(t, anns, 1)
	ann := anns[0].(map[string]any)
	assert.Equal(t, "src/main.go", ann["path"])
}

// ── DeleteAnnotations ─────────────────────────────────────────────────────────

func TestServerClient_DeleteAnnotations(t *testing.T) {
	t.Parallel()
	var seenPath, seenMethod string
	client := newInsightsClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		seenMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))

	err := client.DeleteAnnotations("PROJ", "repo", "abc123", "my-tool")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", seenMethod)
	assert.Equal(t, "/rest/insights/1.0/projects/PROJ/repos/repo/commits/abc123/reports/my-tool/annotations", seenPath)
}

// ── SetMergeCheck ────────────────────────────────────────────────────────────

func TestServerClient_SetMergeCheck(t *testing.T) {
	t.Parallel()
	var seenPath, seenMethod string
	var seenBody map[string]any
	client := newInsightsClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		seenMethod = r.Method
		_ = json.NewDecoder(r.Body).Decode(&seenBody)
		w.WriteHeader(http.StatusNoContent)
	}))

	in := backend.MergeCheckInput{
		Key:         "check-1",
		ReportKey:   "my-tool",
		MustPass:    true,
		MinSeverity: "MEDIUM",
	}
	err := client.SetMergeCheck("PROJ", "repo", "check-1", in)
	require.NoError(t, err)
	assert.Equal(t, "PUT", seenMethod)
	assert.Equal(t, "/rest/insights/latest/projects/PROJ/repos/repo/settings/merge-checks/check-1", seenPath)
	assert.Equal(t, "my-tool", seenBody["reportKey"])
	assert.Equal(t, true, seenBody["mustPass"])
}

// ── GetMergeCheck ─────────────────────────────────────────────────────────────

func TestServerClient_GetMergeCheck(t *testing.T) {
	t.Parallel()
	var seenPath string
	client := newInsightsClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(mergeCheckJSON))
	}))

	got, err := client.GetMergeCheck("PROJ", "repo", "check-1")
	require.NoError(t, err)
	assert.Equal(t, "/rest/insights/latest/projects/PROJ/repos/repo/settings/merge-checks/check-1", seenPath)
	assert.Equal(t, "check-1", got.Key)
	assert.Equal(t, "my-tool", got.ReportKey)
	assert.True(t, got.MustPass)
	assert.Equal(t, "MEDIUM", got.MinSeverity)
}

// ── DeleteMergeCheck ─────────────────────────────────────────────────────────

func TestServerClient_DeleteMergeCheck(t *testing.T) {
	t.Parallel()
	var seenPath, seenMethod string
	client := newInsightsClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		seenMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))

	err := client.DeleteMergeCheck("PROJ", "repo", "check-1")
	require.NoError(t, err)
	assert.Equal(t, "DELETE", seenMethod)
	assert.Equal(t, "/rest/insights/latest/projects/PROJ/repos/repo/settings/merge-checks/check-1", seenPath)
}
