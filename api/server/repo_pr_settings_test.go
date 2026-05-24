package server_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/server"
)

const prSettingsResponseJSON = `{
  "requiredApprovers": 2,
  "requiredAllApprovers": true,
  "requiredAllTasksComplete": false,
  "requiredSuccessfulBuilds": 1,
  "mergeConfig": {
    "defaultStrategy": {"id": "no-ff"},
    "strategies": [{"id": "no-ff"}, {"id": "squash"}]
  }
}`

func TestServerClient_GetRepoPRSettings(t *testing.T) {
	t.Parallel()
	var seenPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(prSettingsResponseJSON))
	}))
	t.Cleanup(srv.Close)

	client := server.NewClient(srv.Client(), srv.URL, "tok", "alice")
	got, err := client.GetRepoPRSettings("MYPROJ", "my-repo")
	require.NoError(t, err)

	assert.Equal(t, "/projects/MYPROJ/repos/my-repo/settings/pull-requests", seenPath)
	assert.Equal(t, 2, got.RequiredApprovers)
	assert.True(t, got.RequiredAllApprovers)
	assert.False(t, got.RequiredAllTasksComplete)
	assert.Equal(t, 1, got.RequiredSuccessfulBuilds)
	assert.Equal(t, "no-ff", got.MergeStrategy)
	assert.Equal(t, []string{"no-ff", "squash"}, got.AllowedStrategies)
}

func TestServerClient_UpdateRepoPRSettings_PartialUpdate(t *testing.T) {
	t.Parallel()
	var seenPOSTBody []byte
	var seenPOSTPath string
	var getCount int

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			getCount++
			_, _ = w.Write([]byte(prSettingsResponseJSON))
			return
		}
		// POST
		seenPOSTPath = r.URL.Path
		seenPOSTBody, _ = io.ReadAll(r.Body)
		// Return the updated state
		_, _ = w.Write([]byte(`{
  "requiredApprovers": 3,
  "requiredAllApprovers": true,
  "requiredAllTasksComplete": false,
  "requiredSuccessfulBuilds": 1,
  "mergeConfig": {
    "defaultStrategy": {"id": "no-ff"},
    "strategies": [{"id": "no-ff"}, {"id": "squash"}]
  }
}`))
	}))
	t.Cleanup(srv.Close)

	approvers := 3
	client := server.NewClient(srv.Client(), srv.URL, "tok", "alice")
	got, err := client.UpdateRepoPRSettings("MYPROJ", "my-repo", backend.RepoPRSettingsInput{
		RequiredApprovers: &approvers,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, getCount)
	assert.Equal(t, "/projects/MYPROJ/repos/my-repo/settings/pull-requests", seenPOSTPath)
	assert.Equal(t, 3, got.RequiredApprovers)

	// Verify the POST body contains the merged fields.
	var body map[string]any
	require.NoError(t, json.Unmarshal(seenPOSTBody, &body))
	assert.Equal(t, float64(3), body["requiredApprovers"])
	// requiredAllApprovers should still be true (from GET response).
	assert.Equal(t, true, body["requiredAllApprovers"])
}

func TestServerClient_UpdateRepoPRSettings_StrategiesUpdate(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(prSettingsResponseJSON))
			return
		}
		body, _ := io.ReadAll(r.Body)
		// Echo the body back so the domain mapping is exercised.
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	strats := []string{"squash", "merge-commit"}
	client := server.NewClient(srv.Client(), srv.URL, "tok", "alice")
	got, err := client.UpdateRepoPRSettings("MYPROJ", "my-repo", backend.RepoPRSettingsInput{
		AllowedStrategies: &strats,
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"squash", "merge-commit"}, got.AllowedStrategies)
}

// TestServer_RepoPRSettings_ImplementsInterface verifies that the Server
// client satisfies RepoPRSettingsClient at compile time.
func TestServer_RepoPRSettings_ImplementsInterface(t *testing.T) {
	t.Parallel()
	var _ backend.RepoPRSettingsClient = (*server.Client)(nil)
}
