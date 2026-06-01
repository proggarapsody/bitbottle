package cloud_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

func TestCloudClient_ListPipelineVariables_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.ListPipelineVariables("ws", "repo")
	require.NoError(t, err)
	assert.Equal(t, "/repositories/ws/repo/pipelines_config/variables/", gotPath)
}

func TestCloudClient_ListPipelineVariables_MapsFields(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[
            {"uuid":"{v-1}","key":"DEPLOY_ENV","value":"prod","secured":false},
            {"uuid":"{v-2}","key":"API_TOKEN","secured":true}
        ]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	vars, err := client.ListPipelineVariables("ws", "repo")
	require.NoError(t, err)
	require.Len(t, vars, 2)
	assert.Equal(t, "v-1", vars[0].UUID)
	assert.Equal(t, "DEPLOY_ENV", vars[0].Key)
	assert.Equal(t, "prod", vars[0].Value)
	assert.False(t, vars[0].Secured)
	assert.Equal(t, "API_TOKEN", vars[1].Key)
	assert.Empty(t, vars[1].Value, "secured variable values are not returned by the API")
	assert.True(t, vars[1].Secured)
}

func TestCloudClient_SetPipelineVariable_PostsWhenKeyAbsent(t *testing.T) {
	t.Parallel()
	var posts, lists int
	var postBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			lists++
			_, _ = w.Write([]byte(`{"values":[]}`)) // empty → upsert must POST
		case http.MethodPost:
			posts++
			_ = json.NewDecoder(r.Body).Decode(&postBody)
			_, _ = w.Write([]byte(`{"uuid":"{v-new}","key":"DEPLOY_ENV","value":"prod","secured":false}`))
		default:
			t.Fatalf("unexpected method %q", r.Method)
		}
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	got, err := client.SetPipelineVariable("ws", "repo", backend.PipelineVariableInput{
		Key: "DEPLOY_ENV", Value: "prod", Secured: false,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, lists)
	assert.Equal(t, 1, posts)
	assert.Equal(t, "DEPLOY_ENV", postBody["key"])
	assert.Equal(t, "prod", postBody["value"])
	assert.Equal(t, false, postBody["secured"])
	assert.Equal(t, "v-new", got.UUID)
}

func TestCloudClient_SetPipelineVariable_PutsWhenKeyExists(t *testing.T) {
	t.Parallel()
	var puts int
	var putPath string
	var putBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"values":[
                {"uuid":"{v-existing}","key":"DEPLOY_ENV","value":"staging","secured":false}
            ]}`))
		case http.MethodPut:
			puts++
			putPath = r.URL.Path
			_ = json.NewDecoder(r.Body).Decode(&putBody)
			_, _ = w.Write([]byte(`{"uuid":"{v-existing}","key":"DEPLOY_ENV","value":"prod","secured":false}`))
		default:
			t.Fatalf("unexpected method %q", r.Method)
		}
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	got, err := client.SetPipelineVariable("ws", "repo", backend.PipelineVariableInput{
		Key: "DEPLOY_ENV", Value: "prod",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, puts)
	assert.Equal(t, "/repositories/ws/repo/pipelines_config/variables/{v-existing}", putPath)
	assert.Equal(t, "prod", putBody["value"])
	assert.Equal(t, "v-existing", got.UUID)
}

func TestCloudClient_DeletePipelineVariable_LooksUpKeyAndDeletes(t *testing.T) {
	t.Parallel()
	var deletePath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"values":[
                {"uuid":"{v-x}","key":"OBSOLETE","secured":false}
            ]}`))
		case http.MethodDelete:
			deletePath = r.URL.Path
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected method %q", r.Method)
		}
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	err := client.DeletePipelineVariable("ws", "repo", "OBSOLETE")
	require.NoError(t, err)
	assert.Equal(t, "/repositories/ws/repo/pipelines_config/variables/{v-x}", deletePath)
}

func TestCloudClient_GetPipelineVariable_ReturnsMatchingVariable(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[
			{"uuid":"{v-1}","key":"DEPLOY_ENV","value":"prod","secured":false},
			{"uuid":"{v-2}","key":"API_TOKEN","secured":true}
		]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	got, err := client.GetPipelineVariable("ws", "repo", "DEPLOY_ENV")
	require.NoError(t, err)
	assert.Equal(t, "v-1", got.UUID)
	assert.Equal(t, "DEPLOY_ENV", got.Key)
	assert.Equal(t, "prod", got.Value)
	assert.False(t, got.Secured)
}

func TestCloudClient_GetPipelineVariable_NotFoundReturnsTypedError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.GetPipelineVariable("ws", "repo", "MISSING")
	require.Error(t, err)
	var de *backend.DomainError
	require.True(t, errors.As(err, &de), "want backend.DomainError, got %T", err)
	assert.Equal(t, backend.ErrNotFound, de.Kind)
	assert.Equal(t, "MISSING", de.ID)
}

func TestCloudClient_DeletePipelineVariable_NotFoundReturnsTypedError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	err := client.DeletePipelineVariable("ws", "repo", "NEVER_EXISTED")
	require.Error(t, err)
	var de *backend.DomainError
	require.True(t, errors.As(err, &de), "want backend.DomainError, got %T", err)
	assert.Equal(t, backend.ErrNotFound, de.Kind)
	// drain helper to keep linter happy
	_ = io.Discard
}
