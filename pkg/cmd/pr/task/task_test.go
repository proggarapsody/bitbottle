package task_test

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/run"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr/task"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const taskConfig = "bb.example.com:\n  oauth_token: tok\n  user: alice\n  git_protocol: ssh\n"

func newTaskRunner() *testhelpers.FakeRunner {
	return testhelpers.NewFakeRunner(testhelpers.RunResponse{
		Stdout: "ssh://git@bb.example.com:7999/myproj/my-service.git\n",
	})
}

func newTaskFactory(t *testing.T, fake backend.Client, runner *testhelpers.FakeRunner) (*factory.Factory, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	f, out, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: taskConfig})
	factorytest.UseBackend(f, fake)
	f.GitRunner = func() run.Runner { return runner }
	return f, out, errOut
}

// ── list tests ───────────────────────────────────────────────────────────────

func TestPRTaskList_ShowsBlockerComments(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRCommentsFn: func(ns, slug string, id int) ([]backend.PRComment, error) {
			assert.Equal(t, 42, id)
			return []backend.PRComment{
				{ID: 1, Author: backend.User{Slug: "alice"}, Text: "fix it", Severity: "BLOCKER", State: "OPEN", Version: 1, CreatedAt: now},
				{ID: 2, Author: backend.User{Slug: "bob"}, Text: "normal comment", Severity: "", State: "", CreatedAt: now},
			}, nil
		},
	}
	f, out, _ := newTaskFactory(t, fake, newTaskRunner())
	cmd := task.NewCmdPRTaskList(f)
	cmd.SetArgs([]string{"42"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "fix it")
	assert.NotContains(t, got, "normal comment", "non-BLOCKER comments should be excluded")
}

func TestPRTaskList_FilterByStateResolved(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRCommentsFn: func(ns, slug string, id int) ([]backend.PRComment, error) {
			return []backend.PRComment{
				{ID: 1, Text: "open task", Severity: "BLOCKER", State: "OPEN", CreatedAt: now},
				{ID: 2, Text: "done task", Severity: "BLOCKER", State: "RESOLVED", CreatedAt: now},
			}, nil
		},
	}
	f, out, _ := newTaskFactory(t, fake, newTaskRunner())
	cmd := task.NewCmdPRTaskList(f)
	cmd.SetArgs([]string{"42", "--state", "resolved"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "done task")
	assert.NotContains(t, got, "open task")
}

func TestPRTaskList_FilterByStateAll(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRCommentsFn: func(ns, slug string, id int) ([]backend.PRComment, error) {
			return []backend.PRComment{
				{ID: 1, Text: "open task", Severity: "BLOCKER", State: "OPEN", CreatedAt: now},
				{ID: 2, Text: "done task", Severity: "BLOCKER", State: "RESOLVED", CreatedAt: now},
			}, nil
		},
	}
	f, out, _ := newTaskFactory(t, fake, newTaskRunner())
	cmd := task.NewCmdPRTaskList(f)
	cmd.SetArgs([]string{"42", "--state", "all"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "open task")
	assert.Contains(t, got, "done task")
}

func TestPRTaskList_JSONIncludesSeverityAndVersion(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC)
	fake := &testhelpers.FakeClient{
		T: t,
		ListPRCommentsFn: func(ns, slug string, id int) ([]backend.PRComment, error) {
			return []backend.PRComment{
				{ID: 9, Author: backend.User{Slug: "alice"}, Text: "do this", Severity: "BLOCKER", State: "OPEN", Version: 3, CreatedAt: now},
			}, nil
		},
	}
	f, out, _ := newTaskFactory(t, fake, newTaskRunner())
	cmd := task.NewCmdPRTaskList(f)
	cmd.SetArgs([]string{"42", "--json"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, `"severity":"BLOCKER"`)
	assert.Contains(t, got, `"version":3`)
	assert.Contains(t, got, `"state":"OPEN"`)
}

// ── create tests ─────────────────────────────────────────────────────────────

