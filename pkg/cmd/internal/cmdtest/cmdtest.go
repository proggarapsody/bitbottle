// Package cmdtest holds shared test fixtures for pkg/cmd subcommand tests.
// Lives outside _test.go so subpackages anywhere under pkg/cmd can import it.
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

// Config is the shared single-host Cloud config used by command tests.
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

// NoPipelineFake wraps backend.Client without implementing
// backend.PipelineClient — simulates a Bitbucket Server backend.
// Embedding the interface (not the concrete FakeClient) prevents
// pipeline-method promotion.
type NoPipelineFake struct {
	backend.Client
}

// NoWorkspaceVarFake wraps backend.Client without implementing
// backend.WorkspaceVariableClient.
type NoWorkspaceVarFake struct {
	backend.Client
}

// NoDeploymentFake wraps backend.Client without implementing
// backend.DeploymentClient — simulates a Bitbucket Server backend.
type NoDeploymentFake struct {
	backend.Client
}

// NoDeployKeyFake wraps backend.Client without implementing
// backend.DeployKeyClient — simulates a backend that doesn't support deploy keys.
type NoDeployKeyFake struct {
	backend.Client
}

// NoDefaultReviewerFake wraps backend.Client without implementing
// backend.DefaultReviewerClient — simulates a backend that doesn't support
// default reviewer management.
type NoDefaultReviewerFake struct {
	backend.Client
}

// NoSSHKeyFake wraps backend.Client without implementing
// backend.SSHKeyClient — simulates a Bitbucket Server backend.
type NoSSHKeyFake struct {
	backend.Client
}
