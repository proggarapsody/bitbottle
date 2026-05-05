package auth_test

import (
	"testing"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/auth"
)

func TestNewCmdAuthStatus_HasHostnameFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := auth.NewCmdAuthStatus(f)
	assert.NotNil(t, cmd.Flag("hostname"))
}

func TestAuthStatus_PrintsConfiguredHosts(t *testing.T) {
	t.Parallel()

	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: authConfig})
	cmd := auth.NewCmdAuthStatus(f)
	cmd.SetArgs([]string{})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "bb.example.com")
	assert.Contains(t, got, "alice")
}

func TestAuthStatus_NoHosts_PrintsNothing(t *testing.T) {
	t.Parallel()

	f, out, _ := factorytest.New(t, factorytest.Opts{})
	cmd := auth.NewCmdAuthStatus(f)
	cmd.SetArgs([]string{})
	require.NoError(t, cmd.Execute())

	// no hosts → must not print any specific hostname; behaviour: "not logged in" message.
	assert.NotContains(t, out.String(), "bb.example.com")
}
