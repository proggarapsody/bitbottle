package list_test

import (
	"strings"
	"testing"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo/list"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdList_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := list.NewCmdList(f)
	format.RegisterOutputFlags(cmd)
	assert.NotNil(t, cmd.Flag("limit"))
	assert.NotNil(t, cmd.Flag("json"))
}

func TestNewCmdList_LimitDefault(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := list.NewCmdList(f)
	format.RegisterOutputFlags(cmd)
	assert.Equal(t, "30", cmd.Flag("limit").DefValue)
}

func TestNewCmdList_NoConfigReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := list.NewCmdList(f)
	format.RegisterOutputFlags(cmd)
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not authenticated")
}

func TestNewCmdList_HasHostnameFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := list.NewCmdList(f)
	format.RegisterOutputFlags(cmd)
	assert.NotNil(t, cmd.Flag("hostname"))
}

func TestNewCmdList_HasJQFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := list.NewCmdList(f)
	format.RegisterOutputFlags(cmd)
	assert.NotNil(t, cmd.Flag("jq"))
}

func TestRepoList_JSON_FieldsOutput(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListReposFn: func(ns string, limit int) ([]backend.Repository, error) {
			return []backend.Repository{
				testhelpers.BackendRepoFactory(testhelpers.BackendRepoWithSlug("svc-a")),
				testhelpers.BackendRepoFactory(testhelpers.BackendRepoWithSlug("svc-b")),
			}, nil
		},
	}

	f, out, _ := newRepoFactory(t, fake)
	cmd := list.NewCmdList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"--json"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, `"slug":"svc-a"`)
	assert.Contains(t, got, `"name":"svc-b"`)
	// OUT2 ships all fields; field selection is deferred.
	assert.Contains(t, got, `"namespace"`)
}

func TestRepoList_JQ_FilterOutput(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListReposFn: func(ns string, limit int) ([]backend.Repository, error) {
			return []backend.Repository{
				testhelpers.BackendRepoFactory(testhelpers.BackendRepoWithSlug("alpha")),
				testhelpers.BackendRepoFactory(testhelpers.BackendRepoWithSlug("beta")),
			}, nil
		},
	}

	f, out, _ := newRepoFactory(t, fake)
	cmd := list.NewCmdList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"--json", "--jq", ".[] | .slug"})
	require.NoError(t, cmd.Execute())

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	assert.Equal(t, []string{`"alpha"`, `"beta"`}, lines)
}

func TestNewCmdList_InvalidLimit(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: repoConfig})
	cmd := list.NewCmdList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"--limit", "0"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--limit")
}
