package cloud_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
)

func newPipelineSSHServer(t *testing.T, handler http.HandlerFunc) *cloud.Client {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return cloud.NewClient(srv.Client(), srv.URL, "tok", "")
}

func TestCloudClient_GetPipelineSSHKeyPair(t *testing.T) {
	t.Parallel()
	var seenMethod, seenPath string
	c := newPipelineSSHServer(t, func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"public_key": "ssh-rsa AAAA...",
			"key_type":   "RSA",
			"created_on": "2024-01-01T00:00:00Z",
		})
	})
	got, err := c.GetPipelineSSHKeyPair("myworkspace", "my-repo")
	require.NoError(t, err)
	assert.Equal(t, http.MethodGet, seenMethod)
	assert.Contains(t, seenPath, "pipelines_config/ssh/key_pair")
	assert.Equal(t, "ssh-rsa AAAA...", got.PublicKey)
	assert.Equal(t, "RSA", got.KeyTypeLabel)
	assert.Equal(t, 2024, got.Created.Year())
}

func TestCloudClient_RegeneratePipelineSSHKeyPair(t *testing.T) {
	t.Parallel()
	var seenMethod string
	c := newPipelineSSHServer(t, func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"public_key": "ssh-rsa BBBB...",
			"key_type":   "RSA",
			"created_on": "2024-06-01T00:00:00Z",
		})
	})
	got, err := c.RegeneratePipelineSSHKeyPair("myworkspace", "my-repo", 4096)
	require.NoError(t, err)
	assert.Equal(t, http.MethodPut, seenMethod)
	assert.Equal(t, "ssh-rsa BBBB...", got.PublicKey)
}

func TestCloudClient_ListPipelineKnownHosts(t *testing.T) {
	t.Parallel()
	var seenPath string
	c := newPipelineSSHServer(t, func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"values": []any{
				map[string]any{
					"uuid":     "{kh-uuid-1}",
					"hostname": "github.com",
					"public_key": map[string]any{
						"key_type":           "RSA",
						"key":                "AAAA...",
						"md5_fingerprint":    "aa:bb",
						"sha256_fingerprint": "SHA256:xxxx",
					},
				},
			},
			"pagelen": 10,
			"size":    1,
		})
	})
	hosts, err := c.ListPipelineKnownHosts("myworkspace", "my-repo")
	require.NoError(t, err)
	assert.Contains(t, seenPath, "pipelines_config/ssh/known_hosts")
	require.Len(t, hosts, 1)
	assert.Equal(t, "kh-uuid-1", hosts[0].UUID)
	assert.Equal(t, "github.com", hosts[0].Hostname)
	assert.Equal(t, "RSA", hosts[0].PublicKey.KeyType)
}

func TestCloudClient_GetPipelineKnownHost(t *testing.T) {
	t.Parallel()
	c := newPipelineSSHServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"uuid":     "{kh-uuid-1}",
			"hostname": "github.com",
			"public_key": map[string]any{
				"key_type":           "RSA",
				"key":                "AAAA...",
				"md5_fingerprint":    "aa:bb",
				"sha256_fingerprint": "SHA256:xxxx",
			},
		})
	})
	host, err := c.GetPipelineKnownHost("myworkspace", "my-repo", "kh-uuid-1")
	require.NoError(t, err)
	assert.Equal(t, "kh-uuid-1", host.UUID)
	assert.Equal(t, "github.com", host.Hostname)
}

func TestCloudClient_AddPipelineKnownHost(t *testing.T) {
	t.Parallel()
	var seenMethod string
	c := newPipelineSSHServer(t, func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"uuid":     "{new-kh}",
			"hostname": "bitbucket.org",
			"public_key": map[string]any{
				"key_type": "RSA",
				"key":      "CCCC...",
			},
		})
	})
	in := backend.PipelineKnownHostInput{
		Hostname:  "bitbucket.org",
		PublicKey: backend.PipelineSSHPublicKey{KeyType: "RSA", Key: "CCCC..."},
	}
	host, err := c.AddPipelineKnownHost("myworkspace", "my-repo", in)
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, seenMethod)
	assert.Equal(t, "bitbucket.org", host.Hostname)
}

func TestCloudClient_DeletePipelineKnownHost(t *testing.T) {
	t.Parallel()
	var seenMethod, seenPath string
	c := newPipelineSSHServer(t, func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})
	err := c.DeletePipelineKnownHost("myworkspace", "my-repo", "kh-uuid-1")
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, seenMethod)
	assert.Contains(t, seenPath, "kh-uuid-1")
}

func TestCloud_PipelineSSHKeyPair_ImplementsInterface(t *testing.T) {
	t.Parallel()
	var _ backend.PipelineSSHKeyPairClient = (*cloud.Client)(nil)
}

func TestCloud_PipelineKnownHosts_ImplementsInterface(t *testing.T) {
	t.Parallel()
	var _ backend.PipelineKnownHostsClient = (*cloud.Client)(nil)
}
