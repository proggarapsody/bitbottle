package cloud_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

func newCloudCIServer(t *testing.T, handler http.HandlerFunc) *cloud.Client {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return cloud.NewClient(srv.Client(), srv.URL, "tok", "")
}

// ── ListCodeInsightsReports ───────────────────────────────────────────────────

func TestCloud_ListCodeInsightsReports(t *testing.T) {
	t.Parallel()
	var gotPath string
	client := newCloudCIServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"values": []any{
				map[string]any{
					"external_id": "tool-1",
					"title":       "Security Scan",
					"result":      "PASSED",
					"report_type": "SECURITY",
				},
			},
		})
	})

	reports, err := client.ListCodeInsightsReports("acme", "myrepo", "abc123")
	require.NoError(t, err)
	assert.Equal(t, "/repositories/acme/myrepo/commit/abc123/reports", gotPath)
	require.Len(t, reports, 1)
	assert.Equal(t, "tool-1", reports[0].Key)
	assert.Equal(t, "Security Scan", reports[0].Title)
	assert.Equal(t, "PASSED", reports[0].Result)
	assert.Equal(t, "SECURITY", reports[0].ReportType)
}

// ── GetCodeInsightsReport ─────────────────────────────────────────────────────

func TestCloud_GetCodeInsightsReport(t *testing.T) {
	t.Parallel()
	var gotPath string
	client := newCloudCIServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"external_id": "my-tool",
			"title":       "My Tool",
			"result":      "FAILED",
			"reporter":    "ci-agent",
		})
	})

	r, err := client.GetCodeInsightsReport("acme", "myrepo", "abc123", "my-tool")
	require.NoError(t, err)
	assert.Equal(t, "/repositories/acme/myrepo/commit/abc123/reports/my-tool", gotPath)
	assert.Equal(t, "my-tool", r.Key)
	assert.Equal(t, "FAILED", r.Result)
	assert.Equal(t, "ci-agent", r.Reporter)
}

// ── PutCodeInsightsReport ─────────────────────────────────────────────────────

func TestCloud_PutCodeInsightsReport(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	var gotBody map[string]any
	client := newCloudCIServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"external_id": "my-tool",
			"title":       "My Tool",
			"result":      "PASSED",
		})
	})

	in := backend.CodeInsightsReportInput{
		Title:      "My Tool",
		Result:     "PASSED",
		ReportType: "TESTING",
		Reporter:   "ci-agent",
	}
	rep, err := client.PutCodeInsightsReport("acme", "myrepo", "abc123", "my-tool", in)
	require.NoError(t, err)
	assert.Equal(t, http.MethodPut, gotMethod)
	assert.Equal(t, "/repositories/acme/myrepo/commit/abc123/reports/my-tool", gotPath)
	assert.Equal(t, "My Tool", gotBody["title"])
	assert.Equal(t, "PASSED", gotBody["result"])
	assert.Equal(t, "my-tool", rep.Key)
}

// ── DeleteCodeInsightsReport ──────────────────────────────────────────────────

func TestCloud_DeleteCodeInsightsReport(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	client := newCloudCIServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})

	err := client.DeleteCodeInsightsReport("acme", "myrepo", "abc123", "my-tool")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/repositories/acme/myrepo/commit/abc123/reports/my-tool", gotPath)
}

// ── ListCodeInsightsAnnotations ───────────────────────────────────────────────

func TestCloud_ListCodeInsightsAnnotations(t *testing.T) {
	t.Parallel()
	var gotPath string
	client := newCloudCIServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"values": []any{
				map[string]any{
					"path":        "main.go",
					"line":        42,
					"summary":     "null ptr",
					"severity":    "HIGH",
					"type":        "BUG",
					"external_id": "ann-1",
				},
			},
		})
	})

	anns, err := client.ListCodeInsightsAnnotations("acme", "myrepo", "abc123", "my-tool")
	require.NoError(t, err)
	assert.Equal(t, "/repositories/acme/myrepo/commit/abc123/reports/my-tool/annotations", gotPath)
	require.Len(t, anns, 1)
	assert.Equal(t, "main.go", anns[0].Path)
	assert.Equal(t, 42, anns[0].Line)
	assert.Equal(t, "null ptr", anns[0].Message)
	assert.Equal(t, "HIGH", anns[0].Severity)
	assert.Equal(t, "ann-1", anns[0].ExternalID)
}

// ── PutCodeInsightsAnnotations ────────────────────────────────────────────────

func TestCloud_PutCodeInsightsAnnotations(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	var gotBody map[string]any
	client := newCloudCIServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
	})

	in := []backend.CodeInsightsAnnotationInput{
		{Path: "main.go", Line: 42, Message: "null ptr", Severity: "HIGH", Type: "BUG"},
	}
	err := client.PutCodeInsightsAnnotations("acme", "myrepo", "abc123", "my-tool", in)
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/repositories/acme/myrepo/commit/abc123/reports/my-tool/annotations", gotPath)
	anns, ok := gotBody["annotations"].([]any)
	require.True(t, ok, "expected annotations key in body")
	require.Len(t, anns, 1)
	ann := anns[0].(map[string]any)
	assert.Equal(t, "main.go", ann["path"])
}

// ── path-escaping ─────────────────────────────────────────────────────────────

func TestCloud_CodeInsights_PathEscaping(t *testing.T) {
	t.Parallel()
	var gotRawPath string
	client := newCloudCIServer(t, func(w http.ResponseWriter, r *http.Request) {
		// r.URL.RawPath preserves percent-encoding from the raw request URI.
		// Fall back to r.URL.Path when no encoding is needed.
		gotRawPath = r.URL.RawPath
		if gotRawPath == "" {
			gotRawPath = r.URL.Path
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// key segment that contains a slash must be percent-encoded so that
	// "tool/subkey" is a single path segment, not two.
	err := client.DeleteCodeInsightsReport("acme", "myrepo", "abc123", "tool/subkey")
	require.NoError(t, err)
	assert.Contains(t, gotRawPath, "tool%2Fsubkey")
}
