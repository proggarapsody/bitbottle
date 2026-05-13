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

func TestAdd_AddsRule(t *testing.T) {
	t.Parallel()
	var gotInput backend.BranchRuleInput
	fake := &testhelpers.FakeClient{
		T: t,
		AddBranchRuleFn: func(ns, slug string, input backend.BranchRuleInput) (backend.BranchRule, error) {
			gotInput = input
			return backend.BranchRule{ID: 42, Kind: input.Kind, Pattern: input.Pattern, Value: input.Value}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := branchrule.NewCmdAdd(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--kind", "push", "--pattern", "main"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "push", gotInput.Kind)
	assert.Equal(t, "main", gotInput.Pattern)
	got := out.String()
	assert.Contains(t, got, "push")
}

func TestAdd_WithValue(t *testing.T) {
	t.Parallel()
	var gotInput backend.BranchRuleInput
	fake := &testhelpers.FakeClient{
		T: t,
		AddBranchRuleFn: func(ns, slug string, input backend.BranchRuleInput) (backend.BranchRule, error) {
			gotInput = input
			return backend.BranchRule{ID: 5, Kind: input.Kind, Pattern: input.Pattern, Value: input.Value}, nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := branchrule.NewCmdAdd(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--kind", "require_approvals_to_merge", "--pattern", "main", "--value", "2"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, 2, gotInput.Value)
}

func TestAdd_RequiresKind(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := branchrule.NewCmdAdd(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--pattern", "main"})
	require.Error(t, cmd.Execute())
}

func TestAdd_RequiresPattern(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := branchrule.NewCmdAdd(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "--kind", "push"})
	require.Error(t, cmd.Execute())
}
