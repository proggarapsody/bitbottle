package context_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/run"
	contextcmd "github.com/proggarapsody/bitbottle/pkg/cmd/context"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const ctxConfigServer = "git.example.com:\n  oauth_token: tok\n  user: alice\n  git_protocol: ssh\n"
const ctxConfigCloud = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: ssh\n"

// inRepoRunner stubs git so that BaseRepo + branch + ahead/behind queries return
// fixed values without touching the filesystem. Order matches the queries the
// command issues:
//  1. config --local --get bitbottle.host  (BaseRepo: missing → fall through)
//  2. remote get-url origin                (BaseRepo: returns SSH remote)
//  3. rev-parse --abbrev-ref HEAD          (current branch)
//  4. rev-list --left-right --count main...HEAD  (ahead/behind)
func inRepoRunner(t *testing.T, remote, branch, leftRight string) *testhelpers.FakeRunner {
	t.Helper()
	return testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Stdout: remote},
		testhelpers.RunResponse{Stdout: branch},
		testhelpers.RunResponse{Stdout: leftRight},
	)
}

func newCtxFactory(t *testing.T, cfg string, fake backend.Client, runner *testhelpers.FakeRunner) (*factory.Factory, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	f, out, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: cfg})
	factorytest.UseBackend(f, fake)
	if runner != nil {
		f.GitRunner = func() run.Runner { return runner }
	}
	return f, out, errOut
}

// ---- flags ----

func TestNewCmdContext_HasJSONJqAndHostnameFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := contextcmd.NewCmdContext(f)
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("jq"))
	assert.NotNil(t, cmd.Flag("hostname"))
}

// ---- inside-a-repo, server backend ----

func TestContext_InRepo_TablePrintsAllFields(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListBranchesFn: func(ns, slug string, _ int) ([]backend.Branch, error) {
			assert.Equal(t, "PROJ", ns)
			assert.Equal(t, "repo", slug)
			return []backend.Branch{
				{Name: "main", IsDefault: true},
				{Name: "feat/x", IsDefault: false},
			}, nil
		},
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{Slug: "alice", DisplayName: "Alice Smith"}, nil
		},
	}
	runner := inRepoRunner(t, "git@git.example.com:PROJ/repo.git", "feat/x", "0\t2\n")
	f, out, _ := newCtxFactory(t, ctxConfigServer, fake, runner)

	cmd := contextcmd.NewCmdContext(f)
	cmd.SetArgs(nil)
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "git.example.com")
	assert.Contains(t, got, "PROJ")
	assert.Contains(t, got, "repo")
	assert.Contains(t, got, "feat/x")
	assert.Contains(t, got, "main")
	assert.Contains(t, got, "alice")
	assert.Contains(t, got, "Alice Smith")
	assert.Contains(t, got, "server")
}

func TestContext_InRepo_JSON_EmitsPRDShape(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListBranchesFn: func(_, _ string, _ int) ([]backend.Branch, error) {
			return []backend.Branch{{Name: "main", IsDefault: true}}, nil
		},
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{Slug: "alice", DisplayName: "Alice Smith"}, nil
		},
	}
	runner := inRepoRunner(t, "git@git.example.com:PROJ/repo.git", "feat/x", "0\t2\n")
	f, out, _ := newCtxFactory(t, ctxConfigServer, fake, runner)

	cmd := contextcmd.NewCmdContext(f)
	cmd.SetArgs([]string{"--json", "host,project,slug,branch,default_branch,ahead,behind,user,backend"})
	require.NoError(t, cmd.Execute())

	var got map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &got))
	assert.Equal(t, "git.example.com", got["host"])
	assert.Equal(t, "PROJ", got["project"])
	assert.Equal(t, "repo", got["slug"])
	assert.Equal(t, "feat/x", got["branch"])
	assert.Equal(t, "main", got["default_branch"])
	assert.EqualValues(t, 2, got["ahead"])
	assert.EqualValues(t, 0, got["behind"])
	assert.Equal(t, "server", got["backend"])
	user, ok := got["user"].(map[string]any)
	require.True(t, ok, "user should be an object")
	assert.Equal(t, "alice", user["slug"])
	assert.Equal(t, "Alice Smith", user["display_name"])
}

