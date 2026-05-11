package pr_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestPRReview_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := pr.NewCmdPRReview(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestPRReview_RequiresActionOrBodyOrInline(t *testing.T) {
	t.Parallel()
	// SubmitReviewFn is nil — the validation must reject before any call.
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRReview(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--approve")
	assert.Contains(t, err.Error(), "required")
}

func TestPRReview_ApproveFlag_SendsApproveAction(t *testing.T) {
	t.Parallel()
	var got backend.SubmitReviewInput
	fake := &testhelpers.FakeClient{
		T: t,
		SubmitReviewFn: func(ns, slug string, id int, in backend.SubmitReviewInput) error {
			got = in
			return nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRReview(f)
	cmd.SetArgs([]string{"42", "--approve", "--body", "lgtm"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "approve", got.Action)
	assert.Equal(t, "lgtm", got.Body)
	assert.Empty(t, got.Inline)
	assert.Contains(t, out.String(), "Submitted review on pull request #42")
}

func TestPRReview_RequestChangesFlag_SendsRequestChangesAction(t *testing.T) {
	t.Parallel()
	var got backend.SubmitReviewInput
	fake := &testhelpers.FakeClient{
		T: t,
		SubmitReviewFn: func(ns, slug string, id int, in backend.SubmitReviewInput) error {
			got = in
			return nil
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRReview(f)
	cmd.SetArgs([]string{"42", "--request-changes", "--body", "nope"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "request_changes", got.Action)
}

func TestPRReview_CommentFlag_SendsCommentAction(t *testing.T) {
	t.Parallel()
	var got backend.SubmitReviewInput
	fake := &testhelpers.FakeClient{
		T: t,
		SubmitReviewFn: func(ns, slug string, id int, in backend.SubmitReviewInput) error {
			got = in
			return nil
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRReview(f)
	cmd.SetArgs([]string{"42", "--comment", "--body", "fyi"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "comment", got.Action)
}

func TestPRReview_BodyWithoutActionDefaultsToComment(t *testing.T) {
	t.Parallel()
	var got backend.SubmitReviewInput
	fake := &testhelpers.FakeClient{
		T: t,
		SubmitReviewFn: func(ns, slug string, id int, in backend.SubmitReviewInput) error {
			got = in
			return nil
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRReview(f)
	cmd.SetArgs([]string{"42", "--body", "hey"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "comment", got.Action, "body without action should default to comment-only review")
}

func TestPRReview_InlineFlag_ParsesIntoInput(t *testing.T) {
	t.Parallel()
	var got backend.SubmitReviewInput
	fake := &testhelpers.FakeClient{
		T: t,
		SubmitReviewFn: func(ns, slug string, id int, in backend.SubmitReviewInput) error {
			got = in
			return nil
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRReview(f)
	cmd.SetArgs([]string{
		"42",
		"--approve",
		"--inline", "pkg/a.go:12:rename",
		"--inline", "pkg/b.go:5-7:extract helper",
	})
	require.NoError(t, cmd.Execute())
	require.Len(t, got.Inline, 2)
	assert.Equal(t, "pkg/a.go", got.Inline[0].Path)
	assert.Equal(t, 12, got.Inline[0].Line)
	assert.Equal(t, "rename", got.Inline[0].Body)
	assert.Equal(t, "pkg/b.go", got.Inline[1].Path)
	assert.Equal(t, 5, got.Inline[1].StartLine)
	assert.Equal(t, 7, got.Inline[1].Line)
	assert.Equal(t, "extract helper", got.Inline[1].Body)
}

func TestPRReview_MutuallyExclusiveActionFlags(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRReview(f)
	cmd.SetArgs([]string{"42", "--approve", "--request-changes"})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestPRReview_APIError_Propagates(t *testing.T) {
	t.Parallel()
	apiErr := errors.New("403 forbidden")
	fake := &testhelpers.FakeClient{
		T: t,
		SubmitReviewFn: func(ns, slug string, id int, in backend.SubmitReviewInput) error {
			return apiErr
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRReview(f)
	cmd.SetArgs([]string{"42", "--approve"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}

func TestPRReview_BadInlineSpec_Errors(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRReview(f)
	cmd.SetArgs([]string{"42", "--approve", "--inline", "missing-body"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "PATH:LINE:BODY")
}
