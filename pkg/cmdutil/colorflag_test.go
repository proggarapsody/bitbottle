package cmdutil_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

// rootWithFlag returns a fresh cobra root with --no-color registered.
func rootWithFlag() *cobra.Command {
	c := &cobra.Command{Use: "root"}
	cmdutil.RegisterNoColorFlag(c)
	return c
}

func TestApplyNoColorFlag_FlagSet_DisablesColor(t *testing.T) {
	t.Parallel()
	c := rootWithFlag()
	require.NoError(t, c.ParseFlags([]string{"--no-color"}))

	ios := iostreams.TestTTY()
	require.True(t, ios.ColorEnabled())

	cmdutil.ApplyNoColorFlag(c, ios)
	assert.False(t, ios.ColorEnabled())
}

func TestApplyNoColorFlag_FlagUnset_LeavesColorAlone(t *testing.T) {
	t.Parallel()
	c := rootWithFlag()
	require.NoError(t, c.ParseFlags([]string{}))

	ios := iostreams.TestTTY()
	require.True(t, ios.ColorEnabled())

	cmdutil.ApplyNoColorFlag(c, ios)
	assert.True(t, ios.ColorEnabled(),
		"omitting --no-color must not flip ColorEnabled")
}

// TestApplyNoColorFlag_PersistentInheritsToChild simulates the real cobra
// shape: --no-color is on root's PersistentFlags, and we read it from a
// child. cobra.Flags() merges parent persistent flags into the child's
// flagset, so this should still pick it up.
func TestApplyNoColorFlag_PersistentInheritsToChild(t *testing.T) {
	t.Parallel()
	root := rootWithFlag()
	leaf := &cobra.Command{Use: "leaf", Run: func(*cobra.Command, []string) {}}
	root.AddCommand(leaf)

	root.SetArgs([]string{"leaf", "--no-color"})
	// Capture the leaf's effective flagset by hooking RunE.
	var captured *cobra.Command
	leaf.Run = func(c *cobra.Command, _ []string) { captured = c }
	require.NoError(t, root.Execute())
	require.NotNil(t, captured)

	ios := iostreams.TestTTY()
	cmdutil.ApplyNoColorFlag(captured, ios)
	assert.False(t, ios.ColorEnabled(),
		"--no-color on a leaf must be visible to ApplyNoColorFlag")
}

// TestApplyNoColorFlag_FlagNotRegistered_NoOp guards the defensive branch:
// calling Apply on a command tree that never registered the flag must not
// panic and must not mutate IOStreams. Tests insurance against future
// refactors that move the registration around.
func TestApplyNoColorFlag_FlagNotRegistered_NoOp(t *testing.T) {
	t.Parallel()
	c := &cobra.Command{Use: "bare"} // no RegisterNoColorFlag call
	ios := iostreams.TestTTY()
	require.True(t, ios.ColorEnabled())

	cmdutil.ApplyNoColorFlag(c, ios)
	assert.True(t, ios.ColorEnabled())
}
