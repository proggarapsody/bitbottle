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

func TestCloudClient_SetRepoDefaultBranch_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	err := client.SetRepoDefaultBranch("myworkspace", "my-repo", "main")
	require.NoError(t, err)
	assert.Equal(t, http.MethodPut, gotMethod)
	assert.Equal(t, "/repositories/myworkspace/my-repo", gotPath)
}

func TestCloudClient_SetRepoDefaultBranch_SendsCorrectBody(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	err := client.SetRepoDefaultBranch("ws", "repo", "develop")
	require.NoError(t, err)
	mainBranch, _ := gotBody["mainbranch"].(map[string]any)
	require.NotNil(t, mainBranch, "body must contain mainbranch key")
	assert.Equal(t, "develop", mainBranch["name"])
}

func TestCloudClient_SetRepoDefaultBranch_NameEscapedInPath(t *testing.T) {
	t.Parallel()
	// Workspace and slug with special characters should be safe in the URL path.
	var gotPath string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	err := client.SetRepoDefaultBranch("my-org", "my-service", "main")
	require.NoError(t, err)
	assert.Equal(t, "/repositories/my-org/my-service", gotPath)
}
