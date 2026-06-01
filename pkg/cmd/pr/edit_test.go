package pr_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func guardOpenPRFn(_, _ string, id int) (backend.PullRequest, error) {
	return backend.PullRequest{ID: id, State: "OPEN"}, nil
}

func TestPREdit_RequiresTitleOrBodyOrRemoveReviewer(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPREdit(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--title")
}

func TestPREdit_PrintsConfirmation(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T:       t,
		GetPRFn: guardOpenPRFn,
		UpdatePRFn: func(ns, slug string, id int, in backend.UpdatePRInput) (backend.PullRequest, error) {
			return backend.PullRequest{ID: id, Title: in.Title, WebURL: "https://bb.example.com/pr/42"}, nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPREdit(f)
	cmd.SetArgs([]string{"42", "--title", "New title"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Updated pull request #42")
	assert.Contains(t, out.String(), "https://bb.example.com/pr/42")
}

func TestPREdit_PassesTitleToAPI(t *testing.T) {
	t.Parallel()
	var gotIn backend.UpdatePRInput
	fake := &testhelpers.FakeClient{
		T:       t,
		GetPRFn: guardOpenPRFn,
		UpdatePRFn: func(ns, slug string, id int, in backend.UpdatePRInput) (backend.PullRequest, error) {
			gotIn = in
			return backend.PullRequest{ID: id, Title: in.Title}, nil
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPREdit(f)
	cmd.SetArgs([]string{"42", "--title", "My new title"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "My new title", gotIn.Title)
}

func TestPREdit_PassesBodyToAPI(t *testing.T) {
	t.Parallel()
	var gotIn backend.UpdatePRInput
	fake := &testhelpers.FakeClient{
		T:       t,
		GetPRFn: guardOpenPRFn,
		UpdatePRFn: func(ns, slug string, id int, in backend.UpdatePRInput) (backend.PullRequest, error) {
			gotIn = in
			return backend.PullRequest{ID: id}, nil
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPREdit(f)
	cmd.SetArgs([]string{"42", "--body", "Updated description"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "Updated description", gotIn.Description)
}

func TestPREdit_RemoveReviewer_CallsRemoveReviewers(t *testing.T) {
	t.Parallel()
	var gotUsers []string
	fake := &testhelpers.FakeClient{
		T:       t,
		GetPRFn: guardOpenPRFn,
		RemoveReviewersFn: func(ns, slug string, id int, users []string) error {
			gotUsers = users
			return nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPREdit(f)
	cmd.SetArgs([]string{"42", "--remove-reviewer", "alice,bob"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, []string{"alice", "bob"}, gotUsers)
	assert.Contains(t, out.String(), "Removed reviewer(s) from pull request #42")
}

func TestPREdit_RemoveReviewerAndTitle_DoesBoth(t *testing.T) {
	t.Parallel()
	var updatedTitle string
	var removedUsers []string
	fake := &testhelpers.FakeClient{
		T:       t,
		GetPRFn: guardOpenPRFn,
		UpdatePRFn: func(ns, slug string, id int, in backend.UpdatePRInput) (backend.PullRequest, error) {
			updatedTitle = in.Title
			return backend.PullRequest{ID: id, Title: in.Title}, nil
		},
		RemoveReviewersFn: func(ns, slug string, id int, users []string) error {
			removedUsers = users
			return nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPREdit(f)
	cmd.SetArgs([]string{"42", "--title", "New title", "--remove-reviewer", "alice"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "New title", updatedTitle)
	assert.Equal(t, []string{"alice"}, removedUsers)
	assert.Contains(t, out.String(), "Updated pull request #42")
	assert.Contains(t, out.String(), "Removed reviewer(s) from pull request #42")
}

func TestPREdit_APIError_PropagatesError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T:       t,
		GetPRFn: guardOpenPRFn,
		UpdatePRFn: func(ns, slug string, id int, in backend.UpdatePRInput) (backend.PullRequest, error) {
			return backend.PullRequest{}, errors.New("403 forbidden")
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPREdit(f)
	cmd.SetArgs([]string{"42", "--title", "title"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}

func TestPREdit_DeclinedPR_GuardBlocksMutation(t *testing.T) {
	t.Parallel()
	updated := false
	fake := &testhelpers.FakeClient{
		T: t,
		GetPRFn: func(_, _ string, id int) (backend.PullRequest, error) {
			return backend.PullRequest{ID: id, State: "SUPERSEDED"}, nil
		},
		UpdatePRFn: func(ns, slug string, id int, in backend.UpdatePRInput) (backend.PullRequest, error) {
			updated = true
			return backend.PullRequest{ID: id}, nil
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPREdit(f)
	cmd.SetArgs([]string{"42", "--title", "title"})
	err := cmd.Execute()
	require.Error(t, err)
	require.True(t, errors.Is(err, backend.ErrInvalidRequest), "want ErrInvalidRequest, got %v", err)
	assert.False(t, updated, "UpdatePR must not be called for a SUPERSEDED PR")
}
