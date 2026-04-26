package pr_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestPRRequestReview_RequiresReviewer(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRRequestReview(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--reviewer")
}

func TestPRRequestReview_PrintsConfirmation(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		RequestReviewFn: func(ns, slug string, id int, users []string) error {
			return nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRRequestReview(f)
	cmd.SetArgs([]string{"42", "--reviewer", "alice"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Requested review on pull request #42")
}

func TestPRRequestReview_PassesUsersToAPI(t *testing.T) {
	t.Parallel()
	var gotUsers []string
	fake := &testhelpers.FakeClient{
		T: t,
		RequestReviewFn: func(ns, slug string, id int, users []string) error {
			gotUsers = users
			return nil
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRRequestReview(f)
	cmd.SetArgs([]string{"42", "--reviewer", "alice,bob"})
	require.NoError(t, cmd.Execute())
	require.Len(t, gotUsers, 2)
	assert.Equal(t, "alice", gotUsers[0])
	assert.Equal(t, "bob", gotUsers[1])
}

func TestPRRequestReview_APIError_PropagatesError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		RequestReviewFn: func(ns, slug string, id int, users []string) error {
			return errors.New("404 not found")
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRRequestReview(f)
	cmd.SetArgs([]string{"42", "--reviewer", "alice"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}
