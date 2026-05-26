package oidc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/oidc"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdConfig_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := oidc.NewCmdConfig(f, nil)
	assert.NotNil(t, cmd.Flag("hostname"))
	assert.NotNil(t, cmd.Flag("json"))
}

func TestNewCmdConfig_RequiresWorkspaceArg(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := oidc.NewCmdConfig(f, nil)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestPipelineOIDCConfig_PrintsKeyValues(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelineOIDCConfigFn: func(workspace string) (backend.PipelineOIDCConfig, error) {
			assert.Equal(t, "myworkspace", workspace)
			return backend.PipelineOIDCConfig{
				Issuer:                 "https://api.bitbucket.org/2.0/workspaces/myworkspace/pipelines-config/identity/oidc",
				JWKSURI:                "https://api.bitbucket.org/2.0/workspaces/myworkspace/pipelines-config/identity/oidc/keys.json",
				SubjectTypesSupported:  []string{"public"},
				ResponseTypesSupported: []string{"id_token"},
				ClaimsSupported:        []string{"sub", "iss"},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := oidc.NewCmdConfig(f, nil)
	cmd.SetArgs([]string{"myworkspace"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "issuer=")
	assert.Contains(t, got, "jwks_uri=")
}

func TestPipelineOIDCConfig_JSON(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelineOIDCConfigFn: func(workspace string) (backend.PipelineOIDCConfig, error) {
			return backend.PipelineOIDCConfig{
				Issuer:  "https://example.com",
				JWKSURI: "https://example.com/keys",
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := oidc.NewCmdConfig(f, nil)
	cmd.SetArgs([]string{"myworkspace", "--json"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), `"issuer"`)
	assert.Contains(t, out.String(), `"jwks_uri"`)
}

func TestPipelineOIDCConfig_UnsupportedOnServer(t *testing.T) {
	t.Parallel()
	type noOIDCFake struct{ backend.Client }
	fake := &noOIDCFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := oidc.NewCmdConfig(f, nil)
	cmd.SetArgs([]string{"myworkspace"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipeline OIDC")
}
