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

// ── ListRunners ───────────────────────────────────────────────────────────────

func TestCloudClient_ListRunners_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/workspaces/myworkspace/pipelines-config/runners", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[
			{"uuid":"{runner-1}","name":"my-runner","state":{"status":"ONLINE"},"platform":{"operating_system":"LINUX","architecture":"X86_64"},"labels":[{"name":"self.hosted"},{"name":"linux"}],"created_on":"2024-01-01T00:00:00.000000+00:00"},
			{"uuid":"{runner-2}","name":"win-runner","state":{"status":"OFFLINE"},"platform":{"operating_system":"WINDOWS","architecture":"X86_64"},"labels":[{"name":"windows"}],"created_on":"2024-02-01T00:00:00.000000+00:00"}
		]}`))
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	runners, err := client.ListRunners("myworkspace")
	require.NoError(t, err)
	require.Len(t, runners, 2)

	assert.Equal(t, "runner-1", runners[0].UUID)
	assert.Equal(t, "my-runner", runners[0].Name)
	assert.Equal(t, "ONLINE", runners[0].State)
	assert.Equal(t, "LINUX", runners[0].Platform.Operating)
	assert.Equal(t, "AMD64", runners[0].Platform.Arch)
	assert.Equal(t, []string{"self.hosted", "linux"}, runners[0].Labels)
	assert.Equal(t, "2024-01-01T00:00:00.000000+00:00", runners[0].CreatedOn)

	assert.Equal(t, "runner-2", runners[1].UUID)
	assert.Equal(t, "OFFLINE", runners[1].State)
	assert.Equal(t, "WINDOWS", runners[1].Platform.Operating)
	assert.Equal(t, "AMD64", runners[1].Platform.Arch)
}

func TestCloudClient_ListRunners_UUIDHasNoBraces(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[
			{"uuid":"{runner-abc}","name":"r","state":{"status":"ONLINE"},"platform":{"operating_system":"LINUX","architecture":"X86_64"},"labels":[],"created_on":"2024-01-01T00:00:00.000000+00:00"}
		]}`))
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	runners, err := client.ListRunners("myworkspace")
	require.NoError(t, err)
	require.Len(t, runners, 1)
	assert.NotContains(t, runners[0].UUID, "{")
	assert.NotContains(t, runners[0].UUID, "}")
}

func TestCloudClient_ListRunners_HTTP403(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"type":"error","error":{"message":"Access denied"}}`))
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.ListRunners("myworkspace")
	require.Error(t, err)
}

// ── CreateRunner ──────────────────────────────────────────────────────────────

func TestCloudClient_CreateRunner_Success(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/workspaces/myworkspace/pipelines-config/runners", r.URL.Path)
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"uuid":"{new-runner}","name":"new-runner","state":{"status":"OFFLINE"},"platform":{"operating_system":"LINUX","architecture":"X86_64"},"labels":[{"name":"self.hosted"}],"created_on":"2024-03-01T00:00:00.000000+00:00"}`))
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	runner, err := client.CreateRunner("myworkspace", backend.CreateRunnerInput{
		Name:   "new-runner",
		Labels: []string{"self.hosted"},
		Platform: backend.RunnerPlatform{
			Operating: "LINUX",
			Arch:      "AMD64",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "new-runner", runner.Name)
	assert.Equal(t, "new-runner", runner.UUID)
	assert.Equal(t, "AMD64", runner.Platform.Arch)

	// Verify request body contains proper arch mapping
	require.NotNil(t, gotBody)
	plat, ok := gotBody["platform"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "X86_64", plat["architecture"])
	assert.Equal(t, "LINUX", plat["operating_system"])
}

func TestCloudClient_CreateRunner_HTTP403(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"type":"error","error":{"message":"Access denied"}}`))
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.CreateRunner("myworkspace", backend.CreateRunnerInput{Name: "r"})
	require.Error(t, err)
}

// ── DeleteRunner ──────────────────────────────────────────────────────────────

func TestCloudClient_DeleteRunner_Success(t *testing.T) {
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
	err := client.DeleteRunner("myworkspace", "runner-abc")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/workspaces/myworkspace/pipelines-config/runners/{runner-abc}", gotPath)
}

func TestCloudClient_DeleteRunner_HTTP403(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"type":"error","error":{"message":"Access denied"}}`))
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	err := client.DeleteRunner("myworkspace", "runner-abc")
	require.Error(t, err)
}
