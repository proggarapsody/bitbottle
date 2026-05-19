// Package app contains the real entry-point logic for the bitbottle binary.
// Extracted from cmd/bitbottle/main.go so that test/script can import it
// for testscript.RunMain in-process dispatch without importing package main.
package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/proggarapsody/bitbottle/internal/aliases"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/root"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// Injected at link time from cmd/bitbottle/main.go via the Version field.
var (
	Version   = "dev"
	BuildDate = "unknown"
	Commit    = "unknown"
)

// Run is the real entry point, returning an exit code.
func Run() int {
	f := factory.New()
	cmd := root.NewCmdRoot(f)
	cmd.Version = Version + " (" + Commit + ") built " + BuildDate

	args, err := expandAlias(f, os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	cmd.SetArgs(args)

	if err := cmd.Execute(); err != nil {
		cmdutil.ExplainError(f.IOStreams, err)
		return 1
	}
	return 0
}

// expandAlias resolves a top-level alias. For shell aliases it execs
// $SHELL -c and never returns. Alias file failures fall back to raw args
// to prevent startup hangs on corrupt aliases.yml.
func expandAlias(f *factory.Factory, args []string) ([]string, error) {
	if len(args) == 0 || isFlag(args[0]) {
		return args, nil
	}
	store, err := f.Aliases()
	if err != nil {
		return args, nil //nolint:nilerr
	}
	exp, ok, err := aliases.Resolve(store, args[0], args[1:])
	if err != nil {
		return nil, err
	}
	if !ok {
		return args, nil
	}
	if exp.Shell != "" {
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
		c := exec.CommandContext(context.Background(), shell, "-c", exp.Shell)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				os.Exit(ee.ExitCode())
			}
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	return exp.Args, nil
}

func isFlag(s string) bool {
	return len(s) > 0 && s[0] == '-'
}
