package server_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// A Server PR PUT replaces the resource. The fake rejects a partial body so
// the title-only user action cannot silently clear its description or reviewers.
func TestServerClient_UpdatePR_PreservesFullPR(t *testing.T) {
	t.Parallel()
	const original = `{"id":42,"version":7,"title":"Original title","description":"Keep this description","state":"OPEN","draft":false,"author":{"user":{"slug":"author"}},"fromRef":{"id":"refs/heads/feature/qa","displayId":"feature/qa"},"toRef":{"id":"refs/heads/main","displayId":"main"},"reviewers":[{"user":{"slug":"reviewer-one"}}]}`
	var put map[string]any
	var methods []string
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			_, _ = io.WriteString(w, original)
			return
		}
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		data, _ := io.ReadAll(r.Body)
		if json.Unmarshal(data, &put) != nil || put["version"] != float64(7) ||
			put["title"] != "Updated title" || put["description"] != "Keep this description" ||
			put["state"] != "OPEN" || put["fromRef"] == nil || put["toRef"] == nil || put["reviewers"] == nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"errors":[{"message":"full PR body required"}]}`)
			return
		}
		_, _ = io.WriteString(w, original)
	})

	_, err := client.UpdatePR("TEAM", "service", 42, backend.UpdatePRInput{Title: "Updated title"})
	require.NoError(t, err)
	require.Equal(t, []string{http.MethodGet, http.MethodPut}, methods)
	require.Equal(t, "refs/heads/feature/qa", put["fromRef"].(map[string]any)["id"])
	require.Equal(t, "refs/heads/main", put["toRef"].(map[string]any)["id"])
	reviewers := put["reviewers"].([]any)
	require.Len(t, reviewers, 1)
	require.Equal(t, "reviewer-one", reviewers[0].(map[string]any)["user"].(map[string]any)["slug"])
}
