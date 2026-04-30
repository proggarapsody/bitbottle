package repo_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// TestRepoSetDefault_WritesAllThreeKeys is the tracer: with HOST/PROJECT/REPO
// the command writes bitbottle.host, bitbottle.project, and bitbottle.slug
// to the local git config in a single invocation.
func TestRepoSetDefault_WritesAllThreeKeys(t *testing.T) {
	t.Parallel()
	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{}, // SetConfig host
		testhelpers.RunResponse{}, // SetConfig project
		testhelpers.RunResponse{}, // SetConfig slug
	)
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: "bb.example.com:\n  oauth_token: tok\n",
		GitRunner:     runner,
	})

	cmd := repo.NewCmdRepoSetDefault(f)
	cmd.SetArgs([]string{"bb.example.com/MYPROJ/myrepo"})
	require.NoError(t, cmd.Execute())

	runner.AssertCalled(t, "config", "--local", "bitbottle.host", "bb.example.com")
	runner.AssertCalled(t, "config", "--local", "bitbottle.project", "MYPROJ")
	runner.AssertCalled(t, "config", "--local", "bitbottle.slug", "myrepo")
}

// TestRepoSetDefault_BareProjectRepo_UsesSingleConfiguredHost verifies that
// when the user omits the host (PROJECT/REPO), the single configured host
// is used — same rule as ResolveRef and -R flag handling.
func TestRepoSetDefault_BareProjectRepo_UsesSingleConfiguredHost(t *testing.T) {
	t.Parallel()
	runner := testhelpers.NewFakeRunner(
		testhelpers.RunResponse{}, testhelpers.RunResponse{}, testhelpers.RunResponse{},
	)
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: "only.example.com:\n  oauth_token: tok\n",
		GitRunner:     runner,
	})

	cmd := repo.NewCmdRepoSetDefault(f)
	cmd.SetArgs([]string{"P/r"})
	require.NoError(t, cmd.Execute())

	runner.AssertCalled(t, "config", "--local", "bitbottle.host", "only.example.com")
}

// TestRepoSetDefault_BareProjectRepo_MultipleHosts_Errors pins the rule:
// without an explicit host AND with multiple configured, the command must
// error rather than guess.
func TestRepoSetDefault_BareProjectRepo_MultipleHosts_Errors(t *testing.T) {
	t.Parallel()
	runner := testhelpers.NewFakeRunner()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: "" +
			"bb1.example.com:\n  oauth_token: tok1\n" +
			"bb2.example.com:\n  oauth_token: tok2\n",
		GitRunner: runner,
	})

	cmd := repo.NewCmdRepoSetDefault(f)
	cmd.SetArgs([]string{"P/r"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "multiple hosts"),
		"expected multi-host error, got: %v", err)
}
