package cloud_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/cloud"
)

func TestCloudClient_ListPipelineCaches_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.ListPipelineCaches("myworkspace", "my-service")
	require.NoError(t, err)
	assert.Equal(t, "/repositories/myworkspace/my-service/pipelines_config/caches/", gotPath)
}

func TestCloudClient_ListPipelineCaches_MapsFields(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[
			{"uuid":"{cache-1}","name":"node_modules","path":"/app/node_modules","file_size_bytes":12345678,"created_on":"2024-01-01T00:00:00.000Z"},
			{"uuid":"{cache-2}","name":"pip-deps","path":"/root/.cache/pip","file_size_bytes":9876543,"created_on":"2024-02-01T00:00:00.000Z"}
		]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	caches, err := client.ListPipelineCaches("myworkspace", "my-service")
	require.NoError(t, err)
	require.Len(t, caches, 2)
	assert.Equal(t, "cache-1", caches[0].UUID)
	assert.Equal(t, "node_modules", caches[0].Name)
	assert.Equal(t, "/app/node_modules", caches[0].Path)
	assert.Equal(t, int64(12345678), caches[0].FileSizeBytes)
	assert.Equal(t, "2024-01-01T00:00:00.000Z", caches[0].CreatedOn)
	assert.Equal(t, "cache-2", caches[1].UUID)
	assert.Equal(t, "pip-deps", caches[1].Name)
}

func TestCloudClient_ListPipelineCaches_UUIDHasNoBraces(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[
			{"uuid":"{cache-abc}","name":"deps","path":"/deps","file_size_bytes":1024,"created_on":"2024-01-01T00:00:00.000Z"}
		]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	caches, err := client.ListPipelineCaches("myworkspace", "my-service")
	require.NoError(t, err)
	require.Len(t, caches, 1)
	assert.NotContains(t, caches[0].UUID, "{")
	assert.NotContains(t, caches[0].UUID, "}")
}

func TestCloudClient_DeletePipelineCache_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	var gotMethod string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	err := client.DeletePipelineCache("myworkspace", "my-service", "cache-abc")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/repositories/myworkspace/my-service/pipelines_config/caches/{cache-abc}", gotPath)
}
