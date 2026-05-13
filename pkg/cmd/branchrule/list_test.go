package branchrule_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/branchrule"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdList_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := branchrule.NewCmdList(f)
	format.RegisterOutputFlags(cmd)
	assert.NotNil(t, cmd.Flag("hostname"))
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("jq"))
}

func TestList_PrintsBranchRules(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListBranchRulesFn: func(ns, slug string) ([]backend.BranchRule, error) {
			return []backend.BranchRule{
				{ID: 1, Kind: "require_approvals_to_merge", Pattern: "main", Value: 2},
				{ID: 2, Kind: "push", Pattern: "main"},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := branchrule.NewCmdList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "require_approvals_to_merge")
	assert.Contains(t, got, "push")
}

func TestList_JSON_Output(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListBranchRulesFn: func(ns, slug string) ([]backend.BranchRule, error) {
			return []backend.BranchRule{
				{ID: 1, Kind: "push", Pattern: "main"},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := branchrule.NewCmdList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--json"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, `"kind":"push"`)
}

func TestList_UnsupportedBackend(t *testing.T) {
	t.Parallel()
	fake := &cmdtest.NoBranchRuleFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := branchrule.NewCmdList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "branch restriction rules")
}
