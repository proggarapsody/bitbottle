package cloud

import (
	"fmt"
	"net/url"

	"github.com/proggarapsody/bitbottle/api/backend"
)

const (
	pipelineOIDCConfigPath = "/workspaces/%s/pipelines-config/identity/oidc/.well-known/openid-configuration"
	pipelineOIDCKeysPath   = "/workspaces/%s/pipelines-config/identity/oidc/keys.json"
)

// cloudPipelineOIDCConfig is the wire representation of the Cloud OIDC
// discovery document.
type cloudPipelineOIDCConfig struct {
	Issuer                 string   `json:"issuer"`
	JWKSURI                string   `json:"jwks_uri"`
	SubjectTypesSupported  []string `json:"subject_types_supported"`
	ResponseTypesSupported []string `json:"response_types_supported"`
	ClaimsSupported        []string `json:"claims_supported"`
}

// cloudPipelineOIDCKey is a single JSON Web Key from the JWKS response.
type cloudPipelineOIDCKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// cloudPipelineOIDCKeys is the wire representation of the JWKS response.
type cloudPipelineOIDCKeys struct {
	Keys []cloudPipelineOIDCKey `json:"keys"`
}

// GetPipelineOIDCConfig returns the OIDC discovery document for a workspace.
// GET /2.0/workspaces/{workspace}/pipelines-config/identity/oidc/.well-known/openid-configuration
func (c *Client) GetPipelineOIDCConfig(workspace string) (backend.PipelineOIDCConfig, error) {
	path := fmt.Sprintf(pipelineOIDCConfigPath, url.PathEscape(workspace))
	var w cloudPipelineOIDCConfig
	if err := c.getJSON(path, &w); err != nil {
		return backend.PipelineOIDCConfig{}, err
	}
	return backend.PipelineOIDCConfig{
		Issuer:                 w.Issuer,
		JWKSURI:                w.JWKSURI,
		SubjectTypesSupported:  w.SubjectTypesSupported,
		ResponseTypesSupported: w.ResponseTypesSupported,
		ClaimsSupported:        w.ClaimsSupported,
	}, nil
}

// GetPipelineOIDCKeys returns the JWKS key-set for a workspace.
// GET /2.0/workspaces/{workspace}/pipelines-config/identity/oidc/keys.json
func (c *Client) GetPipelineOIDCKeys(workspace string) (backend.PipelineOIDCKeys, error) {
	path := fmt.Sprintf(pipelineOIDCKeysPath, url.PathEscape(workspace))
	var w cloudPipelineOIDCKeys
	if err := c.getJSON(path, &w); err != nil {
		return backend.PipelineOIDCKeys{}, err
	}
	keys := make([]backend.PipelineOIDCKey, len(w.Keys))
	for i, k := range w.Keys {
		keys[i] = backend.PipelineOIDCKey{
			Kid: k.Kid,
			Kty: k.Kty,
			Alg: k.Alg,
			Use: k.Use,
			N:   k.N,
			E:   k.E,
		}
	}
	return backend.PipelineOIDCKeys{Keys: keys}, nil
}
