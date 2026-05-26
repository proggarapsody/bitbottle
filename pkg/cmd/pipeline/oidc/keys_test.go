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

func TestNewCmdKeys_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := oidc.NewCmdKeys(f, nil)
	assert.NotNil(t, cmd.Flag("hostname"))
	assert.NotNil(t, cmd.Flag("json"))
}

func TestNewCmdKeys_RequiresWorkspaceArg(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := oidc.NewCmdKeys(f, nil)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestPipelineOIDCKeys_PrintsTable(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelineOIDCKeysFn: func(workspace string) (backend.PipelineOIDCKeys, error) {
			assert.Equal(t, "myworkspace", workspace)
			return backend.PipelineOIDCKeys{
				Keys: []backend.PipelineOIDCKey{
					{Kid: "key-1", Kty: "RSA", Alg: "RS256", Use: "sig", N: "n-val", E: "AQAB"},
					{Kid: "key-2", Kty: "EC", Alg: "ES256", Use: "sig", N: "", E: ""},
				},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := oidc.NewCmdKeys(f, nil)
	cmd.SetArgs([]string{"myworkspace"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "key-1")
	assert.Contains(t, got, "RSA")
	assert.Contains(t, got, "RS256")
	assert.Contains(t, got, "key-2")
}

func TestPipelineOIDCKeys_Empty(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelineOIDCKeysFn: func(workspace string) (backend.PipelineOIDCKeys, error) {
			return backend.PipelineOIDCKeys{Keys: nil}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := oidc.NewCmdKeys(f, nil)
	cmd.SetArgs([]string{"myworkspace"})
	require.NoError(t, cmd.Execute())
	assert.Empty(t, out.String())
}

func TestPipelineOIDCKeys_JSON(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelineOIDCKeysFn: func(workspace string) (backend.PipelineOIDCKeys, error) {
			return backend.PipelineOIDCKeys{
				Keys: []backend.PipelineOIDCKey{
					{Kid: "k1", Kty: "RSA", Alg: "RS256", Use: "sig"},
				},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := oidc.NewCmdKeys(f, nil)
	cmd.SetArgs([]string{"myworkspace", "--json"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), `"keys"`)
}

func TestPipelineOIDCKeys_UnsupportedOnServer(t *testing.T) {
	t.Parallel()
	type noOIDCFake struct{ backend.Client }
	fake := &noOIDCFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := oidc.NewCmdKeys(f, nil)
	cmd.SetArgs([]string{"myworkspace"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipeline OIDC")
}
