package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/auth"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// TestAuthStatus_HostnameFilter_UnknownHost verifies that `auth status
// --hostname X` with an X not present in the config returns a "not logged
// into X" error rather than silently exiting 0.
func TestAuthStatus_HostnameFilter_UnknownHost(t *testing.T) {
	t.Parallel()

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig: authConfig, // contains bb.example.com only
	})
	cmd := auth.NewCmdAuthStatus(f)
	cmd.SetArgs([]string{"--hostname", "other.example.com"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not logged into")
	assert.Contains(t, err.Error(), "other.example.com")
}
