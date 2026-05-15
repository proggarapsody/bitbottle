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

func newServerReviewerGroupClient(t *testing.T, handler http.HandlerFunc) *server.Client {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return server.NewClient(srv.Client(), srv.URL, "tok", "")
}

const serverConditionsListJSON = `{
  "values": [
    {
      "id": 1,
      "sourceMatcher": {"id": "ANY_REF_MATCHER_ID", "displayId": "any-branch", "type": {"id": "ANY_REF"}},
      "targetMatcher": {"id": "ANY_REF_MATCHER_ID", "displayId": "any-branch", "type": {"id": "ANY_REF"}},
      "reviewers": [
        {"slug": "alice", "displayName": "Alice"},
        {"slug": "bob",   "displayName": "Bob"}
      ],
      "requiredApprovals": 1
    }
  ],
  "isLastPage": true
}`

func TestServerClient_ListReviewerGroups_Path(t *testing.T) {
	t.Parallel()
	var gotPath string
	client := newServerReviewerGroupClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[],"isLastPage":true}`))
	})
	_, err := client.ListReviewerGroups("PROJ", "my-repo")
	require.NoError(t, err)
	assert.Equal(t, "/rest/default-reviewers/1.0/projects/PROJ/repos/my-repo/conditions", gotPath)
}

func TestServerClient_ListReviewerGroups_Maps(t *testing.T) {
	t.Parallel()
	client := newServerReviewerGroupClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(serverConditionsListJSON))
	})
	groups, err := client.ListReviewerGroups("PROJ", "my-repo")
	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, 1, groups[0].ID)
	assert.Equal(t, "any-branch", groups[0].Name)
	assert.Equal(t, 1, groups[0].RequiredApprovals)
	require.Len(t, groups[0].Reviewers, 2)
	assert.Equal(t, "alice", groups[0].Reviewers[0].Slug)
	assert.Equal(t, "Alice", groups[0].Reviewers[0].DisplayName)
}

func TestServerClient_CreateReviewerGroup(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	var gotBody map[string]any
	client := newServerReviewerGroupClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": 42,
			"sourceMatcher": {"id": "my-group", "displayId": "my-group", "type": {"id": "ANY_REF"}},
			"targetMatcher": {"id": "ANY_REF_MATCHER_ID", "displayId": "", "type": {"id": "ANY_REF"}},
			"reviewers": [{"slug": "alice", "displayName": "Alice"}],
			"requiredApprovals": 1
		}`))
	})

	group, err := client.CreateReviewerGroup("PROJ", "my-repo", backend.CreateReviewerGroupInput{
		Name:              "my-group",
		UserSlugs:         []string{"alice"},
		RequiredApprovals: 1,
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/rest/default-reviewers/1.0/projects/PROJ/repos/my-repo/conditions", gotPath)
	assert.Equal(t, 42, group.ID)
	assert.Equal(t, "my-group", group.Name)
	// Verify body has sourceMatcher with correct id
	sm := gotBody["sourceMatcher"].(map[string]any)
	assert.Equal(t, "my-group", sm["id"])
}

func TestServerClient_DeleteReviewerGroup(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	client := newServerReviewerGroupClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})
	err := client.DeleteReviewerGroup("PROJ", "my-repo", 7)
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/rest/default-reviewers/1.0/projects/PROJ/repos/my-repo/conditions/7", gotPath)
}
