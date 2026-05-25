package cloud_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

func TestCloudClient_ListWorkspacePipelineVariables_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	_, err := client.ListWorkspacePipelineVariables("myworkspace")
	require.NoError(t, err)
	assert.Equal(t, "/workspaces/myworkspace/pipelines-config/variables/", gotPath)
}

func TestCloudClient_ListWorkspacePipelineVariables_MapsFields(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[
			{"uuid":"{pv-1}","key":"CI_TOKEN","value":"tok123","secured":false},
			{"uuid":"{pv-2}","key":"DEPLOY_SECRET","secured":true}
		]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	vars, err := client.ListWorkspacePipelineVariables("myworkspace")
	require.NoError(t, err)
	require.Len(t, vars, 2)
	assert.Equal(t, "pv-1", vars[0].UUID)
	assert.Equal(t, "CI_TOKEN", vars[0].Key)
	assert.Equal(t, "tok123", vars[0].Value)
	assert.False(t, vars[0].Secured)
	assert.Equal(t, "pv-2", vars[1].UUID)
	assert.True(t, vars[1].Secured)
}

func TestCloudClient_GetWorkspacePipelineVariable_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"uuid":"{pv-1}","key":"CI_TOKEN","value":"tok123","secured":false}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	v, err := client.GetWorkspacePipelineVariable("myworkspace", "pv-1")
	require.NoError(t, err)
	assert.Equal(t, "/workspaces/myworkspace/pipelines-config/variables/{pv-1}", gotPath)
	assert.Equal(t, "pv-1", v.UUID)
	assert.Equal(t, "CI_TOKEN", v.Key)
}

func TestCloudClient_SetWorkspacePipelineVariable_PostsWhenKeyAbsent(t *testing.T) {
	t.Parallel()
	var posts int
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"values":[]}`))
		case http.MethodPost:
			posts++
			_, _ = w.Write([]byte(`{"uuid":"{pv-new}","key":"NEW_KEY","value":"val","secured":false}`))
		}
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	v, err := client.SetWorkspacePipelineVariable("ws", backend.PipelineVariableInput{Key: "NEW_KEY", Value: "val"})
	require.NoError(t, err)
	assert.Equal(t, 1, posts)
	assert.Equal(t, "pv-new", v.UUID)
}

func TestCloudClient_SetWorkspacePipelineVariable_PutsWhenKeyExists(t *testing.T) {
	t.Parallel()
	var puts int
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"values":[{"uuid":"{pv-1}","key":"CI_TOKEN","value":"old","secured":false}]}`))
		case http.MethodPut:
			puts++
			_, _ = w.Write([]byte(`{"uuid":"{pv-1}","key":"CI_TOKEN","value":"new","secured":false}`))
		}
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	v, err := client.SetWorkspacePipelineVariable("ws", backend.PipelineVariableInput{Key: "CI_TOKEN", Value: "new"})
	require.NoError(t, err)
	assert.Equal(t, 1, puts)
	assert.Equal(t, "new", v.Value)
}

func TestCloudClient_DeleteWorkspacePipelineVariable_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	var gotMethod string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	err := client.DeleteWorkspacePipelineVariable("myworkspace", "pv-1")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/workspaces/myworkspace/pipelines-config/variables/{pv-1}", gotPath)
}
