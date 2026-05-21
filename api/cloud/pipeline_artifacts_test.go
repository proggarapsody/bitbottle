package cloud_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/cloud"
)

func TestCloudClient_ListPipelineArtifacts_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.ListPipelineArtifacts("myws", "my-repo", "pipeline-uuid", "step-uuid", 0)
	require.NoError(t, err)
	assert.Equal(t, "/repositories/myws/my-repo/pipelines/{pipeline-uuid}/steps/{step-uuid}/artifacts", gotPath)
}

func TestCloudClient_ListPipelineArtifacts_MapsFields(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[
			{"name":"build.tar.gz","size_bytes":1048576,"links":{"self":{"href":"https://bitbucket.org/dl/build.tar.gz"}}},
			{"name":"test-results.xml","size_bytes":2048,"links":{"self":{"href":"https://bitbucket.org/dl/test-results.xml"}}}
		]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	artifacts, err := client.ListPipelineArtifacts("myws", "my-repo", "p-uuid", "s-uuid", 0)
	require.NoError(t, err)
	require.Len(t, artifacts, 2)
	assert.Equal(t, "build.tar.gz", artifacts[0].Name)
	assert.Equal(t, int64(1048576), artifacts[0].SizeBytes)
	assert.Equal(t, "https://bitbucket.org/dl/build.tar.gz", artifacts[0].URL)
	assert.Equal(t, "test-results.xml", artifacts[1].Name)
	assert.Equal(t, int64(2048), artifacts[1].SizeBytes)
}

func TestCloudClient_ListPipelineArtifacts_BracedUUID(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	// Caller supplies uuid already with braces; should not double-brace
	_, err := client.ListPipelineArtifacts("myws", "repo", "{aabbccdd}", "{eeff0011}", 0)
	require.NoError(t, err)
	assert.NotContains(t, gotPath, "{{")
	assert.NotContains(t, gotPath, "}}")
}

func TestCloudClient_DownloadPipelineArtifact_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write([]byte("binary-content"))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	var buf bytes.Buffer
	err := client.DownloadPipelineArtifact("myws", "my-repo", "pipe-uuid", "step-uuid", "build.tar.gz", &buf)
	require.NoError(t, err)
	assert.Equal(t, http.MethodGet, gotMethod)
	assert.Equal(t, "/repositories/myws/my-repo/pipelines/{pipe-uuid}/steps/{step-uuid}/artifacts/build.tar.gz", gotPath)
	assert.Equal(t, "binary-content", buf.String())
}

func TestCloudClient_DownloadPipelineArtifact_WritesBody(t *testing.T) {
	t.Parallel()
	content := []byte("artifact binary data here")
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(content)
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	var buf bytes.Buffer
	err := client.DownloadPipelineArtifact("myws", "my-repo", "p-uuid", "s-uuid", "artifact.zip", &buf)
	require.NoError(t, err)
	assert.Equal(t, content, buf.Bytes())
}
