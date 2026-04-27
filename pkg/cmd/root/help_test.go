package root_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/root"
)

// TestHelpFunc_RendersArgumentsSection verifies that when a command (or any
// ancestor) carries an Annotations["help:arguments"] entry, the rendered help
// output includes an "ARGUMENTS" section with that text.
func TestHelpFunc_RendersArgumentsSection(t *testing.T) {
	t.Parallel()

	parent := &cobra.Command{Use: "pr", Short: "Manage pull requests"}
	parent.Annotations = map[string]string{
		"help:arguments": "A pull request can be supplied as argument in any of the following formats:\n- by number, e.g. \"123\"\n",
	}
	leaf := &cobra.Command{Use: "view <pr-id>", Short: "View a PR", Run: func(*cobra.Command, []string) {}}
	parent.AddCommand(leaf)

	root.SetHelpFunc(parent)

	out := &bytes.Buffer{}
	parent.SetOut(out)
	parent.SetErr(out)
	parent.SetArgs([]string{"view", "--help"})
	require.NoError(t, parent.Execute())

	got := out.String()
	assert.Contains(t, got, "ARGUMENTS", "help should contain ARGUMENTS section header")
	assert.Contains(t, got, "by number", "help should include the inherited annotation text")
}

// TestHelpFunc_RendersExamples verifies that when a command sets cmd.Example,
// the help renders it under an EXAMPLES section.
func TestHelpFunc_RendersExamples(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List things",
		Example: "  $ bitbottle pr list\n  $ bitbottle pr list --state merged",
		Run:     func(*cobra.Command, []string) {},
	}
	root.SetHelpFunc(cmd)

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"--help"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "EXAMPLES")
	assert.Contains(t, got, "bitbottle pr list --state merged")
}

// TestHelpFunc_NoArgumentsAnnotation_OmitsSection ensures the ARGUMENTS section
// is not rendered when no annotation is set anywhere in the command tree.
func TestHelpFunc_NoArgumentsAnnotation_OmitsSection(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "noargs", Short: "Plain", Run: func(*cobra.Command, []string) {}}
	root.SetHelpFunc(cmd)

	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"--help"})
	require.NoError(t, cmd.Execute())

	assert.False(t, strings.Contains(out.String(), "ARGUMENTS"),
		"ARGUMENTS section should not appear when no annotation is set")
}
