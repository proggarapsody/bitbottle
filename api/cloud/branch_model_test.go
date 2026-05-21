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

func newCloudBranchModelServer(t *testing.T, handler http.HandlerFunc) *cloud.Client {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return cloud.NewClient(srv.Client(), srv.URL, "tok", "")
}

const cloudBranchModelJSON = `{
  "development": {"name": "main", "is_valid": true, "use_mainbranch": true},
  "production": {"name": "production", "is_valid": true, "use_mainbranch": false},
  "branch_types": [
    {"kind": "feature", "prefix": "feature/"},
    {"kind": "hotfix", "prefix": "hotfix/"}
  ]
}`

const cloudBranchModelSettingsJSON = `{
  "development": {"name": "main", "is_valid": true, "use_mainbranch": true},
  "production": {"name": "production", "is_valid": true, "use_mainbranch": false},
  "branch_types": [
    {"enabled": true, "kind": "feature", "prefix": "feature/"},
    {"enabled": true, "kind": "hotfix", "prefix": "hotfix/"}
  ]
}`

func TestCloudClient_GetBranchModel(t *testing.T) {
	t.Parallel()
	client := newCloudBranchModelServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/repositories/myws/my-repo/branching-model", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cloudBranchModelJSON))
	})
	model, err := client.GetBranchModel("myws", "my-repo")
	require.NoError(t, err)
	assert.Equal(t, "main", model.Development.Name)
	assert.True(t, model.Development.UseMainbranch)
	require.NotNil(t, model.Production)
	assert.Equal(t, "production", model.Production.Name)
	require.Len(t, model.BranchTypes, 2)
	assert.Equal(t, "feature", model.BranchTypes[0].Kind)
	assert.Equal(t, "feature/", model.BranchTypes[0].Prefix)
}

func TestCloudClient_GetBranchModelSettings(t *testing.T) {
	t.Parallel()
	client := newCloudBranchModelServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/repositories/myws/my-repo/branching-model/settings", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cloudBranchModelSettingsJSON))
	})
	settings, err := client.GetBranchModelSettings("myws", "my-repo")
	require.NoError(t, err)
	assert.Equal(t, "main", settings.Development.Name)
	assert.True(t, settings.Development.UseMainbranch)
	require.Len(t, settings.BranchTypes, 2)
	assert.True(t, settings.BranchTypes[0].Enabled)
	assert.Equal(t, "hotfix/", settings.BranchTypes[1].Prefix)
}

func TestCloudClient_UpdateBranchModelSettings(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	client := newCloudBranchModelServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/repositories/myws/my-repo/branching-model/settings", r.URL.Path)
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cloudBranchModelSettingsJSON))
	})
	in := backend.BranchModelSettingsInput{
		Development: &backend.BranchModelSettingsBranch{
			UseMainbranch: true,
		},
		BranchTypes: []backend.BranchTypeSettings{
			{Enabled: true, Kind: "feature", Prefix: "feat/"},
		},
	}
	settings, err := client.UpdateBranchModelSettings("myws", "my-repo", in)
	require.NoError(t, err)
	assert.Equal(t, "main", settings.Development.Name)
	assert.NotNil(t, gotBody["development"])
	assert.NotNil(t, gotBody["branch_types"])
}

func TestCloudClient_GetBranchModel_EmptyArgs(t *testing.T) {
	t.Parallel()
	client := newCloudBranchModelServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach server")
	})
	_, err := client.GetBranchModel("", "my-repo")
	require.Error(t, err)
}