func TestPRTaskCreate_RequiresBody(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newTaskFactory(t, fake, newTaskRunner())
	cmd := task.NewCmdPRTaskCreate(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--body")
}

func TestPRTaskCreate_PostsBlockerComment(t *testing.T) {
	t.Parallel()
	var gotIn backend.AddPRCommentInput
	fake := &testhelpers.FakeClient{
		T: t,
		AddPRCommentFn: func(ns, slug string, id int, in backend.AddPRCommentInput) (backend.PRComment, error) {
			gotIn = in
			return backend.PRComment{ID: 77, Text: in.Text, Severity: "BLOCKER", State: "OPEN"}, nil
		},
	}
	f, out, _ := newTaskFactory(t, fake, newTaskRunner())
	cmd := task.NewCmdPRTaskCreate(f)
	cmd.SetArgs([]string{"42", "--body", "fix the bug"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "BLOCKER", gotIn.Severity)
	assert.Equal(t, "fix the bug", gotIn.Text)
	assert.Contains(t, out.String(), "Created task #77")
}

func TestPRTaskCreate_WithParentFlag(t *testing.T) {
	t.Parallel()
	var gotIn backend.AddPRCommentInput
	fake := &testhelpers.FakeClient{
		T: t,
		AddPRCommentFn: func(ns, slug string, id int, in backend.AddPRCommentInput) (backend.PRComment, error) {
			gotIn = in
			return backend.PRComment{ID: 78}, nil
		},
	}
	f, _, _ := newTaskFactory(t, fake, newTaskRunner())
	cmd := task.NewCmdPRTaskCreate(f)
	cmd.SetArgs([]string{"42", "--body", "sub-task", "--parent", "5"})
	require.NoError(t, cmd.Execute())

	require.NotNil(t, gotIn.Parent)
	assert.Equal(t, 5, *gotIn.Parent)
	assert.Equal(t, "BLOCKER", gotIn.Severity)
}

func TestPRTaskCreate_PropagatesAPIError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		AddPRCommentFn: func(ns, slug string, id int, in backend.AddPRCommentInput) (backend.PRComment, error) {
			return backend.PRComment{}, errors.New("500 server error")
		},
	}
	f, _, _ := newTaskFactory(t, fake, newTaskRunner())
	cmd := task.NewCmdPRTaskCreate(f)
	cmd.SetArgs([]string{"42", "--body", "x"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

// ── resolve / reopen tests ───────────────────────────────────────────────────

// fakeStateSetterClient wraps FakeClient + implements PRCommentStateSetter.
type fakeStateSetterClient struct {
	*testhelpers.FakeClient
	SetPRCommentStateFn func(ns, slug string, id, commentID int, state string) error
}

func (f *fakeStateSetterClient) SetPRCommentState(ns, slug string, id, commentID int, state string) error {
	if f.SetPRCommentStateFn != nil {
		return f.SetPRCommentStateFn(ns, slug, id, commentID, state)
	}
	if f.T != nil {
		f.T.Fatalf("unexpected call to fakeStateSetterClient.SetPRCommentState")
	}
	return nil
}

var _ backend.PRCommentStateSetter = (*fakeStateSetterClient)(nil)

func TestPRTaskResolve_CallsSetStateResolved(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug, gotState string
	var gotPRID, gotTaskID int
	fake := &fakeStateSetterClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		SetPRCommentStateFn: func(ns, slug string, id, commentID int, state string) error {
			gotNS, gotSlug, gotState = ns, slug, state
			gotPRID, gotTaskID = id, commentID
			return nil
		},
	}
	f, _, _ := newTaskFactory(t, fake, newTaskRunner())
	cmd := task.NewCmdPRTaskResolve(f)
	cmd.SetArgs([]string{"42", "77"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-service", gotSlug)
	assert.Equal(t, 42, gotPRID)
	assert.Equal(t, 77, gotTaskID)
	assert.Equal(t, "RESOLVED", gotState)
}

func TestPRTaskReopen_CallsSetStateOpen(t *testing.T) {
	t.Parallel()
	var gotState string
	fake := &fakeStateSetterClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		SetPRCommentStateFn: func(ns, slug string, id, commentID int, state string) error {
			gotState = state
			return nil
		},
	}
	f, _, _ := newTaskFactory(t, fake, newTaskRunner())
	cmd := task.NewCmdPRTaskReopen(f)
	cmd.SetArgs([]string{"42", "77"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "OPEN", gotState)
}

func TestPRTaskResolve_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	// Plain FakeClient does NOT implement PRCommentStateSetter.
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newTaskFactory(t, fake, newTaskRunner())
	cmd := task.NewCmdPRTaskResolve(f)
	cmd.SetArgs([]string{"42", "77"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

func TestPRTaskReopen_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newTaskFactory(t, fake, newTaskRunner())
	cmd := task.NewCmdPRTaskReopen(f)
	cmd.SetArgs([]string{"42", "77"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

func TestPRTaskResolve_InvalidTaskID(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newTaskFactory(t, fake, newTaskRunner())
	cmd := task.NewCmdPRTaskResolve(f)
	cmd.SetArgs([]string{"42", "notanumber"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "TASK_ID")
}
