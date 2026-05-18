package root_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/root"
)

// TestRootPersistentDebug_Registered verifies that --debug is wired as a
// persistent flag on the root command. The transport.timeout error hint
// promises --debug; without it users would get "unknown flag: --debug".
func TestRootPersistentDebug_Registered(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: rootConfig})
	cmd := root.NewCmdRoot(f)

	flag := cmd.PersistentFlags().Lookup("debug")
	require.NotNil(t, flag, "expected persistent --debug flag on root")
	assert.Equal(t, "false", flag.DefValue)
}

// TestRootPersistentDebug_PropagatesToFactory verifies that --debug sets
// f.DebugHTTP before any subcommand RunE executes.
func TestRootPersistentDebug_PropagatesToFactory(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: rootConfig})
	cmd := root.NewCmdRoot(f)

	var seen bool
	probe := &cobra.Command{
		Use: "probe-debug",
		RunE: func(_ *cobra.Command, _ []string) error {
			seen = f.DebugHTTP
			return nil
		},
	}
	cmd.AddCommand(probe)

	require.False(t, f.DebugHTTP, "precondition: DebugHTTP should default to false")
	cmd.SetArgs([]string{"probe-debug", "--debug"})
	require.NoError(t, cmd.Execute())
	assert.True(t, seen, "expected --debug to set f.DebugHTTP before subcommand RunE")
}
