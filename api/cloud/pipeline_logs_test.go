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

func TestCloudClient_GetPipelineStepLog_CorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("log line\n"))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	rc, err := client.GetPipelineStepLog("myws", "my-repo", "pipe-uuid", "step-uuid")
	require.NoError(t, err)
	defer rc.Close() //nolint:errcheck
	assert.Equal(t, "/repositories/myws/my-repo/pipelines/{pipe-uuid}/steps/{step-uuid}/log", gotPath)
}

func TestCloudClient_GetPipelineStepLog_ReturnsContent(t *testing.T) {
	t.Parallel()
	const logBody = "::group::Build\nRunning build...\n::endgroup::\n"
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(logBody))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	rc, err := client.GetPipelineStepLog("ws", "repo", "p-1", "s-1")
	require.NoError(t, err)
	defer rc.Close() //nolint:errcheck
	got, err := io.ReadAll(rc)
	require.NoError(t, err)
	assert.Equal(t, logBody, string(got))
}

func TestCloudClient_GetPipelineStepLog_BracesAddedToUUIDs(t *testing.T) {
	t.Parallel()
	// When UUIDs are passed without braces, the path should still include them.
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	rc, err := client.GetPipelineStepLog("ws", "repo", "aabbccdd-0000-0000-0000-000000000001", "11111111-0000-0000-0000-000000000002")
	require.NoError(t, err)
	defer rc.Close() //nolint:errcheck
	assert.Equal(t, "/repositories/ws/repo/pipelines/{aabbccdd-0000-0000-0000-000000000001}/steps/{11111111-0000-0000-0000-000000000002}/log", gotPath)
}
