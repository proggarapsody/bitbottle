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

func TestPRCommentList_InlineFlagFiltersToInlineOnly(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRCommentsFn: func(ns, slug string, id int) ([]backend.PRComment, error) {
			return []backend.PRComment{
				{ID: 1, Author: backend.User{Slug: "alice"}, Text: "general", CreatedAt: now},
				{ID: 2, Author: backend.User{Slug: "bob"}, Text: "inline nit", CreatedAt: now,
					Inline: &backend.PRCommentInline{Path: "main.go", Side: "new", Line: 42}},
			}, nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentList(f)
	cmd.SetArgs([]string{"42", "--inline"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.NotContains(t, got, "general", "general comment should be filtered out")
	assert.Contains(t, got, "inline nit")
	assert.Contains(t, got, "main.go:42")
}

func TestPRCommentList_LocationColumnShownWhenInlinePresent(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRCommentsFn: func(ns, slug string, id int) ([]backend.PRComment, error) {
			return []backend.PRComment{
				{ID: 1, Author: backend.User{Slug: "alice"}, Text: "ok", CreatedAt: now},
				{ID: 2, Author: backend.User{Slug: "bob"}, Text: "fix", CreatedAt: now,
					Inline: &backend.PRCommentInline{Path: "src/foo.go", Side: "new", Line: 7, StartLine: 3}},
			}, nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentList(f)
	cmd.SetArgs([]string{"42"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "src/foo.go:3-7", "multi-line inline rendered as path:start-end")
	assert.Contains(t, got, "alice")
	assert.Contains(t, got, "bob")
}

func TestPRCommentList_JSONExposesInlineAndThreadFields(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	updated := now.Add(2 * time.Hour)
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRCommentsFn: func(ns, slug string, id int) ([]backend.PRComment, error) {
			return []backend.PRComment{
				{ID: 9, Author: backend.User{Slug: "bob"}, Text: "reply", CreatedAt: now, UpdatedAt: updated,
					ParentID: 7, Resolved: true,
					Inline: &backend.PRCommentInline{Path: "main.go", Side: "old", Line: 10}},
			}, nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentList(f)
	cmd.SetArgs([]string{"42", "--json", "id,parentId,resolved,updatedAt,inline"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, `"parentId":7`)
	assert.Contains(t, got, `"resolved":true`)
	assert.Contains(t, got, `"updatedAt":"`+updated.Format(time.RFC3339)+`"`)
	assert.Contains(t, got, `"path":"main.go"`)
	assert.Contains(t, got, `"side":"old"`)
	assert.Contains(t, got, `"line":10`)
}

func TestPRCommentList_ExistingJSONFieldsUnchanged(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRCommentsFn: func(ns, slug string, id int) ([]backend.PRComment, error) {
			return []backend.PRComment{
				{ID: 1, Author: backend.User{Slug: "alice"}, Text: "hi", CreatedAt: now},
			}, nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentList(f)
	cmd.SetArgs([]string{"42", "--json", "id,author,text,createdAt"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, `"id":1`)
	assert.Contains(t, got, `"author":"alice"`)
	assert.Contains(t, got, `"text":"hi"`)
	assert.NotContains(t, got, "parentId", "additive change must not leak unrequested fields")
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
