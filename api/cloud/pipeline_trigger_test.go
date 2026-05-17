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

func newCloudTriggerServer(t *testing.T, handler http.HandlerFunc) *cloud.Client {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return cloud.NewClient(srv.Client(), srv.URL, "tok", "")
}

func TestCloudClient_TriggerPipeline_Success(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	client := newCloudTriggerServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/repositories/myws/my-repo/pipelines/", r.URL.Path)
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"uuid": "{abc-123}",
			"state": {"name": "PENDING"},
			"links": {"self": [{"href": "https://api.bitbucket.org/2.0/repositories/myws/my-repo/pipelines/%7Babc-123%7D"}]}
		}`))
	})
	result, err := client.TriggerPipeline("myws", "my-repo", backend.PipelineTriggerInput{
		Branch: "main",
	})
	require.NoError(t, err)
	assert.Equal(t, "abc-123", result.UUID)
	assert.Equal(t, "PENDING", result.State)
	assert.Contains(t, result.Link, "pipelines")
	// Verify the request body
	target, _ := gotBody["target"].(map[string]any)
	require.NotNil(t, target)
	assert.Equal(t, "branch", target["ref_type"])
	assert.Equal(t, "pipeline_ref_target", target["type"])
	assert.Equal(t, "main", target["ref_name"])
}

func TestCloudClient_TriggerPipeline_WithVariables(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	client := newCloudTriggerServer(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"uuid": "{def-456}",
			"state": {"name": "PENDING"},
			"links": {"self": [{"href": "https://api.bitbucket.org/2.0/repositories/myws/my-repo/pipelines/%7Bdef-456%7D"}]}
		}`))
	})
	result, err := client.TriggerPipeline("myws", "my-repo", backend.PipelineTriggerInput{
		Branch: "feature/x",
		Variables: []backend.PipelineVariable{
			{Key: "FOO", Value: "bar", Secured: false},
			{Key: "SECRET", Value: "s3cr3t", Secured: true},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "def-456", result.UUID)
	vars, _ := gotBody["variables"].([]any)
	require.Len(t, vars, 2)
	v0, _ := vars[0].(map[string]any)
	assert.Equal(t, "FOO", v0["key"])
	assert.Equal(t, "bar", v0["value"])
}

func TestCloudClient_TriggerPipeline_OmitsVariablesWhenEmpty(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	client := newCloudTriggerServer(t, func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"uuid": "{abc-123}",
			"state": {"name": "PENDING"},
			"links": {"self": [{"href": "https://api.bitbucket.org/2.0/repositories/myws/my-repo/pipelines/%7Babc-123%7D"}]}
		}`))
	})
	_, err := client.TriggerPipeline("myws", "my-repo", backend.PipelineTriggerInput{
		Branch: "main",
	})
	require.NoError(t, err)
	_, hasVariables := gotBody["variables"]
	assert.False(t, hasVariables, "request body should not contain 'variables' key when no variables are provided, got: %v", gotBody)
}

func TestCloudClient_TriggerPipeline_NoSelfLink(t *testing.T) {
	t.Parallel()
	client := newCloudTriggerServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"uuid": "{ghi-789}",
			"state": {"name": "PENDING"},
			"links": {"self": []}
		}`))
	})
	result, err := client.TriggerPipeline("myws", "my-repo", backend.PipelineTriggerInput{Branch: "main"})
	require.NoError(t, err)
	assert.Equal(t, "ghi-789", result.UUID)
	assert.Empty(t, result.Link)
}
