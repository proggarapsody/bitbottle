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

func TestPREdit_RequiresTitleOrBody(t *testing.T) {
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
		T: t,
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
		T: t,
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
		T: t,
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

func TestPREdit_APIError_PropagatesError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
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
