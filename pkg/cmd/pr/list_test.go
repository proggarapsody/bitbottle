package pr_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdPRList_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := pr.NewCmdPRList(f)
	format.RegisterOutputFlags(cmd)
	assert.NotNil(t, cmd.Flag("state"))
	assert.NotNil(t, cmd.Flag("limit"))
	assert.NotNil(t, cmd.Flag("json"))
}

func TestNewCmdPRList_StateDefault(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := pr.NewCmdPRList(f)
	format.RegisterOutputFlags(cmd)
	assert.Equal(t, "open", cmd.Flag("state").DefValue)
}

func TestNewCmdPRList_LimitDefault(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := pr.NewCmdPRList(f)
	format.RegisterOutputFlags(cmd)
	assert.Equal(t, "30", cmd.Flag("limit").DefValue)
}

func TestNewCmdPRList_NoRemoteReturnsError(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := pr.NewCmdPRList(f)
	format.RegisterOutputFlags(cmd)
	err := cmd.Execute()
	require.Error(t, err)
	// No git remote and no PROJECT/REPO arg — must error.
	assert.NotNil(t, err)
}

func TestNewCmdPRList_AcceptsMaxOneArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := pr.NewCmdPRList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"PROJ/repo", "extra"})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestNewCmdPRList_HasJQFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := pr.NewCmdPRList(f)
	format.RegisterOutputFlags(cmd)
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
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--json"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, `"id":1`)
	assert.Contains(t, got, `"title":"Fix auth"`)
	// OUT2 ships all fields; field selection is deferred.
	assert.Contains(t, got, `"state"`)
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
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--json", "--jq", ".[] | .id"})
	require.NoError(t, cmd.Execute())

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	assert.Equal(t, []string{"10", "20"}, lines)
}

func TestNewCmdPRList_InvalidLimit(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := pr.NewCmdPRList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--limit", "0"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--limit")
}

func TestNewCmdPRList_RejectsInvalidState(t *testing.T) {
	t.Parallel()
	for _, bad := range []string{"INVALID", "", "opn"} {
		f, _, _ := factorytest.New(t, factorytest.Opts{})
		cmd := pr.NewCmdPRList(f)
		format.RegisterOutputFlags(cmd)
		cmd.SetArgs([]string{"MYPROJ/my-service", "--state", bad})
		err := cmd.Execute()
		require.Error(t, err, "state %q should be rejected", bad)
		assert.Contains(t, err.Error(), "invalid value")
	}
}

func TestPRList_ValidStatesMapCorrectly(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"open":     "OPEN",
		"opened":   "OPEN",
		"closed":   "DECLINED",
		"declined": "DECLINED",
		"merged":   "MERGED",
		"MERGED":   "MERGED", // case-insensitive accept, still maps right
	}
	for in, want := range cases {
		var gotState string
		fake := &testhelpers.FakeClient{
			T: t,
			ListPRsFn: func(ns, slug, state string, limit int) ([]backend.PullRequest, error) {
				gotState = state
				return nil, nil
			},
		}
		f, _, _ := newPRFactory(t, fake, newPRRunner())
		cmd := pr.NewCmdPRList(f)
		format.RegisterOutputFlags(cmd)
		cmd.SetArgs([]string{"MYPROJ/my-service", "--state", in})
		require.NoError(t, cmd.Execute(), "state %q should be accepted", in)
		assert.Equal(t, want, gotState, "state %q mapped wrong", in)
	}
}

func TestPRList_PartialResults(t *testing.T) {
	t.Parallel()
	listErr := errors.New("429 Too Many Requests")
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRsFn: func(ns, slug, state string, limit int) ([]backend.PullRequest, error) {
			return []backend.PullRequest{
				testhelpers.BackendPRFactory(testhelpers.BackendPRWithID(1), testhelpers.BackendPRWithTitle("partial PR")),
			}, listErr
		},
	}
	f, out, errOut := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, out.String(), "partial PR")
	assert.Contains(t, errOut.String(), "warning: partial results")
}
