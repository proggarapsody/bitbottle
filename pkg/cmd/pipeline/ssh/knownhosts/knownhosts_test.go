package knownhosts_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/ssh/knownhosts"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdList_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := knownhosts.NewCmdList(f, nil)
	assert.NotNil(t, cmd.Flag("hostname"))
}

func TestKnownHostsList_PrintsHosts(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineKnownHostsFn: func(ns, slug string) ([]backend.PipelineKnownHost, error) {
			assert.Equal(t, "myworkspace", ns)
			assert.Equal(t, "my-service", slug)
			return []backend.PipelineKnownHost{
				{UUID: "uuid-1", Hostname: "github.com", PublicKey: backend.PipelineSSHPublicKey{KeyType: "RSA", MD5: "aa:bb:cc"}},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := knownhosts.NewCmdList(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "github.com")
}

func TestKnownHostsList_JSON(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPipelineKnownHostsFn: func(ns, slug string) ([]backend.PipelineKnownHost, error) {
			return []backend.PipelineKnownHost{
				{UUID: "uuid-1", Hostname: "github.com", PublicKey: backend.PipelineSSHPublicKey{KeyType: "RSA"}},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := knownhosts.NewCmdList(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service", "--json"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), `"hostname"`)
}

func TestKnownHostsView_PrintsHost(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelineKnownHostFn: func(ns, slug, uuid string) (backend.PipelineKnownHost, error) {
			assert.Equal(t, "uuid-1", uuid)
			return backend.PipelineKnownHost{
				UUID:      "uuid-1",
				Hostname:  "github.com",
				PublicKey: backend.PipelineSSHPublicKey{KeyType: "RSA", MD5: "aa:bb:cc"},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := knownhosts.NewCmdKHView(f, nil)
	cmd.SetArgs([]string{"uuid-1", "myworkspace/my-service"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "github.com")
}

func TestKnownHostsAdd_PrintsAdded(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		AddPipelineKnownHostFn: func(ns, slug string, in backend.PipelineKnownHostInput) (backend.PipelineKnownHost, error) {
			assert.Equal(t, "github.com", in.Hostname)
			return backend.PipelineKnownHost{
				UUID:      "new-uuid",
				Hostname:  "github.com",
				PublicKey: backend.PipelineSSHPublicKey{KeyType: "RSA"},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := knownhosts.NewCmdAdd(f, nil)
	cmd.SetArgs([]string{"github.com", "myworkspace/my-service", "--key", "AAAA...", "--key-type", "RSA"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Added known host")
}

func TestKnownHostsDelete_RequiresConfirmInNonTTY(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := knownhosts.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"uuid-1", "myworkspace/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestKnownHostsDelete_WithConfirm(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		DeletePipelineKnownHostFn: func(ns, slug, uuid string) error {
			assert.Equal(t, "uuid-1", uuid)
			return nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := knownhosts.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"uuid-1", "myworkspace/my-service", "--confirm"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Deleted known host")
}

func TestKnownHostsList_UnsupportedOnHost(t *testing.T) {
	t.Parallel()
	fake := &noKnownHostsFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := knownhosts.NewCmdList(f, nil)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pipeline known hosts")
}

// noKnownHostsFake wraps backend.Client without implementing PipelineKnownHostsClient.
type noKnownHostsFake struct {
	backend.Client
}
