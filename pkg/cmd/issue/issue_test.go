package issue_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/run"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/issue"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const cloudConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

// runner returns a FakeRunner with the standard `git remote get-url origin`
// response so commands without an explicit PROJECT/REPO can resolve.
func runner() *testhelpers.FakeRunner {
	return testhelpers.NewFakeRunner(testhelpers.RunResponse{
		Stdout: "https://bitbucket.org/acme/repo.git\n",
	})
}

func newFactory(t *testing.T, fake backend.Client) (*factory.Factory, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	f, out, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)
	r := runner()
	f.GitRunner = func() run.Runner { return r }
	return f, out, errOut
}

// noIssueFake wraps backend.Client without satisfying IssueClient. The
// embedding (not the concrete struct) prevents method promotion.
type noIssueFake struct {
	backend.Client
}

// ---- list ----

func TestList_PrintsIssueIDsAndStates(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListIssuesFn: func(ns, slug, state string, limit int) ([]backend.Issue, error) {
			assert.Equal(t, "open", state, "default --state must be 'open'")
			return []backend.Issue{
				{ID: 7, Title: "Bug", State: "open", Kind: "bug", Reporter: backend.User{Slug: "alice"}},
				{ID: 8, Title: "Idea", State: "new", Kind: "enhancement", Reporter: backend.User{Slug: "carol"}},
			}, nil
		},
	}
	f, out, _ := newFactory(t, fake)
	cmd := issue.NewCmdIssueList(f)
	cmd.SetArgs([]string{"acme/repo"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "7")
	assert.Contains(t, got, "Bug")
	assert.Contains(t, got, "8")
	assert.Contains(t, got, "Idea")
}

