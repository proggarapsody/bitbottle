package server_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/server"
)

func TestServerClient_ListTags_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"values":[],"isLastPage":true}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.ListTags("PROJ", "my-repo", 10)
	require.NoError(t, err)
	assert.Equal(t, "/projects/PROJ/repos/my-repo/tags", gotPath)
}

func TestServerClient_ListTags_MapsFields(t *testing.T) {
	t.Parallel()
	client := fixtureClient(t, "testdata/tag_list.json", 200)
	tags, err := client.ListTags("PROJ", "my-repo", 25)
	require.NoError(t, err)
	require.Len(t, tags, 2)
	assert.Equal(t, "v1.0.0", tags[0].Name)
	assert.Equal(t, "abc1234def567890", tags[0].Hash)
	assert.Equal(t, "Release v1.0.0", tags[0].Message)
	// lightweight tag has empty message
	assert.Equal(t, "v0.9.0", tags[1].Name)
	assert.Equal(t, "", tags[1].Message)
}

func TestServerClient_CreateTag_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"displayId":"v1.0.0","latestCommit":"abc","displayMessage":""}`))
	}))
	t.Cleanup(srv.Close)
	client := server.NewClient(srv.Client(), srv.URL, "tok", "")
	_, err := client.CreateTag("PROJ", "my-repo", backend.CreateTagInput{
		Name:    "v1.0.0",
		StartAt: "main",
	})
	require.NoError(t, err)
	assert.Equal(t, "/projects/PROJ/repos/my-repo/tags", gotPath)
	assert.Equal(t, http.MethodPost, gotMethod)
}

func TestServerClient_CreateTag_MapsResponse(t *testing.T) {
	t.Parallel()
	client := fixtureClient(t, "testdata/tag_create.json", 200)
	tag, err := client.CreateTag("PROJ", "my-repo", backend.CreateTagInput{
		Name:    "v2.0.0",
		StartAt: "main",
		Message: "Annotated release",
	})
	require.NoError(t, err)
	assert.Equal(t, "v2.0.0", tag.Name)
	assert.Equal(t, "cafe1234beef5678", tag.Hash)
	assert.Equal(t, "Annotated release", tag.Message)
}

func TestServerClient_DeleteTag_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	var gotBody []byte
	client, _ := newServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusNoContent)
	})
	err := client.DeleteTag("PROJ", "my-repo", "v1.0.0")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/projects/PROJ/repos/my-repo/tags/v1.0.0", gotPath)
	assert.Contains(t, string(gotBody), "v1.0.0")
}
