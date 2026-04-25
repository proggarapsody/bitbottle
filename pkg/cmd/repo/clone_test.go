package repo_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdRepoClone_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := repo.NewCmdRepoClone(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestRepoClone_SSH_InvokesGitClone(t *testing.T) {
	t.Parallel()

	runner := testhelpers.NewFakeRunner(testhelpers.RunResponse{})
	f, _, _ := newRepoRunnerFactory(t, nil, repoConfigSSH, runner)
	cmd := repo.NewCmdRepoClone(f)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	require.NoError(t, cmd.Execute())

	require.NotEmpty(t, runner.Calls, "expected at least one git call")
	last := runner.Calls[len(runner.Calls)-1]
	require.NotEmpty(t, last.Args)
	assert.Equal(t, "clone", last.Args[0])
	require.GreaterOrEqual(t, len(last.Args), 2)
	assert.True(t, strings.HasPrefix(last.Args[1], "ssh://"), "expected SSH clone URL, got %q", last.Args[1])
}

func TestRepoClone_HTTPS_InvokesGitClone(t *testing.T) {
	t.Parallel()

	runner := testhelpers.NewFakeRunner(testhelpers.RunResponse{})
	f, _, _ := newRepoRunnerFactory(t, nil, repoConfigHTTPS, runner)
	cmd := repo.NewCmdRepoClone(f)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	require.NoError(t, cmd.Execute())

	require.NotEmpty(t, runner.Calls, "expected at least one git call")
	last := runner.Calls[len(runner.Calls)-1]
	require.NotEmpty(t, last.Args)
	assert.Equal(t, "clone", last.Args[0])
	require.GreaterOrEqual(t, len(last.Args), 2)
	assert.True(t, strings.HasPrefix(last.Args[1], "https://"), "expected HTTPS clone URL, got %q", last.Args[1])
}

func TestRepoClone_Cloud_SSH_UsesCloudURL(t *testing.T) {
	t.Parallel()

	runner := testhelpers.NewFakeRunner(testhelpers.RunResponse{})
	f, _, _ := newRepoRunnerFactory(t, nil, repoConfigCloudSSH, runner)
	cmd := repo.NewCmdRepoClone(f)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())

	require.NotEmpty(t, runner.Calls, "expected at least one git call")
	last := runner.Calls[len(runner.Calls)-1]
	require.GreaterOrEqual(t, len(last.Args), 2)
	url := last.Args[1]
	assert.Contains(t, url, "bitbucket.org", "cloud SSH URL must reference bitbucket.org, got %q", url)
}

func TestRepoClone_GitError_PropagatesError(t *testing.T) {
	t.Parallel()

	runner := testhelpers.NewFakeRunner(testhelpers.RunResponse{Err: errors.New("exit 128")})
	f, _, _ := newRepoRunnerFactory(t, nil, repoConfigSSH, runner)
	cmd := repo.NewCmdRepoClone(f)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exit 128")
}
