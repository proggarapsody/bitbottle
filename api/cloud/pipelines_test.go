package cloud_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

func TestCloudClient_ListPipelines_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.ListPipelines("myworkspace", "my-service", 10)
	require.NoError(t, err)
	assert.Equal(t, "/repositories/myworkspace/my-service/pipelines/", gotPath)
}

func TestCloudClient_ListPipelines_MapsCompletedResultState(t *testing.T) {
	t.Parallel()
	// COMPLETED + result.name should flatten to the result name (SUCCESSFUL, FAILED, etc.)
	client, _ := cloudFixtureClient(t, "testdata/pipeline_list.json", 200)
	pipelines, err := client.ListPipelines("myworkspace", "my-service", 10)
	require.NoError(t, err)
	require.Len(t, pipelines, 2)
	assert.Equal(t, "SUCCESSFUL", pipelines[0].State)
}

func TestCloudClient_ListPipelines_MapsInProgressState(t *testing.T) {
	t.Parallel()
	client, _ := cloudFixtureClient(t, "testdata/pipeline_list.json", 200)
	pipelines, err := client.ListPipelines("myworkspace", "my-service", 10)
	require.NoError(t, err)
	require.Len(t, pipelines, 2)
	assert.Equal(t, "IN_PROGRESS", pipelines[1].State)
}

func TestCloudClient_ListPipelines_MapsBuildNumber(t *testing.T) {
	t.Parallel()
	client, _ := cloudFixtureClient(t, "testdata/pipeline_list.json", 200)
	pipelines, err := client.ListPipelines("myworkspace", "my-service", 10)
	require.NoError(t, err)
	require.Len(t, pipelines, 2)
	assert.Equal(t, 42, pipelines[0].BuildNumber)
	assert.Equal(t, 43, pipelines[1].BuildNumber)
}

func TestCloudClient_ListPipelines_MapsRefName(t *testing.T) {
	t.Parallel()
	client, _ := cloudFixtureClient(t, "testdata/pipeline_list.json", 200)
	pipelines, err := client.ListPipelines("myworkspace", "my-service", 10)
	require.NoError(t, err)
	require.Len(t, pipelines, 2)
	assert.Equal(t, "main", pipelines[0].RefName)
	assert.Equal(t, "branch", pipelines[0].RefType)
}

func TestCloudClient_GetPipeline_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	uuid := "{aabbccdd-1234-5678-abcd-000000000001}"
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		data, _ := os.ReadFile("testdata/pipeline_get.json")
		_, _ = w.Write(data)
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.GetPipeline("myworkspace", "my-service", uuid)
	require.NoError(t, err)
	assert.Equal(t, "/repositories/myworkspace/my-service/pipelines/"+uuid, gotPath)
}

func TestCloudClient_ListPipelines_UUIDHasNoBraces(t *testing.T) {
	t.Parallel()
	client, _ := cloudFixtureClient(t, "testdata/pipeline_list.json", 200)
	pipelines, err := client.ListPipelines("myworkspace", "my-service", 10)
	require.NoError(t, err)
	require.NotEmpty(t, pipelines)
	for _, p := range pipelines {
		assert.NotContains(t, p.UUID, "{", "UUID should not contain opening brace")
		assert.NotContains(t, p.UUID, "}", "UUID should not contain closing brace")
	}
}

func TestCloudClient_GetPipeline_UUIDHasNoBraces(t *testing.T) {
	t.Parallel()
	client, _ := cloudFixtureClient(t, "testdata/pipeline_get.json", 200)
	p, err := client.GetPipeline("myworkspace", "my-service", "{aabbccdd-1234-5678-abcd-000000000001}")
	require.NoError(t, err)
	assert.NotContains(t, p.UUID, "{")
	assert.NotContains(t, p.UUID, "}")
}

func TestCloudClient_RunPipeline_PostsCorrectBody(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		data, _ := os.ReadFile("testdata/pipeline_get.json")
		_, _ = w.Write(data)
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.RunPipeline("myworkspace", "my-service", backend.RunPipelineInput{Branch: "main"})
	require.NoError(t, err)
	target, _ := gotBody["target"].(map[string]any)
	require.NotNil(t, target)
	assert.Equal(t, "pipeline_ref_target", target["type"])
	assert.Equal(t, "branch", target["ref_type"])
	assert.Equal(t, "main", target["ref_name"])
}
