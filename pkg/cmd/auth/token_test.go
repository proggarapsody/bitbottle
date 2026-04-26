package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/auth"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func TestAuthToken_PrintsToken(t *testing.T) {
	t.Parallel()

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: authConfig,
	})
	cmd := auth.NewCmdAuthToken(f)
	cmd.SetArgs([]string{})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "tok")
}

func TestAuthToken_NoToken_ReturnsError(t *testing.T) {
	t.Parallel()

	const noTokenConfig = "bb.example.com:\n  oauth_token: \"\"\n  user: alice\n  git_protocol: ssh\n"

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: noTokenConfig,
	})
	cmd := auth.NewCmdAuthToken(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no token")
}
