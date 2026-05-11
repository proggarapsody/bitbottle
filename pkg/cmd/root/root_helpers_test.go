package root_test

// root_helpers_test.go — shared test helpers for all root sub-command tests.

import (
	"bytes"
	"testing"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/run"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const rootConfig = "bb.example.com:\n  oauth_token: tok\n  user: alice\n  git_protocol: ssh\n"

func newStatusRunner() *testhelpers.FakeRunner {
	return testhelpers.NewFakeRunner(
		testhelpers.RunResponse{Stdout: "ssh://git@bb.example.com:7999/myproj/my-service.git\n"},
	)
}

func newStatusFactory(t *testing.T, fake backend.Client) (*factory.Factory, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	f, out, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: rootConfig})
	factorytest.UseBackend(f, fake)
	f.GitRunner = func() run.Runner { return newStatusRunner() }
	return f, out, errOut
}
