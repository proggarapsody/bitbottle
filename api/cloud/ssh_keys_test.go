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

// newSSHKeyServer builds a test TLS server that handles GET /user as well as
// the caller-provided handler for SSH-key endpoints.
func newSSHKeyServer(t *testing.T, sshHandler http.HandlerFunc) *cloud.Client {
	t.Helper()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/user" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"nickname":"alice","account_id":"abc123"}`)
			return
		}
		sshHandler(w, r)
	}))
	t.Cleanup(srv.Close)
	return cloud.NewClient(srv.Client(), srv.URL, "tok", "")
}

const cloudSSHKeyListJSON = `{
  "values": [
    {"id":1,"label":"Laptop key","key":"ssh-rsa AAAA1"},
    {"id":2,"label":"Work key","key":"ssh-rsa AAAA2"}
  ]
}`

func TestCloudClient_ListSSHKeys(t *testing.T) {
	t.Parallel()
	client := newSSHKeyServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/users/alice/ssh-keys", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(cloudSSHKeyListJSON))
	})
	keys, err := client.ListSSHKeys()
	require.NoError(t, err)
	require.Len(t, keys, 2)
	assert.Equal(t, 1, keys[0].ID)
	assert.Equal(t, "Laptop key", keys[0].Label)
	assert.Equal(t, "ssh-rsa AAAA1", keys[0].Key)
	assert.Equal(t, 2, keys[1].ID)
	assert.Equal(t, "Work key", keys[1].Label)
}

func TestCloudClient_AddSSHKey(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	client := newSSHKeyServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/users/alice/ssh-keys", r.URL.Path)
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":3,"label":"My Key","key":"ssh-rsa AAAA3"}`))
	})
	sk, err := client.AddSSHKey(backend.SSHKeyInput{
		Key:   "ssh-rsa AAAA3",
		Label: "My Key",
	})
	require.NoError(t, err)
	assert.Equal(t, 3, sk.ID)
	assert.Equal(t, "My Key", sk.Label)
	assert.Equal(t, "ssh-rsa AAAA3", sk.Key)
	assert.Equal(t, "ssh-rsa AAAA3", gotBody["key"])
	assert.Equal(t, "My Key", gotBody["label"])
}

func TestCloudClient_AddSSHKey_MissingKey(t *testing.T) {
	t.Parallel()
	client := newSSHKeyServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("unexpected HTTP call")
	})
	_, err := client.AddSSHKey(backend.SSHKeyInput{Label: "no key"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "key required")
}

func TestCloudClient_DeleteSSHKey(t *testing.T) {
	t.Parallel()
	var gotMethod, gotPath string
	client := newSSHKeyServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})
	err := client.DeleteSSHKey(7)
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t, "/users/alice/ssh-keys/7", gotPath)
}
