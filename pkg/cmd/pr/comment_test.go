package pr_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
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
	format.RegisterOutputFlags(cmd)
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
	format.RegisterOutputFlags(cmd)
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
	format.RegisterOutputFlags(cmd)
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
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "--json"})
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
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "--json"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, `"id":1`)
	assert.Contains(t, got, `"author":"alice"`)
	assert.Contains(t, got, `"text":"hi"`)
	// OUT2 ships all fields uniformly; field selection is deferred.
	assert.Contains(t, got, "parentId")
}

func TestPRCommentAdd_RequiresBody(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentAdd(f)
	format.RegisterOutputFlags(cmd)
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
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "--body", "Looks good"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "Looks good", gotIn.Text)
	assert.Contains(t, out.String(), "Added comment #7")
}

func TestPRCommentAdd_InlineFlagBuildsAnchor(t *testing.T) {
	t.Parallel()
	var gotIn backend.AddPRCommentInput
	fake := &testhelpers.FakeClient{
		T: t,
		AddPRCommentFn: func(ns, slug string, id int, in backend.AddPRCommentInput) (backend.PRComment, error) {
			gotIn = in
			return backend.PRComment{ID: 7}, nil
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentAdd(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "--body", "nit", "--inline", "main.go:88"})
	require.NoError(t, cmd.Execute())

	require.NotNil(t, gotIn.Inline, "Inline must be populated when --inline is set")
	assert.Equal(t, "main.go", gotIn.Inline.Path)
	assert.Equal(t, "new", gotIn.Inline.Side)
	assert.Equal(t, 88, gotIn.Inline.Line)
}

func TestPRCommentAdd_InlineSideOldFlag(t *testing.T) {
	t.Parallel()
	var gotIn backend.AddPRCommentInput
	fake := &testhelpers.FakeClient{
		T: t,
		AddPRCommentFn: func(ns, slug string, id int, in backend.AddPRCommentInput) (backend.PRComment, error) {
			gotIn = in
			return backend.PRComment{ID: 8}, nil
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentAdd(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "--body", "x", "--inline", "main.go:5", "--side", "old"})
	require.NoError(t, cmd.Execute())

	require.NotNil(t, gotIn.Inline)
	assert.Equal(t, "old", gotIn.Inline.Side)
}

func TestPRCommentAdd_ParentFlagBuildsReply(t *testing.T) {
	t.Parallel()
	var gotIn backend.AddPRCommentInput
	fake := &testhelpers.FakeClient{
		T: t,
		AddPRCommentFn: func(ns, slug string, id int, in backend.AddPRCommentInput) (backend.PRComment, error) {
			gotIn = in
			return backend.PRComment{ID: 9}, nil
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentAdd(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "--body", "agreed", "--parent", "7"})
	require.NoError(t, cmd.Execute())

	require.NotNil(t, gotIn.Parent)
	assert.Equal(t, 7, *gotIn.Parent)
	assert.Nil(t, gotIn.Inline, "reply without --inline must not set the anchor")
}

func TestPRCommentAdd_BadInlineSpecRejectedBeforeAPICall(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t} // unset AddPRCommentFn → fatal if reached
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentAdd(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "--body", "x", "--inline", "no-colon"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--inline")
}

func TestPRCommentEdit_RequiresBody(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentEdit(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "99"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--body")
}

func TestPRCommentEdit_PassesBodyToAPI(t *testing.T) {
	t.Parallel()
	var gotID, gotComment int
	var gotBody string
	fake := &testhelpers.FakeClient{
		T: t,
		EditPRCommentFn: func(ns, slug string, id, commentID int, body string) (backend.PRComment, error) {
			gotID, gotComment, gotBody = id, commentID, body
			return backend.PRComment{ID: commentID, Text: body}, nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentEdit(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "99", "--body", "updated"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, 42, gotID)
	assert.Equal(t, 99, gotComment)
	assert.Equal(t, "updated", gotBody)
	assert.Contains(t, out.String(), "Updated comment #99")
}

func TestPRCommentDelete_CallsAPI(t *testing.T) {
	t.Parallel()
	var gotID, gotComment int
	fake := &testhelpers.FakeClient{
		T: t,
		DeletePRCommentFn: func(ns, slug string, id, commentID int) error {
			gotID, gotComment = id, commentID
			return nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentDelete(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "99"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, 42, gotID)
	assert.Equal(t, 99, gotComment)
	assert.Contains(t, out.String(), "Deleted comment #99")
}

// fakeResolverClient embeds FakeClient + implements PRCommentResolver.
type fakeResolverClient struct {
	*testhelpers.FakeClient
	ResolvePRCommentFn func(ns, slug string, id, commentID int) error
}

func (f *fakeResolverClient) ResolvePRComment(ns, slug string, id, commentID int) error {
	if f.ResolvePRCommentFn != nil {
		return f.ResolvePRCommentFn(ns, slug, id, commentID)
	}
	if f.T != nil {
		f.T.Fatalf("unexpected call to fakeResolverClient.ResolvePRComment")
	}
	return nil
}

var _ backend.PRCommentResolver = (*fakeResolverClient)(nil)

func TestPRCommentResolve_CallsAPIOnCloud(t *testing.T) {
	t.Parallel()
	var gotID, gotComment int
	fake := &fakeResolverClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		ResolvePRCommentFn: func(ns, slug string, id, commentID int) error {
			gotID, gotComment = id, commentID
			return nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentResolve(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "99"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, 42, gotID)
	assert.Equal(t, 99, gotComment)
	assert.Contains(t, out.String(), "Resolved comment #99")
}

func TestPRCommentResolve_UnsupportedOnServer(t *testing.T) {
	t.Parallel()
	// Plain FakeClient does NOT implement PRCommentResolver.
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentResolve(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "99"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
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
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "--body", "hi"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}

// ── fakeReactorClient embeds FakeClient + CommentReactor ────────────────────

type fakeReactorClient struct {
	*testhelpers.FakeClient
}

var _ backend.CommentReactor = (*fakeReactorClient)(nil)

func (c *fakeReactorClient) ListCommentReactions(ns, slug string, prID, commentID int) ([]backend.CommentReaction, error) {
	if c.ListCommentReactionsFn != nil {
		return c.ListCommentReactionsFn(ns, slug, prID, commentID)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to fakeReactorClient.ListCommentReactions; set ListCommentReactionsFn in your test")
	}
	return nil, nil
}

func (c *fakeReactorClient) AddCommentReaction(ns, slug string, prID, commentID int, emoji string) error {
	if c.AddCommentReactionFn != nil {
		return c.AddCommentReactionFn(ns, slug, prID, commentID, emoji)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to fakeReactorClient.AddCommentReaction; set AddCommentReactionFn in your test")
	}
	return nil
}

func (c *fakeReactorClient) RemoveCommentReaction(ns, slug string, prID, commentID int, emoji string) error {
	if c.RemoveCommentReactionFn != nil {
		return c.RemoveCommentReactionFn(ns, slug, prID, commentID, emoji)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to fakeReactorClient.RemoveCommentReaction; set RemoveCommentReactionFn in your test")
	}
	return nil
}

// ── pr comment react ─────────────────────────────────────────────────────────

func TestPRCommentReact_NormalisesAndCallsAPI(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug, gotEmoji string
	var gotPRID, gotCommentID int
	fake := &fakeReactorClient{
		FakeClient: &testhelpers.FakeClient{
			T: t,
			AddCommentReactionFn: func(ns, slug string, prID, commentID int, emoji string) error {
				gotNS, gotSlug, gotEmoji = ns, slug, emoji
				gotPRID, gotCommentID = prID, commentID
				return nil
			},
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentReact(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "99", "--emoji", ":thumbsup:"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "thumbs_up", gotEmoji, "emoji should be normalised to underscore form")
	assert.Equal(t, 42, gotPRID)
	assert.Equal(t, 99, gotCommentID)
	assert.NotEmpty(t, gotNS)
	assert.NotEmpty(t, gotSlug)
	assert.Empty(t, out.String(), "no output on success")
}

func TestPRCommentReact_RequiresEmoji(t *testing.T) {
	t.Parallel()
	fake := &fakeReactorClient{FakeClient: &testhelpers.FakeClient{T: t}}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentReact(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "99"})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestPRCommentReact_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	// Plain FakeClient does NOT implement CommentReactor.
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentReact(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "99", "--emoji", "heart"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

// ── pr comment unreact ───────────────────────────────────────────────────────

func TestPRCommentUnreact_NormalisesAndCallsAPI(t *testing.T) {
	t.Parallel()
	var gotEmoji string
	var gotPRID, gotCommentID int
	fake := &fakeReactorClient{
		FakeClient: &testhelpers.FakeClient{
			T: t,
			RemoveCommentReactionFn: func(ns, slug string, prID, commentID int, emoji string) error {
				gotEmoji = emoji
				gotPRID, gotCommentID = prID, commentID
				return nil
			},
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentUnreact(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "99", "--emoji", "thumbs_up"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "thumbs_up", gotEmoji)
	assert.Equal(t, 42, gotPRID)
	assert.Equal(t, 99, gotCommentID)
	assert.Empty(t, out.String(), "no output on success")
}

func TestPRCommentUnreact_RequiresEmoji(t *testing.T) {
	t.Parallel()
	fake := &fakeReactorClient{FakeClient: &testhelpers.FakeClient{T: t}}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentUnreact(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "99"})
	err := cmd.Execute()
	require.Error(t, err)
}

// ── pr comment list --reactions ───────────────────────────────────────────────

func TestPRCommentList_ReactionsFlag_FetchesAndRendersReactions(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	fake := &fakeReactorClient{
		FakeClient: &testhelpers.FakeClient{
			T: t,
			ListPRCommentsFn: func(ns, slug string, id int) ([]backend.PRComment, error) {
				return []backend.PRComment{
					{ID: 1, Author: backend.User{Slug: "alice"}, Text: "LGTM", CreatedAt: now},
					{ID: 2, Author: backend.User{Slug: "bob"}, Text: "nice", CreatedAt: now},
				}, nil
			},
			ListCommentReactionsFn: func(ns, slug string, prID, commentID int) ([]backend.CommentReaction, error) {
				if commentID == 1 {
					return []backend.CommentReaction{
						{Emoji: "thumbs_up", Users: []backend.User{{Slug: "bob"}}},
					}, nil
				}
				return nil, nil
			},
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "--reactions"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	// In non-TTY mode the header row is suppressed, but the emoji glyph is always rendered.
	assert.Contains(t, got, "👍", "thumbs_up should render as emoji glyph")
}

func TestPRCommentList_ReactionsFlag_JSONIncludesReactions(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	fake := &fakeReactorClient{
		FakeClient: &testhelpers.FakeClient{
			T: t,
			ListPRCommentsFn: func(ns, slug string, id int) ([]backend.PRComment, error) {
				return []backend.PRComment{
					{ID: 1, Author: backend.User{Slug: "alice"}, Text: "LGTM", CreatedAt: now},
				}, nil
			},
			ListCommentReactionsFn: func(ns, slug string, prID, commentID int) ([]backend.CommentReaction, error) {
				return []backend.CommentReaction{
					{Emoji: "heart", Users: []backend.User{{Slug: "carol"}}},
				}, nil
			},
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "--reactions", "--json"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, `"reactions"`)
	assert.Contains(t, got, `"heart"`)
	assert.Contains(t, got, `"carol"`)
}

func TestPRCommentList_ReactionsFlag_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	// Plain FakeClient does NOT implement CommentReactor.
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRCommentsFn: func(ns, slug string, id int) ([]backend.PRComment, error) {
			return []backend.PRComment{
				{ID: 1, Author: backend.User{Slug: "alice"}, Text: "LGTM", CreatedAt: now},
			}, nil
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRCommentList(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"42", "--reactions"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}
