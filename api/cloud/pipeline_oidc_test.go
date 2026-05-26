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

func newPipelineOIDCServer(t *testing.T, handler http.HandlerFunc) *cloud.Client {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return cloud.NewClient(srv.Client(), srv.URL, "tok", "")
}

func TestCloudClient_GetPipelineOIDCConfig(t *testing.T) {
	t.Parallel()
	var seenMethod, seenPath string
	c := newPipelineOIDCServer(t, func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                   "https://api.bitbucket.org/2.0/workspaces/myworkspace/pipelines-config/identity/oidc",
			"jwks_uri":                 "https://api.bitbucket.org/2.0/workspaces/myworkspace/pipelines-config/identity/oidc/keys.json",
			"subject_types_supported":  []string{"public"},
			"response_types_supported": []string{"id_token"},
			"claims_supported":         []string{"sub", "iss", "iat", "exp"},
		})
	})
	got, err := c.GetPipelineOIDCConfig("myworkspace")
	require.NoError(t, err)
	assert.Equal(t, http.MethodGet, seenMethod)
	assert.Contains(t, seenPath, "/workspaces/myworkspace/pipelines-config/identity/oidc/.well-known/openid-configuration")
	assert.Equal(t, "https://api.bitbucket.org/2.0/workspaces/myworkspace/pipelines-config/identity/oidc", got.Issuer)
	assert.Contains(t, got.JWKSURI, "keys.json")
	assert.Equal(t, []string{"public"}, got.SubjectTypesSupported)
	assert.Equal(t, []string{"id_token"}, got.ResponseTypesSupported)
	assert.Contains(t, got.ClaimsSupported, "sub")
}

func TestCloudClient_GetPipelineOIDCConfig_PathEscape(t *testing.T) {
	t.Parallel()
	var seenURI string
	c := newPipelineOIDCServer(t, func(w http.ResponseWriter, r *http.Request) {
		seenURI = r.RequestURI
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":   "https://example.com",
			"jwks_uri": "https://example.com/keys",
		})
	})
	_, err := c.GetPipelineOIDCConfig("my workspace")
	require.NoError(t, err)
	assert.Contains(t, seenURI, "my%20workspace")
}

func TestCloudClient_GetPipelineOIDCKeys(t *testing.T) {
	t.Parallel()
	var seenMethod, seenPath string
	c := newPipelineOIDCServer(t, func(w http.ResponseWriter, r *http.Request) {
		seenMethod = r.Method
		seenPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"keys": []map[string]any{
				{
					"kid": "key-id-1",
					"kty": "RSA",
					"alg": "RS256",
					"use": "sig",
					"n":   "modulus-base64",
					"e":   "AQAB",
				},
			},
		})
	})
	got, err := c.GetPipelineOIDCKeys("myworkspace")
	require.NoError(t, err)
	assert.Equal(t, http.MethodGet, seenMethod)
	assert.Contains(t, seenPath, "/workspaces/myworkspace/pipelines-config/identity/oidc/keys.json")
	require.Len(t, got.Keys, 1)
	assert.Equal(t, "key-id-1", got.Keys[0].Kid)
	assert.Equal(t, "RSA", got.Keys[0].Kty)
	assert.Equal(t, "RS256", got.Keys[0].Alg)
	assert.Equal(t, "sig", got.Keys[0].Use)
	assert.Equal(t, "modulus-base64", got.Keys[0].N)
	assert.Equal(t, "AQAB", got.Keys[0].E)
}

func TestCloudClient_GetPipelineOIDCKeys_Empty(t *testing.T) {
	t.Parallel()
	c := newPipelineOIDCServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []any{}})
	})
	got, err := c.GetPipelineOIDCKeys("myworkspace")
	require.NoError(t, err)
	assert.Empty(t, got.Keys)
}
