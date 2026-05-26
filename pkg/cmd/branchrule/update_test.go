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

func TestUpdate_UpdatesPattern(t *testing.T) {
	t.Parallel()
	var gotID int
	var gotInput backend.UpdateBranchRuleInput
	fake := &testhelpers.FakeClient{
		T: t,
		UpdateBranchRuleFn: func(ns, slug string, id int, in backend.UpdateBranchRuleInput) (backend.BranchRule, error) {
			gotID = id
			gotInput = in
			return backend.BranchRule{ID: id, Kind: "push", Pattern: *in.Pattern}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := branchrule.NewCmdUpdate(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "5", "--pattern", "release/*"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, 5, gotID)
	require.NotNil(t, gotInput.Pattern)
	assert.Equal(t, "release/*", *gotInput.Pattern)
	assert.Nil(t, gotInput.Value)
	assert.Nil(t, gotInput.Users)
	assert.Nil(t, gotInput.Groups)
	got := out.String()
	assert.Contains(t, got, "push")
}

func TestUpdate_UpdatesValue(t *testing.T) {
	t.Parallel()
	var gotInput backend.UpdateBranchRuleInput
	fake := &testhelpers.FakeClient{
		T: t,
		UpdateBranchRuleFn: func(ns, slug string, id int, in backend.UpdateBranchRuleInput) (backend.BranchRule, error) {
			gotInput = in
			return backend.BranchRule{ID: id, Kind: "require_approvals_to_merge", Pattern: "main", Value: *in.Value}, nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := branchrule.NewCmdUpdate(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "3", "--value", "3"})
	require.NoError(t, cmd.Execute())
	require.NotNil(t, gotInput.Value)
	assert.Equal(t, 3, *gotInput.Value)
}

func TestUpdate_UpdatesUsers(t *testing.T) {
	t.Parallel()
	var gotInput backend.UpdateBranchRuleInput
	fake := &testhelpers.FakeClient{
		T: t,
		UpdateBranchRuleFn: func(ns, slug string, id int, in backend.UpdateBranchRuleInput) (backend.BranchRule, error) {
			gotInput = in
			return backend.BranchRule{ID: id, Kind: "push", Pattern: "main"}, nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := branchrule.NewCmdUpdate(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "2", "--users", "alice,bob"})
	require.NoError(t, cmd.Execute())
	require.NotNil(t, gotInput.Users)
	assert.Equal(t, []string{"alice", "bob"}, *gotInput.Users)
}

func TestUpdate_UpdatesGroups(t *testing.T) {
	t.Parallel()
	var gotInput backend.UpdateBranchRuleInput
	fake := &testhelpers.FakeClient{
		T: t,
		UpdateBranchRuleFn: func(ns, slug string, id int, in backend.UpdateBranchRuleInput) (backend.BranchRule, error) {
			gotInput = in
			return backend.BranchRule{ID: id, Kind: "push", Pattern: "main"}, nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := branchrule.NewCmdUpdate(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "2", "--groups", "dev-team"})
	require.NoError(t, cmd.Execute())
	require.NotNil(t, gotInput.Groups)
	assert.Equal(t, []string{"dev-team"}, *gotInput.Groups)
}

func TestUpdate_RequiresAtLeastOneFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := branchrule.NewCmdUpdate(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", "5"})
	require.Error(t, cmd.Execute())
}

func TestUpdate_RequiresID(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := branchrule.NewCmdUpdate(f)
	format.RegisterOutputFlags(cmd)
	// no ID provided — only repo arg
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.Error(t, cmd.Execute())
}
