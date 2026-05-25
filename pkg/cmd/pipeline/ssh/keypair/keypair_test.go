package keypair_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/ssh/keypair"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdView_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := keypair.NewCmdView(f, nil)
	assert.NotNil(t, cmd.Flag("hostname"))
}

func TestSSHKeyPairView_PrintsKeyType(t *testing.T) {
	t.Parallel()
	ts := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelineSSHKeyPairFn: func(ns, slug string) (backend.PipelineSSHKeyPair, error) {
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "my-service", slug)
			return backend.PipelineSSHKeyPair{
				PublicKey:    "ssh-rsa AAAA...",
				KeyTypeLabel: "RSA",
				Created:      ts,
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := keypair.NewCmdView(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "RSA")
}

func TestSSHKeyPairView_JSON(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelineSSHKeyPairFn: func(ns, slug string) (backend.PipelineSSHKeyPair, error) {
			return backend.PipelineSSHKeyPair{
				PublicKey:    "ssh-rsa AAAA...",
				KeyTypeLabel: "RSA",
				Created:      time.Now(),
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := keypair.NewCmdView(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "--json"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), `"public_key"`)
	assert.Contains(t, out.String(), `"key_type"`)
}

func TestSSHKeyPairView_UnsupportedOnHost(t *testing.T) {
	t.Parallel()
	fake := &noSSHKeyPairFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := keypair.NewCmdView(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipeline SSH key pair")
}

func TestSSHKeyPairRegenerate_RequiresConfirmInNonTTY(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := keypair.NewCmdRegenerate(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestSSHKeyPairRegenerate_WithConfirm(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		RegeneratePipelineSSHKeyPairFn: func(ns, slug string, bits int) (backend.PipelineSSHKeyPair, error) {
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "my-service", slug)
			assert.Equal(t, 4096, bits)
			return backend.PipelineSSHKeyPair{
				PublicKey:    "ssh-rsa BBBB...",
				KeyTypeLabel: "RSA",
				Created:      time.Now(),
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := keypair.NewCmdRegenerate(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "--confirm", "--bits", "4096"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Regenerated")
}

// noSSHKeyPairFake wraps backend.Client without implementing PipelineSSHKeyPairClient.
type noSSHKeyPairFake struct {
	backend.Client
}
