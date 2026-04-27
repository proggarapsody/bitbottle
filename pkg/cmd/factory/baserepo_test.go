package factory_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// TestBaseRepo_InfersFromGitRemote verifies BaseRepo auto-detects the
// repository from `git remote get-url origin`.
func TestBaseRepo_InfersFromGitRemote(t *testing.T) {
	t.Parallel()

	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Stdout: "ssh://git@bb.example.com:7999/MYPROJ/myrepo.git\n"},
	)
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: "bb.example.com:\n  oauth_token: tok\n",
		GitRunner:     runner,
	})

	ref, err := f.BaseRepo()
	require.NoError(t, err)
	assert.Equal(t, "bb.example.com", ref.Host)
	assert.Equal(t, "MYPROJ", ref.Project)
	assert.Equal(t, "myrepo", ref.Slug)
}

// TestBaseRepo_NoGitRepo_ReturnsCleanError verifies the error message does
// not leak raw exec output (e.g. "exit status 128").
func TestBaseRepo_NoGitRepo_ReturnsCleanError(t *testing.T) {
	t.Parallel()

	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Err: errors.New("exit status 128")},
	)
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: "bb.example.com:\n  oauth_token: tok\n",
		GitRunner:     runner,
	})

	_, err := f.BaseRepo()
	require.Error(t, err)
	assert.False(t, strings.Contains(err.Error(), "exit status 128"),
		"error should not leak raw git exit status, got: %s", err.Error())
	assert.Contains(t, err.Error(), "no git remotes",
		"error should clearly say no git remotes were found")
}

// TestBaseRepo_NotAuthenticated_HintsAuthLogin verifies that with no hosts
// configured the error tells the user to run auth login.
func TestBaseRepo_NotAuthenticated_HintsAuthLogin(t *testing.T) {
	t.Parallel()

	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Err: errors.New("exit status 128")},
	)
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		// No InitialConfig — no hosts configured.
		GitRunner: runner,
	})

	_, err := f.BaseRepo()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "auth login",
		"error should hint at running auth login when no hosts configured")
}
