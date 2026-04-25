package pr_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdPRList_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRList(f)
	assert.NotNil(t, cmd.Flag("state"))
	assert.NotNil(t, cmd.Flag("limit"))
	assert.NotNil(t, cmd.Flag("json"))
}

func TestNewCmdPRList_StateDefault(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRList(f)
	assert.Equal(t, "open", cmd.Flag("state").DefValue)
}

func TestNewCmdPRList_LimitDefault(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRList(f)
	assert.Equal(t, "30", cmd.Flag("limit").DefValue)
}

func TestNewCmdPRList_NoRemoteReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRList(f)
	err := cmd.Execute()
	require.Error(t, err)
	// No git remote and no PROJECT/REPO arg — must error.
	assert.NotNil(t, err)
}

func TestNewCmdPRList_AcceptsMaxOneArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRList(f)
	cmd.SetArgs([]string{"PROJ/repo", "extra"})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestNewCmdPRList_HasJQFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRList(f)
	assert.NotNil(t, cmd.Flag("jq"))
}

func TestPRList_JSON_FieldsOutput(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListPRsFn: func(ns, slug, state string, limit int) ([]backend.PullRequest, error) {
			return []backend.PullRequest{
				testhelpers.BackendPRFactory(testhelpers.BackendPRWithID(1), testhelpers.BackendPRWithTitle("Fix auth")),
				testhelpers.BackendPRFactory(testhelpers.BackendPRWithID(2), testhelpers.BackendPRWithTitle("Bump deps")),
			}, nil
		},
	}

	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRList(f)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--json", "id,title"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, `"id":1`)
	assert.Contains(t, got, `"title":"Fix auth"`)
	assert.NotContains(t, got, `"state"`)
}

func TestPRList_JQ_FilterOutput(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListPRsFn: func(ns, slug, state string, limit int) ([]backend.PullRequest, error) {
			return []backend.PullRequest{
				testhelpers.BackendPRFactory(testhelpers.BackendPRWithID(10)),
				testhelpers.BackendPRFactory(testhelpers.BackendPRWithID(20)),
			}, nil
		},
	}

	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRList(f)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--json", "id", "--jq", ".[] | .id"})
	require.NoError(t, cmd.Execute())

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	assert.Equal(t, []string{"10", "20"}, lines)
}
