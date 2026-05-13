package sshkey_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/sshkey"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestDelete_DeletesKey(t *testing.T) {
	t.Parallel()
	deleted := false
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteSSHKeyFn: func(id int) error {
			assert.Equal(t, 7, id)
			deleted = true
			return nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := sshkey.NewCmdDelete(f)
	cmd.SetArgs([]string{"--hostname", "bitbucket.org", "7"})
	require.NoError(t, cmd.Execute())
	assert.True(t, deleted)
	assert.Contains(t, out.String(), "7")
}

func TestDelete_InvalidID(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := sshkey.NewCmdDelete(f)
	cmd.SetArgs([]string{"--hostname", "bitbucket.org", "notanumber"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "positive integer")
}

func TestDelete_MissingID(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := sshkey.NewCmdDelete(f)
	cmd.SetArgs([]string{"--hostname", "bitbucket.org"})
	require.Error(t, cmd.Execute())
}
