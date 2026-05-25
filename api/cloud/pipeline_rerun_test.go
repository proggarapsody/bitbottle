package cloud_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/cloud"
)

// pipelineRerunMultiHandler returns an httptest.TLSServer that serves GET and
// POST requests from separate handlers. The returned server closes on test
// cleanup.
func pipelineRerunMultiHandler(t *testing.T, onGET, onPOST http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			onGET(w, r)
		case http.MethodPost:
			onPOST(w, r)
		default:
			http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestRerunPipeline_GetsSourceThenPostsWithCommitHash(t *testing.T) {
	t.Parallel()

	const sourceUUID = "aabbccdd-1234-5678-abcd-000000000001"

	getResp := `{
		"target": {
			"type":     "pipeline_ref_target",
			"ref_name": "main",
			"ref_type": "branch",
			"commit":   {"hash": "abc123"}
		}
	}`
	postResp := `{
		"uuid":         "{bbbbcccc-0000-0000-0000-000000000002}",
		"build_number": 99,
		"state":        {"name": "PENDING", "result": {"name": ""}},
		"target":       {"ref_type": "branch", "ref_name": "main"},
		"created_on":   "2026-05-01T00:00:00.000Z",
		"duration_in_seconds": 0
	}`

	var gotPostBody map[string]any
	srv := pipelineRerunMultiHandler(t,
		func(w http.ResponseWriter, r *http.Request) {
			// Verify GET path includes braced UUID
			assert.Contains(t, r.URL.Path, "{"+sourceUUID+"}")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(getResp))
		},
		func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewDecoder(r.Body).Decode(&gotPostBody)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(postResp))
		},
	)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	pl, err := client.RerunPipeline("myworkspace", "my-service", sourceUUID)
	require.NoError(t, err)

	// Returned pipeline UUID has no braces.
	assert.Equal(t, "bbbbcccc-0000-0000-0000-000000000002", pl.UUID)
	assert.Equal(t, 99, pl.BuildNumber)

	// POST body must include commit hash.
	target, _ := gotPostBody["target"].(map[string]any)
	require.NotNil(t, target)
	commit, _ := target["commit"].(map[string]any)
	require.NotNil(t, commit, "expected commit key in POST body when hash is non-empty")
	assert.Equal(t, "abc123", commit["hash"])
	assert.Equal(t, "commit", commit["type"])
}

func TestRerunPipeline_FallsBackToRefOnlyWhenNoCommitHash(t *testing.T) {
	t.Parallel()

	const sourceUUID = "aabbccdd-1234-5678-abcd-000000000003"

	getResp := `{
		"target": {
			"type":     "pipeline_ref_target",
			"ref_name": "develop",
			"ref_type": "branch",
			"commit":   {}
		}
	}`
	postResp := `{
		"uuid":         "{ccccdddd-0000-0000-0000-000000000004}",
		"build_number": 100,
		"state":        {"name": "PENDING", "result": {"name": ""}},
		"target":       {"ref_type": "branch", "ref_name": "develop"},
		"created_on":   "2026-05-02T00:00:00.000Z",
		"duration_in_seconds": 0
	}`

	var gotPostBody map[string]any
	srv := pipelineRerunMultiHandler(t,
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(getResp))
		},
		func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewDecoder(r.Body).Decode(&gotPostBody)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(postResp))
		},
	)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	pl, err := client.RerunPipeline("myworkspace", "my-service", sourceUUID)
	require.NoError(t, err)
	assert.Equal(t, "ccccdddd-0000-0000-0000-000000000004", pl.UUID)

	// POST body must NOT include a commit key (or it must be null/absent).
	target, _ := gotPostBody["target"].(map[string]any)
	require.NotNil(t, target)
	_, hasCommit := target["commit"]
	assert.False(t, hasCommit, "commit key must be absent when source pipeline has no commit hash")
}

func TestRerunPipeline_ErrorOnGetFailure(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.RerunPipeline("myworkspace", "my-service", "nonexistent-uuid")
	require.Error(t, err)
}
