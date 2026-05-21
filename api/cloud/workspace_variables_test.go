package cloud_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

func TestCloudClient_ListWorkspaceVariables_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	_, err := client.ListWorkspaceVariables("myworkspace")
	require.NoError(t, err)
	assert.Equal(t, "/workspaces/myworkspace/pipelines-config/variables/", gotPath)
}

func TestCloudClient_ListWorkspaceVariables_MapsFields(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[
			{"uuid":"{wv-1}","key":"CI_TOKEN","value":"tok123","secured":false},
			{"uuid":"{wv-2}","key":"DEPLOY_SECRET","secured":true}
		]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	vars, err := client.ListWorkspaceVariables("myworkspace")
	require.NoError(t, err)
	require.Len(t, vars, 2)
	assert.Equal(t, "wv-1", vars[0].UUID)
	assert.Equal(t, "CI_TOKEN", vars[0].Key)
	assert.Equal(t, "tok123", vars[0].Value)
	assert.False(t, vars[0].Secured)
	assert.Equal(t, "wv-2", vars[1].UUID)
	assert.True(t, vars[1].Secured)
	assert.Empty(t, vars[1].Value, "secured variables never have values returned")
}

func TestCloudClient_SetWorkspaceVariable_PostsWhenKeyAbsent(t *testing.T) {
	t.Parallel()
	var posts int
	var postBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"values":[]}`))
		case http.MethodPost:
			posts++
			_ = json.NewDecoder(r.Body).Decode(&postBody)
			_, _ = w.Write([]byte(`{"uuid":"{wv-new}","key":"CI_TOKEN","value":"abc","secured":false}`))
		default:
			t.Fatalf("unexpected method %q", r.Method)
		}
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	got, err := client.SetWorkspaceVariable("myworkspace", backend.PipelineVariableInput{
		Key: "CI_TOKEN", Value: "abc", Secured: false,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, posts)
	assert.Equal(t, "CI_TOKEN", postBody["key"])
	assert.Equal(t, "abc", postBody["value"])
	assert.Equal(t, false, postBody["secured"])
	assert.Equal(t, "wv-new", got.UUID)
}

func TestCloudClient_SetWorkspaceVariable_PutsWhenKeyExists(t *testing.T) {
	t.Parallel()
	var putPath string
	var putBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"values":[{"uuid":"{wv-existing}","key":"CI_TOKEN","value":"old","secured":false}]}`))
		case http.MethodPut:
			putPath = r.URL.Path
			_ = json.NewDecoder(r.Body).Decode(&putBody)
			_, _ = w.Write([]byte(`{"uuid":"{wv-existing}","key":"CI_TOKEN","value":"new","secured":false}`))
		default:
			t.Fatalf("unexpected method %q", r.Method)
		}
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	got, err := client.SetWorkspaceVariable("myworkspace", backend.PipelineVariableInput{
		Key: "CI_TOKEN", Value: "new",
	})
	require.NoError(t, err)
	assert.Equal(t, "/workspaces/myworkspace/pipelines-config/variables/{wv-existing}", putPath)
	assert.Equal(t, "new", putBody["value"])
	assert.Equal(t, "wv-existing", got.UUID)
}

func TestCloudClient_DeleteWorkspaceVariable_LooksUpKeyAndDeletes(t *testing.T) {
	t.Parallel()
	var deletePath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"values":[{"uuid":"{wv-old}","key":"OBSOLETE","secured":false}]}`))
		case http.MethodDelete:
			deletePath = r.URL.Path
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected method %q", r.Method)
		}
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	err := client.DeleteWorkspaceVariable("myworkspace", "OBSOLETE")
	require.NoError(t, err)
	assert.Equal(t, "/workspaces/myworkspace/pipelines-config/variables/{wv-old}", deletePath)
}

func TestCloudClient_DeleteWorkspaceVariable_NotFoundReturnsTypedError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	err := client.DeleteWorkspaceVariable("myworkspace", "NEVER_EXISTED")
	require.Error(t, err)
	var de *backend.DomainError
	require.True(t, errors.As(err, &de), "want backend.DomainError, got %T", err)
	assert.Equal(t, backend.ErrNotFound, de.Kind)
}
