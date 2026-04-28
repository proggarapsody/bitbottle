package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/proggarapsody/bitbottle/internal/aliases"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/root"
)

// Injected at build time by goreleaser via -X ldflags.
var (
	version   = "dev"
	buildDate = "unknown"
	commit    = "unknown"
)

func main() {
	f := factory.New()
	cmd := root.NewCmdRoot(f)
	cmd.Version = version + " (" + commit + ") built " + buildDate

	args, err := expandAlias(f, os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	cmd.SetArgs(args)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// expandAlias resolves a top-level alias against the on-disk store. Returns
// the (possibly rewritten) argv that cobra should parse. For shell aliases,
// it execs $SHELL -c and never returns.
//
// Failures to load the alias file fall back to the raw args — startup must
// not block on a corrupt aliases.yml.
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
