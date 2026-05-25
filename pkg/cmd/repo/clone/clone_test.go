package clone_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo/clone"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdClone_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := clone.NewCmdClone(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestRepoClone_SSH_InvokesGitClone(t *testing.T) {
	t.Parallel()

	runner := testhelpers.NewFakeRunner(testhelpers.RunResponse{})
	f, _, _ := newRepoRunnerFactory(t, nil, repoConfigSSH, runner)
	cmd := clone.NewCmdClone(f)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	require.NoError(t, cmd.Execute())

	cloneCall := findCloneCall(t, runner)
	require.GreaterOrEqual(t, len(cloneCall.Args), 2)
	assert.True(t, strings.HasPrefix(cloneCall.Args[1], "ssh://"), "expected SSH clone URL, got %q", cloneCall.Args[1])
}

func TestRepoClone_HTTPS_InvokesGitClone(t *testing.T) {
	t.Parallel()

	runner := testhelpers.NewFakeRunner(testhelpers.RunResponse{})
	f, _, _ := newRepoRunnerFactory(t, nil, repoConfigHTTPS, runner)
	cmd := clone.NewCmdClone(f)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	require.NoError(t, cmd.Execute())

	cloneCall := findCloneCall(t, runner)
	require.GreaterOrEqual(t, len(cloneCall.Args), 2)
	assert.True(t, strings.HasPrefix(cloneCall.Args[1], "https://"), "expected HTTPS clone URL, got %q", cloneCall.Args[1])
}

func TestRepoClone_Cloud_SSH_UsesCloudURL(t *testing.T) {
	t.Parallel()

	runner := testhelpers.NewFakeRunner(testhelpers.RunResponse{})
	f, _, _ := newRepoRunnerFactory(t, nil, repoConfigCloudSSH, runner)
	cmd := clone.NewCmdClone(f)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())

	cloneCall := findCloneCall(t, runner)
	require.GreaterOrEqual(t, len(cloneCall.Args), 2)
	url := cloneCall.Args[1]
	assert.Contains(t, url, "bitbucket.org", "cloud SSH URL must reference bitbucket.org, got %q", url)
}

// findCloneCall returns the first runner call with Args[0] == "clone".
func findCloneCall(t *testing.T, runner *testhelpers.FakeRunner) testhelpers.Call {
	t.Helper()
	for _, c := range runner.Calls {
		if len(c.Args) > 0 && c.Args[0] == "clone" {
			return c
		}
	}
	t.Fatalf("no git clone call found in %d runner calls", len(runner.Calls))
	return testhelpers.Call{}
}

func TestRepoClone_UsesAPICloneURL_SSH(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		GetRepoFn: func(ns, slug string) (backend.Repository, error) {
			return backend.Repository{
				Slug:      slug,
				Namespace: ns,
				CloneURLs: []backend.CloneURL{
					{Name: "ssh", URL: "ssh://git@bb.example.com:7999/MYPROJ/my-service.git"},
					{Name: "http", URL: "https://bb.example.com/scm/myproj/my-service.git"},
				},
			}, nil
		},
	}
	runner := testhelpers.NewFakeRunner(testhelpers.RunResponse{})
	f, _, _ := newRepoRunnerFactory(t, fake, repoConfigSSH, runner)
	cmd := clone.NewCmdClone(f)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	require.NoError(t, cmd.Execute())

	cloneCall := findCloneCall(t, runner)
	require.GreaterOrEqual(t, len(cloneCall.Args), 2)
	assert.Equal(t, "ssh://git@bb.example.com:7999/MYPROJ/my-service.git", cloneCall.Args[1])
}

func TestRepoClone_UsesAPICloneURL_HTTPS(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		GetRepoFn: func(ns, slug string) (backend.Repository, error) {
			return backend.Repository{
				Slug:      slug,
				Namespace: ns,
				CloneURLs: []backend.CloneURL{
					{Name: "https", URL: "https://bb.example.com/scm/myproj/my-service.git"},
					{Name: "ssh", URL: "ssh://git@bb.example.com:7999/MYPROJ/my-service.git"},
				},
			}, nil
		},
	}
	runner := testhelpers.NewFakeRunner(testhelpers.RunResponse{})
	f, _, _ := newRepoRunnerFactory(t, fake, repoConfigSSH, runner)
	cmd := clone.NewCmdClone(f)
	cmd.SetArgs([]string{"MYPROJ/my-service", "--https"})
	require.NoError(t, cmd.Execute())

	cloneCall := findCloneCall(t, runner)
	require.GreaterOrEqual(t, len(cloneCall.Args), 2)
	assert.Equal(t, "https://bb.example.com/scm/myproj/my-service.git", cloneCall.Args[1])
}

func TestRepoClone_FallsBackToHeuristic_WhenAPIReturnsEmpty(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		GetRepoFn: func(ns, slug string) (backend.Repository, error) {
			return backend.Repository{Slug: slug, Namespace: ns}, nil
		},
	}
	runner := testhelpers.NewFakeRunner(testhelpers.RunResponse{})
	f, _, _ := newRepoRunnerFactory(t, fake, repoConfigSSH, runner)
	cmd := clone.NewCmdClone(f)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	require.NoError(t, cmd.Execute())

	cloneCall := findCloneCall(t, runner)
	require.GreaterOrEqual(t, len(cloneCall.Args), 2)
	assert.True(t, strings.HasPrefix(cloneCall.Args[1], "ssh://"),
		"fallback heuristic should produce SSH URL, got %q", cloneCall.Args[1])
}

func TestRepoClone_WritesGitConfig_AfterClone(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		GetRepoFn: func(ns, slug string) (backend.Repository, error) {
			return backend.Repository{Slug: slug, Namespace: ns}, nil
		},
	}
	runner := testhelpers.NewFakeRunner(testhelpers.RunResponse{})
	f, _, _ := newRepoRunnerFactory(t, fake, repoConfigSSH, runner)
	cmd := clone.NewCmdClone(f)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	require.NoError(t, cmd.Execute())

	// Look for a config --local bitbottle.host call with -C my-service.
	found := false
	for _, c := range runner.Calls {
		if len(c.Args) >= 6 && c.Args[0] == "-C" && c.Args[1] == "my-service" &&
			c.Args[2] == "config" && c.Args[3] == "--local" && c.Args[4] == "bitbottle.host" {
			assert.Equal(t, "bb.example.com", c.Args[5])
			found = true
			break
		}
	}
	assert.True(t, found, "expected git -C my-service config --local bitbottle.host bb.example.com call")
}

func TestRepoClone_GitError_PropagatesError(t *testing.T) {
	t.Parallel()

	runner := testhelpers.NewFakeRunner(testhelpers.RunResponse{Err: errors.New("exit 128")})
	f, _, _ := newRepoRunnerFactory(t, nil, repoConfigSSH, runner)
	cmd := clone.NewCmdClone(f)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exit 128")
}
