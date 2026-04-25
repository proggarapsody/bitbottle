package branch_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/branch"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const branchConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

func TestNewCmdBranchList_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := branch.NewCmdBranchList(f)
	assert.NotNil(t, cmd.Flag("limit"))
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("jq"))
}

func TestNewCmdBranchList_LimitDefault(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := branch.NewCmdBranchList(f)
	assert.Equal(t, "30", cmd.Flag("limit").DefValue)
}

func TestNewCmdBranchList_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := branch.NewCmdBranchList(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestBranchList_PrintsNames(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListBranchesFn: func(ns, slug string, limit int) ([]backend.Branch, error) {
			return []backend.Branch{
				{Name: "main", IsDefault: true, LatestHash: "abc1234def"},
				{Name: "feature/login", IsDefault: false, LatestHash: "deadbeef12"},
			}, nil
		},
	}

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   branchConfig,
		BackendOverride: fake,
	})
	cmd := branch.NewCmdBranchList(f)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "main")
	assert.Contains(t, got, "feature/login")
}

func TestBranchList_TruncatesHash(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListBranchesFn: func(ns, slug string, limit int) ([]backend.Branch, error) {
			return []backend.Branch{
				{Name: "main", LatestHash: "abc1234def567890"},
			}, nil
		},
	}

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   branchConfig,
		BackendOverride: fake,
	})
	cmd := branch.NewCmdBranchList(f)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "abc1234d")
	assert.NotContains(t, got, "abc1234def567890")
}

func TestBranchList_JSON_FieldsOutput(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListBranchesFn: func(ns, slug string, limit int) ([]backend.Branch, error) {
			return []backend.Branch{
				{Name: "main", IsDefault: true, LatestHash: "abc1234"},
			}, nil
		},
	}

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   branchConfig,
		BackendOverride: fake,
	})
	cmd := branch.NewCmdBranchList(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "--json", "name,default"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, `"name":"main"`)
	assert.Contains(t, got, `"default":true`)
}

func TestBranchList_JQ_FilterOutput(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListBranchesFn: func(ns, slug string, limit int) ([]backend.Branch, error) {
			return []backend.Branch{
				{Name: "main"},
				{Name: "develop"},
			}, nil
		},
	}

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   branchConfig,
		BackendOverride: fake,
	})
	cmd := branch.NewCmdBranchList(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "--json", "name", "--jq", ".[] | .name"})
	require.NoError(t, cmd.Execute())

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	// jq output uses json.Marshal, so strings are quoted
	assert.Equal(t, []string{`"main"`, `"develop"`}, lines)
}
