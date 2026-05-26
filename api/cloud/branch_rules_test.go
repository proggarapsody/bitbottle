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

func newCloudBranchRuleServer(t *testing.T, handler http.HandlerFunc) *cloud.Client {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return cloud.NewClient(srv.Client(), srv.URL, "tok", "")
}

const cloudBranchRuleListJSON = `{
  "values": [
    {"id":1,"kind":"require_approvals_to_merge","pattern":"main","value":2},
    {"id":2,"kind":"push","pattern":"main"}
  ]
}`

func TestCloudClient_ListBranchRules(t *testing.T) {
	t.Parallel()
	client := newCloudBranchRuleServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repositories/myws/my-repo/branch-restrictions", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cloudBranchRuleListJSON))
	})
	rules, err := client.ListBranchRules("myws", "my-repo")
	require.NoError(t, err)
	require.Len(t, rules, 2)
	assert.Equal(t, 1, rules[0].ID)
	assert.Equal(t, "require_approvals_to_merge", rules[0].Kind)
	assert.Equal(t, "main", rules[0].Pattern)
	assert.Equal(t, 2, rules[0].Value)
	assert.Equal(t, 2, rules[1].ID)
	assert.Equal(t, "push", rules[1].Kind)
}

func TestCloudClient_AddBranchRule(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	client := newCloudBranchRuleServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/repositories/myws/my-repo/branch-restrictions", r.URL.Path)
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":3,"kind":"require_approvals_to_merge","pattern":"main","value":1}`))
	})
	rule, err := client.AddBranchRule("myws", "my-repo", backend.BranchRuleInput{
		Kind:    "require_approvals_to_merge",
		Pattern: "main",
		Value:   1,
	})
	require.NoError(t, err)
	assert.Equal(t, 3, rule.ID)
	assert.Equal(t, "require_approvals_to_merge", rule.Kind)
	assert.Equal(t, "main", rule.Pattern)
	assert.Equal(t, 1, rule.Value)
	assert.Equal(t, "require_approvals_to_merge", gotBody["kind"])
	assert.Equal(t, "main", gotBody["pattern"])
}

func TestCloudClient_DeleteBranchRule(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	client := newCloudBranchRuleServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})
	err := client.DeleteBranchRule("myws", "my-repo", 7)
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/repositories/myws/my-repo/branch-restrictions/7", gotPath)
}

func TestCloudClient_UpdateBranchRule(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	var gotBody map[string]any
	// Two requests: GET (fetch current) then PUT (apply patch)
	reqCount := 0
	client := newCloudBranchRuleServer(t, func(w http.ResponseWriter, r *http.Request) {
		reqCount++
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			gotMethod = r.Method
			gotPath = r.URL.Path
			_, _ = w.Write([]byte(`{"id":5,"kind":"push","pattern":"main","value":0}`))
			return
		}
		assert.Equal(t, http.MethodPut, r.Method)
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		_, _ = w.Write([]byte(`{"id":5,"kind":"push","pattern":"release/*","value":0}`))
	})
	newPattern := "release/*"
	rule, err := client.UpdateBranchRule("myws", "my-repo", 5, backend.UpdateBranchRuleInput{
		Pattern: &newPattern,
	})
	require.NoError(t, err)
	assert.Equal(t, 5, rule.ID)
	assert.Equal(t, "push", rule.Kind)
	assert.Equal(t, "release/*", rule.Pattern)
	assert.Equal(t, http.MethodGet, gotMethod)
	assert.Equal(t, "/repositories/myws/my-repo/branch-restrictions/5", gotPath)
	assert.Equal(t, 2, reqCount)
	assert.Equal(t, "release/*", gotBody["pattern"])
	assert.Equal(t, "push", gotBody["kind"])
}
