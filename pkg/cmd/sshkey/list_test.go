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

func TestNewCmdList_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := sshkey.NewCmdList(f)
	format.RegisterOutputFlags(cmd)
	assert.NotNil(t, cmd.Flag("hostname"))
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("jq"))
}

func TestList_PrintsSSHKeys(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListSSHKeysFn: func() ([]backend.SSHKey, error) {
			return []backend.SSHKey{
				{ID: 1, Label: "Laptop key", Key: "ssh-rsa AAAA1"},
				{ID: 2, Label: "Work key", Key: "ssh-rsa AAAA2"},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := sshkey.NewCmdList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"--hostname", "bitbucket.org"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "Laptop key")
	assert.Contains(t, got, "Work key")
}

func TestList_JSON_Output(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListSSHKeysFn: func() ([]backend.SSHKey, error) {
			return []backend.SSHKey{
				{ID: 1, Label: "Laptop key", Key: "ssh-rsa AAAA1"},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := sshkey.NewCmdList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"--hostname", "bitbucket.org", "--json"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, `"label":"Laptop key"`)
}

func TestList_UnsupportedBackend(t *testing.T) {
	t.Parallel()
	fake := &cmdtest.NoSSHKeyFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := sshkey.NewCmdList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"--hostname", "bitbucket.org"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SSH keys")
}