func TestList_StateAllOmitsFilter(t *testing.T) {
	t.Parallel()
	var gotState string
	fake := &testhelpers.FakeClient{
		T: t,
		ListIssuesFn: func(ns, slug, state string, limit int) ([]backend.Issue, error) {
			gotState = state
			return nil, nil
		},
	}
	f, _, _ := newFactory(t, fake)
	cmd := issue.NewCmdIssueList(f)
	cmd.SetArgs([]string{"acme/repo", "--state", "all"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "", gotState, "--state all must pass empty filter to backend")
}

func TestList_StateOnHoldNormalisesToSpace(t *testing.T) {
	t.Parallel()
	var gotState string
	fake := &testhelpers.FakeClient{
		T: t,
		ListIssuesFn: func(ns, slug, state string, limit int) ([]backend.Issue, error) {
			gotState = state
			return nil, nil
		},
	}
	f, _, _ := newFactory(t, fake)
	cmd := issue.NewCmdIssueList(f)
	cmd.SetArgs([]string{"acme/repo", "--state", "on-hold"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "on hold", gotState,
		"on-hold (CLI-friendly) must map to Bitbucket's 'on hold' (with space)")
}

func TestList_ServerBackend_ReturnsUnsupported(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, noIssueFake{Client: &testhelpers.FakeClient{T: t}})
	r := runner()
	f.GitRunner = func() run.Runner { return r }

	cmd := issue.NewCmdIssueList(f)
	cmd.SetArgs([]string{"acme/repo"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Cloud only")
}

// ---- view ----

func TestView_AcceptsIDOnly(t *testing.T) {
	t.Parallel()
	var gotID int
	fake := &testhelpers.FakeClient{
		T: t,
		GetIssueFn: func(ns, slug string, id int) (backend.Issue, error) {
			gotID = id
			return backend.Issue{ID: id, Title: "T", Reporter: backend.User{Slug: "a"}}, nil
		},
	}
	f, _, _ := newFactory(t, fake)
	cmd := issue.NewCmdIssueView(f)
	cmd.SetArgs([]string{"acme/repo", "42"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, 42, gotID)
}

func TestView_RejectsNonNumericID(t *testing.T) {
	t.Parallel()
	f, _, _ := newFactory(t, &testhelpers.FakeClient{T: t})
	cmd := issue.NewCmdIssueView(f)
	cmd.SetArgs([]string{"acme/repo", "not-a-number"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid issue ID")
}

func TestView_AcceptsRepoArgPlusID(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug string
	fake := &testhelpers.FakeClient{
		T: t,
		GetIssueFn: func(ns, slug string, id int) (backend.Issue, error) {
			gotNS, gotSlug = ns, slug
			return backend.Issue{ID: id, Reporter: backend.User{Slug: "a"}}, nil
		},
	}
	f, _, _ := newFactory(t, fake)
	cmd := issue.NewCmdIssueView(f)
	cmd.SetArgs([]string{"otherws/otherrepo", "9"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "otherws", gotNS)
	assert.Equal(t, "otherrepo", gotSlug)
}

// ---- create ----

func TestCreate_RequiresTitle(t *testing.T) {
	t.Parallel()
	f, _, _ := newFactory(t, &testhelpers.FakeClient{T: t})
	cmd := issue.NewCmdIssueCreate(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err, "missing --title must error before any backend call")
}

func TestCreate_PassesAllFields(t *testing.T) {
	t.Parallel()
	var got backend.CreateIssueInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreateIssueFn: func(ns, slug string, in backend.CreateIssueInput) (backend.Issue, error) {
			got = in
			return backend.Issue{ID: 100, Title: in.Title, Reporter: backend.User{Slug: "a"}}, nil
		},
	}
	f, out, _ := newFactory(t, fake)
	cmd := issue.NewCmdIssueCreate(f)
	cmd.SetArgs([]string{
		"acme/repo",
		"--title", "Crash",
		"--body", "Repro:\n1. ...",
		"--kind", "bug",
		"--priority", "critical",
	})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "Crash", got.Title)
	assert.Equal(t, "Repro:\n1. ...", got.Content)
	assert.Equal(t, "bug", got.Kind)
	assert.Equal(t, "critical", got.Priority)
	assert.Contains(t, out.String(), "Crash")
}

// ---- close ----

func TestClose_SendsClosedState(t *testing.T) {
	t.Parallel()
	var gotID int
	var gotInput backend.UpdateIssueInput
	fake := &testhelpers.FakeClient{
		T: t,
		UpdateIssueFn: func(ns, slug string, id int, in backend.UpdateIssueInput) (backend.Issue, error) {
			gotID = id
			gotInput = in
			return backend.Issue{ID: id, State: "closed", Reporter: backend.User{Slug: "a"}}, nil
		},
	}
	f, out, _ := newFactory(t, fake)
	cmd := issue.NewCmdIssueClose(f)
	cmd.SetArgs([]string{"acme/repo", "42"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, 42, gotID)
	assert.Equal(t, "closed", gotInput.State)
	assert.Empty(t, gotInput.Title, "close must only mutate state")
	assert.Contains(t, out.String(), "Closed issue #42")
}

func TestClose_RejectsNonNumericID(t *testing.T) {
	t.Parallel()
	f, _, _ := newFactory(t, &testhelpers.FakeClient{T: t})
	cmd := issue.NewCmdIssueClose(f)
	cmd.SetArgs([]string{"acme/repo", "abc"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid issue ID")
}

// ---- formatter / colour ----

func TestIssueStateColor_TTY(t *testing.T) {
	t.Parallel()
	colorize := issue.IssueStateColor(testTTYIOStreams())
	cases := map[string]string{
		"open":      "\033[32mopen\033[0m",
		"new":       "\033[32mnew\033[0m",
		"closed":    "\033[35mclosed\033[0m",
		"resolved":  "\033[35mresolved\033[0m",
		"wontfix":   "\033[31mwontfix\033[0m",
		"invalid":   "\033[31minvalid\033[0m",
		"duplicate": "\033[31mduplicate\033[0m",
		"on hold":   "\033[33mon hold\033[0m",
		// Edge cases
		"":      "",
		"OPEN":  "OPEN", // case-sensitive
		"weird": "weird",
	}
	for state, want := range cases {
		assert.Equal(t, want, colorize(state), "state=%q", state)
	}
}
