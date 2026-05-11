package cloud_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/cloud"
)

func buildDashboardResponse(prs []map[string]any) []byte {
	body, _ := json.Marshal(map[string]any{
		"pagelen": 50,
		"values":  prs,
		"page":    1,
		"size":    len(prs),
	})
	return body
}

func makeDashboardPR(id int, title, fullName, role, state string) map[string]any {
	return map[string]any{
		"id":    id,
		"title": title,
		"state": state,
		"source": map[string]any{
			"repository": map[string]any{"full_name": fullName},
			"commit":     map[string]any{"hash": "abc123"},
		},
		"author": map[string]any{
			"display_name": "Alice",
			"nickname":     "alice",
		},
		"links": map[string]any{
			"html": map[string]any{"href": "https://bitbucket.org/" + fullName + "/pull-requests/" + string(rune('0'+id))},
		},
	}
}

func TestCloudClient_ListMyPRs_ReturnsAuthorAndReviewerPRs(t *testing.T) {
	t.Parallel()

	authorPR := makeDashboardPR(1, "Author PR", "workspace/repo", "AUTHOR", "OPEN")
	reviewerPR := makeDashboardPR(2, "Reviewer PR", "workspace/repo", "REVIEWER", "OPEN")

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		role := r.URL.Query().Get("role")
		switch role {
		case "AUTHOR":
			_, _ = w.Write(buildDashboardResponse([]map[string]any{authorPR}))
		case "REVIEWER":
			_, _ = w.Write(buildDashboardResponse([]map[string]any{reviewerPR}))
		default:
			_, _ = w.Write(buildDashboardResponse(nil))
		}
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	entries, err := client.ListMyPRs("workspace", "repo")
	require.NoError(t, err)
	assert.Len(t, entries, 2)
}

func TestCloudClient_ListMyPRs_AuthorWinsOnConflict(t *testing.T) {
	t.Parallel()

	// Same PR ID appears in both AUTHOR and REVIEWER lists
	authorPR := makeDashboardPR(42, "Author version", "workspace/repo", "AUTHOR", "OPEN")
	reviewerPR := makeDashboardPR(42, "Reviewer version", "workspace/repo", "REVIEWER", "OPEN")

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		role := r.URL.Query().Get("role")
		switch role {
		case "AUTHOR":
			_, _ = w.Write(buildDashboardResponse([]map[string]any{authorPR}))
		case "REVIEWER":
			_, _ = w.Write(buildDashboardResponse([]map[string]any{reviewerPR}))
		default:
			_, _ = w.Write(buildDashboardResponse(nil))
		}
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	entries, err := client.ListMyPRs("workspace", "repo")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "AUTHOR", entries[0].Role)
	assert.Equal(t, "Author version", entries[0].Title)
}

func TestCloudClient_ListMyPRs_QueryIncludesStateOpen(t *testing.T) {
	t.Parallel()
	var gotQueries []string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQueries = append(gotQueries, r.URL.RawQuery)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(buildDashboardResponse(nil))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	_, err := client.ListMyPRs("workspace", "repo")
	require.NoError(t, err)
	for _, q := range gotQueries {
		assert.Contains(t, q, "state=OPEN")
	}
}
