package cloud_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/cloud"
)

func TestCloudClient_ListPipelineSteps_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	uuid := "{aabbccdd-1234-5678-abcd-000000000001}"
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.ListPipelineSteps("myworkspace", "my-service", uuid)
	require.NoError(t, err)
	assert.Equal(t, "/repositories/myworkspace/my-service/pipelines/"+uuid+"/steps/", gotPath)
}

func TestCloudClient_ListPipelineSteps_MapsFields(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[
            {"uuid":"{step-1}","name":"Build","duration_in_seconds":42,
             "state":{"name":"COMPLETED","result":{"name":"SUCCESSFUL"}}},
            {"uuid":"{step-2}","name":"Test","duration_in_seconds":17,
             "state":{"name":"IN_PROGRESS"}}
        ]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	steps, err := client.ListPipelineSteps("ws", "repo", "{p-uuid}")
	require.NoError(t, err)
	require.Len(t, steps, 2)
	assert.Equal(t, "step-1", steps[0].UUID, "leading/trailing braces stripped")
	assert.Equal(t, "Build", steps[0].Name)
	assert.Equal(t, "SUCCESSFUL", steps[0].State, "COMPLETED flattens to result name")
	assert.Equal(t, 42, steps[0].Duration)
	assert.Equal(t, "IN_PROGRESS", steps[1].State, "non-COMPLETED stays as-is")
}

func TestCloudClient_GetPipelineStepLog_StreamsBytes(t *testing.T) {
	t.Parallel()
	const payload = "step 1: starting\nstep 1: done\n"
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(payload))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	rc, err := client.GetPipelineStepLog("ws", "repo", "{p-uuid}", "{s-uuid}")
	require.NoError(t, err)
	defer rc.Close() //nolint:errcheck
	got, err := io.ReadAll(rc)
	require.NoError(t, err)
	assert.Equal(t, payload, string(got))
	assert.Equal(t, "/repositories/ws/repo/pipelines/{p-uuid}/steps/{s-uuid}/log", gotPath)
}

func TestCloudClient_GetPipelineStepLog_404Surfaces(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"type":"error","error":{"message":"not found"}}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.GetPipelineStepLog("ws", "repo", "{p}", "{s}")
	require.Error(t, err)
}
