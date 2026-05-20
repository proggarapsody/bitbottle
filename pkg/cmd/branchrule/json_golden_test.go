package branchrule_test

// json_golden_test.go — golden-file tests for --json field stability.
//
// To regenerate golden files after an intentional field rename:
//
//	BITBOTTLE_UPDATE_GOLDEN=1 go test ./pkg/cmd/branchrule/...

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/branchrule"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestBranchRuleList_JSONGolden(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListBranchRulesFn: func(ns, slug string) ([]backend.BranchRule, error) {
			return []backend.BranchRule{
				{
					ID:      1,
					Kind:    "require_approvals_to_merge",
					Pattern: "main",
					Value:   2,
				},
			}, nil
		},
	}

	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := branchrule.NewCmdList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--json"})
	require.NoError(t, cmd.Execute())

	testhelpers.AssertGolden(t, "json/branchrule-list", out.String())
}
