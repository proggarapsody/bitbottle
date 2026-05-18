package root_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/root"
)

// TestRootPersistentSkipTLSVerify_Registered verifies the global -k /
// --skip-tls-verify flag is wired on the root command. Without it, the
// `network.tls_unknown_authority` error hint promises a flag that
// doesn't exist, breaking the user's copy-paste recovery path.
// Regression: the user ran `pr approve N -R … -k` and got
// "unknown shorthand flag: 'k'" despite the error hint telling them to.
func TestRootPersistentSkipTLSVerify_Registered(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: rootConfig})
	cmd := root.NewCmdRoot(f)

	long := cmd.PersistentFlags().Lookup("skip-tls-verify")
	require.NotNil(t, long, "expected persistent --skip-tls-verify flag")
	assert.Equal(t, "k", long.Shorthand, "expected -k shorthand on --skip-tls-verify")
	assert.Equal(t, "false", long.DefValue)
}

// TestRootPersistentSkipTLSVerify_PropagatesToFactory verifies that
// passing `-k` on the root sets f.SkipTLSOverride before any
// subcommand RunE executes. Inject a no-op subcommand whose RunE
// snapshots the override value — that's the same lifecycle moment when
// a real command would dial through f.Backend / f.HTTPClient.
func TestRootPersistentSkipTLSVerify_PropagatesToFactory(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: rootConfig})
	cmd := root.NewCmdRoot(f)

	var seenOverride bool
	probe := &cobra.Command{
		Use: "probe-tls",
		RunE: func(_ *cobra.Command, _ []string) error {
			seenOverride = f.SkipTLSOverride
			return nil
		},
	}
	cmd.AddCommand(probe)

	require.False(t, f.SkipTLSOverride, "precondition: override should default to false")
	cmd.SetArgs([]string{"probe-tls", "-k"})
	require.NoError(t, cmd.Execute())
	assert.True(t, seenOverride, "expected -k to set f.SkipTLSOverride before subcommand RunE")
	assert.True(t, f.SkipTLSOverride, "expected f.SkipTLSOverride to remain set after RunE")
}
