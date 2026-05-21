package cloud_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

const listDeploymentsJSON = `{"values":[{"uuid":"{dep-1}","state":{"name":"COMPLETED"},"environment":{"uuid":"{env-1}","name":"Production","environment_type":{"name":"Production"},"rank":3},"release":{"name":"v1.2.3","url":"https://bitbucket.org/ws/repo/src/abc","commit":{"hash":"abc123"}}}]}`

const listEnvironmentsJSON = `{"values":[{"uuid":"{env-1}","name":"Staging","environment_type":{"name":"Test"},"rank":1},{"uuid":"{env-2}","name":"Production","environment_type":{"name":"Production"},"rank":3}]}`

func TestCloudClient_ListDeployments_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	_, err := client.ListDeployments("ws", "repo", 10)
	require.NoError(t, err)
	assert.Equal(t, "/repositories/ws/repo/deployments", gotPath)
}

func TestCloudClient_ListDeployments_MapsFields(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(listDeploymentsJSON))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	deps, err := client.ListDeployments("ws", "repo", 10)
	require.NoError(t, err)
	require.Len(t, deps, 1)
	assert.Equal(t, "dep-1", deps[0].UUID)
	assert.Equal(t, "COMPLETED", deps[0].State)
	assert.Equal(t, "env-1", deps[0].Environment.UUID)
	assert.Equal(t, "Production", deps[0].Environment.Name)
	assert.Equal(t, 3, deps[0].Environment.Rank)
	assert.Equal(t, "v1.2.3", deps[0].Release.Name)
	assert.Equal(t, "abc123", deps[0].Release.CommitHash)
}

func TestCloudClient_GetDeployment_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"uuid":"{dep-1}","state":{"name":"PENDING"},"environment":{"uuid":"{env-1}","name":"Test","environment_type":{"name":"Test"},"rank":1},"release":{"name":"v0.1","url":"","commit":{"hash":"deadbeef"}}}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	dep, err := client.GetDeployment("ws", "repo", "dep-1")
	require.NoError(t, err)
	assert.Equal(t, "/repositories/ws/repo/deployments/{dep-1}", gotPath)
	assert.Equal(t, "dep-1", dep.UUID)
	assert.Equal(t, "PENDING", dep.State)
}

func TestCloudClient_ListEnvironments_MapsFields(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(listEnvironmentsJSON))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	envs, err := client.ListEnvironments("ws", "repo")
	require.NoError(t, err)
	assert.Equal(t, "/repositories/ws/repo/environments", gotPath)
	require.Len(t, envs, 2)
	assert.Equal(t, "env-1", envs[0].UUID)
	assert.Equal(t, "Staging", envs[0].Name)
	assert.Equal(t, "Test", envs[0].Type)
	assert.Equal(t, 1, envs[0].Rank)
	assert.Equal(t, "env-2", envs[1].UUID)
	assert.Equal(t, "Production", envs[1].Name)
	assert.Equal(t, 3, envs[1].Rank)
}

func TestCloudClient_CreateEnvironment_PostsCorrectBody(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"uuid":"{env-new}","name":"Staging","environment_type":{"name":"Test"},"rank":2}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	env, err := client.CreateEnvironment("ws", "repo", backend.CreateEnvironmentInput{
		Name: "Staging", Type: "Test", Rank: 2,
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/repositories/ws/repo/environments", gotPath)
	assert.Equal(t, "Staging", gotBody["name"])
	envType, _ := gotBody["environment_type"].(map[string]any)
	assert.Equal(t, "Test", envType["name"])
	assert.Equal(t, "env-new", env.UUID)
}

func TestCloudClient_DeleteEnvironment_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	err := client.DeleteEnvironment("ws", "repo", "env-1")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/repositories/ws/repo/environments/{env-1}", gotPath)
}

