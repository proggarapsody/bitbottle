package extension

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/extensions"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdUpgrade returns the `extension upgrade` subcommand.
func NewCmdUpgrade(f *factory.Factory) *cobra.Command {
	var all bool
	var force bool

	cmd := &cobra.Command{
		Use:   "upgrade [NAME]",
		Short: "Upgrade installed extensions",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			extDir := filepath.Join(f.ConfigDir(), "extensions")
			mgr := extensions.New(extDir, nil)

			if all {
				results := mgr.UpgradeAll(force)
				for name, err := range results {
					if err != nil {
						fmt.Fprintf(f.IOStreams.ErrOut, "✗ %s: %v\n", name, err)
					} else {
						fmt.Fprintf(f.IOStreams.Out, "✓ %s: up to date or upgraded\n", name)
					}
				}
				return nil
			}

			if len(args) == 0 {
				return fmt.Errorf("specify an extension name or --all")
			}

			name := args[0]
			old, newV, err := mgr.Upgrade(name, force)
			if err != nil {
				return err
			}
			if old == "" && newV == "" {
				fmt.Fprintf(f.IOStreams.Out, "%s: local install — skipping\n", name)
				return nil
			}
			if old == newV {
				fmt.Fprintf(f.IOStreams.Out, "%s: already at %s\n", name, newV)
				return nil
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ %s: upgraded %s → %s\n", name, old, newV)
			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Upgrade all installed extensions")
	cmd.Flags().BoolVar(&force, "force", false, "Reinstall even if already at the latest version")

	return cmd
}
