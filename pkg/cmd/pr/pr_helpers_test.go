package pr_test

// pr_helpers_test.go — shared test helpers for all pr sub-command tests.
//
// Helpers here live in the pr_test package (external test package) so they
// are usable across every *_test.go file in this directory without exposing
// anything in the production package.

import (
	"bytes"
	"testing"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// prConfig is the shared single-host config used by all PR command tests.
const prConfig = "bb.example.com:\n  oauth_token: tok\n  user: alice\n  git_protocol: ssh\n"

// newPRRunner returns a FakeRunner pre-seeded with the single response that
// resolveRepoRef needs: the output of `git remote get-url origin`.
// All PR sub-commands that operate on an existing PR (approve, view, diff,
// merge, checkout) call resolveRepoRef as their first git operation.
func newPRRunner(extra ...testhelpers.RunResponse) *testhelpers.FakeRunner {
	responses := []testhelpers.RunResponse{
		{Stdout: "ssh://git@bb.example.com:7999/myproj/my-service.git\n"},
	}
	return testhelpers.NewFakeRunner(append(responses, extra...)...)
}

// newPRFactory constructs the standard factory used by PR sub-command tests.
// It wires in the shared prConfig, the supplied backend fake, and the supplied
// runner (use newPRRunner() for the common single-response case).
func newPRFactory(t *testing.T, fake backend.Client, runner *testhelpers.FakeRunner) (*factory.Factory, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	return factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   prConfig,
		BackendOverride: fake,
		GitRunner:       runner,
	})
}
