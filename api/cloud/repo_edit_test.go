package cloud_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

func TestCloudClient_EditRepo_DescriptionUpdate(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		body, _ := os.ReadFile("testdata/repo_edit.json")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	desc := "updated description"
	repo, err := client.EditRepo("myworkspace", "my-service", backend.EditRepoInput{
		Description: &desc,
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPut, gotMethod)
	assert.Equal(t, "/repositories/myworkspace/my-service", gotPath)
	assert.Equal(t, desc, gotBody["description"])
	assert.Equal(t, "my-service", repo.Slug)
	assert.Equal(t, "myworkspace", repo.Namespace)
}

func TestCloudClient_EditRepo_CloudOnlyWebsite(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		body, _ := os.ReadFile("testdata/repo_edit.json")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "")

	website := "https://example.com"
	_, err := client.EditRepo("myworkspace", "my-service", backend.EditRepoInput{
		Website: &website,
	})
	require.NoError(t, err)
	assert.Equal(t, website, gotBody["website"])
	// description is nil — must not appear in body
	_, hasDesc := gotBody["description"]
	assert.False(t, hasDesc, "nil description should not be serialised")
}
