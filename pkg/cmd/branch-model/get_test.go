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

func TestNewCmdGet_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := branchmodel.NewCmdGet(f)
	assert.NotNil(t, cmd.Flag("hostname"))
	assert.NotNil(t, cmd.Flag("json"))
}

func TestGet_PrintsBranchModel(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetBranchModelFn: func(ws, slug string) (backend.BranchModel, error) {
			assert.Equal(t, "myworkspace", ws)
			assert.Equal(t, "my-service", slug)
			prod := backend.BranchModelBranch{Name: "production", IsValid: true}
			return backend.BranchModel{
				Development: backend.BranchModelBranch{Name: "main", UseMainbranch: true},
				Production:  &prod,
				BranchTypes: []backend.BranchType{
					{Kind: "feature", Prefix: "feature/"},
					{Kind: "hotfix", Prefix: "hotfix/"},
				},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := branchmodel.NewCmdGet(f)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "main")
	assert.Contains(t, got, "production")
	assert.Contains(t, got, "feature/")
	assert.Contains(t, got, "hotfix/")
}

func TestGet_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetBranchModelFn: func(ws, slug string) (backend.BranchModel, error) {
			return backend.BranchModel{
				Development: backend.BranchModelBranch{Name: "develop"},
				BranchTypes: []backend.BranchType{{Kind: "bugfix", Prefix: "bugfix/"}},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := branchmodel.NewCmdGet(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "--json"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, `"name":"develop"`)
	assert.Contains(t, got, `"kind":"bugfix"`)
}

func TestGet_UnsupportedBackend(t *testing.T) {
	t.Parallel()
	fake := &cmdtest.NoBranchModelFake{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := branchmodel.NewCmdGet(f)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "branching model")
}
