package server_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/server"
)

const serverDeployKeyListJSON = `{
  "values": [
    {"id":1,"label":"CI key","key":{"id":10,"text":"ssh-rsa AAAA1","label":"CI key"}},
    {"id":2,"label":"Deploy key","key":{"id":20,"text":"ssh-rsa AAAA2","label":"Deploy key"}}
  ],
  "isLastPage": true
}`

func newServerDeployKeyClient(t *testing.T, handler http.HandlerFunc) *server.Client {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return server.NewClient(srv.Client(), srv.URL, "tok", "")
}

func TestServerClient_ListDeployKeys_Path(t *testing.T) {
	t.Parallel()
	var gotPath string
	client := newServerDeployKeyClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"values":[],"isLastPage":true}`))
	})
	_, err := client.ListDeployKeys("PROJ", "my-repo")
	require.NoError(t, err)
	assert.Equal(t, "/projects/PROJ/repos/my-repo/ssh", gotPath)
}

func TestServerClient_ListDeployKeys_Maps(t *testing.T) {
	t.Parallel()
	client := newServerDeployKeyClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(serverDeployKeyListJSON))
	})
	keys, err := client.ListDeployKeys("PROJ", "my-repo")
	require.NoError(t, err)
	require.Len(t, keys, 2)
	assert.Equal(t, 1, keys[0].ID)
	assert.Equal(t, "CI key", keys[0].Label)
	assert.Equal(t, "ssh-rsa AAAA1", keys[0].Key)
	assert.Equal(t, 2, keys[1].ID)
	assert.Equal(t, "ssh-rsa AAAA2", keys[1].Key)
}

func TestServerClient_AddDeployKey(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	client := newServerDeployKeyClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/projects/PROJ/repos/my-repo/ssh", r.URL.Path)
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":3,"label":"New Key","key":{"id":30,"text":"ssh-rsa AAAA3","label":"New Key"}}`))
	})
	dk, err := client.AddDeployKey("PROJ", "my-repo", backend.DeployKeyInput{
		Key:   "ssh-rsa AAAA3",
		Label: "New Key",
	})
	require.NoError(t, err)
	assert.Equal(t, 3, dk.ID)
	assert.Equal(t, "New Key", dk.Label)
	assert.Equal(t, "ssh-rsa AAAA3", dk.Key)
	keyObj, _ := gotBody["key"].(map[string]any)
	assert.Equal(t, "ssh-rsa AAAA3", keyObj["text"])
	assert.Equal(t, "New Key", keyObj["label"])
}

func TestServerClient_DeleteDeployKey(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	client := newServerDeployKeyClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})
	err := client.DeleteDeployKey("PROJ", "my-repo", 7)
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/projects/PROJ/repos/my-repo/ssh/7", gotPath)
}
