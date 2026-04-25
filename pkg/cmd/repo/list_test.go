package repo_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdRepoList_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := repo.NewCmdRepoList(f)
	assert.NotNil(t, cmd.Flag("limit"))
	assert.NotNil(t, cmd.Flag("json"))
}

func TestNewCmdRepoList_LimitDefault(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := repo.NewCmdRepoList(f)
	assert.Equal(t, "30", cmd.Flag("limit").DefValue)
}

func TestNewCmdRepoList_NoConfigReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := repo.NewCmdRepoList(f)
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not authenticated")
}

func TestNewCmdRepoList_HasHostnameFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := repo.NewCmdRepoList(f)
	assert.NotNil(t, cmd.Flag("hostname"))
}

func TestNewCmdRepoList_HasJQFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := repo.NewCmdRepoList(f)
	assert.NotNil(t, cmd.Flag("jq"))
}

func TestRepoList_JSON_FieldsOutput(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListReposFn: func(limit int) ([]backend.Repository, error) {
			return []backend.Repository{
				testhelpers.BackendRepoFactory(testhelpers.BackendRepoWithSlug("svc-a")),
				testhelpers.BackendRepoFactory(testhelpers.BackendRepoWithSlug("svc-b")),
			}, nil
		},
	}

	f, out, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoList(f)
	cmd.SetArgs([]string{"--json", "slug,name"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, `"slug":"svc-a"`)
	assert.Contains(t, got, `"name":"svc-b"`)
	assert.NotContains(t, got, `"namespace"`)
}

func TestRepoList_JQ_FilterOutput(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListReposFn: func(limit int) ([]backend.Repository, error) {
			return []backend.Repository{
				testhelpers.BackendRepoFactory(testhelpers.BackendRepoWithSlug("alpha")),
				testhelpers.BackendRepoFactory(testhelpers.BackendRepoWithSlug("beta")),
			}, nil
		},
	}

	f, out, _ := newRepoFactory(t, fake)
	cmd := repo.NewCmdRepoList(f)
	cmd.SetArgs([]string{"--json", "slug", "--jq", ".[] | .slug"})
	require.NoError(t, cmd.Execute())

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	assert.Equal(t, []string{`"alpha"`, `"beta"`}, lines)
}
