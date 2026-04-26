package cloud_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

func TestCloudClient_ListTags_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.ListTags("myworkspace", "my-service", 10)
	require.NoError(t, err)
	assert.Equal(t, "/repositories/myworkspace/my-service/refs/tags", gotPath)
}

func TestCloudClient_ListTags_MapsFields(t *testing.T) {
	t.Parallel()
	client, _ := cloudFixtureClient(t, "testdata/tag_list.json", 200)
	tags, err := client.ListTags("myworkspace", "my-service", 25)
	require.NoError(t, err)
	require.Len(t, tags, 2)
	assert.Equal(t, "v1.0.0", tags[0].Name)
	assert.Equal(t, "abc1234def567890", tags[0].Hash)
	assert.Equal(t, "Release v1.0.0", tags[0].Message)
	assert.Equal(t, "https://bitbucket.org/myworkspace/my-service/src/v1.0.0", tags[0].WebURL)
	// lightweight tag has empty message
	assert.Equal(t, "v0.9.0", tags[1].Name)
	assert.Equal(t, "", tags[1].Message)
}

func TestCloudClient_CreateTag_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"v1.0.0","target":{"hash":"abc"},"message":"","links":{"html":{"href":""}}}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.CreateTag("myworkspace", "my-service", backend.CreateTagInput{
		Name:    "v1.0.0",
		StartAt: "main",
	})
	require.NoError(t, err)
	assert.Equal(t, "/repositories/myworkspace/my-service/refs/tags", gotPath)
	assert.Equal(t, http.MethodPost, gotMethod)
}

func TestCloudClient_CreateTag_MapsResponse(t *testing.T) {
	t.Parallel()
	client, _ := cloudFixtureClient(t, "testdata/tag_create.json", 201)
	tag, err := client.CreateTag("myworkspace", "my-service", backend.CreateTagInput{
		Name:    "v2.0.0",
		StartAt: "main",
		Message: "Annotated release",
	})
	require.NoError(t, err)
	assert.Equal(t, "v2.0.0", tag.Name)
	assert.Equal(t, "cafe1234beef5678", tag.Hash)
	assert.Equal(t, "Annotated release", tag.Message)
	assert.Equal(t, "https://bitbucket.org/myworkspace/my-service/src/v2.0.0", tag.WebURL)
}

func TestCloudClient_CreateTag_OmitsMessageWhenEmpty(t *testing.T) {
	t.Parallel()
	var gotBody []byte
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"v1.0.0","target":{"hash":"abc"},"message":"","links":{"html":{"href":""}}}`))
	})
	_, err := client.CreateTag("myworkspace", "my-service", backend.CreateTagInput{
		Name:    "v1.0.0",
		StartAt: "main",
		Message: "", // empty — should be omitted
	})
	require.NoError(t, err)
	assert.NotContains(t, string(gotBody), `"message"`)
}

func TestCloudClient_DeleteTag_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotRawPath, gotMethod string
	client, _ := newCloudClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotRawPath = r.RequestURI
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	})
	err := client.DeleteTag("myworkspace", "my-service", "v1.0.0")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Contains(t, gotRawPath, "/refs/tags/v1.0.0")
}
