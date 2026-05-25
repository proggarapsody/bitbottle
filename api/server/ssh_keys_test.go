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

// buildSSHKeyServer creates a TLS test server wiring the ssh REST root.
func buildSSHKeyServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *server.Client) {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	c := server.NewClient(srv.Client(), srv.URL, "tok", "alice")
	return srv, c
}

func TestServerClient_ListSSHKeys(t *testing.T) {
	t.Parallel()
	var seenPath string
	_, c := buildSSHKeyServer(t, func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"values": []any{
				map[string]any{"id": 1, "label": "Laptop", "text": "ssh-rsa AAAA1"},
				map[string]any{"id": 2, "label": "Work", "text": "ssh-rsa AAAA2"},
			},
			"isLastPage": true,
			"size":       2,
		})
	})
	keys, err := c.ListSSHKeys()
	require.NoError(t, err)
	assert.Contains(t, seenPath, "/keys")
	require.Len(t, keys, 2)
	assert.Equal(t, 1, keys[0].ID)
	assert.Equal(t, "Laptop", keys[0].Label)
	assert.Equal(t, "ssh-rsa AAAA1", keys[0].Key)
	assert.Equal(t, 2, keys[1].ID)
	assert.Equal(t, "Work", keys[1].Label)
}

func TestServerClient_AddSSHKey(t *testing.T) {
	t.Parallel()
	var seenMethod, seenPath string
	var seenBody map[string]any
	_, c := buildSSHKeyServer(t, func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &seenBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": 3, "label": "New Key", "text": "ssh-rsa AAAA3",
		})
	})
	got, err := c.AddSSHKey(backend.SSHKeyInput{Key: "ssh-rsa AAAA3", Label: "New Key"})
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, seenMethod)
	assert.Contains(t, seenPath, "/keys")
	assert.Equal(t, "ssh-rsa AAAA3", seenBody["text"])
	assert.Equal(t, "New Key", seenBody["label"])
	assert.Equal(t, 3, got.ID)
	assert.Equal(t, "New Key", got.Label)
	assert.Equal(t, "ssh-rsa AAAA3", got.Key)
}

func TestServerClient_AddSSHKey_MissingKey(t *testing.T) {
	t.Parallel()
	_, c := buildSSHKeyServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("unexpected HTTP call")
	})
	_, err := c.AddSSHKey(backend.SSHKeyInput{Label: "no key"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "key required")
}

func TestServerClient_DeleteSSHKey(t *testing.T) {
	t.Parallel()
	var seenMethod, seenPath string
	_, c := buildSSHKeyServer(t, func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})
	err := c.DeleteSSHKey(42)
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, seenMethod)
	assert.Contains(t, seenPath, "/keys/42")
}

func TestServer_SSHKey_ImplementsInterface(t *testing.T) {
	t.Parallel()
	var _ backend.SSHKeyClient = (*server.Client)(nil)
}
