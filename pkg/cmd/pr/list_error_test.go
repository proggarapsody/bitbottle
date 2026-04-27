package pr_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// TestPRList_NotInGitRepo_ReturnsCleanError ensures `pr list` outside a git
// repo doesn't leak "exit status 128" or stale phrasing in the error.
func TestPRList_NotInGitRepo_ReturnsCleanError(t *testing.T) {
	t.Parallel()

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: "bb.example.com:\n  oauth_token: tok\n",
		GitRunner: testhelpers.NewFakeRunner(
			testhelpers.RunResponse{Err: errors.New("exit status 128")},
		),
	})

	root := pr.NewCmdPR(f)
	root.SetArgs([]string{"list"})
	err := root.Execute()
	require.Error(t, err)

	msg := err.Error()
	assert.False(t, strings.Contains(msg, "exit status 128"),
		"error must not leak raw git exit status, got: %s", msg)
	assert.Contains(t, msg, "no git remotes",
		"error should clearly say no git remotes were found")
}
