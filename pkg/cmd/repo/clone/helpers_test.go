package clone_test

import (
	"bytes"
	"testing"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/run"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const (
	repoConfig         = "bb.example.com:\n  oauth_token: tok\n  user: alice\n  git_protocol: ssh\n"
	repoConfigSSH      = repoConfig
	repoConfigHTTPS    = "bb.example.com:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"
	repoConfigCloudSSH = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: ssh\n"
)

func newRepoRunnerFactory(t *testing.T, fake backend.Client, cfg string, runner *testhelpers.FakeRunner) (*factory.Factory, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	f, out, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: cfg})
	if fake != nil {
		factorytest.UseBackend(f, fake)
	}
	f.GitRunner = func() run.Runner { return runner }
	return f, out, errOut
}
