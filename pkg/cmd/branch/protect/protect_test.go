package protect_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/run"
	"github.com/proggarapsody/bitbottle/pkg/cmd/branch/protect"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// serverConfig is a hosts.yml stub that registers a BBS-style host with
// backend_type=server, so f.Backend(...) returns the Server adapter (and
// any client we plug in via UseBackend has its BranchProtector view
// resolved correctly).
const serverConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n  backend_type: server\n"

const cloudConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n  backend_type: cloud\n"

// runner stubs `git remote get-url origin` so commands can resolve a
// repository when the user passes PROJECT/REPO positionally.
func runner() *testhelpers.FakeRunner {
	return testhelpers.NewFakeRunner(testhelpers.RunResponse{
		Stdout: "https://bitbucket.org/MYPROJ/my-service.git\n",
	})
}

func newFactory(t *testing.T, fake backend.Client, hostsYAML string) (*factory.Factory, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	f, out, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: hostsYAML, BackendType: "server"})
	factorytest.UseBackend(f, fake)
	r := runner()
	f.GitRunner = func() run.Runner { return r }
	return f, out, errOut
}

// noProtectFake wraps backend.Client without satisfying BranchProtector,
// simulating a Cloud backend invocation. The interface embedding (not the
// concrete struct) prevents method promotion.
type noProtectFake struct {
	backend.Client
}

// ---- list ----

func TestList_PrintsRows(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListBranchProtectionsFn: func(ns, slug string, limit int) ([]backend.BranchProtection, error) {
			assert.Equal(t, "MYPROJ", ns)
			assert.Equal(t, "my-service", slug)
			return []backend.BranchProtection{
				{ID: 1, Type: "fast-forward-only", MatcherID: "main", MatcherKind: "BRANCH", Users: []string{"alice"}, Groups: []string{"devs"}},
				{ID: 2, Type: "no-deletes", MatcherID: "release/*", MatcherKind: "PATTERN"},
			}, nil
		},
	}
	f, out, _ := newFactory(t, fake, serverConfig)
	cmd := protect.NewCmdList(f, nil)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "fast-forward-only")
	assert.Contains(t, got, "main")
	assert.Contains(t, got, "release/*")
	assert.Contains(t, got, "alice")
	assert.Contains(t, got, "@devs", "groups must be prefixed with @ in the EXEMPT column")
}

func TestList_JSON(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListBranchProtectionsFn: func(ns, slug string, limit int) ([]backend.BranchProtection, error) {
			return []backend.BranchProtection{{ID: 5, Type: "read-only", MatcherID: "main", MatcherKind: "BRANCH"}}, nil
		},
	}
	f, out, _ := newFactory(t, fake, serverConfig)
	cmd := protect.NewCmdList(f, nil)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--json", "id,type,matcher"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, `"id":5`)
	assert.Contains(t, got, `"type":"read-only"`)
	assert.Contains(t, got, `"matcher":"main"`)
}

func TestList_CloudBackend_ReturnsUnsupported(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, noProtectFake{Client: &testhelpers.FakeClient{T: t}})
	r := runner()
	f.GitRunner = func() run.Runner { return r }

	cmd := protect.NewCmdList(f, nil)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Server / Data Center only")
}

// ---- create ----

func TestCreate_RequiresExactlyOneOfBranchOrPattern(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		args []string
	}{
		{"neither", []string{"MYPROJ/my-service", "--type", "no-deletes"}},
		{"both", []string{"MYPROJ/my-service", "--type", "no-deletes", "--branch", "main", "--pattern", "release/*"}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fake := &testhelpers.FakeClient{T: t} // no Fn — must not be called
			f, _, _ := newFactory(t, fake, serverConfig)
			cmd := protect.NewCmdCreate(f, nil)
			cmd.SetArgs(tc.args)
			err := cmd.Execute()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "exactly one of --branch or --pattern")
		})
	}
}

func TestCreate_RejectsUnknownType(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newFactory(t, fake, serverConfig)
	cmd := protect.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--type", "garbage", "--branch", "main"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be one of")
}

func TestCreate_BranchPath(t *testing.T) {
	t.Parallel()
	var got backend.CreateBranchProtectionInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreateBranchProtectionFn: func(ns, slug string, in backend.CreateBranchProtectionInput) (backend.BranchProtection, error) {
			got = in
			return backend.BranchProtection{ID: 9, Type: in.Type, MatcherID: in.MatcherID, MatcherKind: in.MatcherKind}, nil
		},
	}
	f, out, _ := newFactory(t, fake, serverConfig)
	cmd := protect.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--type", "fast-forward-only", "--branch", "main", "--user", "alice", "--group", "devs"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "fast-forward-only", got.Type)
	assert.Equal(t, "main", got.MatcherID)
	assert.Equal(t, "BRANCH", got.MatcherKind)
	assert.Equal(t, []string{"alice"}, got.Users)
	assert.Equal(t, []string{"devs"}, got.Groups)
	assert.Contains(t, out.String(), "Created restriction 9")
}

func TestCreate_PatternPath(t *testing.T) {
	t.Parallel()
	var got backend.CreateBranchProtectionInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreateBranchProtectionFn: func(ns, slug string, in backend.CreateBranchProtectionInput) (backend.BranchProtection, error) {
			got = in
			return backend.BranchProtection{ID: 10, MatcherID: in.MatcherID, MatcherKind: in.MatcherKind, Type: in.Type}, nil
		},
	}
	f, _, _ := newFactory(t, fake, serverConfig)
	cmd := protect.NewCmdCreate(f, nil)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--type", "no-deletes", "--pattern", "release/*"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "release/*", got.MatcherID)
	assert.Equal(t, "PATTERN", got.MatcherKind)
}

// ---- delete ----

func TestDelete_RejectsNonNumericID(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newFactory(t, fake, serverConfig)
	cmd := protect.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"MYPROJ/my-service", "abc"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ID must be numeric")
}

func TestDelete_HappyPath(t *testing.T) {
	t.Parallel()
	var gotID int
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteBranchProtectionFn: func(ns, slug string, id int) error {
			gotID = id
			assert.Equal(t, "MYPROJ", ns)
			assert.Equal(t, "my-service", slug)
			return nil
		},
	}
	f, out, _ := newFactory(t, fake, serverConfig)
	cmd := protect.NewCmdDelete(f, nil)
	cmd.SetArgs([]string{"MYPROJ/my-service", "42"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, 42, gotID)
	assert.Contains(t, out.String(), "Deleted restriction 42")
}
