package sshkey_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/sshkey"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestAdd_AddsKey(t *testing.T) {
	t.Parallel()
	var gotInput backend.SSHKeyInput
	fake := &testhelpers.FakeClient{
		T: t,
		AddSSHKeyFn: func(input backend.SSHKeyInput) (backend.SSHKey, error) {
			gotInput = input
			return backend.SSHKey{ID: 42, Label: input.Label, Key: input.Key}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := sshkey.NewCmdAdd(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"--hostname", "bitbucket.org", "--key", "ssh-rsa AAAA", "--label", "CI"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "ssh-rsa AAAA", gotInput.Key)
	assert.Equal(t, "CI", gotInput.Label)
	got := out.String()
	assert.Contains(t, got, "CI")
}

func TestAdd_RequiresKey(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := sshkey.NewCmdAdd(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"--hostname", "bitbucket.org"})
	require.Error(t, cmd.Execute())
}
