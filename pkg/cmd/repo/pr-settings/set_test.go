package prsettings_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	prsettings "github.com/proggarapsody/bitbottle/pkg/cmd/repo/pr-settings"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestRepoPRSettingsSet_UpdatesApprovers(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug string
	var gotIn backend.RepoPRSettingsInput
	fake := &testhelpers.FakeClient{
		T: t,
		UpdateRepoPRSettingsFn: func(ns, slug string, in backend.RepoPRSettingsInput) (backend.RepoPRSettings, error) {
			gotNS, gotSlug, gotIn = ns, slug, in
			return backend.RepoPRSettings{RequiredApprovers: 3}, nil
		},
	}
	f, out, _ := newFactory(t, fake)
	cmd := prsettings.NewCmdPRSettings(f)
	cmd.SetArgs([]string{"set", "MYPROJ/my-repo", "--required-approvers", "3"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-repo", gotSlug)
	require.NotNil(t, gotIn.RequiredApprovers)
	assert.Equal(t, 3, *gotIn.RequiredApprovers)
	assert.Nil(t, gotIn.MergeStrategy)
	assert.Contains(t, out.String(), "updated")
}

func TestRepoPRSettingsSet_UpdatesMergeStrategy(t *testing.T) {
	t.Parallel()
	var gotIn backend.RepoPRSettingsInput
	fake := &testhelpers.FakeClient{
		T: t,
		UpdateRepoPRSettingsFn: func(ns, slug string, in backend.RepoPRSettingsInput) (backend.RepoPRSettings, error) {
			gotIn = in
			return backend.RepoPRSettings{MergeStrategy: "squash"}, nil
		},
	}
	f, _, _ := newFactory(t, fake)
	cmd := prsettings.NewCmdPRSettings(f)
	cmd.SetArgs([]string{"set", "MYPROJ/my-repo", "--merge-strategy", "squash"})
	require.NoError(t, cmd.Execute())

	require.NotNil(t, gotIn.MergeStrategy)
	assert.Equal(t, "squash", *gotIn.MergeStrategy)
	assert.Nil(t, gotIn.RequiredApprovers)
}

func TestRepoPRSettingsSet_AllowedStrategies_Parsed(t *testing.T) {
	t.Parallel()
	var gotIn backend.RepoPRSettingsInput
	fake := &testhelpers.FakeClient{
		T: t,
		UpdateRepoPRSettingsFn: func(ns, slug string, in backend.RepoPRSettingsInput) (backend.RepoPRSettings, error) {
			gotIn = in
			return backend.RepoPRSettings{}, nil
		},
	}
	f, _, _ := newFactory(t, fake)
	cmd := prsettings.NewCmdPRSettings(f)
	cmd.SetArgs([]string{"set", "MYPROJ/my-repo", "--allowed-strategies", "squash,no-ff"})
	require.NoError(t, cmd.Execute())

	require.NotNil(t, gotIn.AllowedStrategies)
	assert.Equal(t, []string{"squash", "no-ff"}, *gotIn.AllowedStrategies)
}

func TestRepoPRSettingsSet_NoFlags_ReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := newFactory(t, &testhelpers.FakeClient{T: t})
	cmd := prsettings.NewCmdPRSettings(f)
	cmd.SetArgs([]string{"set", "MYPROJ/my-repo"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one flag")
}

func TestRepoPRSettingsSet_APIError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		UpdateRepoPRSettingsFn: func(ns, slug string, in backend.RepoPRSettingsInput) (backend.RepoPRSettings, error) {
			return backend.RepoPRSettings{}, errors.New("403 forbidden")
		},
	}
	f, _, _ := newFactory(t, fake)
	cmd := prsettings.NewCmdPRSettings(f)
	cmd.SetArgs([]string{"set", "MYPROJ/my-repo", "--required-approvers", "1"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}

func TestRepoPRSettingsSet_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		UpdateRepoPRSettingsFn: func(ns, slug string, in backend.RepoPRSettingsInput) (backend.RepoPRSettings, error) {
			return backend.RepoPRSettings{RequiredApprovers: 5, MergeStrategy: "no-ff"}, nil
		},
	}
	f, out, _ := newFactory(t, fake)
	cmd := prsettings.NewCmdPRSettings(f)
	cmd.SetArgs([]string{"set", "MYPROJ/my-repo", "--required-approvers", "5", "--json"})
	require.NoError(t, cmd.Execute())

	output := out.String()
	assert.Contains(t, output, `"requiredApprovers"`)
	assert.Contains(t, output, "5")
}
