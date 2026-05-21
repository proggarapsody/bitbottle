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

func TestCloudClient_TransferRepo_IssuesCorrectPath(t *testing.T) {
	t.Parallel()
	var gotPath, gotMethod string
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"full_name":"target-ws/my-repo","name":"my-repo","slug":"my-repo","scm":"git","links":{"html":{"href":"https://bitbucket.org/target-ws/my-repo"}}}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	_, err := client.TransferRepo("source-ws", "my-repo", "target-ws")
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/repositories/source-ws/my-repo/transfer", gotPath)
}

func TestCloudClient_TransferRepo_SendsNewOwner(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"full_name":"target-ws/repo","name":"repo","slug":"repo","scm":"git","links":{"html":{"href":""}}}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	_, err := client.TransferRepo("source-ws", "repo", "target-ws")
	require.NoError(t, err)
	newOwner, _ := gotBody["new_owner"].(map[string]any)
	require.NotNil(t, newOwner, "body must contain new_owner")
	assert.Equal(t, "target-ws", newOwner["username"])
}

func TestCloudClient_TransferRepo_MapsReturnedRepo(t *testing.T) {
	t.Parallel()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"full_name":"target-ws/cool-service","name":"cool-service","slug":"cool-service","scm":"git","is_private":false,"description":"Moved repo","links":{"html":{"href":"https://bitbucket.org/target-ws/cool-service"}}}`))
	}))
	t.Cleanup(srv.Close)
	client := cloud.NewClient(srv.Client(), srv.URL, "tok", "ws")

	repo, err := client.TransferRepo("source-ws", "cool-service", "target-ws")
	require.NoError(t, err)
	assert.Equal(t, "cool-service", repo.Slug)
	assert.Equal(t, "cool-service", repo.Name)
	assert.Equal(t, "target-ws", repo.Namespace)
	assert.Equal(t, "git", repo.SCM)
	assert.Equal(t, "Moved repo", repo.Description)
	assert.Contains(t, repo.WebURL, "target-ws/cool-service")
}
