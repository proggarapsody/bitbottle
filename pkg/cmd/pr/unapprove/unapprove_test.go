package unapprove_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr/unapprove"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func openPRFn(id int) func(ns, slug string, id int) (backend.PullRequest, error) {
	return func(_, _ string, _ int) (backend.PullRequest, error) {
		return backend.PullRequest{ID: id, State: "OPEN"}, nil
	}
}

func TestPRUnapprove_PrintsConfirmation(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T:       t,
		GetPRFn: openPRFn(42),
		UnapprovePRFn: func(ns, slug string, id int) error {
			return nil
		},
	}
	f, out, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := unapprove.NewCmdUnapprove(f)
	cmd.SetArgs([]string{"42"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Removed approval from pull request #42")
}

func TestPRUnapprove_APIError_PropagatesError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T:       t,
		GetPRFn: openPRFn(42),
		UnapprovePRFn: func(ns, slug string, id int) error {
			return errors.New("403 forbidden")
		},
	}
	f, _, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := unapprove.NewCmdUnapprove(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}

func TestPRUnapprove_DeclinedPR_GuardBlocksMutation(t *testing.T) {
	t.Parallel()
	called := false
	fake := &testhelpers.FakeClient{
		T: t,
		GetPRFn: func(_, _ string, _ int) (backend.PullRequest, error) {
			return backend.PullRequest{ID: 42, State: "MERGED"}, nil
		},
		UnapprovePRFn: func(_, _ string, _ int) error {
			called = true
			return nil
		},
	}
	f, _, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := unapprove.NewCmdUnapprove(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	require.True(t, errors.Is(err, backend.ErrConflict), "want ErrConflict, got %v", err)
	assert.False(t, called, "UnapprovePR must not be called for a MERGED PR")
}
