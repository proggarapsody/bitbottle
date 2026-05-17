package deploykey_test

// json_golden_test.go — golden-file tests for --json field stability.
//
// To regenerate golden files after an intentional field rename:
//
//	BITBOTTLE_UPDATE_GOLDEN=1 go test ./pkg/cmd/deploykey/...

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/deploykey"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestDeployKeyList_JSONGolden(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListDeployKeysFn: func(ns, slug string) ([]backend.DeployKey, error) {
			return []backend.DeployKey{
				{
					ID:       1,
					Label:    "CI deploy key",
					Key:      "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC",
					ReadOnly: true,
				},
			}, nil
		},
	}

	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := deploykey.NewCmdList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--json"})
	require.NoError(t, cmd.Execute())

	testhelpers.AssertGolden(t, "json/deploykey-list", out.String())
}
