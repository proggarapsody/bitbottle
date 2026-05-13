package extension

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	extensions "github.com/proggarapsody/bitbottle/internal/extensions"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdExec returns the `extension exec` subcommand.
func NewCmdExec(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:                "exec NAME [args...]",
		Short:              "Execute an installed extension",
		Args:               cobra.MinimumNArgs(1),
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			extArgs := args[1:]

			extDir := filepath.Join(f.ConfigDir(), "extensions")
			mgr := extensions.New(extDir, nil)

			// Read token from env; extension decides what to do if empty.
			token := os.Getenv("BB_TOKEN")

			// Version string injected at build time; fall back to "dev".
			version := resolveVersion(cmd)

			err := mgr.Exec(name, extArgs, token, version)
			if err != nil {
				var ee *exec.ExitError
				if errors.As(err, &ee) {
					os.Exit(ee.ExitCode())
				}
				return fmt.Errorf("extension exec: %w", err)
			}
			return nil
		},
	}
	return cmd
}

// resolveVersion walks up to the root command to read its version string,
// falling back to "dev" if it is not set (e.g. in unit tests).
func resolveVersion(cmd *cobra.Command) string {
	root := cmd.Root()
	if root != nil && root.Version != "" {
		return root.Version
	}
	return "dev"
}
