package cmdutil_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

// TestEnablePagerForAnnotated_WrapsAnnotatedCommand verifies that a command
// carrying Annotations["pager"]="true" gets its RunE wrapped so that on a
// TTY the output is streamed through $PAGER. We use `tr a-z A-Z` as a
// discriminating pager — if the bytes get transformed, we know the pager
// pipeline ran; if they appear lowercase the command bypassed the pager.
//
// This lifts pager wiring out of individual commands (pr/diff, commit/log)
// into a single root-level concern, so adding pager support to a new
// command is just a one-line annotation, not a 5-line copy-paste.
func TestEnablePagerForAnnotated_WrapsAnnotatedCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("spawns a pager subprocess")
	}
	t.Setenv("PAGER", "tr a-z A-Z")

	ios := iostreams.TestTTY()

	root := &cobra.Command{Use: "root", SilenceErrors: true, SilenceUsage: true}
	pagedCmd := &cobra.Command{
		Use:         "paged",
		Annotations: map[string]string{"pager": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprint(ios.Out, "hello pager")
			return nil
		},
	}
	root.AddCommand(pagedCmd)

	cmdutil.EnablePagerForAnnotated(root, ios)

	root.SetArgs([]string{"paged"})
	require.NoError(t, root.Execute())

	got := ios.Out.(*bytes.Buffer).String()
	assert.True(t, strings.Contains(got, "HELLO PAGER"),
		"output should be transformed by pager, got: %q", got)
}

// TestEnablePagerForAnnotated_LeavesUnmarkedCommandsUntouched verifies the
// negative case: short-output commands (list-style) must NOT be piped
// through the pager. They get used inside scripts and their output
// goes to terminals directly. If pager activation leaked to all
// commands, every `bitbottle pr list` would block waiting for `q`.
func TestEnablePagerForAnnotated_LeavesUnmarkedCommandsUntouched(t *testing.T) {
	// Set a transforming pager — if the unmarked command got wrapped
	// despite no annotation, this would catch it.
	t.Setenv("PAGER", "tr a-z A-Z")

	ios := iostreams.TestTTY()

	root := &cobra.Command{Use: "root", SilenceErrors: true, SilenceUsage: true}
	plain := &cobra.Command{
		Use: "plain",
		// no annotation
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprint(ios.Out, "hello plain")
			return nil
		},
	}
	root.AddCommand(plain)

	cmdutil.EnablePagerForAnnotated(root, ios)

	root.SetArgs([]string{"plain"})
	require.NoError(t, root.Execute())

	got := ios.Out.(*bytes.Buffer).String()
	assert.Equal(t, "hello plain", got,
		"unannotated command must not pipe through pager; got: %q", got)
}
