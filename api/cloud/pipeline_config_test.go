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

// newPipelineConfigServer builds a TLS test server for pipelines_config endpoints.
func newPipelineConfigServer(t *testing.T, handler http.HandlerFunc) *cloud.Client {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return cloud.NewClient(srv.Client(), srv.URL, "tok", "")
}

func TestCloudClient_GetPipelinesConfig(t *testing.T) {
	t.Parallel()
	var seenMethod, seenPath string
	c := newPipelineConfigServer(t, func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"enabled": true,
			"type":    "repository_pipeline_settings",
		})
	})
	got, err := c.GetPipelinesConfig("myworkspace", "my-repo")
	require.NoError(t, err)
	assert.Equal(t, http.MethodGet, seenMethod)
	assert.Contains(t, seenPath, "/repositories/myworkspace/my-repo/pipelines_config")
	assert.True(t, got.Enabled)
}

func TestCloudClient_GetPipelinesConfig_Disabled(t *testing.T) {
	t.Parallel()
	c := newPipelineConfigServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"enabled": false,
			"type":    "repository_pipeline_settings",
		})
	})
	got, err := c.GetPipelinesConfig("myworkspace", "my-repo")
	require.NoError(t, err)
	assert.False(t, got.Enabled)
}

func TestCloudClient_UpdatePipelinesConfig_Enable(t *testing.T) {
	t.Parallel()
	var seenMethod, seenPath string
	var seenBody map[string]any
	c := newPipelineConfigServer(t, func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &seenBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"enabled": true,
			"type":    "repository_pipeline_settings",
		})
	})
	got, err := c.UpdatePipelinesConfig("myworkspace", "my-repo", backend.PipelineConfig{Enabled: true})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPut, seenMethod)
	assert.Contains(t, seenPath, "/repositories/myworkspace/my-repo/pipelines_config")
	assert.Equal(t, true, seenBody["enabled"])
	assert.True(t, got.Enabled)
}

func TestCloudClient_UpdatePipelinesConfig_Disable(t *testing.T) {
	t.Parallel()
	var seenBody map[string]any
	c := newPipelineConfigServer(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &seenBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"enabled": false,
			"type":    "repository_pipeline_settings",
		})
	})
	got, err := c.UpdatePipelinesConfig("myworkspace", "my-repo", backend.PipelineConfig{Enabled: false})
	require.NoError(t, err)
	assert.Equal(t, false, seenBody["enabled"])
	assert.False(t, got.Enabled)
}

func TestCloud_PipelineConfig_ImplementsInterface(t *testing.T) {
	t.Parallel()
	var _ backend.PipelineConfigClient = (*cloud.Client)(nil)
}
