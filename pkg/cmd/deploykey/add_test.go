package deploykey_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/deploykey"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestAdd_AddsKey(t *testing.T) {
	t.Parallel()
	var gotInput backend.DeployKeyInput
	fake := &testhelpers.FakeClient{
		T: t,
		AddDeployKeyFn: func(ns, slug string, input backend.DeployKeyInput) (backend.DeployKey, error) {
			gotInput = input
			return backend.DeployKey{ID: 42, Label: input.Label, Key: input.Key}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := deploykey.NewCmdAdd(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--key", "ssh-rsa AAAA", "--label", "CI"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "ssh-rsa AAAA", gotInput.Key)
	assert.Equal(t, "CI", gotInput.Label)
	got := out.String()
	assert.Contains(t, got, "CI")
}

func TestAdd_RequiresKey(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := deploykey.NewCmdAdd(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.Error(t, cmd.Execute())
}

func TestAdd_PermissionReadWrite(t *testing.T) {
	t.Parallel()
	var gotInput backend.DeployKeyInput
	fake := &testhelpers.FakeClient{
		T: t,
		AddDeployKeyFn: func(ns, slug string, input backend.DeployKeyInput) (backend.DeployKey, error) {
			gotInput = input
			return backend.DeployKey{ID: 7, Label: input.Label, Key: input.Key}, nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := deploykey.NewCmdAdd(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--key", "ssh-rsa BBBB", "--permission", "read-write"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "read_write", gotInput.Permission)
}

func TestAdd_PermissionRead(t *testing.T) {
	t.Parallel()
	var gotInput backend.DeployKeyInput
	fake := &testhelpers.FakeClient{
		T: t,
		AddDeployKeyFn: func(ns, slug string, input backend.DeployKeyInput) (backend.DeployKey, error) {
			gotInput = input
			return backend.DeployKey{ID: 8, Label: input.Label, Key: input.Key}, nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := deploykey.NewCmdAdd(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--key", "ssh-rsa CCCC", "--permission", "read"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "read", gotInput.Permission)
}

func TestAdd_PermissionInvalid(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := deploykey.NewCmdAdd(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--key", "ssh-rsa DDDD", "--permission", "write"})
	require.Error(t, cmd.Execute())
}
