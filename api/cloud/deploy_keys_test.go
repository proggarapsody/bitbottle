package cloud_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

func newCloudDeployKeyServer(t *testing.T, handler http.HandlerFunc) *cloud.Client {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return cloud.NewClient(srv.Client(), srv.URL, "tok", "")
}

const cloudDeployKeyListJSON = `{
  "values": [
    {"id":1,"label":"CI key","key":"ssh-rsa AAAA1","read_only":false},
    {"id":2,"label":"Deploy key","key":"ssh-rsa AAAA2","read_only":true}
  ]
}`

func TestCloudClient_ListDeployKeys(t *testing.T) {
	t.Parallel()
	client := newCloudDeployKeyServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/repositories/myws/my-repo/deploy-keys", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cloudDeployKeyListJSON))
	})
	keys, err := client.ListDeployKeys("myws", "my-repo")
	require.NoError(t, err)
	require.Len(t, keys, 2)
	assert.Equal(t, 1, keys[0].ID)
	assert.Equal(t, "CI key", keys[0].Label)
	assert.Equal(t, "ssh-rsa AAAA1", keys[0].Key)
	assert.False(t, keys[0].ReadOnly)
	assert.Equal(t, 2, keys[1].ID)
	assert.True(t, keys[1].ReadOnly)
}

func TestCloudClient_AddDeployKey(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	client := newCloudDeployKeyServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/repositories/myws/my-repo/deploy-keys", r.URL.Path)
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":3,"label":"My Key","key":"ssh-rsa AAAA3","read_only":false}`))
	})
	dk, err := client.AddDeployKey("myws", "my-repo", backend.DeployKeyInput{
		Key:   "ssh-rsa AAAA3",
		Label: "My Key",
	})
	require.NoError(t, err)
	assert.Equal(t, 3, dk.ID)
	assert.Equal(t, "My Key", dk.Label)
	assert.Equal(t, "ssh-rsa AAAA3", dk.Key)
	assert.Equal(t, "ssh-rsa AAAA3", gotBody["key"])
	assert.Equal(t, "My Key", gotBody["label"])
}

func TestCloudClient_DeleteDeployKey(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	client := newCloudDeployKeyServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})
	err := client.DeleteDeployKey("myws", "my-repo", 5)
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/repositories/myws/my-repo/deploy-keys/5", gotPath)
}
