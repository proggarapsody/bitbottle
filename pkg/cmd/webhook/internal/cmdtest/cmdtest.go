// Package cmdtest holds shared test fixtures for webhook subcommand tests.
// It is not in `_test.go` so subpackages under pkg/cmd/webhook/* can import it.
package cmdtest

import (
	"bytes"
	"testing"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/run"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// Config is the shared single-host Cloud config for webhook tests. Webhooks
// are supported on both backends, so tests pick whichever host they need.
const Config = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

// NewRunner returns a FakeRunner pre-seeded with a remote URL response.
func NewRunner() *testhelpers.FakeRunner {
	return testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Stdout: "https://bitbucket.org/myworkspace/my-service.git\n"},
	)
}

// NewFactory wires the shared config, a backend.Client, and a runner.
func NewFactory(t *testing.T, fake backend.Client, runner *testhelpers.FakeRunner) (*factory.Factory, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	f, out, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: Config})
	factorytest.UseBackend(f, fake)
	f.GitRunner = func() run.Runner { return runner }
	return f, out, errOut
}
