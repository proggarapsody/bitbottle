package pipeline_test

import (
	"bytes"
	"testing"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/run"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// pipelineConfig is the shared single-host Cloud config for pipeline tests.
const pipelineConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

// newPipelineRunner returns a FakeRunner pre-seeded with a remote URL response.
func newPipelineRunner() *testhelpers.FakeRunner {
	return testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Stdout: "https://bitbucket.org/myworkspace/my-service.git\n"},
	)
}

// newPipelineFactory wires the shared config, a FakePipelineClient, and a runner.
func newPipelineFactory(t *testing.T, fake backend.Client, runner *testhelpers.FakeRunner) (*factory.Factory, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	f, out, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: pipelineConfig})
	factorytest.UseBackend(f, fake)
	f.GitRunner = func() run.Runner { return runner }
	return f, out, errOut
}