func TestContext_InRepo_JQ_FiltersUserSlug(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListBranchesFn: func(_, _ string, _ int) ([]backend.Branch, error) {
			return []backend.Branch{{Name: "main", IsDefault: true}}, nil
		},
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{Slug: "alice", DisplayName: "Alice"}, nil
		},
	}
	runner := inRepoRunner(t, "git@git.example.com:PROJ/repo.git", "main", "0\t0\n")
	f, out, _ := newCtxFactory(t, ctxConfigServer, fake, runner)

	cmd := contextcmd.NewCmdContext(f)
	cmd.SetArgs([]string{"--json", "user", "--jq", ".user.slug"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, `"alice"`, strings.TrimSpace(out.String()))
}

// ---- outside-a-repo ----

func TestContext_OutsideRepo_StillResolvesHostAndUser(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{Slug: "alice", DisplayName: "Alice"}, nil
		},
	}
	// Outside-a-repo: BaseRepo()'s `git remote get-url origin` fails. The
	// FakeRunner returns no error by default once responses are exhausted, so
	// queue an explicit failure for the remote lookup.
	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Err: assertGitError("not a git repository")},
	)
	f, out, _ := newCtxFactory(t, ctxConfigServer, fake, runner)

	cmd := contextcmd.NewCmdContext(f)
	cmd.SetArgs([]string{"--json", "host,project,slug,branch,default_branch,ahead,behind,user,backend"})
	require.NoError(t, cmd.Execute())

	var got map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &got))
	assert.Equal(t, "git.example.com", got["host"])
	assert.Equal(t, "", got["project"])
	assert.Equal(t, "", got["slug"])
	assert.Equal(t, "", got["branch"])
	assert.Equal(t, "", got["default_branch"])
	assert.EqualValues(t, 0, got["ahead"])
	assert.EqualValues(t, 0, got["behind"])
	assert.Equal(t, "server", got["backend"])
	user, ok := got["user"].(map[string]any)
	require.True(t, ok, "user should still resolve outside a repo")
	assert.Equal(t, "alice", user["slug"])
}

// ---- backend type detection ----

func TestContext_CloudHost_BackendIsCloud(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{Slug: "alice", DisplayName: "Alice"}, nil
		},
	}
	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Err: assertGitError("not a git repository")},
	)
	f, out, _ := newCtxFactory(t, ctxConfigCloud, fake, runner)

	cmd := contextcmd.NewCmdContext(f)
	cmd.SetArgs([]string{"--json", "backend"})
	require.NoError(t, cmd.Execute())

	var got map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &got))
	assert.Equal(t, "cloud", got["backend"])
}

// ---- --hostname override ----

func TestContext_HostnameFlag_OverridesConfigPick(t *testing.T) {
	t.Parallel()
	const multi = "git.example.com:\n  oauth_token: tok\n  user: alice\n  git_protocol: ssh\n" +
		"bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: ssh\n"
	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{Slug: "alice", DisplayName: "Alice"}, nil
		},
	}
	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Err: assertGitError("not a git repository")},
	)
	f, out, _ := newCtxFactory(t, multi, fake, runner)

	cmd := contextcmd.NewCmdContext(f)
	cmd.SetArgs([]string{"--hostname", "bitbucket.org", "--json", "host,backend"})
	require.NoError(t, cmd.Execute())

	var got map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &got))
	assert.Equal(t, "bitbucket.org", got["host"])
	assert.Equal(t, "cloud", got["backend"])
}

// assertGitError fabricates an error matching what git would emit so the
// `outside-a-repo` branch in BaseRepo triggers naturally.
func assertGitError(msg string) error {
	return &gitErr{msg: msg}
}

type gitErr struct{ msg string }

func (e *gitErr) Error() string { return e.msg }
