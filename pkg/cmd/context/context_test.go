package context_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/internal/run"
	contextcmd "github.com/proggarapsody/bitbottle/pkg/cmd/context"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const ctxConfigServer = "git.example.com:\n  oauth_token: tok\n  user: alice\n  git_protocol: ssh\n"
const ctxConfigCloud = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: ssh\n"

// inRepoRunner stubs git for the two calls Build() makes once BaseRepo has
// already resolved (we override BaseRepo directly via the factory in tests
// instead of routing through git config + remote get-url, which keeps the
// queue ordering invariant under future BaseRepo refactors). Order matches:
//  1. rev-parse --abbrev-ref HEAD          (current branch)
//  2. rev-list --left-right --count base...HEAD  (ahead/behind)
func inRepoRunner(t *testing.T, branch, leftRight string) *testhelpers.FakeRunner {
	t.Helper()
	return testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Stdout: branch},
		testhelpers.RunResponse{Stdout: leftRight},
	)
}

// newCtxFactory wires a Factory with a stub backend and (optionally) a stub
// git runner for the branch / ahead-behind probes. baseRepo, when non-zero,
// overrides f.BaseRepo so the test never depends on git-remote inference; an
// empty baseRepo leaves f.BaseRepo erroring (the outside-a-repo path).
func newCtxFactory(t *testing.T, cfg string, fake backend.Client, runner *testhelpers.FakeRunner, baseRepo bbrepo.RepoRef) (*factory.Factory, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	f, out, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: cfg})
	factorytest.UseBackend(f, fake)
	if runner != nil {
		f.GitRunner = func() run.Runner { return runner }
	}
	if baseRepo != (bbrepo.RepoRef{}) {
		f.BaseRepo = func() (bbrepo.RepoRef, error) { return baseRepo, nil }
	} else {
		f.BaseRepo = func() (bbrepo.RepoRef, error) {
			return bbrepo.RepoRef{}, errors.New("not a git repository")
		}
	}
	return f, out, errOut
}

// ---- flags ----

func TestNewCmdContext_HasJSONJqAndHostnameFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := contextcmd.NewCmdContext(f)
	format.RegisterOutputFlags(cmd)
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
	runner := inRepoRunner(t, "feat/x", "0\t2\n")
	f, out, _ := newCtxFactory(t, ctxConfigServer, fake, runner, bbrepo.RepoRef{
		Host: "git.example.com", Project: "PROJ", Slug: "repo",
	})

	cmd := contextcmd.NewCmdContext(f)
	format.RegisterOutputFlags(cmd)
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
	runner := inRepoRunner(t, "feat/x", "0\t2\n")
	f, out, _ := newCtxFactory(t, ctxConfigServer, fake, runner, bbrepo.RepoRef{
		Host: "git.example.com", Project: "PROJ", Slug: "repo",
	})

	cmd := contextcmd.NewCmdContext(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"--json"})
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
	runner := inRepoRunner(t, "main", "0\t0\n")
	f, out, _ := newCtxFactory(t, ctxConfigServer, fake, runner, bbrepo.RepoRef{
		Host: "git.example.com", Project: "PROJ", Slug: "repo",
	})

	cmd := contextcmd.NewCmdContext(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"--json", "--jq", ".user.slug"})
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
	// Outside-a-repo: BaseRepo errors. We override f.BaseRepo to drive that
	// path directly so the test does not depend on git's exit codes for
	// `remote get-url origin`.
	f, out, _ := newCtxFactory(t, ctxConfigServer, fake, nil, bbrepo.RepoRef{})

	cmd := contextcmd.NewCmdContext(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"--json"})
	require.NoError(t, cmd.Execute())

	var got map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &got))
	assert.Equal(t, "git.example.com", got["host"])
	assert.Equal(t, "", got["project"])
	assert.Equal(t, "", got["slug"])
	assert.Equal(t, "", got["branch"])
	assert.Equal(t, "", got["default_branch"])
	assert.Equal(t, "server", got["backend"])
	// ahead/behind are unknowable outside a repo and must not be reported
	// as the literal numeric value 0 — that would lie to agents reading the
	// JSON who would conclude "in sync" when the truth is "unknown". The
	// pointer-with-omitempty contract means the keys are absent.
	_, hasAhead := got["ahead"]
	_, hasBehind := got["behind"]
	assert.False(t, hasAhead, "ahead must be omitted outside a repo")
	assert.False(t, hasBehind, "behind must be omitted outside a repo")
	user, ok := got["user"].(map[string]any)
	require.True(t, ok, "user should still resolve outside a repo")
	assert.Equal(t, "alice", user["slug"])
}

