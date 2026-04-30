package completion_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/root"
)

const completionConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

func executeCompletion(t *testing.T, shell string) (string, error) {
	t.Helper()
	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: completionConfig,
	})
	rootCmd := root.NewCmdRoot(f)
	rootCmd.SetArgs([]string{"completion", shell})
	err := rootCmd.Execute()
	return out.String(), err
}

func executeCompletionFlag(t *testing.T, shell string) (string, error) {
	t.Helper()
	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: completionConfig,
	})
	rootCmd := root.NewCmdRoot(f)
	rootCmd.SetArgs([]string{"completion", "--shell", shell})
	err := rootCmd.Execute()
	return out.String(), err
}

func TestCompletion_FlagStillSupported(t *testing.T) {
	t.Parallel()

	got, err := executeCompletionFlag(t, "bash")
	require.NoError(t, err)
	assert.NotEmpty(t, got)
}

func TestCompletion_Bash(t *testing.T) {
	t.Parallel()

	got, err := executeCompletion(t, "bash")
	require.NoError(t, err)
	assert.NotEmpty(t, got)
}

func TestCompletion_Zsh(t *testing.T) {
	t.Parallel()

	got, err := executeCompletion(t, "zsh")
	require.NoError(t, err)
	assert.NotEmpty(t, got)
}

func TestCompletion_Fish(t *testing.T) {
	t.Parallel()

	got, err := executeCompletion(t, "fish")
	require.NoError(t, err)
	assert.NotEmpty(t, got)
}

func TestCompletion_PowerShell(t *testing.T) {
	t.Parallel()

	got, err := executeCompletion(t, "powershell")
	require.NoError(t, err)
	assert.NotEmpty(t, got)
}

func TestCompletion_UnknownShell_ReturnsError(t *testing.T) {
	t.Parallel()

	_, err := executeCompletion(t, "tcsh")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tcsh")
}
