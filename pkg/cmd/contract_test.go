// Package cmd_test holds the SCRIPT-TRUST CLI contract test. It walks the
// whole cobra command tree, runs every leaf command with intentionally-bad
// input, and asserts the script-facing contract:
//
//	a leaf that prints a diagnostic to stderr MUST exit non-zero.
//
// This is the regression guard against the "print-then-return-nil" anti-
// pattern (an error printed to stderr while the process exits 0), which
// silently breaks any script that relies on exit codes (`set -e`, `&&`).
package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/root"
)

// leafCommands returns every runnable leaf command in the tree (a command
// with a Run/RunE and no runnable subcommands). Pure command groups (e.g.
// `pr`, `repo`) are skipped — they print help and exit 0 by design.
func leafCommands(rootCmd *cobra.Command) []*cobra.Command {
	var leaves []*cobra.Command
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		runnableChild := false
		for _, sub := range c.Commands() {
			if sub.IsAdditionalHelpTopicCommand() || sub.Hidden {
				continue
			}
			runnableChild = true
			walk(sub)
		}
		if !runnableChild && (c.RunE != nil || c.Run != nil) {
			leaves = append(leaves, c)
		}
	}
	walk(rootCmd)
	return leaves
}

// TestLeafCommands_ErrorPathsExitNonZero is the contract guard. For each
// leaf it runs two intentionally-bad invocations and asserts that whenever
// the command emits anything to stderr, Execute returns a non-nil error
// (which app.Run maps to exit code 1).
func TestLeafCommands_ErrorPathsExitNonZero(t *testing.T) {
	t.Parallel()

	// Bad-input variants. No host is configured (empty hosts.yml), so any
	// command that gets past arg validation fails host/repo resolution; any
	// command that validates args first fails there. Either way an error
	// path is exercised. We never expect a successful (exit-0) run here.
	variants := [][]string{
		nil,                       // no positionals — trips RangeArgs / missing required flag
		{"@@@not-a-valid-ref@@@"}, // bogus single positional
		{"@@@bad@@@", "x", "y"},   // bogus positional + extra trailing args
	}

	for _, leaf := range leafCommands(buildRoot(t)) {
		path := leaf.CommandPath()
		if skipLeaf(path) {
			continue
		}
		// Some leaves legitimately succeed with no args and a stub backend
		// (e.g. read-only listings against the noop 404 client print an
		// error AND return it). We only assert the *negative* contract:
		// stderr-without-error is forbidden. Successful runs are fine.
		for _, args := range variants {
			t.Run(path+"/"+strings.Join(args, "_"), func(t *testing.T) {
				t.Parallel()
				errOut, err := runLeaf(t, path, args)
				if strings.TrimSpace(errOut) != "" && err == nil {
					t.Fatalf("contract violation: %q printed to stderr but exited 0 (returned nil):\n%s",
						path, errOut)
				}
			})
		}
	}
}

// skipLeaf excludes leaves that shell out to external processes (npx, the
// user's $SHELL, a long-running MCP server) or otherwise have side effects
// that make them unsuitable for in-process fuzzing. Their exit-code
// contract is covered by their own package tests; the script-trust surface
// is the API/repo command family.
func skipLeaf(path string) bool {
	skipPrefixes := []string{
		"bitbottle skill",      // execs npx to install skills
		"bitbottle extension",  // execs/clones external extension binaries
		"bitbottle mcp",        // starts a long-running stdio server
		"bitbottle completion", // emits a shell script; not an error surface
		"bitbottle alias",      // can exec the user's $SHELL
	}
	for _, p := range skipPrefixes {
		if path == p || strings.HasPrefix(path, p+" ") {
			return true
		}
	}
	return false
}

// runLeaf builds a fresh root (so each invocation gets clean buffers and an
// unmutated factory) and executes the command identified by path with args.
// It returns the stderr capture and the Execute error.
func runLeaf(t *testing.T, path string, args []string) (errOut string, err error) {
	t.Helper()
	rootCmd, errBuf := newRootWithCapture(t)
	// path is "bitbottle sub leaf …"; drop the program name.
	parts := strings.Fields(path)
	invocation := append(append([]string{}, parts[1:]...), args...)
	rootCmd.SetArgs(invocation)
	// A leaf that panics on bad input is itself a contract violation; turn
	// it into a test failure with a clear message rather than crashing the
	// whole run.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("contract violation: %q panicked on args %v: %v", path, args, r)
		}
	}()
	err = rootCmd.Execute()
	return errBuf.String(), err
}

func buildRoot(t *testing.T) *cobra.Command {
	t.Helper()
	rootCmd, _ := newRootWithCapture(t)
	return rootCmd
}

// newRootWithCapture wires a hermetic factory (empty hosts.yml, no git
// remote, 404 HTTP stub) and the global output flags that app.Run relies
// on, returning the root command and the stderr buffer.
func newRootWithCapture(t *testing.T) (*cobra.Command, *bytes.Buffer) {
	t.Helper()
	f, _, errOut := factorytest.New(t, factorytest.Opts{})
	// factorytest leaves ConfigDir unset; a few leaves (extension list/…)
	// read it. Point it at a temp dir so the harness mirrors production.
	dir := t.TempDir()
	f.ConfigDir = func() string { return dir }
	rootCmd := root.NewCmdRoot(f)
	// NewCmdRoot already sets SilenceErrors/SilenceUsage and registers the
	// global output flags (--json/--yaml/--jq/--template), matching app.Run.
	return rootCmd, errOut
}