func TestContext_InRepo_AheadBehindGitFailure_OmitsFromJSON(t *testing.T) {
	t.Parallel()
	// When `git rev-list --left-right --count base...HEAD` fails (e.g. base
	// not yet fetched), the ahead/behind values are genuinely unknown — they
	// must NOT be reported as 0/0 because agents reading the JSON would
	// conclude "in sync". The keys must be omitted via omitempty pointers.
	fake := &testhelpers.FakeClient{
		T: t,
		ListBranchesFn: func(_, _ string, _ int) ([]backend.Branch, error) {
			return []backend.Branch{{Name: "main", IsDefault: true}}, nil
		},
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{Slug: "alice", DisplayName: "Alice"}, nil
		},
	}
	runner := testhelpers.NewFakeRunner(
		// rev-parse --abbrev-ref HEAD → returns the branch.
		testhelpers.RunResponse{Stdout: "feat/x"},
		// rev-list --left-right --count → fails (e.g. unknown ref).
		testhelpers.RunResponse{Err: errors.New("fatal: ambiguous argument 'main...HEAD'")},
	)
	f, out, _ := newCtxFactory(t, ctxConfigServer, fake, runner, bbrepo.RepoRef{
		Host: "git.example.com", Project: "PROJ", Slug: "repo",
	})

	cmd := contextcmd.NewCmdContext(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"--json"})
	require.NoError(t, cmd.Execute())

	var got map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &got))
	_, hasAhead := got["ahead"]
	_, hasBehind := got["behind"]
	assert.False(t, hasAhead, "ahead must be omitted when git fails")
	assert.False(t, hasBehind, "behind must be omitted when git fails")
}

func TestContext_InRepo_AheadBehindGitFailure_TableSaysUnknown(t *testing.T) {
	t.Parallel()
	// Mirror the JSON-omission test: in TTY output the user must see an
	// explicit "unknown" advisory — not "0 / 0" — so they know to run git
	// fetch and retry rather than acting as if HEAD were synchronised.
	fake := &testhelpers.FakeClient{
		T: t,
		ListBranchesFn: func(_, _ string, _ int) ([]backend.Branch, error) {
			return []backend.Branch{{Name: "main", IsDefault: true}}, nil
		},
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{Slug: "alice", DisplayName: "Alice"}, nil
		},
	}
	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Stdout: "feat/x"},
		testhelpers.RunResponse{Err: errors.New("fatal: ambiguous argument 'main...HEAD'")},
	)
	f, out, _ := newCtxFactory(t, ctxConfigServer, fake, runner, bbrepo.RepoRef{
		Host: "git.example.com", Project: "PROJ", Slug: "repo",
	})

	cmd := contextcmd.NewCmdContext(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs(nil)
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "Ahead/Behind:")
	assert.Contains(t, got, "unknown")
	assert.Contains(t, got, "git fetch")
	// Importantly the literal "0 / 0" must NOT appear — that is the lie we
	// are explicitly avoiding.
	assert.NotContains(t, got, "0 / 0")
}

func TestContext_OutsideRepo_Table_PrintsAdvisory(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{Slug: "alice", DisplayName: "Alice"}, nil
		},
	}
	// Outside-a-repo human-readable table: humans cannot tell whether the
	// repo fields are empty by design or because of a bug. Emit a one-line
	// advisory so they can recover with -R/--hostname.
	f, out, _ := newCtxFactory(t, ctxConfigServer, fake, nil, bbrepo.RepoRef{})

	cmd := contextcmd.NewCmdContext(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs(nil)
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "Host:")
	assert.Contains(t, got, "git.example.com")
	assert.Contains(t, got, "Backend:")
	assert.Contains(t, got, "User:")
	assert.Contains(t, got, "alice")
	// The advisory must mention the recovery flags so the human knows
	// what to type next.
	assert.Contains(t, got, "outside a git repo")
	assert.Contains(t, got, "--hostname")
	assert.Contains(t, got, "-R")
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
	f, out, _ := newCtxFactory(t, ctxConfigCloud, fake, nil, bbrepo.RepoRef{})

	cmd := contextcmd.NewCmdContext(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"--json"})
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
	f, out, _ := newCtxFactory(t, multi, fake, nil, bbrepo.RepoRef{})

	cmd := contextcmd.NewCmdContext(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"--hostname", "bitbucket.org", "--json"})
	require.NoError(t, cmd.Execute())

	var got map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &got))
	assert.Equal(t, "bitbucket.org", got["host"])
	assert.Equal(t, "cloud", got["backend"])
}
