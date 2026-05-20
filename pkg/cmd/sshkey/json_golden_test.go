package sshkey_test

// json_golden_test.go — golden-file tests for --json field stability.
//
// To regenerate golden files after an intentional field rename:
//
//	BITBOTTLE_UPDATE_GOLDEN=1 go test ./pkg/cmd/sshkey/...

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/sshkey"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestSSHKeyList_JSONGolden(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListSSHKeysFn: func() ([]backend.SSHKey, error) {
			return []backend.SSHKey{
				{
					ID:    1,
					Label: "Laptop key",
					Key:   "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDexample",
				},
			}, nil
		},
	}

	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := sshkey.NewCmdList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"--hostname", "bitbucket.org", "--json"})
	require.NoError(t, cmd.Execute())

	testhelpers.AssertGolden(t, "json/sshkey-list", out.String())
}
