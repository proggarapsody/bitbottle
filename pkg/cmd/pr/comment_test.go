package pr_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestPRCommentList_RendersTable(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRCommentsFn: func(ns, slug string, id int) ([]backend.PRComment, error) {
			assert.Equal(t, 42, id)
			return []backend.PRComment{
				{ID: 1, Author: backend.User{Slug: "alice"}, Text: "LGTM", CreatedAt: now},
				{ID: 2, Author: backend.User{Slug: "bob"}, Text: "please add tests", CreatedAt: now.Add(time.Hour)},
			}, nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentList(f)
	cmd.SetArgs([]string{"42"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "alice")
	assert.Contains(t, got, "LGTM")
	assert.Contains(t, got, "bob")
	assert.Contains(t, got, "please add tests")
}

func TestPRCommentAdd_RequiresBody(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentAdd(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--body")
}

func TestPRCommentAdd_PassesBodyToAPI(t *testing.T) {
	t.Parallel()
	var gotIn backend.AddPRCommentInput
	fake := &testhelpers.FakeClient{
		T: t,
		AddPRCommentFn: func(ns, slug string, id int, in backend.AddPRCommentInput) (backend.PRComment, error) {
			gotIn = in
			return backend.PRComment{ID: 7}, nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentAdd(f)
	cmd.SetArgs([]string{"42", "--body", "Looks good"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "Looks good", gotIn.Text)
	assert.Contains(t, out.String(), "Added comment #7")
}

func TestPRCommentAdd_PropagatesAPIError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		AddPRCommentFn: func(ns, slug string, id int, in backend.AddPRCommentInput) (backend.PRComment, error) {
			return backend.PRComment{}, errors.New("403 forbidden")
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentAdd(f)
	cmd.SetArgs([]string{"42", "--body", "hi"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}
