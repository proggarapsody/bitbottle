package extension

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/extensions"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdInstall returns the `extension install` subcommand.
func NewCmdInstall(f *factory.Factory) *cobra.Command {
	var localPath string
	var force bool

	cmd := &cobra.Command{
		Use:   "install [USER/REPO]",
		Short: "Install a bitbottle extension",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			extDir := filepath.Join(f.ConfigDir(), "extensions")
			mgr := extensions.New(extDir, nil)

			if localPath != "" {
				if err := mgr.InstallLocal(localPath, force); err != nil {
					return err
				}
				fmt.Fprintf(f.IOStreams.Out, "✓ Installed extension from %s\n", localPath)
				return nil
			}

			if len(args) == 0 {
				return fmt.Errorf("requires USER/REPO argument or --local PATH")
			}
			ownerRepo := args[0]
			if err := mgr.InstallFromGitHub(ownerRepo, force); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Installed extension %s\n", ownerRepo)
			return nil
		},
	}

	cmd.Flags().StringVar(&localPath, "local", "", "Path to local extension directory (installs via symlink)")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite an already-installed extension")

	return cmd
}
