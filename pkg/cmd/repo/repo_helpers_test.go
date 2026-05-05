package repo_test

// repo_helpers_test.go — shared test helpers for repo sub-command tests.

import (
	"bytes"
	"testing"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/run"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// Config fixtures shared across repo test files.
// repoConfig / repoConfigSSH share the same SSH value; repoConfigSSH is kept
// as an alias so clone_test.go reads clearly without a separate declaration.
const (
	repoConfig         = "bb.example.com:\n  oauth_token: tok\n  user: alice\n  git_protocol: ssh\n"
	repoConfigSSH      = repoConfig
	repoConfigHTTPS    = "bb.example.com:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"
	repoConfigCloudSSH = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: ssh\n"
)

// newRepoFactory is the standard factory for repo command tests that talk to
// the API but do not invoke git (create, delete, view).
func newRepoFactory(t *testing.T, fake backend.Client) (*factory.Factory, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	f, out, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: repoConfig})
	factorytest.UseBackend(f, fake)
	return f, out, errOut
}

// newRepoRunnerFactory is the standard factory for repo command tests that
// invoke git (clone) but also talk to the API.
func newRepoRunnerFactory(t *testing.T, fake backend.Client, cfg string, runner *testhelpers.FakeRunner) (*factory.Factory, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	f, out, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: cfg})
	factorytest.UseBackend(f, fake)
	f.GitRunner = func() run.Runner { return runner }
	return f, out, errOut
}