func TestCloudClient_ListEnvVariables_MapsFields(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[
			{"uuid":"{var-1}","key":"DB_HOST","value":"localhost","secured":false},
			{"uuid":"{var-2}","key":"DB_PASS","secured":true}
		]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	vars, err := client.ListEnvVariables("ws", "repo", "env-1")
	require.NoError(t, err)
	assert.Equal(t, "/repositories/ws/repo/deployments_config/environments/{env-1}/variables", gotPath)
	require.Len(t, vars, 2)
	assert.Equal(t, "var-1", vars[0].UUID)
	assert.Equal(t, "DB_HOST", vars[0].Key)
	assert.Equal(t, "localhost", vars[0].Value)
	assert.False(t, vars[0].Secured)
	assert.True(t, vars[1].Secured)
	assert.Empty(t, vars[1].Value)
}

func TestCloudClient_SetEnvVariable_PostsWhenKeyAbsent(t *testing.T) {
	t.Parallel()
	var gotMethod string
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"values":[]}`))
		case http.MethodPost:
			gotMethod = r.Method
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			_, _ = w.Write([]byte(`{"uuid":"{var-new}","key":"NEW_VAR","value":"val","secured":false}`))
		default:
			t.Fatalf("unexpected method %q", r.Method)
		}
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	got, err := client.SetEnvVariable("ws", "repo", "env-1", backend.EnvVariableInput{
		Key: "NEW_VAR", Value: "val", Secured: false,
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "NEW_VAR", gotBody["key"])
	assert.Equal(t, "val", gotBody["value"])
	assert.Equal(t, "var-new", got.UUID)
}

func TestCloudClient_SetEnvVariable_PutsWhenKeyExists(t *testing.T) {
	t.Parallel()
	var putPath string
	var putBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"values":[{"uuid":"{var-ex}","key":"DB_HOST","value":"old","secured":false}]}`))
		case http.MethodPut:
			putPath = r.URL.Path
			_ = json.NewDecoder(r.Body).Decode(&putBody)
			_, _ = w.Write([]byte(`{"uuid":"{var-ex}","key":"DB_HOST","value":"new","secured":false}`))
		default:
			t.Fatalf("unexpected method %q", r.Method)
		}
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	got, err := client.SetEnvVariable("ws", "repo", "env-1", backend.EnvVariableInput{
		Key: "DB_HOST", Value: "new",
	})
	require.NoError(t, err)
	assert.Equal(t, "/repositories/ws/repo/deployments_config/environments/{env-1}/variables/{var-ex}", putPath)
	assert.Equal(t, "new", putBody["value"])
	assert.Equal(t, "var-ex", got.UUID)
}

func TestCloudClient_DeleteEnvVariable_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	err := client.DeleteEnvVariable("ws", "repo", "env-1", "var-1")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/repositories/ws/repo/deployments_config/environments/{env-1}/variables/{var-1}", gotPath)
}

func TestCloudClient_ListEnvVariables_EmptyList(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	vars, err := client.ListEnvVariables("ws", "repo", "env-1")
	require.NoError(t, err)
	assert.Empty(t, vars)
}

func TestCloudClient_SetEnvVariable_SecuredFlagForwarded(t *testing.T) {
	t.Parallel()
	var postBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"values":[]}`))
		case http.MethodPost:
			_ = json.NewDecoder(r.Body).Decode(&postBody)
			_, _ = w.Write([]byte(`{"uuid":"{sec-1}","key":"SECRET","secured":true}`))
		default:
			t.Fatalf("unexpected method %q", r.Method)
		}
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	got, err := client.SetEnvVariable("ws", "repo", "env-1", backend.EnvVariableInput{
		Key: "SECRET", Value: "s3cr3t", Secured: true,
	})
	require.NoError(t, err)
	assert.Equal(t, true, postBody["secured"])
	assert.True(t, got.Secured)
}

func TestCloudClient_DeleteEnvVariable_NotFoundDoesNotError(t *testing.T) {
	t.Parallel()
	// DeleteEnvVariable takes a UUID directly — it just DELETEs without a lookup.
	// A 204 response should produce no error.
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")
	err := client.DeleteEnvVariable("ws", "repo", "env-1", "var-missing")
	require.NoError(t, err)
}
