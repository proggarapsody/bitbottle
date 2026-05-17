package branch_test

// json_golden_test.go — golden-file tests for --json field stability.
//
// To regenerate golden files after an intentional field rename:
//
//	BITBOTTLE_UPDATE_GOLDEN=1 go test ./pkg/cmd/branch/...

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/branch"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestBranchList_JSONGolden(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListBranchesFn: func(ns, slug string, limit int) ([]backend.Branch, error) {
			return []backend.Branch{
				{
					Name:       "main",
					IsDefault:  true,
					LatestHash: "abc1234def567890",
				},
			}, nil
		},
	}

	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := branch.NewCmdBranchList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--json"})
	require.NoError(t, cmd.Execute())

	testhelpers.AssertGolden(t, "json/branch-list", out.String())
}
