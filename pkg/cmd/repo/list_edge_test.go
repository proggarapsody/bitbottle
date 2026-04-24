package repo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
)

// TestRepoList_MultipleHostsAmbiguous verifies that when more than one host is
// configured and --hostname is not specified the command returns an error
// telling the user to use --hostname.
func TestRepoList_MultipleHostsAmbiguous(t *testing.T) {
	t.Parallel()

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: "a.example.com:\n  oauth_token: tok\n  git_protocol: ssh\nb.example.com:\n  oauth_token: tok2\n  git_protocol: ssh\n",
	})

	cmd := repo.NewCmdRepoList(f)
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--hostname")
}
