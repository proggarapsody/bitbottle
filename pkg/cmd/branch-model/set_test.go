package branchmodel_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	branchmodel "github.com/proggarapsody/bitbottle/pkg/cmd/branch-model"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdSet_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := branchmodel.NewCmdSet(f)
	assert.NotNil(t, cmd.Flag("hostname"))
	assert.NotNil(t, cmd.Flag("dev-branch"))
	assert.NotNil(t, cmd.Flag("prod-branch"))
	assert.NotNil(t, cmd.Flag("prod-enabled"))
	assert.NotNil(t, cmd.Flag("branch-type-prefix"))
	assert.NotNil(t, cmd.Flag("json"))
}

func TestSet_UpdatesDevBranch(t *testing.T) {
	t.Parallel()
	var gotInput backend.BranchModelSettingsInput
	fake := &testhelpers.FakeClient{
		T: t,
		GetBranchModelSettingsFn: func(ws, slug string) (backend.BranchModelSettings, error) {
			return backend.BranchModelSettings{
				Development: backend.BranchModelSettingsBranch{Name: "main", UseMainbranch: true},
				BranchTypes: []backend.BranchTypeSettings{
					{Enabled: true, Kind: "feature", Prefix: "feature/"},
				},
			}, nil
		},
		UpdateBranchModelSettingsFn: func(ws, slug string, in backend.BranchModelSettingsInput) (backend.BranchModelSettings, error) {
			assert.Equal(t, "myworkspace", ws)
			assert.Equal(t, "my-service", slug)
			gotInput = in
			return backend.BranchModelSettings{
				Development: backend.BranchModelSettingsBranch{Name: "develop"},
				BranchTypes: []backend.BranchTypeSettings{},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := branchmodel.NewCmdSet(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "--dev-branch", "develop"})
	require.NoError(t, cmd.Execute())
	require.NotNil(t, gotInput.Development)
	assert.Equal(t, "develop", gotInput.Development.Name)
	assert.Contains(t, out.String(), "develop")
}

func TestSet_BranchTypePrefixes_Merged(t *testing.T) {
	t.Parallel()
	var gotBranchTypes []backend.BranchTypeSettings
	fake := &testhelpers.FakeClient{
		T: t,
		GetBranchModelSettingsFn: func(ws, slug string) (backend.BranchModelSettings, error) {
			return backend.BranchModelSettings{
				BranchTypes: []backend.BranchTypeSettings{
					{Enabled: true, Kind: "feature", Prefix: "feature/"},
					{Enabled: true, Kind: "hotfix", Prefix: "hotfix/"},
				},
			}, nil
		},
		UpdateBranchModelSettingsFn: func(ws, slug string, in backend.BranchModelSettingsInput) (backend.BranchModelSettings, error) {
			gotBranchTypes = in.BranchTypes
			return backend.BranchModelSettings{}, nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := branchmodel.NewCmdSet(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "--branch-type-prefix", "feature=feat/,hotfix=hf/"})
	require.NoError(t, cmd.Execute())
	require.Len(t, gotBranchTypes, 2)
	prefixes := map[string]string{}
	for _, bt := range gotBranchTypes {
		prefixes[bt.Kind] = bt.Prefix
	}
	assert.Equal(t, "feat/", prefixes["feature"])
	assert.Equal(t, "hf/", prefixes["hotfix"])
}

func TestSet_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetBranchModelSettingsFn: func(ws, slug string) (backend.BranchModelSettings, error) {
			return backend.BranchModelSettings{}, nil
		},
		UpdateBranchModelSettingsFn: func(ws, slug string, in backend.BranchModelSettingsInput) (backend.BranchModelSettings, error) {
			return backend.BranchModelSettings{
				Development: backend.BranchModelSettingsBranch{Name: "main"},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := branchmodel.NewCmdSet(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "--dev-branch", "main", "--json"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), `"name":"main"`)
}

func TestSet_UnsupportedBackend(t *testing.T) {
	t.Parallel()
	fake := &cmdtest.NoBranchModelFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := branchmodel.NewCmdSet(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "--dev-branch", "main"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "branching model")
}

func TestParseBranchTypePrefixes_Invalid(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetBranchModelSettingsFn: func(ws, slug string) (backend.BranchModelSettings, error) {
			return backend.BranchModelSettings{}, nil
		},
		UpdateBranchModelSettingsFn: func(ws, slug string, in backend.BranchModelSettingsInput) (backend.BranchModelSettings, error) {
			return backend.BranchModelSettings{}, nil
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := branchmodel.NewCmdSet(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "--branch-type-prefix", "no-equals-sign"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "kind=prefix")
}
