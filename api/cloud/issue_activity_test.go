package cloud_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/cloud"
)

func TestListIssueActivity(t *testing.T) {
	t.Parallel()
	page := map[string]any{
		"pagelen": 20,
		"values": []any{
			map[string]any{
				"id":         1,
				"kind":       "status",
				"created_on": "2024-01-15T10:00:00+00:00",
				"user": map[string]any{
					"account_id":   "abc",
					"display_name": "Alice",
					"nickname":     "alice",
				},
				"changes": map[string]any{
					"status": map[string]any{"old": "new", "new": "open"},
				},
			},
			map[string]any{
				"id":         2,
				"kind":       "priority",
				"created_on": "2024-01-16T12:00:00+00:00",
				"user": map[string]any{
					"account_id":   "def",
					"display_name": "Bob",
					"nickname":     "bob",
				},
				"changes": map[string]any{
					"priority": map[string]any{"old": "major", "new": "minor"},
				},
			},
		},
	}

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/issues/42/changes")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(page)
	}))
	defer srv.Close()

	c := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	changes, err := c.ListIssueActivity("myws", "myrepo", 42, 0)
	require.NoError(t, err)
	require.Len(t, changes, 2)

	assert.Equal(t, 1, changes[0].ID)
	assert.Equal(t, "status", changes[0].Kind)
	assert.Equal(t, "new", changes[0].OldVal)
	assert.Equal(t, "open", changes[0].NewVal)
	assert.Equal(t, "alice", changes[0].User.Slug)
	assert.Equal(t, "Alice", changes[0].User.DisplayName)
	assert.True(t, changes[0].CreatedOn.Equal(time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)), "CreatedOn mismatch: %v", changes[0].CreatedOn)

	assert.Equal(t, 2, changes[1].ID)
	assert.Equal(t, "priority", changes[1].Kind)
	assert.Equal(t, "major", changes[1].OldVal)
	assert.Equal(t, "minor", changes[1].NewVal)
}
