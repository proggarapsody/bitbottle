package mcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestGetPipelineOIDCConfig_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelineOIDCConfigFn: func(workspace string) (backend.PipelineOIDCConfig, error) {
			assert.Equal(t, "myws", workspace)
			return backend.PipelineOIDCConfig{
				Issuer:  "https://api.bitbucket.org/2.0/workspaces/myws/pipelines-config/identity/oidc",
				JWKSURI: "https://api.bitbucket.org/2.0/workspaces/myws/pipelines-config/identity/oidc/keys.json",
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.getPipelineOIDCConfig(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "issuer", "jwks_uri")
}

func TestGetPipelineOIDCConfig_MissingWorkspace(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.getPipelineOIDCConfig(context.Background(), makeReq(map[string]any{}))
	require.NoError(t, err)
	assertErrorResult(t, result, "workspace")
}

func TestGetPipelineOIDCConfig_UnsupportedOnHost(t *testing.T) {
	t.Parallel()
	type noOIDCFake struct{ backend.Client }
	const serverConfig = "git.example.com:\n  oauth_token: tok\n"
	base := &testhelpers.FakeClient{T: t}
	noOIDC := &noOIDCFake{Client: base}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, noOIDC)
	h := newHandlers(f)
	result, err := h.getPipelineOIDCConfig(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "unsupported")
}

func TestGetPipelineOIDCKeys_Success(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelineOIDCKeysFn: func(workspace string) (backend.PipelineOIDCKeys, error) {
			assert.Equal(t, "myws", workspace)
			return backend.PipelineOIDCKeys{
				Keys: []backend.PipelineOIDCKey{
					{Kid: "key-1", Kty: "RSA", Alg: "RS256", Use: "sig"},
				},
			}, nil
		},
	}
	h := newHandlersWithFake(t, singleCloudConfig, fake)
	result, err := h.getPipelineOIDCKeys(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
	}))
	require.NoError(t, err)
	assertJSONContains(t, result, "keys", "key-1")
}

func TestGetPipelineOIDCKeys_MissingWorkspace(t *testing.T) {
	t.Parallel()
	h := newHandlersWithFake(t, singleCloudConfig, &testhelpers.FakeClient{T: t})
	result, err := h.getPipelineOIDCKeys(context.Background(), makeReq(map[string]any{}))
	require.NoError(t, err)
	assertErrorResult(t, result, "workspace")
}

func TestGetPipelineOIDCKeys_UnsupportedOnHost(t *testing.T) {
	t.Parallel()
	type noOIDCFake struct{ backend.Client }
	const serverConfig = "git.example.com:\n  oauth_token: tok\n"
	base := &testhelpers.FakeClient{T: t}
	noOIDC := &noOIDCFake{Client: base}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, noOIDC)
	h := newHandlers(f)
	result, err := h.getPipelineOIDCKeys(context.Background(), makeReq(map[string]any{
		"workspace": "myws",
	}))
	require.NoError(t, err)
	assertErrorResult(t, result, "unsupported")
}
