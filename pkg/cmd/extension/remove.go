package extension

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/extensions"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdRemove returns the `extension remove` subcommand.
func NewCmdRemove(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "remove NAME",
		Short: "Remove an installed extension",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			extDir := filepath.Join(f.ConfigDir(), "extensions")
			mgr := extensions.New(extDir, nil)

			if err := mgr.Remove(name); err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "✓ Removed extension %s\n", name)
			return nil
		},
	}
}
