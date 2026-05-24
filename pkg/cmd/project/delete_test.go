package project_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/project"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestProjectDelete_WithConfirm(t *testing.T) {
	t.Parallel()
	var gotKey string
	fake := &testhelpers.FakeClient{
		T: t,
		DeleteServerProjectFn: func(key string) error {
			gotKey = key
			return nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "git.example.com:\n  oauth_token: tok\n"})
	factorytest.UseBackend(f, fake)

	cmd := project.NewCmdProject(f)
	cmd.SetArgs([]string{"delete", "PRJ", "--confirm", "--hostname", "git.example.com"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "PRJ", gotKey)
	assert.Contains(t, out.String(), "Deleted project PRJ")
}

func TestProjectDelete_NonTTY_RequiresConfirm(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "git.example.com:\n  oauth_token: tok\n"})
	// Default factorytest IOStreams is non-TTY.

	cmd := project.NewCmdProject(f)
	cmd.SetArgs([]string{"delete", "PRJ", "--hostname", "git.example.com"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "--confirm"))
}
